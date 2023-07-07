// pkg/internal/tests/common/config/config_test.go
package config

import (
	"os"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/config"
)

func TestGetEnv(t *testing.T) {
	// Set an environment variable for testing
	os.Setenv("TEST_VAR", "test value")
	defer os.Unsetenv("TEST_VAR") // Ensure cleanup after test

	value := config.GetEnv("TEST_VAR", "default value")

	if value != "test value" {
		t.Errorf("GetEnv(\"TEST_VAR\", \"default value\") = %s; want \"test value\"", value)
	}
}
