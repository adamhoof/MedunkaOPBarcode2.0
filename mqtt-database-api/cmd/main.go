package main

import (
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-database-api/pkg/api"
	mqtt "github.com/eclipse/paho.mqtt.golang"
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

	options := mqtt.NewClientOptions()
	options.AddBroker(os.Getenv("MQTT_SERVER_AND_PORT"))
	options.SetClientID("mqttDatabaseAPI")
	options.SetAutoReconnect(true)
	options.SetConnectRetry(true)
	options.SetCleanSession(false)
	options.SetConnectRetryInterval(time.Second * 2)
	options.SetOrderMatters(false)

	mqttClient := mqtt.NewClient(options)

	for {
		token := mqttClient.Connect()
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("mqtt client failed to connect, retrying...", token.Error())
		time.Sleep(5 * time.Second)

	}
	log.Println("mqtt client connected")

	token := mqttClient.Subscribe(os.Getenv("MQTT_PRODUCT_DATA_REQUEST"), 0, api.GetProductData(&postgresqlHandler))
	if token.WaitTimeout(2*time.Second) && token.Error() != nil {
		log.Fatal(token.Error())
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
