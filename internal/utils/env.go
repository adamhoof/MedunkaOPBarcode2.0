package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetEnvOrPanic returns the value or crashes the app.
// We do not want to run if critical config is missing.
func GetEnvOrPanic(key string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		panic(fmt.Sprintf("CRITICAL: Environment variable %s is not set", key))
	}
	return val
}

// GetEnvAsInt returns a parsed int or panics.
func GetEnvAsInt(key string) int {
	val := GetEnvOrPanic(key)
	parsed, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Environment variable %s must be an int: %v", key, err))
	}
	return parsed
}

// GetEnvAsInt64 returns a parsed int64 or panics.
func GetEnvAsInt64(key string) int64 {
	val := GetEnvOrPanic(key)
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Environment variable %s must be an int64: %v", key, err))
	}
	return parsed
}

// GetEnvAsDuration returns a parsed duration or panics.
func GetEnvAsDuration(key string) time.Duration {
	val := GetEnvOrPanic(key)
	parsed, err := time.ParseDuration(val)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Environment variable %s must be a duration: %v", key, err))
	}
	return parsed
}

// ReadSecretOrFail reads a path from env, then reads the file content.
func ReadSecretOrFail(envName string) string {
	path := GetEnvOrPanic(envName)
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
