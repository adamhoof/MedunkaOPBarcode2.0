package database_update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/parser"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func (e ErrorResponse) Error() string {
	return e.Message
}

type SuccessResponse struct {
	Message string `json:"message"`
}

const (
	requestTimeout = 15 * time.Minute
)

func extractFileFromRequest(request *http.Request) (string, error) {
	receivedFile, _, err := request.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("failed to read multipart file from request: %s", err)
	}
	defer func() {
		err = receivedFile.Close()
		if err != nil {
			log.Printf("failed to close received multipart file: %s\n", err)
		}
	}()

	targetPath := os.Getenv("HTTP_SERVER_CSV_FILE_PATH")
	if strings.TrimSpace(targetPath) == "" {
		return "", errors.New("HTTP_SERVER_CSV_FILE_PATH is not set")
	}

	if err = os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory for upload: %w", err)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %s", err)
	}

	err = file.Chmod(0666)
	if err != nil {
		return "", fmt.Errorf("failed to change permissions of file %s: %s", file.Name(), err)
	}

	if _, err = io.Copy(file, receivedFile); err != nil {
		return "", fmt.Errorf("failed to write file %s: %s", file.Name(), err)
	}

	return file.Name(), nil
}

func handleError(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
	if err != nil {
		log.Printf("failed to encode message into json: %s\n", err)
	}
}
func HandleDatabaseUpdateRequest(handler database.Handler, catalogParser parser.CatalogParser) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if catalogParser == nil {
			handleError(w, errors.New("catalog parser is not configured"), http.StatusInternalServerError)
			return
		}

		fileLocation, err := extractFileFromRequest(request)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}
		defer func() {
			err = os.Remove(fileLocation)
			if err != nil {
				log.Printf("failed to remove file: %s\n", err)
			}
		}()

		ctx, cancel := context.WithTimeout(request.Context(), requestTimeout)
		defer cancel()

		file, err := os.Open(fileLocation)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				log.Printf("failed to close upload file: %s", closeErr)
			}
		}()

		stream, parseErrors := catalogParser.ParseStream(file)

		importErr := make(chan error, 1)
		go func() {
			importErr <- handler.ImportCatalog(ctx, stream)
		}()

		select {
		case err = <-importErr:
			if err != nil {
				handleError(w, err, http.StatusInternalServerError)
				return
			}
		case parseErr, ok := <-parseErrors:
			if ok && parseErr != nil {
				handleError(w, parseErr, http.StatusInternalServerError)
				return
			}
		case <-ctx.Done():
			handleError(w, ctx.Err(), http.StatusRequestTimeout)
			return
		}

		w.WriteHeader(http.StatusOK)
		response := SuccessResponse{Message: "All done!"}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("failed to write to client: %s\n", err)
		}
	}
}
