package mqtt_client

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

func CreateDefault(name string) mqtt.Client {
	options := mqtt.NewClientOptions()
	options.AddBroker(os.Getenv("MQTT_SERVER_AND_PORT"))
	options.SetClientID(name)
	options.SetAutoReconnect(true)
	options.SetConnectRetry(true)
	options.SetCleanSession(false)
	options.SetOrderMatters(false)

	return mqtt.NewClient(options)
}

func ConnectDefault(mqttClient *mqtt.Client) {
	for {
		token := (*mqttClient).Connect()
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("mqtt client failed to connect, retrying...", token.Error())
		time.Sleep(5 * time.Second)

	}
	log.Println("mqtt client connected")
}
