package main

import (
	"bytes"
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	conf, err := config.LoadConfig("/home/adamhoof/MedunkaOPBarcode2.0/Config.json")
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
			mdbFile, err := findMDBFile(conf.CLIControlApp.MDBFileLocation)
			if err != nil {
				log.Fatal(err)
			}

			err = sendFileToServer(conf.HTTPDatabaseUpdate.Host, conf.HTTPDatabaseUpdate.Port, conf.HTTPDatabaseUpdate.Endpoint, mdbFile)
			if err != nil {
				log.Fatal(err)
			}
		default:
			fmt.Println("Unknown command. Please try again.")
		}
	}
}

func findMDBFile(dirPath string) (string, error) {
	var mdbFile string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".mdb" {
			mdbFile = path
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if mdbFile == "" {
		return "", fmt.Errorf("no mdb file found in the specified directory")
	}

	return mdbFile, nil
}

func sendFileToServer(host string, port string, endpoint string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	return nil
}
