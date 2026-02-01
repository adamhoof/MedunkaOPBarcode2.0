package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/parser"
	"github.com/google/uuid"
)

const MaxUploadSize = 5 * 1024 * 1024 * 1024

// HandleDatabaseUpdate accepts a multipart file upload, identifies the correct parser via factory,
// saves the stream to a temporary file, and spawns a background worker to process the import.
func HandleDatabaseUpdate(db database.Handler, jobStore *sync.Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "file too large", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		selectedParser, err := parser.New(header.Filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid file type: %v", err), http.StatusBadRequest)
			return
		}

		jobID := uuid.New().String()
		jobStore.Store(jobID, JobStatus{State: "pending", Message: "Upload starting"})

		ext := filepath.Ext(header.Filename)
		tempFile, err := os.CreateTemp("", fmt.Sprintf("upload_%s_*%s", jobID, ext))
		if err != nil {
			jobStore.Store(jobID, JobStatus{State: "failed", Message: "Disk error"})
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(tempFile, file); err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			jobStore.Store(jobID, JobStatus{State: "failed", Message: "Upload interrupted"})
			http.Error(w, "upload failed", http.StatusInternalServerError)
			return
		}
		tempFile.Close()

		go processImportJob(db, selectedParser, tempFile.Name(), jobID, jobStore)

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"job_id":  jobID,
			"message": "Upload accepted",
		})
	}
}

// processImportJob runs in a background goroutine to parse the saved file and bulk import it into the database,
// updating the job status in the shared memory store throughout the process.
func processImportJob(db database.Handler, p parser.CatalogParser, filePath string, jobID string, jobStore *sync.Map) {
	defer os.Remove(filePath)

	jobStore.Store(jobID, JobStatus{State: "processing", Message: "Importing data..."})

	file, err := os.Open(filePath)
	if err != nil {
		jobStore.Store(jobID, JobStatus{State: "failed", Message: "Failed to open file"})
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	stream, parseErrors := p.ParseStream(file)

	err = db.ImportCatalog(ctx, stream)
	if err != nil {
		jobStore.Store(jobID, JobStatus{State: "failed", Message: fmt.Sprintf("Database error: %v", err)})
		return
	}

	if parseErr := <-parseErrors; parseErr != nil {
		jobStore.Store(jobID, JobStatus{State: "failed", Message: fmt.Sprintf("Parsing error: %v", parseErr)})
		return
	}

	jobStore.Store(jobID, JobStatus{State: "completed", Message: "Import finished successfully"})
}
