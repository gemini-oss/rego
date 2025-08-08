// pkg/internal/tests/snipeit/snipeit_test.go
package snipeit_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/snipeit"
)

// TestMode determines whether to use live API or fixtures
type TestMode int

const (
	TestModeLive TestMode = iota
	TestModeFixture
	TestModeRecord // Record live responses as fixtures
)

// IntegrationTest provides infrastructure for testing with real API data
type IntegrationTest struct {
	t      *testing.T
	client *snipeit.Client
	mode   TestMode
}

// NewIntegrationTest creates a new integration test helper
func NewIntegrationTest(t *testing.T) *IntegrationTest {
	mode := TestModeFixture // default to fixtures

	// Check environment to determine test mode
	testMode := os.Getenv("SNIPEIT_TEST_MODE")
	switch testMode {
	case "live":
		mode = TestModeLive
	case "record":
		mode = TestModeRecord
	}

	it := &IntegrationTest{
		t:    t,
		mode: mode,
	}

	// Setup client if using live API
	if mode == TestModeLive || mode == TestModeRecord {
		url := os.Getenv("SNIPEIT_URL")
		token := os.Getenv("SNIPEIT_TOKEN")

		if url == "" || token == "" {
			t.Skip("SNIPEIT_URL and SNIPEIT_TOKEN required for live tests")
		}

		// Set environment variables for NewClient
		os.Setenv("SNIPEIT_URL", url)
		os.Setenv("SNIPEIT_TOKEN", token)

		client := snipeit.NewClient(log.INFO)
		it.client = client
	}

	return it
}

// GetHardwareList returns hardware list from API or fixture
func (it *IntegrationTest) GetHardwareList() (*snipeit.HardwareList, error) {
	switch it.mode {
	case TestModeLive:
		return it.client.Assets().GetAllAssets()

	case TestModeRecord:
		// Get from API and save
		list, err := it.client.Assets().GetAllAssets()
		if err != nil {
			return nil, err
		}

		// Save to fixture
		if err := it.saveFixture("hardware_list.json", list); err != nil {
			it.t.Logf("Warning: failed to save fixture: %v", err)
		}

		return list, nil

	case TestModeFixture:
		var list snipeit.HardwareList
		if err := it.loadFixture("hardware_list.json", &list); err != nil {
			return nil, err
		}
		return &list, nil

	default:
		return nil, fmt.Errorf("unknown test mode: %v", it.mode)
	}
}

// GetLicenseList returns license list from API or fixture
func (it *IntegrationTest) GetLicenseList() (*snipeit.LicenseList, error) {
	switch it.mode {
	case TestModeLive:
		return it.client.Licenses().GetAllLicenses()

	case TestModeRecord:
		list, err := it.client.Licenses().GetAllLicenses()
		if err != nil {
			return nil, err
		}

		if err := it.saveFixture("license_list.json", list); err != nil {
			it.t.Logf("Warning: failed to save fixture: %v", err)
		}

		return list, nil

	case TestModeFixture:
		var list snipeit.LicenseList
		if err := it.loadFixture("license_list.json", &list); err != nil {
			return nil, err
		}
		return &list, nil

	default:
		return nil, fmt.Errorf("unknown test mode: %v", it.mode)
	}
}

// saveFixture saves data as a JSON fixture
func (it *IntegrationTest) saveFixture(name string, data any) error {
	dir := "testdata"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Add timestamp to fixture
	type fixtureWrapper struct {
		Timestamp string `json:"timestamp"`
		Data      any    `json:"data"`
	}

	wrapped := fixtureWrapper{
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
	}

	b, err := json.MarshalIndent(wrapped, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, name)
	return os.WriteFile(path, b, 0644)
}

// loadFixture loads data from a JSON fixture
func (it *IntegrationTest) loadFixture(name string, v any) error {
	path := filepath.Join("testdata", name)

	// Check if fixture exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("fixture not found: %s (run with SNIPEIT_TEST_MODE=record to create)", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Unwrap the fixture
	var wrapped struct {
		Timestamp string          `json:"timestamp"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &wrapped); err != nil {
		// Try direct unmarshal for backward compatibility
		return json.Unmarshal(data, v)
	}

	// Check fixture age
	if timestamp, err := time.Parse(time.RFC3339, wrapped.Timestamp); err == nil {
		age := time.Since(timestamp)
		if age > 30*24*time.Hour {
			it.t.Logf("Warning: fixture %s is %v old", name, age)
		}
	}

	return json.Unmarshal(wrapped.Data, v)
}

// TestStructCompatibility validates that our structs match the API response
func TestStructCompatibility(t *testing.T) {
	it := NewIntegrationTest(t)

	t.Run("Hardware", func(t *testing.T) {
		list, err := it.GetHardwareList()
		if err != nil {
			t.Fatalf("Failed to get hardware list: %v", err)
		}

		// Validate we got data
		if list.Total == 0 {
			t.Error("No hardware items returned")
		}

		// Check for specific fields that have caused issues
		if list.Rows != nil && *list.Rows != nil && len(*list.Rows) > 0 {
			item := (*list.Rows)[0]

			// Log field types for debugging
			t.Logf("Sample hardware item fields:")
			t.Logf("  ID: %d", item.ID)
			t.Logf("  Serial: %s", item.Serial)

			if item.WarrantyExpires != nil {
				t.Logf("  WarrantyExpires: %s", item.WarrantyExpires.Format("2006-01-02"))
			}

			if item.CustomFields != nil {
				t.Logf("  CustomFields: %d fields", len(*item.CustomFields))
			}
		}
	})

	t.Run("License", func(t *testing.T) {
		list, err := it.GetLicenseList()
		if err != nil {
			t.Fatalf("Failed to get license list: %v", err)
		}

		if list.Total == 0 {
			t.Error("No license items returned")
		}
	})
}

// TestFieldTypeValidation ensures field types match API responses
func TestFieldTypeValidation(t *testing.T) {
	// Test specific problematic fields
	tests := []struct {
		name     string
		testFunc func() error
	}{
		{
			name: "warranty_expires_as_object",
			testFunc: func() error {
				// Test data with warranty_expires as object
				data := `{
					"warranty_expires": {
						"date": "2028-06-10",
						"formatted": "06/10/2028"
					}
				}`

				var hw snipeit.Hardware[snipeit.HardwareGET]
				return json.Unmarshal([]byte(data), &hw)
			},
		},
		{
			name: "custom_fields_structure",
			testFunc: func() error {
				// Test actual custom_fields structure
				data := `{
					"custom_fields": {
						"MAC Address": {
							"field": "_snipeit_mac_address_1",
							"value": "AC:07:75:1A:21:78",
							"field_format": "MAC",
							"element": "text"
						}
					}
				}`

				var hw snipeit.Hardware[snipeit.HardwareGET]
				return json.Unmarshal([]byte(data), &hw)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.testFunc(); err != nil {
				t.Errorf("Field validation failed: %v", err)
			}
		})
	}
}
