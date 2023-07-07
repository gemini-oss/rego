// pkg/common/config/config.go
package config

import (
	"os"
	"strconv"
)

// A simple function to get environment variables with a default value
func GetEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// A helper function to get environment variables as integers
func GetEnvAsInt(key string, defaultVal int) int {
	valueStr := GetEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
