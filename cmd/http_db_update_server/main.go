package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
)

func main() {
	postgreSQLHandler, err := database.New("postgres")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := postgreSQLHandler.Close(); closeErr != nil {
			log.Printf("failed to close database connection: %s", closeErr)
		}
	}()

	jobStore := &sync.Map{}

	maxUploadSize := utils.GetEnvAsInt64("HTTP_MAX_UPLOAD_SIZE")

	http.HandleFunc(utils.GetEnvOrPanic("HTTP_SERVER_UPDATE_ENDPOINT"), HandleDatabaseUpdate(postgreSQLHandler, jobStore, maxUploadSize))
	http.HandleFunc("/job-status", HandleJobStatus(jobStore))

	host := utils.GetEnvOrPanic("HTTP_SERVER_HOST")
	port := utils.GetEnvOrPanic("HTTP_SERVER_PORT")
	certPath := utils.GetEnvOrPanic("TLS_CERT_PATH")
	keyPath := utils.GetEnvOrPanic("TLS_KEY_PATH")

	log.Printf("Starting server on %s:%s", host, port)

	err = http.ListenAndServeTLS(fmt.Sprintf("%s:%s", host, port), certPath, keyPath, nil)
	if err != nil {
		log.Fatal("unable to start")
	}
}
