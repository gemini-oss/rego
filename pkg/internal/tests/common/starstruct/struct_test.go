// pkg/internal/tests/common/starstruct/starstruct_test.go
package starstruct_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/starstruct"
)

// Example struct for testing
type TestStruct struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Tags    []string `json:"tags"`
	Address struct {
		City  string `json:"city"`
		State string `json:"state"`
	} `json:"address"`
}

// TestPrettyJSON tests the PrettyJSON function for various struct inputs.
func TestPrettyJSON(t *testing.T) {
	testStruct := TestStruct{
		Name: "Anthony Dardano",
		Age:  0,
		Tags: []string{"Staff Enterprise Infrastructure Engineer", "DJ"},
		Address: struct {
			City  string "json:\"city\""
			State string "json:\"state\""
		}{City: "N/A", State: "FL"},
	}

	expectedJSON := `{
  "name": "Anthony Dardano",
  "age": 0,
  "tags": [
    "Staff Enterprise Infrastructure Engineer",
    "DJ"
  ],
  "address": {
    "city": "N/A",
    "state": "FL"
  }
}`

	got, err := starstruct.PrettyJSON(testStruct)
	if err != nil {
		t.Errorf("PrettyJSON() error = %v, wantErr false", err)
		return
	}

	var gotJSON, wantJSON interface{}
	json.Unmarshal([]byte(got), &gotJSON)
	json.Unmarshal([]byte(expectedJSON), &wantJSON)

	if !reflect.DeepEqual(gotJSON, wantJSON) {
		t.Errorf("PrettyJSON() mismatch. Got = %v, want = %v", got, expectedJSON)
	}
}

// TestStructToMap tests the StructToMap function for various struct inputs.
func TestToMap(t *testing.T) {
	testStruct := TestStruct{
		Name: "Anthony Dardano",
		Age:  0,
		Tags: []string{"Staff Enterprise Infrastructure Engineer", "DJ"},
		Address: struct {
			City  string `json:"city"`
			State string `json:"state"`
		}{City: "N/A", State: "FL"},
	}

	expectedMap := map[string]interface{}{
		"name":    "Anthony Dardano",
		"age":     0,
		"tags":    []interface{}{"Staff Enterprise Infrastructure Engineer", "DJ"},
		"address": map[string]interface{}{"city": "N/A", "state": "FL"},
	}

	got, err := starstruct.ToMap(testStruct, true)
	if err != nil {
		t.Errorf("ToMap() error = %v", err)
		return
	}

	if !reflect.DeepEqual(got, expectedMap) {
		t.Errorf("ToMap() = %#v, want %#v", got, expectedMap)
		// Additional detailed logging
		for k, v := range got {
			if ev, ok := expectedMap[k]; ok {
				if !reflect.DeepEqual(v, ev) {
					t.Errorf("Mismatch in key '%v': got %#v, want %#v", k, v, ev)
				}
			} else {
				t.Errorf("Key '%v' found in got, but not in want", k)
			}
		}
		for k := range expectedMap {
			if _, ok := got[k]; !ok {
				t.Errorf("Key '%v' found in want, but not in got", k)
			}
		}
	}
}

// TestFlattenStructFields tests the FlattenStructFields function for various struct inputs.
func TestFlattenStructFields(t *testing.T) {
	testStruct := TestStruct{
		Name: "Anthony Dardano",
		Age:  0,
		Tags: []string{"Staff Enterprise Infrastructure Engineer", "DJ"},
		Address: struct {
			City  string "json:\"city\""
			State string "json:\"state\""
		}{City: "N/A", State: "FL"},
	}

	fields := []string{"name", "age", "tags", "address.city", "address.state"}
	expectedSlice := [][]string{
		{"name", "Anthony Dardano"},
		{"age", "0"},
		{"tags.0", "Staff Enterprise Infrastructure Engineer"},
		{"tags.1", "DJ"},
		{"address.city", "N/A"},
		{"address.state", "FL"},
	}

	got, err := starstruct.FlattenStructFields(testStruct, &fields)
	if err != nil {
		t.Errorf("FlattenStructFields() error = %v, wantErr false", err)
		return
	}
	if !reflect.DeepEqual(got, expectedSlice) {
		t.Errorf("FlattenStructFields() = %v, want %v", got, expectedSlice)
	}
}
