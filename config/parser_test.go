package config

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {

	expectedConfig := Config{
		Database: DatabaseConfig{
			Host: "yeet.com",
			Port: 1111,
		},
		MQTT: MQTTConfig{
			ServerAndPort:      "mqtt.example.com",
			ProductDataRequest: "request",
		},
	}

	tmpConfigFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal("Failed to create temporary config file")
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tmpConfigFile.Name())

	configContent, err := json.Marshal(expectedConfig)
	if err != nil {
		t.Fatal("Failed to marshal expected config")
	}

	if _, err := tmpConfigFile.Write(configContent); err != nil {
		t.Fatal("Failed to write config content to temporary file")
	}

	config, err := LoadConfig(tmpConfigFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !reflect.DeepEqual(config, &expectedConfig) {
		t.Errorf("Unexpected values in the config: got %v, want %v", config, expectedConfig)
	}
}
