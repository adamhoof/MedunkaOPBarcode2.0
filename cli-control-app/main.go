package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	var dirPath string
	flag.StringVar(&dirPath, "dir", "", "The directory where the mdb file is located")
	flag.Parse()

	if dirPath == "" {
		log.Println("dir flag is required")
	}
	dirPath = "/home/adamhoof/Desktop/67668305_2022.mdb"

	mdbFile, err := findMDBFile(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	err = sendFileToServer("http://localhost:8080/upload", mdbFile)
	if err != nil {
		log.Fatal(err)
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

func sendFileToServer(url, filePath string) error {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read file: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/upload", bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("unable to create new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %v", resp.Status)
	}

	log.Println("File uploaded successfully")
	return nil
}
