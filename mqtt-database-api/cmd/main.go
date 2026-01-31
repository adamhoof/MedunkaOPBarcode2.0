package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-client"
	"github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-database-api/pkg/api"
)

func main() {
	postgresqlHandler, err := database.NewPostgres()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := postgresqlHandler.Close(); closeErr != nil {
			log.Printf("failed to close database connection: %s", closeErr)
		}
	}()

	mqttClient := mqtt_client.CreateDefault("mqttDatabaseAPI")
	mqtt_client.ConnectDefault(&mqttClient)

	token := mqttClient.Subscribe(os.Getenv("MQTT_PRODUCT_DATA_REQUEST"), 1, api.HandleProductDataRequest(postgresqlHandler))
	if token.WaitTimeout(2*time.Second) && token.Error() != nil {
		log.Fatal(token.Error())
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
