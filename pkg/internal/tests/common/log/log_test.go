// pkg/internal/tests/common/log/log.go
package log_test

import (
	"bytes"
	"fmt"
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

func TestLoggerWithOptions(t *testing.T) {
	// Test WithStdout option
	t.Run("WithStdout", func(t *testing.T) {
		var buf bytes.Buffer
		l := log.NewLogger("{test}", log.INFO, log.WithStdout())
		l.SetOutput(&buf)

		l.Info("stdout only message")
		output := buf.String()

		if !strings.Contains(output, "stdout only message") {
			t.Errorf("Expected output to contain message, got %q", output)
		}

		// Verify no file was created
		if _, err := os.Stat("./rego.log"); !os.IsNotExist(err) {
			t.Error("Expected no log file to be created with WithStdout option")
			os.Remove("./rego.log")
		}
	})

	// Test WithLogFile option
	t.Run("WithLogFile", func(t *testing.T) {
		customLogPath := "./custom_test.log"
		defer os.Remove(customLogPath)

		l := log.NewLogger("{test}", log.INFO, log.WithLogFile(customLogPath))
		l.Info("custom file message")

		// Check if custom file was created
		content, err := os.ReadFile(customLogPath)
		if err != nil {
			t.Fatalf("Failed to read custom log file: %v", err)
		}

		if !strings.Contains(string(content), "custom file message") {
			t.Errorf("Custom log file should contain message, got %s", string(content))
		}

		l.Delete()
	})

	// Test WithColor option
	t.Run("WithColor", func(t *testing.T) {
		var buf bytes.Buffer

		// Test with color disabled
		l1 := log.NewLogger("{test}", log.INFO, log.WithStdout(), log.WithColor(false))
		l1.SetOutput(&buf)
		l1.Info("no color message")
		noColorOutput := buf.String()

		if strings.Contains(noColorOutput, "\033[") {
			t.Error("Expected no color codes with WithColor(false)")
		}

		// Test with color enabled
		buf.Reset()
		l2 := log.NewLogger("{test}", log.INFO, log.WithStdout(), log.WithColor(true))
		l2.SetOutput(&buf)
		l2.Info("color message")
		colorOutput := buf.String()

		if !strings.Contains(colorOutput, "\033[") {
			t.Error("Expected color codes with WithColor(true)")
		}
	})

	// Test multiple options
	t.Run("MultipleOptions", func(t *testing.T) {
		var buf bytes.Buffer
		l := log.NewLogger("{test}", log.DEBUG, log.WithStdout(), log.WithColor(false))
		l.SetOutput(&buf)

		l.Debug("multi option message")
		output := buf.String()

		if !strings.Contains(output, "multi option message") {
			t.Errorf("Expected output to contain message, got %q", output)
		}

		if strings.Contains(output, "\033[") {
			t.Error("Expected no color codes with WithColor(false)")
		}

		// Verify no file was created
		if _, err := os.Stat("./rego.log"); !os.IsNotExist(err) {
			t.Error("Expected no log file with WithStdout option")
			os.Remove("./rego.log")
		}
	})

	// Test backwards compatibility (no options)
	t.Run("BackwardsCompatibility", func(t *testing.T) {
		defer os.Remove("./rego.log")

		l := log.NewLogger("{test}", log.INFO)
		l.Info("backwards compat message")

		// Check if default log file was created
		if _, err := os.Stat("./rego.log"); os.IsNotExist(err) {
			t.Error("Expected default log file to be created when no options provided")
		}

		l.Delete()
	})
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogger("{test}", log.INFO, log.WithStdout())
	l.SetOutput(&buf)

	// Test Printf methods
	t.Run("Printf", func(t *testing.T) {
		buf.Reset()
		l.Printf("formatted %s %d", "string", 42)
		if !strings.Contains(buf.String(), "formatted string 42") {
			t.Error("Printf failed")
		}
	})

	t.Run("Infof", func(t *testing.T) {
		buf.Reset()
		l.Infof("info %s", "formatted")
		if !strings.Contains(buf.String(), "info formatted") {
			t.Error("Infof failed")
		}
	})

	t.Run("Debugf", func(t *testing.T) {
		l.Verbosity = log.DEBUG
		buf.Reset()
		l.Debugf("debug %d", 123)
		if !strings.Contains(buf.String(), "debug 123") {
			t.Error("Debugf failed")
		}
	})

	t.Run("Warningf", func(t *testing.T) {
		l.Verbosity = log.INFO
		buf.Reset()
		l.Warningf("warning %v", "test")
		if !strings.Contains(buf.String(), "warning test") {
			t.Error("Warningf failed")
		}
	})

	t.Run("Errorf", func(t *testing.T) {
		buf.Reset()
		l.Errorf("error %s", "formatted")
		if !strings.Contains(buf.String(), "error formatted") {
			t.Error("Errorf failed")
		}
	})

	t.Run("Tracef", func(t *testing.T) {
		l.Verbosity = log.TRACE
		buf.Reset()
		l.Tracef("trace %s", "formatted")
		if !strings.Contains(buf.String(), "trace formatted") {
			t.Error("Tracef failed")
		}
	})
}

func TestPanicBehavior(t *testing.T) {
	t.Run("Panic", func(t *testing.T) {
		var buf bytes.Buffer
		l := log.NewLogger("{test}", log.INFO, log.WithStdout())
		l.SetOutput(&buf)

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic")
			} else if r != "panic message" {
				t.Errorf("Expected 'panic message', got %v", r)
			}
		}()

		l.Panic("panic message")
	})

	t.Run("Panicf", func(t *testing.T) {
		var buf bytes.Buffer
		l := log.NewLogger("{test}", log.INFO, log.WithStdout())
		l.SetOutput(&buf)

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic")
			} else if r != "formatted panic 42" {
				t.Errorf("Expected 'formatted panic 42', got %v", r)
			}
		}()

		l.Panicf("formatted panic %d", 42)
	})
}

