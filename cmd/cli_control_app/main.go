package main

import (
	"fmt"
	"log"

	"github.com/adamhoof/MedunkaOPBarcode2.0/cmd/cli_control_app/commands"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
)

type cliConfig struct {
	BaseTopic           string
	LightControlTopic   string
	FirmwareUpdateTopic string
	StatusTopic         string
	MdbPath             string
	HttpHost            string
	HttpPort            string
	HttpUpdateEndpoint  string
	HttpStatusEndpoint  string
	TlsCaPath           string
}

func loadCliConfig() cliConfig {
	return cliConfig{
		BaseTopic:           utils.GetEnvOrPanic("MQTT_BASE_TOPIC"),
		FirmwareUpdateTopic: utils.GetEnvOrPanic("MQTT_FIRMWARE_UPDATE_TOPIC"),
		StatusTopic:         utils.GetEnvOrPanic("MQTT_STATUS_TOPIC"),
		MdbPath:             utils.GetEnvOrPanic("MDB_FILEPATH"),
		HttpHost:            utils.GetEnvOrPanic("HTTP_SERVER_HOST"),
		HttpPort:            utils.GetEnvOrPanic("HTTP_SERVER_PORT"),
		HttpUpdateEndpoint:  utils.GetEnvOrPanic("HTTP_SERVER_UPDATE_ENDPOINT"),
		HttpStatusEndpoint:  utils.GetEnvOrPanic("HTTP_SERVER_UPDATE_STATUS_ENDPOINT"),
		TlsCaPath:           utils.GetEnvOrPanic("TLS_CA_PATH"),
	}
}

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
	cfg := loadCliConfig()

	mqttClient := utils.CreateSecureMQTTClient(utils.GetEnvOrPanic("MQTT_CLIAPP_CLIENT_ID"))
	utils.ConnectOrFail(mqttClient)
	defer mqttClient.Disconnect(250)

	log.Println("CLI App Ready. Connected to MQTT.")

	httpClient := utils.CreateSecureHTTPClient(cfg.TlsCaPath)

	availableCommands := []Command{
		{"ls", "list all available commands"},
		{"dbu", "convert local MDB to CSV and upload to server"},
		{"sfwe", "send firmware enable command to all stations"},
		{"sfwd", "send firmware disable command to all stations"},
		{"sw", "send wake command to all stations"},
		{"ss", "send sleep command to all stations"},
		{"e", "exit"},
	}

	if err := commands.DatabaseUpdate(httpClient, cfg.MdbPath, cfg.HttpHost, cfg.HttpPort, cfg.HttpUpdateEndpoint, cfg.HttpStatusEndpoint); err != nil {
		log.Printf("Startup update failed: %v\n", err)
	}

	for {
		fmt.Print("HekrMejMej > ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			log.Printf("failed to scan line: %s\n", err)
			continue
		}

		fmt.Println()
		switch input {
		case "ls":
			printAllCommands(availableCommands)
		case "dbu":
			if err := commands.DatabaseUpdate(httpClient, cfg.MdbPath, cfg.HttpHost, cfg.HttpPort, cfg.HttpUpdateEndpoint, cfg.HttpStatusEndpoint); err != nil {
				log.Printf("Update failed: %v\n", err)
			}
		case "sfwe":
			commands.SendGlobalCommand(mqttClient, cfg.BaseTopic, cfg.FirmwareUpdateTopic, "enable", false)
		case "sfwd":
			commands.SendGlobalCommand(mqttClient, cfg.BaseTopic, cfg.FirmwareUpdateTopic, "disable", false)
		case "sw":
			commands.SendGlobalCommand(mqttClient, cfg.BaseTopic, cfg.StatusTopic, "wake", true)
		case "ss":
			commands.SendGlobalCommand(mqttClient, cfg.BaseTopic, cfg.StatusTopic, "sleep", true)
		case "e":
			return
		default:
			log.Printf("invalid command\n")
			printAllCommands(availableCommands)
		}
	}
}
