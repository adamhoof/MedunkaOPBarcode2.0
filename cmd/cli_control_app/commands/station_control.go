package commands

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SendGlobalCommand publishes a simple command to a global topic.
func SendGlobalCommand(client mqtt.Client, baseTopic, subTopic, state string, retained bool) {
	topic := fmt.Sprintf("%s/%s/%s", baseTopic, subTopic, state)
	token := client.Publish(topic, 1, retained, "1")
	if token.Wait() && token.Error() != nil {
		fmt.Printf("failed to publish to %s: %v\n", topic, token.Error())
		return
	}
	fmt.Printf("published to %s\n", topic)
}
