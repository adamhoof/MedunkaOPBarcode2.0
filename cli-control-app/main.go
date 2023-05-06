package main

import (
	"bytes"
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	file_parser "github.com/adamhoof/MedunkaOPBarcode2.0/file-parser"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	conf, err := config.LoadConfig(os.Args[1])
	if err != nil {
		log.Fatal(err)
		return
	}

	for {
		fmt.Print("> ")
		var input string
		_, err = fmt.Scanln(&input)
		if err != nil {
			fmt.Print(err)
			continue
		}

		switch input {
		case "update":
			if err := fileReadyToBeUsed(conf.CLIControlApp.MDBFileLocation); err != nil {
				log.Println(err)
			}

			mdbFileParser := file_parser.MDBFileParser{}
			if err := mdbFileParser.ToCSV(conf.CLIControlApp.MDBFileLocation, conf.HTTPDatabaseUpdate.OutputCSVLocation); err != nil {
				log.Println("failed to parse mdb to csv: ", err)
				continue
			}

			err = sendFileToServer(conf.HTTPDatabaseUpdate.Host, conf.HTTPDatabaseUpdate.Port, conf.HTTPDatabaseUpdate.Endpoint, conf.HTTPDatabaseUpdate.OutputCSVLocation)
			if err != nil {
				log.Println(err)
			}
		default:
			fmt.Println("Unknown command. Please try again.")
		}
	}
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
	defer file.Close()

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
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %s\n%s", resp.Status, string(bodyBytes))
	}

	fmt.Println(string(bodyBytes))
	return nil
}
