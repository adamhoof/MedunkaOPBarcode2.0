package commands

import (
	"encoding/json"
	mqtt_client "github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-client"
	"log"
	"time"
)

type LightCommand struct {
	State bool `json:"state"`
}

func TurnOnLight(topic string) {
	mqttClient := mqtt_client.CreateDefault("light_controller")
	mqtt_client.ConnectDefault(&mqttClient)

	lightCommand := LightCommand{State: true}
	lightCommandAsJson, err := json.Marshal(&lightCommand)
	if err != nil {
		log.Println("unable to serialize light command into json: ", err)
		return
	}

	for {
		token := mqttClient.Publish(topic, 1, false, lightCommandAsJson)
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

	lightCommand := LightCommand{State: false}
	lightCommandAsJson, err := json.Marshal(&lightCommand)
	if err != nil {
		log.Println("unable to serialize light command into json: ", err)
		return
	}

	for {
		token := mqttClient.Publish(topic, 1, false, lightCommandAsJson)
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("failed to publish message...", token.Error())
		time.Sleep(1 * time.Second)
	}
	mqttClient.Disconnect(0)
	log.Println("mqtt client disconnected")
}
