package main

import (
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/http-database-update-server/pkg/database-update"
	"log"
	"net/http"
	"os"
)

func main() {
	postgreSQLHandler := database.PostgreSQLHandler{}
	http.HandleFunc(os.Getenv("HTTP_SERVER_UPDATE_ENDPOINT"), database_update.HandleDatabaseUpdateRequest(&postgreSQLHandler))

	log.Printf("Starting server on %s:%s", "0.0.0.0", os.Getenv("HTTP_SERVER_PORT"))

	err := http.ListenAndServe(fmt.Sprintf("%s:%s", "0.0.0.0", os.Getenv("HTTP_SERVER_PORT")), nil)
	if err != nil {
		log.Fatal("unable to start")
	}
}
