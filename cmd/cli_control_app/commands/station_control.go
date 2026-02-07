package commands

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SendCommand publishes a payload directly to the given topic.
func SendCommand(client mqtt.Client, topic, payload string, retained bool) {
	token := client.Publish(topic, 1, retained, payload)
	if token.Wait() && token.Error() != nil {
		fmt.Printf("failed to publish to %s: %v\n", topic, token.Error())
		return
	}
	fmt.Printf("published '%s' to %s\n", payload, topic)
}
