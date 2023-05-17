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

type HTTPDatabaseUpdateServerConfig struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	Endpoint  string `json:"endpoint"`
	Delimiter string `json:"delimiter"`
	TableName string `json:"tableName"`
	//provide only directory of where to put this db update .csv file (it is created temporary and thus assigned a random name)
	OutputCSVDirectory string `json:"outputCSVDirectory"`
}

type CLIControlAppConfig struct {
	//provide the full path to the files
	MDBFileLocation        string `json:"mdbFileLocation"`
	OutputCSVLocation      string `json:"outputCSVLocation"`
	ShellMDBParserLocation string `json:"shellMDBFileParser"`
}
type Config struct {
	Database           DatabaseConfig                 `json:"database"`
	MQTT               MQTTConfig                     `json:"mqtt"`
	Serial             SerialConfig                   `json:"serial"`
	HTTPDatabaseUpdate HTTPDatabaseUpdateServerConfig `json:"updateServer"`
	CLIControlApp      CLIControlAppConfig            `json:"cliControl"`
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
