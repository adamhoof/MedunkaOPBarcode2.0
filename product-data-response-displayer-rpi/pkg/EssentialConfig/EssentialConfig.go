package essential

import (
	"encoding/json"
	"os"
)

type Config struct {
	SerialPortName                     string `json:"SerialPortName"`
	SerialPortBaudRate                 int    `json:"SerialPortBaudRate"`
	BarcodeReadingTerminationDelimiter byte   `json:"BarcodeReadingTerminationDelimiter"`
	MQTTClientName                     string `json:"MQTTClientName"`
	MQTTServer                         string `json:"MQTTServer"`
	MQTTServerUsername                 string `json:"MQTTServerUsername"`
	MQTTServerPassword                 string `json:"MQTTServerPassword"`
}

func LoadConfigFromJsonFile(pathToJsonFile string, unpackInto *Config) {
	file, err := os.ReadFile(pathToJsonFile)
	if err != nil {
		panic("Failed to read .json file, make sure you have this file, check spelling, path to file: " + err.Error())
	}

	err = json.Unmarshal(file, &unpackInto)
	if err != nil {
		panic("Failed to extract Config.json into struct: " + err.Error())
	}
}
