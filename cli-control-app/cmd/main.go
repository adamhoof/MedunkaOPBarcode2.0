package main

import (
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/cli-control-app/pkg/commands"
	"github.com/adamhoof/MedunkaOPBarcode2.0/cli-control-app/pkg/device_controller"
	"log"
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
		{"update", "update database table containing product info"},
		{"mqttTest", "test out mqtt connection by sending sample product data request"},
		{"controlDevice", "control specific device (command prompts for name of device later)"},
	}

	for {
		fmt.Print("HekrMejMej > ")
		if _, err := fmt.Scanln(&input); err != nil {
			log.Printf("failed to scan line: %s\n", err)
			continue
		}

		switch input {
		case "update":
			if err := commands.DatabaseUpdate(); err != nil {
				log.Println(err)
			}
		case "mqttTest":
			commands.MQTTProductDataRequestTest("mqtt_test", "8595020340103", true)
		case "controlDevice":
			fmt.Print("Which device do you want to control?: ")
			if _, err := fmt.Scanln(&input); err != nil {
				log.Printf("failed to scan line: %s\n", err)
				continue
			}
			// TODO search for device, if it does exist, enter function
			device_controller.EnterDeviceControlMode(input)
		case "ls":
			printAllCommands(availableCommands)
		default:
			log.Printf("invalid command, try again...\n")
		}
	}
}
