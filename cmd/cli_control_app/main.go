package main

import (
	"fmt"
	"log"

	"github.com/adamhoof/MedunkaOPBarcode2.0/cmd/cli_control_app/commands"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
)

type cliConfig struct {
	ControlTopic       string
	MdbPath            string
	FirmwarePath       string
	HttpHost           string
	HttpPort           string
	HttpUpdateEndpoint string
	HttpStatusEndpoint string
	TlsCaPath          string
	TlsCertPath        string
	TlsKeyPath         string
}

func loadCliConfig() cliConfig {
	return cliConfig{
		ControlTopic:       utils.GetEnvOrPanic("MQTT_CONTROL_TOPIC"),
		MdbPath:            utils.GetEnvOrPanic("MDB_FILEPATH"),
		FirmwarePath:       utils.GetEnvOrPanic("FIRMWARE_FILEPATH"),
		HttpHost:           utils.GetEnvOrPanic("HTTP_SERVER_HOST"),
		HttpPort:           utils.GetEnvOrPanic("HTTP_SERVER_PORT"),
		HttpUpdateEndpoint: utils.GetEnvOrPanic("HTTP_SERVER_UPDATE_ENDPOINT"),
		HttpStatusEndpoint: utils.GetEnvOrPanic("HTTP_SERVER_UPDATE_STATUS_ENDPOINT"),
		TlsCaPath:          utils.GetEnvOrPanic("TLS_CA_PATH"),
		TlsCertPath:        utils.GetEnvOrPanic("TLS_CLIENT_CERT_PATH"),
		TlsKeyPath:         utils.GetEnvOrPanic("TLS_CLIENT_KEY_PATH"),
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

	httpClient := utils.CreateSecureHTTPClient(cfg.TlsCaPath, cfg.TlsCertPath, cfg.TlsKeyPath)

	availableCommands := []Command{
		{"ls", "list all available commands"},
		{"upd", "convert local MDB to CSV and upload to server"},
		{"sfw", "send firmware enable command to all stations"},
		{"sw", "send wake command to all stations"},
		{"ss", "send sleep command to all stations"},
		{"scfs", "send conf_scanner command to all stations"},
		{"e", "exit"},
	}

	if err := commands.DatabaseUpdate(httpClient, cfg.MdbPath, cfg.HttpHost, cfg.HttpPort, cfg.HttpUpdateEndpoint, cfg.HttpStatusEndpoint); err != nil {
		log.Printf("Startup update failed: %v\n", err)
	}

	fmt.Printf("\ntype 'ls' to see available commands\n\n")

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
		case "upd":
			if err := commands.DatabaseUpdate(httpClient, cfg.MdbPath, cfg.HttpHost, cfg.HttpPort, cfg.HttpUpdateEndpoint, cfg.HttpStatusEndpoint); err != nil {
				log.Printf("Update failed: %v\n", err)
			}
		case "sfw":
			downloadUrl, waitAndStop, err := commands.FirmwareUpdateServer(cfg.FirmwarePath, cfg.TlsCertPath, cfg.TlsKeyPath)
			if err != nil {
				log.Printf("Failed to start server: %v\n", err)
				continue
			}

			log.Printf("Broadcasting update command...")
			commandPayload := fmt.Sprintf("%s", downloadUrl)
			commands.SendCommand(mqttClient, cfg.ControlTopic, commandPayload, false)

			waitAndStop()

			fmt.Println("Update sequence complete.")
		case "sw":
			commands.SendCommand(mqttClient, cfg.ControlTopic, "wake", true)
		case "ss":
			commands.SendCommand(mqttClient, cfg.ControlTopic, "sleep", true)
		case "scfs":
			commands.SendCommand(mqttClient, cfg.ControlTopic, "conf_scanner", false)
		case "e":
			return
		default:
			log.Printf("invalid command\n")
			printAllCommands(availableCommands)
		}
	}
}