func TestLogLevelFunction(t *testing.T) {
	tests := []struct {
		level    int
		colored  bool
		expected string
	}{
		{log.TRACE, false, "TRACE"},
		{log.DEBUG, false, "DEBUG"},
		{log.INFO, false, "INFO"},
		{log.WARNING, false, "WARNING"},
		{log.ERROR, false, "ERROR"},
		{log.FATAL, false, "FATAL"},
		{log.PANIC, false, "PANIC"},
		{999, false, "UNKNOWN"},
		{log.INFO, true, log.Green + "INFO" + log.Reset},
		{log.ERROR, true, log.Red + "ERROR" + log.Reset},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := log.LogLevel(tt.level, tt.colored)
			if result != tt.expected {
				t.Errorf("LogLevel(%d, %v) = %q, want %q", tt.level, tt.colored, result, tt.expected)
			}
		})
	}
}

func TestLogRotation(t *testing.T) {
	// Simple rotation test first
	t.Run("SimpleRotation", func(t *testing.T) {
		testLogPath := "./test_simple_rotate.log"
		defer func() {
			os.Remove(testLogPath)
			os.Remove(testLogPath + ".1")
		}()

		// Create logger with rotating file, writing to both file and stdout
		l := log.NewLogger("{test}", log.INFO, log.WithRotatingFile(testLogPath, 1, 2, false))

		// Write > 1MB to trigger rotation
		for i := 0; i < 50; i++ {
			l.Info(strings.Repeat("A", 25000))
		}

		l.Close()

		// Check if rotation happened
		if _, err := os.Stat(testLogPath + ".1"); os.IsNotExist(err) {
			// Debug: check what files exist
			files, _ := os.ReadDir(".")
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "test_simple_rotate") {
					info, _ := f.Info()
					t.Logf("Found file: %s (size: %d bytes)", f.Name(), info.Size())
				}
			}
			t.Error("Expected rotated file .1 to exist after writing > 1MB")
		} else {
			t.Log("Rotation successful - .1 file exists")
		}
	})

	// Test WithRotatingLogFile option
	t.Run("WithRotatingLogFile", func(t *testing.T) {
		testLogPath := "./test_rotate.log"
		defer func() {
			// Clean up all rotated files
			os.Remove(testLogPath)
			for i := 1; i <= 3; i++ {
				os.Remove(fmt.Sprintf("%s.%d", testLogPath, i))
				os.Remove(fmt.Sprintf("%s.%d.gz", testLogPath, i))
			}
		}()

		// Create logger with small rotation size (1MB) for testing
		l := log.NewLogger("{test}", log.INFO, log.WithRotatingFile(testLogPath, 1, 3, false))

		// Write more than 1MB to trigger rotation
		totalWritten := 0
		for i := 0; i < 50; i++ {
			msg := strings.Repeat("X", 25000)
			l.Info(msg)
			totalWritten += len(msg) + 200 // approximate overhead per log line
		}

		l.Close()

		// Check if rotation occurred
		if _, err := os.Stat(testLogPath + ".1"); os.IsNotExist(err) {
			t.Error("Expected rotated file .1 to exist")
			t.Logf("Wrote approximately %d bytes (%d MB)", totalWritten, totalWritten/1024/1024)
		}
	})

	// Test rotation with compression
	t.Run("WithRotatingLogFileCompressed", func(t *testing.T) {
		testLogPath := "./test_rotate_compress.log"
		defer func() {
			// Clean up all rotated files
			os.Remove(testLogPath)
			for i := 1; i <= 3; i++ {
				os.Remove(fmt.Sprintf("%s.%d", testLogPath, i))
				os.Remove(fmt.Sprintf("%s.%d.gz", testLogPath, i))
			}
		}()

		// Create logger with compression enabled
		l := log.NewLogger("{test}", log.INFO, log.WithRotatingFile(testLogPath, 1, 3, true))

		// Write more than 1MB to trigger rotation and compression
		for i := 0; i < 50; i++ {
			l.Info(strings.Repeat("Y", 25000))
		}

		l.Close()

		// Check if compressed file exists
		compressedFile := testLogPath + ".1.gz"
		if _, err := os.Stat(compressedFile); os.IsNotExist(err) {
			t.Error("Expected compressed rotated file .1.gz to exist")
		}

		// Check that uncompressed rotated file was removed
		if _, err := os.Stat(testLogPath + ".1"); !os.IsNotExist(err) {
			t.Error("Uncompressed rotated file should have been removed after compression")
		}
	})

	// Test default logger has rotation enabled
	t.Run("DefaultLoggerRotation", func(t *testing.T) {
		// Clean up any existing rego.log
		defer func() {
			os.Remove("./rego.log")
			for i := 1; i <= 5; i++ {
				os.Remove(fmt.Sprintf("./rego.log.%d.gz", i))
			}
		}()

		l := log.NewLogger("{test}", log.INFO) // No options = default with rotation

		// Just verify it was created without error
		l.Info("test message")
		l.Close()

		// Check if log file exists
		if _, err := os.Stat("./rego.log"); os.IsNotExist(err) {
			t.Error("Default logger should create rego.log")
		}
	})

	// Test max backups limit
	t.Run("MaxBackupsLimit", func(t *testing.T) {
		testLogPath := "./test_max_backups.log"
		maxBackups := 2
		defer func() {
			// Clean up all possible files
			os.Remove(testLogPath)
			for i := 1; i <= 5; i++ {
				os.Remove(fmt.Sprintf("%s.%d", testLogPath, i))
			}
		}()

		l := log.NewLogger("{test}", log.INFO, log.WithRotatingFile(testLogPath, 1, maxBackups, false))

		// Trigger multiple rotations - write 5MB total (5 rotations)
		for i := 0; i < 200; i++ {
			l.Info(strings.Repeat("Z", 25000)) // ~25KB per line
		}

		l.Close()

		// Check that only maxBackups files exist
		for i := 1; i <= maxBackups; i++ {
			if _, err := os.Stat(fmt.Sprintf("%s.%d", testLogPath, i)); os.IsNotExist(err) {
				t.Errorf("Expected backup file .%d to exist", i)
			}
		}

		// Check that older backups were removed
		if _, err := os.Stat(fmt.Sprintf("%s.%d", testLogPath, maxBackups+1)); !os.IsNotExist(err) {
			t.Errorf("Backup file .%d should have been removed (max backups: %d)", maxBackups+1, maxBackups)
		}
	})
}

func TestColorStripping(t *testing.T) {
	// Test that colors are stripped from file output
	t.Run("FileOutputStripsColors", func(t *testing.T) {
		testFile := "./test_color_strip.log"
		defer os.Remove(testFile)

		l := log.NewLogger("{test}", log.INFO, log.WithLogFile(testFile))
		l.Color = true // Enable colors

		l.Info("This should not have colors in file")
		l.Close()

		// Read file content
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		// Check that ANSI codes are not present
		if strings.Contains(string(content), "\033[") {
			t.Error("File output should not contain ANSI color codes")
		}

		// Check that the message is present
		if !strings.Contains(string(content), "This should not have colors in file") {
			t.Error("File should contain the log message")
		}
	})
}
