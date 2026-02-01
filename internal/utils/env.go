package utils

import (
	"fmt"
	"os"
	"strings"
)

// ReadEnvOrFail returns the value or crashes the app.
// We do not want to run if critical config is missing.
func ReadEnvOrFail(key string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		panic(fmt.Sprintf("CRITICAL: Environment variable %s is not set", key))
	}
	return val
}

// ReadSecretOrFail reads a path from env, then reads the file content.
func ReadSecretOrFail(envName string) string {
	path := ReadEnvOrFail(envName)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Failed to read secret file at %s (env: %s): %v", path, envName, err))
	}
	value := strings.TrimSpace(string(content))
	if value == "" {
		panic(fmt.Sprintf("CRITICAL: Secret file at %s (env: %s) is empty", path, envName))
	}
	return value
}
