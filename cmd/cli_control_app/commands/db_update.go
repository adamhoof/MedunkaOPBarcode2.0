package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type ResponseContent struct {
	Message string `json:"message"`
}

// DatabaseUpdate converts MDB to CSV and uploads it to the server.
func DatabaseUpdate(client *http.Client, mdbPath, httpHost, httpPort, httpUpdateEndpoint string) error {
	if _, err := os.Stat(mdbPath); os.IsNotExist(err) {
		return fmt.Errorf("MDB file not found at: %s", mdbPath)
	}

	csvPath := "/tmp/export.csv"
	defer func() {
		_ = os.Remove(csvPath)
	}()

	fmt.Println("Converting MDB to CSV...")
	cmd := exec.Command("./mdb_to_csv.sh", mdbPath, csvPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to convert MDB: %w", err)
	}

	fmt.Println("Uploading CSV to server...")
	if err := sendFileToServer(client, httpHost, httpPort, httpUpdateEndpoint, csvPath); err != nil {
		return fmt.Errorf("failed to send file to server: %w", err)
	}
	return nil
}

func sendFileToServer(client *http.Client, host, port, endpoint, fileLocation string) error {
	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("failed to close file: %s\n", err)
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(fileLocation))
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s:%s%s", host, port, endpoint)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %s", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %s\n", err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}

	responseContent := ResponseContent{}
	if err = json.Unmarshal(bodyBytes, &responseContent); err != nil {
		return fmt.Errorf("failed to parse server response: %s", err)
	}

	fmt.Printf("Success: %s (Status: %s)\n", responseContent.Message, resp.Status)
	return nil
}
