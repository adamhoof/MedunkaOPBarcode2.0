package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Dbname   string `json:"dbName"`
}

type MQTTConfig struct {
	ServerAndPort      string `json:"serverAndPort"`
	ProductDataRequest string `json:"productDataRequest"`
}

type SerialConfig struct {
	PortName   string `json:"portName"`
	BaudRate   int    `json:"baudRate"`
	Terminator int    `json:"terminator"`
}

type ControlAndUpdateServerConfig struct {
	Host               string `json:"host"`
	Port               string `json:"port"`
	FileUploadEndpoint string `json:"fileUploadEndpoint"`
	Delimiter          string `json:"delimiter"`
	TableName          string `json:"tableName"`
	PathToUiStuff      string `json:"pathToUiStuff"`
}
type Config struct {
	Database         DatabaseConfig               `json:"database"`
	MQTT             MQTTConfig                   `json:"mqtt"`
	Serial           SerialConfig                 `json:"serial"`
	ControlAndUpdate ControlAndUpdateServerConfig `json:"updateServer"`
}

func LoadConfig(pathToJsonFile string) (config *Config, err error) {
	file, err := os.ReadFile(pathToJsonFile)
	if err != nil {
		return config, fmt.Errorf("failed to read file:%s %s", "\n",
			err.Error())
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal into Config struct:%s %s", "\n",
			err.Error())
	}

	return config, err
}
