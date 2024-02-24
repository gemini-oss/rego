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
func TestStructToMap(t *testing.T) {
	testStruct := TestStruct{
		Name: "Anthony Dardano",
		Age:  0,
		Tags: []string{"Staff Enterprise Infrastructure Engineer", "DJ"},
		Address: struct {
			City  string "json:\"city\""
			State string "json:\"state\""
		}{City: "N/A", State: "FL"},
	}

	expectedMap := map[string]string{
		"name":    "Anthony Dardano",
		"tags":    "Staff Enterprise Infrastructure Engineerᕙ(▀̿̿Ĺ̯̿̿▀̿ ̿)ᕗDJ",
		"address": "{N/A FL}",
	}

	got := starstruct.StructToMap(testStruct)
	if !reflect.DeepEqual(got, expectedMap) {
		t.Errorf("StructToMap() = %v, want %v", got, expectedMap)
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
