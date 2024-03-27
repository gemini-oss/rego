// pkg/common/config/config.go
package config

import (
	"os"
	"strconv"
)

// Get environment variable (no default value)
func GetEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""
}

// Get environment variable as integer (no default value)
func GetEnvAsInt(key string) int {
	valueStr := GetEnv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return 0
}
