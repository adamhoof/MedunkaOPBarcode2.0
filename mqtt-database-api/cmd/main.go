package main

import (
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	"github.com/adamhoof/MedunkaOPBarcode2.0/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-database-api/pkg/api"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

func main() {
	conf, err := config.LoadConfig(os.Args[1])
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("config ok")

	postgresqlHandler := database.PostgreSQLHandler{}
	err = postgresqlHandler.Connect(&conf.Database)
	if err != nil {
		log.Fatal("connection to database unsuccessful: ", err)
	}

	log.Println("database connected")

	options := mqtt.ClientOptions{
		ClientID:             "mqtt-database-api",
		CleanSession:         false,
		Order:                false,
		AutoReconnect:        true,
		ConnectRetryInterval: 2 * time.Second,
		ConnectRetry:         true,
	}
	options.AddBroker(conf.MQTT.ServerAndPort)

	mqttClient := mqtt.NewClient(&options)

	for {
		token := mqttClient.Connect()
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("mqtt client failed to connect, retrying...", token.Error())
		time.Sleep(5 * time.Second)

	}
	log.Println("mqtt client connected")

	token := mqttClient.Subscribe(conf.MQTT.ProductDataRequest, 0, api.GetProductData(&postgresqlHandler, conf))
	if token.Error() != nil {
		log.Fatal(token.Error())
	}
}
