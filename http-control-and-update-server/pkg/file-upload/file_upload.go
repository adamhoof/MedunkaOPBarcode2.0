package file_upload

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	_, err = io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	extension := strings.ToLower(filepath.Ext(header.Filename))
	switch extension {
	case ".csv":
		// Call CSV parser
		fmt.Println("CSV file uploaded.")
	case ".xlsx":
		// Call XLSX parser
		fmt.Println("XLSX file uploaded.")
	default:
		http.Error(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}
