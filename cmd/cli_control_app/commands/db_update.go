package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type UploadResponse struct {
	Message string `json:"message"`
	JobID   string `json:"job_id"`
}

type JobStatus struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

// DatabaseUpdate converts MDB to CSV and uploads it to the server.
func DatabaseUpdate(client *http.Client, mdbPath, httpHost, httpPort, httpUpdateEndpoint, statusEndpoint string) error {
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
	uploadResponse, err := sendFileToServer(client, httpHost, httpPort, httpUpdateEndpoint, csvPath)
	if err != nil {
		return fmt.Errorf("failed to send file to server: %w", err)
	}
	if uploadResponse.JobID == "" {
		return fmt.Errorf("server response missing job_id")
	}

	fmt.Printf("Job ID: %s\n", uploadResponse.JobID)
	if err := pollJobStatus(client, httpHost, httpPort, statusEndpoint, uploadResponse.JobID); err != nil {
		return err
	}
	return nil
}

func sendFileToServer(client *http.Client, host, port, endpoint, fileLocation string) (UploadResponse, error) {
	file, err := os.Open(fileLocation)
	if err != nil {
		return UploadResponse{}, err
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
		return UploadResponse{}, err
	}
	if _, err = io.Copy(part, file); err != nil {
		return UploadResponse{}, err
	}

	if err := writer.Close(); err != nil {
		return UploadResponse{}, err
	}

	serverUrl := fmt.Sprintf("https://%s:%s%s", host, port, endpoint)
	req, err := http.NewRequest("POST", serverUrl, body)
	if err != nil {
		return UploadResponse{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to do request: %s", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %s\n", err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to read response body: %s", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		return UploadResponse{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	uploadResponse := UploadResponse{}
	if err = json.Unmarshal(bodyBytes, &uploadResponse); err != nil {
		return UploadResponse{}, fmt.Errorf("failed to parse server response: %s", err)
	}

	fmt.Printf("Upload response: %s (Status: %s)\n", uploadResponse.Message, resp.Status)
	return uploadResponse, nil
}

func pollJobStatus(client *http.Client, host, port, statusEndpoint, jobID string) error {
	statusURL := fmt.Sprintf("https://%s:%s%s?id=%s", host, port, statusEndpoint, url.QueryEscape(jobID))
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)
	var lastStatus string

	for {
		select {
		case <-ticker.C:
			req, err := http.NewRequest("GET", statusURL, nil)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("error: failed to reach server: %v\n", err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				return fmt.Errorf("status check failed (%s): %s", resp.Status, string(body))
			}

			var status JobStatus
			err = json.NewDecoder(resp.Body).Decode(&status)
			resp.Body.Close()
			if err != nil {
				return fmt.Errorf("failed to decode status response: %w", err)
			}

			currentStatus := fmt.Sprintf("%s:%s", status.State, status.Message)
			if currentStatus != lastStatus {
				fmt.Printf("Status: %s, %s\n", status.State, status.Message)
				lastStatus = currentStatus
			}

			switch status.State {
			case "completed":
				return nil
			case "failed":
				return fmt.Errorf("job failed: %s", status.Message)
			}

		case <-timeout:
			return fmt.Errorf("job status polling timed out")
		}
	}
}
