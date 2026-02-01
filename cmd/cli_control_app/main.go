package main

import (
	"fmt"
	"log"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
)

type Command struct {
	name        string
	description string
}

func printAllCommands(availableCommands []Command) {
	for _, command := range availableCommands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
}

func main() {
	var input string

	availableCommands := []Command{
		{"ls", "list all available commands with their description"},
		{"upd", "update database table containing product info"},
		{"tmqtt", "test out mqtt connection by sending sample product data request"},
		{"cdev", "control specific device (command prompts for name of device later)"},
	}

	mqttRequestTopic := utils.GetEnvOrPanic("MQTT_TOPIC_REQUEST")
	lightControlTopic := utils.GetEnvOrPanic("LIGHT_CONTROL_TOPIC")
	firmwareUpdateTopic := utils.GetEnvOrPanic("FIRMWARE_UPDATE_TOPIC")
	mdbPath := utils.GetEnvOrPanic("MDB_PATH")
	httpHost := utils.GetEnvOrPanic("HTTP_SERVER_HOST")
	httpPort := utils.GetEnvOrPanic("HTTP_SERVER_PORT")
	httpUpdateEndpoint := utils.GetEnvOrPanic("HTTP_SERVER_UPDATE_ENDPOINT")

	for {
		fmt.Print("HekrMejMej > ")
		if _, err := fmt.Scanln(&input); err != nil {
			log.Printf("failed to scan line: %s\n", err)
			continue
		}

		fmt.Println()
		switch input {
		case "upd":
			if err := DatabaseUpdate(mdbPath, httpHost, httpPort, httpUpdateEndpoint); err != nil {
				log.Println(err)
			}
		case "tmqtt":
			MQTTProductDataRequestTest("mqtt_test", mqttRequestTopic, "8595020340103", true)
		case "cdev":
			fmt.Print("Which device do you want to control?: ")
			if _, err := fmt.Scanln(&input); err != nil {
				log.Printf("failed to scan line: %s\n", err)
				continue
			}
			// TODO search for device, if it does exist, enter function
			EnterDeviceControlMode(input, lightControlTopic, firmwareUpdateTopic)
		case "ls":
			printAllCommands(availableCommands)
		default:
			log.Printf("invalid command, try again...\n")
		}
	}
}
