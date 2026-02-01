package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/parser"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
)

func main() {
	postgreSQLHandler, err := database.NewPostgres()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := postgreSQLHandler.Close(); closeErr != nil {
			log.Printf("failed to close database connection: %s", closeErr)
		}
	}()

	mmdbParser, err := parser.NewMMDB()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc(utils.ReadEnvOrFail("HTTP_SERVER_UPDATE_ENDPOINT"), HandleDatabaseUpdateRequest(postgreSQLHandler, mmdbParser))

	host := utils.ReadEnvOrFail("HTTP_SERVER_HOST")
	port := utils.ReadEnvOrFail("HTTP_SERVER_PORT")
	certPath := utils.ReadEnvOrFail("TLS_CERT_PATH")
	keyPath := utils.ReadEnvOrFail("TLS_KEY_PATH")

	log.Printf("Starting server on %s:%s", host, port)

	err = http.ListenAndServeTLS(fmt.Sprintf("%s:%s", host, port), certPath, keyPath, nil)
	if err != nil {
		log.Fatal("unable to start")
	}
}
