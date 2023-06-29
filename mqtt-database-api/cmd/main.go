package main

import (
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-client"
	"github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-database-api/pkg/api"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	postgresqlHandler := database.PostgreSQLHandler{}
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOSTNAME"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
	err := postgresqlHandler.Connect(connectionString)
	if err != nil {
		log.Fatal("connection to database unsuccessful: ", err)
	}
	log.Println("database connected")

	mqttClient := mqtt_client.CreateDefault("mqttDatabaseAPI")
	mqtt_client.ConnectDefault(&mqttClient)

	token := mqttClient.Subscribe(os.Getenv("MQTT_PRODUCT_DATA_REQUEST"), 1, api.HandleProductDataRequest(&postgresqlHandler))
	if token.WaitTimeout(2*time.Second) && token.Error() != nil {
		log.Fatal(token.Error())
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
