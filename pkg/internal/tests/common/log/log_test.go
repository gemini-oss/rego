// pkg/internal/tests/common/log/log.go
package log_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/log"
)

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogger("{test}", log.INFO)
	l.Color = false
	l.SetOutput(&buf) // Redirect output to buffer for testing

	testCases := []struct {
		name     string
		logFunc  func()
		expected string
		level    int
	}{
		{
			name:     "Trace level",
			logFunc:  func() { l.Trace("Trace message") },
			expected: "TRACE - Trace message",
			level:    log.TRACE,
		},
		{
			name:     "Debug level",
			logFunc:  func() { l.Debug("Debug message") },
			expected: "DEBUG - Debug message",
			level:    log.DEBUG,
		},
		{
			name:     "Info level",
			logFunc:  func() { l.Print("Info message") },
			expected: "INFO - Info message",
			level:    log.INFO,
		},
		{
			name:     "Warning level",
			logFunc:  func() { l.Warning("Warning message") },
			expected: "WARNING - Warning message",
			level:    log.WARNING,
		},
		{
			name:     "Error level",
			logFunc:  func() { l.Error("Error message") },
			expected: "ERROR - Error message",
			level:    log.ERROR,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l.Verbosity = tc.level // Set the verbosity for each test case
			buf.Reset()            // Clear buffer before each test
			tc.logFunc()           // Call the log function
			output := buf.String()

			if !strings.Contains(output, tc.expected) {
				t.Errorf("Expected output to contain %q, got %q", tc.expected, output)
			}
		})
	}

	// Clean up
	l.Delete()
}

func TestFileOutput(t *testing.T) {
	logFilePath := "./test.log"
	defer os.Remove(logFilePath) // Clean up after test

	l := log.NewLogger("{test}", log.INFO)
	l.SetNewFile(logFilePath)

	l.Println("Test message to file")

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "Test message to file") {
		t.Errorf("Log file should contain 'Test message to file', got %s", string(content))
	}

	// Clean up
	l.Delete()
}

func TestLogLevelChange(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogger("{test}", log.INFO)
	l.Color = false
	l.SetOutput(&buf)

	l.Verbosity = log.DEBUG
	l.Debug("Debug message")
	if !strings.Contains(buf.String(), "DEBUG - Debug message") {
		t.Errorf("Logger should output debug messages when verbosity is DEBUG")
	}

	buf.Reset()
	l.Verbosity = log.ERROR
	l.Debug("This should not appear")
	if strings.Contains(buf.String(), "This should not appear") {
		t.Errorf("Logger should not output debug messages when verbosity is ERROR")
	}

	// Clean up
	l.Delete()
}

func TestLoggerColorOutput(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogger("{test}", log.INFO)
	l.Color = true    // Enable color output
	l.SetOutput(&buf) // Redirect output to buffer for testing

	testCases := []struct {
		name      string
		logFunc   func()
		expected  string
		level     int
		colorCode string
	}{
		{
			name:      "Trace level with color",
			logFunc:   func() { l.Trace("Trace message") },
			expected:  "Trace message",
			level:     log.TRACE,
			colorCode: log.Blue,
		},
		{
			name:      "Debug level with color",
			logFunc:   func() { l.Debug("Debug message") },
			expected:  "Debug message",
			level:     log.DEBUG,
			colorCode: log.Cyan,
		},
		{
			name:      "Info level with color",
			logFunc:   func() { l.Print("Info message") },
			expected:  "Info message",
			level:     log.INFO,
			colorCode: log.Green,
		},
		{
			name:      "Warning level with color",
			logFunc:   func() { l.Warning("Warning message") },
			expected:  "Warning message",
			level:     log.WARNING,
			colorCode: log.Yellow,
		},
		{
			name:      "Error level with color",
			logFunc:   func() { l.Error("Error message") },
			expected:  "Error message",
			level:     log.ERROR,
			colorCode: log.Red,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l.Verbosity = tc.level // Set the verbosity for each test case
			buf.Reset()            // Clear buffer before each test
			tc.logFunc()           // Call the log function
			output := buf.String()

			if !strings.Contains(output, tc.expected) {
				t.Errorf("Expected output to contain %q, got %q", tc.expected, output)
			}
			if !strings.Contains(output, tc.colorCode) {
				t.Errorf("Expected output to contain color code %q, got %q", tc.colorCode, output)
			}
		})
	}

	// Clean up
	l.Delete()
}
