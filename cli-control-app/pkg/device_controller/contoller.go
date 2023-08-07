package device_controller

import (
	"fmt"
	commands "github.com/adamhoof/MedunkaOPBarcode2.0/cli-control-app/pkg/commands"
	"log"
	"os"
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
func EnterDeviceControlMode(deviceName string) {
	availableCommands := []Command{
		{"ls", "list all available commands with their description"},
		{"lon", "turn on light of device"},
		{"loff", "turn off light of device"},
		{"firmwareUpdate", "put device into firmware update mode"},
		{"e", "exit device control mode"},
	}
	fmt.Printf("Hello fella, you are currently in the control mode of device: %s\ntype ls to list available commands...\n", deviceName)

	for {
		fmt.Printf("%s > ", deviceName)
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			log.Printf("failed to scan line: %s\n", err)
			continue
		}

		fmt.Println()
		switch input {
		case "ls":
			printAllCommands(availableCommands)
		case "lon":
			commands.TurnOnLight(deviceName + "/" + os.Getenv("LIGHT_CONTROL_TOPIC"))
		//send mqtt command
		case "loff":
			commands.TurnOffLight(deviceName + "/" + os.Getenv("LIGHT_CONTROL_TOPIC"))
			//send mqtt command
		case "firmwareUpdate":
			commands.UpdateFirmware(deviceName + "/" + os.Getenv("FIRMWARE_UPDATE_TOPIC"))
			//send mqtt command
		case "e":
			fmt.Printf("It was pleasure to communicate, your %s\n", deviceName)
			return
		default:
			log.Printf("invalid command, try again...\n")
		}
	}
}
