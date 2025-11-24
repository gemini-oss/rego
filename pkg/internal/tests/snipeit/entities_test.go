package snipeit_test

import (
	"encoding/json"
	"testing"

	"github.com/gemini-oss/rego/pkg/snipeit"
)

// TestTimestampUnmarshal tests the Timestamp type with various date formats
func TestTimestampUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string // expected date in 2006-01-02 format
		wantErr  bool
	}{
		{
			name:     "object_with_date",
			json:     `{"date": "2028-06-10", "formatted": "06/10/2028"}`,
			expected: "2028-06-10",
		},
		{
			name:     "object_with_datetime",
			json:     `{"datetime": "2025-07-07 01:19:07", "formatted": "07/07/2025 1:19AM"}`,
			expected: "2025-07-07",
		},
		{
			name:     "iso8601_string",
			json:     `"2025-01-01T00:00:00.000000Z"`,
			expected: "2025-01-01",
		},
		{
			name:     "snipetime_string",
			json:     `"2025-01-01 15:04:05"`,
			expected: "2025-01-01",
		},
		{
			name:     "date_only_string",
			json:     `"2025-01-01"`,
			expected: "2025-01-01",
		},
		{
			name:    "invalid_string",
			json:    `"not a date"`,
			wantErr: true,
		},
		{
			name:    "wrong_type",
			json:    `123`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts snipeit.Timestamp
			err := json.Unmarshal([]byte(tt.json), &ts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := ts.Format("2006-01-02")
				if got != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, got)
				}
			}
		})
	}
}

// TestDateInfoUnmarshal tests the DateInfo type
func TestDateInfoUnmarshal(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expectedDate   string
		expectedFormat string
	}{
		{
			name:           "with_datetime",
			json:           `{"datetime": "2025-01-01 00:00:00", "formatted": "01/01/2025"}`,
			expectedDate:   "2025-01-01 00:00:00",
			expectedFormat: "01/01/2025",
		},
		{
			name:           "with_date_only",
			json:           `{"date": "2025-01-01", "formatted": "01/01/2025"}`,
			expectedDate:   "2025-01-01",
			expectedFormat: "01/01/2025",
		},
		{
			name:           "prefer_datetime_over_date",
			json:           `{"datetime": "2025-01-01 00:00:00", "date": "2025-01-02", "formatted": "01/01/2025"}`,
			expectedDate:   "2025-01-01 00:00:00",
			expectedFormat: "01/01/2025",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var di snipeit.DateInfo
			if err := json.Unmarshal([]byte(tt.json), &di); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if di.Date != tt.expectedDate {
				t.Errorf("Date: expected %s, got %s", tt.expectedDate, di.Date)
			}

			if di.Formatted != tt.expectedFormat {
				t.Errorf("Formatted: expected %s, got %s", tt.expectedFormat, di.Formatted)
			}
		})
	}
}

