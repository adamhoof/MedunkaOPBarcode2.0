package config

import (
	"encoding/json"
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

type UpdateServerConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	UploadPath string `json:"uploadPath"`
	Delimiter  string `json:"delimiter"`
}
type Config struct {
	Database     DatabaseConfig     `json:"database"`
	MQTT         MQTTConfig         `json:"mqtt"`
	Serial       SerialConfig       `json:"serial"`
	UpdateServer UpdateServerConfig `json:"updateServer"`
}

func LoadConfig(pathToJsonFile string) (config *Config, err error) {
	file, err := os.ReadFile(pathToJsonFile)
	if err != nil {
		panic("Failed to read .json file, make sure you have this file, check spelling, path to file: " + err.Error())
	}

	err = json.Unmarshal(file, &config)
	return config, err
}
