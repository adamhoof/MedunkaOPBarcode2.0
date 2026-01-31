package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adamhoof/MedunkaOPBarcode2.0/http-database-update-server/pkg/database-update"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/parser"
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

	http.HandleFunc(os.Getenv("HTTP_SERVER_UPDATE_ENDPOINT"), database_update.HandleDatabaseUpdateRequest(postgreSQLHandler, mmdbParser))

	log.Printf("Starting server on %s:%s", "0.0.0.0", os.Getenv("HTTP_SERVER_PORT"))

	err = http.ListenAndServe(fmt.Sprintf("%s:%s", "0.0.0.0", os.Getenv("HTTP_SERVER_PORT")), nil)
	if err != nil {
		log.Fatal("unable to start")
	}
}
