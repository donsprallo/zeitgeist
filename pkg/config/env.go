package config

import (
	"os"
	"strconv"
)

// GetEnvStr load a string value from environment key. If environment key
// does not exist, a fallback value is returned.
func GetEnvStr(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetEnvInt load a integer value from environment key. Of environment key
// does not exist, a fallback value is returned.
func GetEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}
