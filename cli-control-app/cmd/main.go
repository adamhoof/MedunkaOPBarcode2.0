package main

import (
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/cli-control-app/pkg/commands"
	"github.com/adamhoof/MedunkaOPBarcode2.0/cli-control-app/pkg/device_controller"
	"log"
)

func main() {
	var input string

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
			commands.MQTTRequestTest()
		case "controlDevice":
			fmt.Print("Which device do you want to control?: ")
			if _, err := fmt.Scanln(&input); err != nil {
				log.Printf("failed to scan line: %s\n", err)
				continue
			}
			// TODO search for device, if it does exist, enter function
			device_controller.EnterDeviceControlMode(input)
		default:
			log.Printf("invalid command, try again...\n")
		}
	}
}
