package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	file_parser "github.com/adamhoof/MedunkaOPBarcode2.0/file-parser"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type ResponseContent struct {
	Message string `json:"message"`
}

func main() {
	for {
		fmt.Print("> ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			log.Printf("failed to scan line: %s\n", err)
			continue
		}

		switch input {
		case "update":
			if err := update(); err != nil {
				log.Println(err)
			}
		default:
			log.Printf("invalid command, try again...\n")
		}
	}
}

func update() error {
	if err := fileReadyToBeUsed(os.Getenv("MDB_PATH")); err != nil {
		return fmt.Errorf("file not ready to be used: %w", err)
	}

	mdbFileParser := file_parser.MDBFileParser{}
	if err := mdbFileParser.ToCSV(os.Getenv("MDB_PATH"), os.Getenv("CSV_OUTPUT_PATH"), os.Getenv("SHELL_MDB_FILE_PARSER_PATH")); err != nil {
		return fmt.Errorf("failed to parse mdb to csv: %w", err)
	}

	if err := sendFileToServer(os.Getenv("HTTP_SERVER_HOST"), os.Getenv("HTTP_SERVER_PORT"), os.Getenv("HTTP_SERVER_UPDATE_ENDPOINT"), os.Getenv("CSV_OUTPUT_PATH")); err != nil {
		return fmt.Errorf("failed to send file to server: %w", err)
	}
	return nil
}
func fileReadyToBeUsed(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("permission denied: cannot read file at %s", filePath)
		}
		if err := file.Close(); err != nil {
			log.Println("failed to close file")
		}
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found at the specified path: %s", filePath)
	}
	return err
}

func sendFileToServer(host string, port string, endpoint string, fileLocation string) error {

	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Printf("failed to close file: %s\n", err)
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(fileLocation))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:%s%s", host, port, endpoint)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %s", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Printf("failed to close response body: %s\n", err)
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

	fmt.Printf("%s\n%s\n", responseContent.Message, resp.Status)
	return nil
}
