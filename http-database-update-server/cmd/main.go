package main

import (
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	"github.com/adamhoof/MedunkaOPBarcode2.0/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/http-database-update-server/pkg/database-update"
	"log"
	"net/http"
)

func main() {
	conf, err := config.LoadConfig("/home/adamhoof/MedunkaOPBarcode2.0/Config.json")
	if err != nil {
		log.Fatal(err)
	}

	postgreSQLHandler := database.PostgreSQLHandler{}
	http.HandleFunc(conf.HTTPDatabaseUpdate.Endpoint, database_update.HandleDatabaseUpdateRequest(conf, &postgreSQLHandler))

	log.Printf("Starting server on %s:%s", conf.HTTPDatabaseUpdate.Host, conf.HTTPDatabaseUpdate.Port)

	err = http.ListenAndServe(fmt.Sprintf("%s:%s", conf.HTTPDatabaseUpdate.Host, conf.HTTPDatabaseUpdate.Port), nil)
	if err != nil {
		log.Fatal("unable to start")
	}
	log.Printf("Listening at%s:%s%s", conf.HTTPDatabaseUpdate.Host, conf.HTTPDatabaseUpdate.Port, conf.HTTPDatabaseUpdate.Endpoint)
}
