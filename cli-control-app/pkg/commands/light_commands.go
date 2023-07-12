package commands

import (
	mqtt_client "github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-client"
	"log"
	"time"
)

func TurnOnLight(topic string) {
	mqttClient := mqtt_client.CreateDefault("light_controller")
	mqtt_client.ConnectDefault(&mqttClient)

	for {
		token := mqttClient.Publish(topic, 1, false, "true")
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("failed to publish message...", token.Error())
		time.Sleep(1 * time.Second)
	}
	mqttClient.Disconnect(0)
	log.Println("mqtt client disconnected")
}

func TurnOffLight(topic string) {
	mqttClient := mqtt_client.CreateDefault("light_controller")
	mqtt_client.ConnectDefault(&mqttClient)

	for {
		token := mqttClient.Publish(topic, 1, false, "false")
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("failed to publish message...", token.Error())
		time.Sleep(1 * time.Second)
	}
	mqttClient.Disconnect(0)
	log.Println("mqtt client disconnected")
}
