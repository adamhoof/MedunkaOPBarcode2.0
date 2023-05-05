package file_upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	serverURL  = "http://localhost:8080"
	uploadPath = "/upload"
	outputFile = "uploaded_file.mdb"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error reading file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	out, err := os.Create(outputFile)
	if err != nil {
		http.Error(w, "Error creating output file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error writing file to disk", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded and saved as %s", outputFile)
}
