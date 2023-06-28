package commands

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

func UpdateDeviceFirmware(deviceName string) {
	options := mqtt.NewClientOptions()
	options.AddBroker(os.Getenv("MQTT_SERVER_AND_PORT"))
	options.SetClientID("firmware_updater")
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

	for {
		token := mqttClient.Publish(fmt.Sprintf("%s/%s", deviceName, "firmware_update"), 1, false, true)
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("failed to publish message...", token.Error())
		time.Sleep(1 * time.Second)
	}
	mqttClient.Disconnect(0)
	log.Println("mqtt client disconnected")
}