// TestBoolIntMarshalUnmarshal tests the BoolInt type
func TestBoolIntMarshalUnmarshal(t *testing.T) {
	// Test unmarshalling
	unmarshalTests := []struct {
		name     string
		json     string
		expected bool
	}{
		{"zero", `0`, false},
		{"one", `1`, true},
		{"string_zero", `"0"`, false},
		{"string_one", `"1"`, true},
		{"bool_false", `false`, false},
		{"bool_true", `true`, true},
		{"null", `null`, false},
	}

	for _, tt := range unmarshalTests {
		t.Run("unmarshal_"+tt.name, func(t *testing.T) {
			var b snipeit.BoolInt
			if err := json.Unmarshal([]byte(tt.json), &b); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if bool(b) != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, bool(b))
			}
		})
	}

	// Test marshalling
	marshalTests := []struct {
		name     string
		value    snipeit.BoolInt
		expected string
	}{
		{"false", false, "0"},
		{"true", true, "1"},
	}

	for _, tt := range marshalTests {
		t.Run("marshal_"+tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

// TestHardwareFieldTypes tests that Hardware struct handles various field type variations
func TestHardwareFieldTypes(t *testing.T) {
	// Test the warranty_expires field specifically
	t.Run("warranty_expires_as_object", func(t *testing.T) {
		testJSON := `{
			"id": 1,
			"warranty_expires": {
				"date": "2028-06-10",
				"formatted": "06/10/2028"
			}
		}`

		var hw snipeit.Hardware[snipeit.HardwareGET]
		if err := json.Unmarshal([]byte(testJSON), &hw); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if hw.WarrantyExpires == nil {
			t.Fatal("warranty_expires is nil")
		}

		date := hw.WarrantyExpires.Format("2006-01-02")
		if date != "2028-06-10" {
			t.Errorf("Expected 2028-06-10, got %s", date)
		}
	})

	// Test custom_fields structure
	t.Run("custom_fields_structure", func(t *testing.T) {
		testJSON := `{
			"id": 123,
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
		if err := json.Unmarshal([]byte(testJSON), &hw); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if hw.CustomFields == nil || *hw.CustomFields == nil {
			t.Fatal("custom_fields is nil")
		}

		fields := *hw.CustomFields
		macField, ok := fields["MAC Address"]
		if !ok {
			t.Fatal("MAC Address field not found")
		}

		if macField["value"] != "AC:07:75:1A:21:78" {
			t.Errorf("Unexpected MAC value: %s", macField["value"])
		}
	})

	// Test method-specific field types
	t.Run("archived_field_type_difference", func(t *testing.T) {
		// GET returns string
		getJSON := `{"id": 1, "archived": "false"}`
		var hwGet snipeit.Hardware[snipeit.HardwareGET]
		if err := json.Unmarshal([]byte(getJSON), &hwGet); err != nil {
			t.Fatalf("GET unmarshal failed: %v", err)
		}
		if hwGet.Method.Archived != "false" {
			t.Errorf("Expected archived='false', got %s", hwGet.Method.Archived)
		}

		// POST uses bool
		hwPost := snipeit.Hardware[snipeit.HardwarePOST]{
			Method: snipeit.HardwarePOST{
				Archived: true,
			},
		}

		// Just verify the field exists and is the right type
		if !hwPost.Method.Archived {
			t.Error("POST archived field should be true")
		}
	})
}

// TestPaginatedList tests the generic paginated list structure
func TestPaginatedList(t *testing.T) {
	testJSON := `{
		"total": 2,
		"rows": [
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"}
		]
	}`

	var list snipeit.PaginatedList[snipeit.Record]
	if err := json.Unmarshal([]byte(testJSON), &list); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Test basic fields
	if list.Total != 2 {
		t.Errorf("Expected total=2, got %d", list.Total)
	}

	if list.Rows == nil || *list.Rows == nil {
		t.Fatal("Rows is nil")
	}

	if len(*list.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(*list.Rows))
	}

	// Test helper methods
	if list.TotalCount() != 2 {
		t.Errorf("TotalCount() = %d, want 2", list.TotalCount())
	}

	// Test Elements()
	elements := list.Elements()
	if elements == nil || len(*elements) != 2 {
		t.Error("Elements() returned wrong data")
	}

	// Test first item
	firstItem := (*list.Rows)[0]
	if firstItem.ID != 1 || firstItem.Name != "Item 1" {
		t.Errorf("First item data mismatch")
	}
}

// TestLicenseFieldTypes tests License-specific field type variations
func TestLicenseFieldTypes(t *testing.T) {
	t.Run("seats_field_type_difference", func(t *testing.T) {
		// GET returns int
		getJSON := `{"id": 1, "name": "Test License", "seats": 10}`
		var licGet snipeit.License[snipeit.LicenseGET]
		if err := json.Unmarshal([]byte(getJSON), &licGet); err != nil {
			t.Fatalf("GET unmarshal failed: %v", err)
		}
		if licGet.Method.Seats != 10 {
			t.Errorf("Expected seats=10, got %d", licGet.Method.Seats)
		}

		// POST uses string - test marshalling
		licPost := snipeit.License[snipeit.LicensePOST]{
			LicenseBase: &snipeit.LicenseBase{
				SnipeIT: &snipeit.SnipeIT{
					Record: &snipeit.Record{Name: "Test License"},
				},
			},
			Method: snipeit.LicensePOST{
				Seats: "10",
			},
		}

		data, err := json.Marshal(licPost)
		if err != nil {
			t.Fatalf("POST marshal failed: %v", err)
		}

		// Verify it marshals as string
		var check map[string]interface{}
		_ = json.Unmarshal(data, &check)
		if seats, ok := check["seats"].(string); !ok || seats != "10" {
			t.Error("POST seats field should marshal as string '10'")
		}
	})
}

// TestMessagesUnmarshal tests the Messages type that can be string or map
func TestMessagesUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		isString  bool
		stringVal string
		mapKeys   []string
	}{
		{
			name:      "string_message",
			json:      `"Success"`,
			isString:  true,
			stringVal: "Success",
		},
		{
			name:     "map_message",
			json:     `{"name": ["The name field is required"], "email": ["Invalid email"]}`,
			isString: false,
			mapKeys:  []string{"name", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg snipeit.Messages
			if err := json.Unmarshal([]byte(tt.json), &msg); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if msg.IsString != tt.isString {
				t.Errorf("IsString = %v, want %v", msg.IsString, tt.isString)
			}

			if tt.isString && msg.StringValue != tt.stringVal {
				t.Errorf("StringValue = %s, want %s", msg.StringValue, tt.stringVal)
			}

			if !tt.isString && len(msg.MapValue) != len(tt.mapKeys) {
				t.Errorf("MapValue has %d keys, want %d", len(msg.MapValue), len(tt.mapKeys))
			}
		})
	}
}
