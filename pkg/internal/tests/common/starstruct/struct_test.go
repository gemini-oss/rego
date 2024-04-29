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

var (
	defaultTestStruct = TestStruct{
		Name: "Anthony Dardano",
		Age:  0,
		Tags: []string{"Staff Enterprise Infrastructure Engineer", "DJ"},
		Address: struct {
			City  string `json:"city"`
			State string `json:"state"`
		}{City: "N/A", State: "FL"},
	}
)

// TestPrettyJSON tests the PrettyJSON function for various struct inputs.
func TestPrettyJSON(t *testing.T) {
	testStruct := defaultTestStruct

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
	testStruct := defaultTestStruct

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
	testStruct := defaultTestStruct

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

// TestFlattenStructFieldsSelective tests the FlattenStructFields function for various struct inputs, and ensures the output is only what's desired
func TestFlattenStructFieldsSelective(t *testing.T) {
	testStruct := defaultTestStruct

	fields := []string{"name", "tags", "address.state"}
	expectedSlice := [][]string{
		{"name", "Anthony Dardano"},
		{"tags.0", "Staff Enterprise Infrastructure Engineer"},
		{"tags.1", "DJ"},
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

// TestGenerateFieldNames tests the FlattenStructFields function for dynamic field generation using TestStruct.
func TestGenerateFieldNames(t *testing.T) {
	testStruct := defaultTestStruct

	// Testing with direct struct
	fields, err := starstruct.GenerateFieldNames("", reflect.ValueOf(testStruct))
	if err != nil {
		t.Errorf("GenerateFieldNames() error = %v, wantErr false", err)
		return
	}

	expectedFields := []string{"name", "age", "tags", "address.city", "address.state"}
	if !reflect.DeepEqual(*fields, expectedFields) {
		t.Errorf("GenerateFieldNames() fields = %v, want %v", *fields, expectedFields)
	}

	// Testing with a pointer to a struct (single level of indirection)
	ptrToStruct := &testStruct
	fields, err = starstruct.GenerateFieldNames("", reflect.ValueOf(ptrToStruct))
	if err != nil {
		t.Errorf("GenerateFieldNames() error with pointer to struct = %v, wantErr false", err)
		return
	}
	if !reflect.DeepEqual(*fields, expectedFields) {
		t.Errorf("GenerateFieldNames() got = %v, want %v for pointer to struct", *fields, expectedFields)
	}

	// Testing with a pointer to a pointer to a struct (two levels of indirection)
	ptrToPtrToStruct := &ptrToStruct
	fields, err = starstruct.GenerateFieldNames("", reflect.ValueOf(ptrToPtrToStruct))
	if err != nil {
		t.Errorf("GenerateFieldNames() error with pointer to pointer to struct = %v, wantErr false", err)
		return
	}
	if !reflect.DeepEqual(*fields, expectedFields) {
		t.Errorf("GenerateFieldNames() got = %v, want %v for pointer to pointer to struct", *fields, expectedFields)
	}

	// Test with a slice of pointers to structs
	sliceOfPtrsToStructs := []*TestStruct{ptrToStruct, ptrToStruct}
	fields, err = starstruct.GenerateFieldNames("", reflect.ValueOf(sliceOfPtrsToStructs))
	if err != nil {
		t.Errorf("GenerateFieldNames() error with slice of pointers to structs = %v, wantErr false", err)
		return
	}
	if !reflect.DeepEqual(*fields, expectedFields) {
		t.Errorf("GenerateFieldNames() got = %v, want %v for slice of pointers to structs", *fields, expectedFields)
	}

	// Test with non-struct and non-slice (should return error)
	nonStructInput := 123
	_, err = starstruct.GenerateFieldNames("", reflect.ValueOf(nonStructInput))
	if err == nil {
		t.Errorf("GenerateFieldNames() did not return error for non-struct and non-slice input")
	}
}

// TestKeyOrderingInMap checks the numerical ordering of keys representing indices in a slice.
func TestKeyOrderingInMap(t *testing.T) {
	testStruct := struct {
		Items []string `json:"items"`
	}{
		Items: []string{"first", "second", "third", "fourth"},
	}

	expectedSlice := [][]string{
		{"items.0", "first"},
		{"items.1", "second"},
		{"items.2", "third"},
		{"items.3", "fourth"},
	}

	fields := []string{"items"}
	gotSlice, err := starstruct.FlattenStructFields(testStruct, &fields)
	if err != nil {
		t.Errorf("FlattenStructFields() error = %v", err)
		return
	}

	if !reflect.DeepEqual(gotSlice, expectedSlice) {
		t.Errorf("FlattenStructFields() got = %v, want %v", gotSlice, expectedSlice)
	}
}

// TestComplexNestedStructureOrdering ensures that nested structures are ordered properly.
func TestComplexNestedStructureOrdering(t *testing.T) {
	testStruct := struct {
		Config struct {
			Settings []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"settings"`
		} `json:"config"`
	}{
		Config: struct {
			Settings []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"settings"`
		}{
			Settings: []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}{
				{Key: "timeout", Value: "30s"},
				{Key: "retry", Value: "5"},
			},
		},
	}

	expectedSlice := [][]string{
		{"config.settings.0.key", "timeout"},
		{"config.settings.0.value", "30s"},
		{"config.settings.1.key", "retry"},
		{"config.settings.1.value", "5"},
	}

	fields := []string{"config.settings"}
	gotSlice, err := starstruct.FlattenStructFields(testStruct, &fields)
	if err != nil {
		t.Errorf("FlattenStructFields() error = %v", err)
		return
	}

	if !reflect.DeepEqual(gotSlice, expectedSlice) {
		t.Errorf("FlattenStructFields() got = %v, want %v", gotSlice, expectedSlice)
	}
}

// TestFromTableToStructs tests converting a [][]string with headers as the first row into a slice of dynamic structs.
func TestFromTableToStructs(t *testing.T) {
	// Define a sample input where the first row is headers and subsequent rows are data
	table := [][]string{
		{"Name", "Age", "City"},
		{"Anthony", "30", "Miami"},
		{"Dardano", "25", "New York"},
	}

	// Call the FromTableToStructs function
	results, err := starstruct.TableToStructs(table)
	if err != nil {
		t.Errorf("TableToStructs() error = %v, wantErr nil", err)
		return
	}

	// Define what we expect the resulting structs to look like in JSON format for simplicity
	expected := []map[string]interface{}{
		{"Name": "Anthony", "Age": "30", "City": "Miami"},
		{"Name": "Dardano", "Age": "25", "City": "New York"},
	}

	// Check if the results match the expected output
	if len(results) != len(expected) {
		t.Errorf("TableToStructs() got %d results, want %d", len(results), len(expected))
		return
	}

	for i, result := range results {
		// Convert each struct to a map for easy comparison
		resultMap, err := starstruct.ToMap(result, true)
		if err != nil {
			t.Errorf("ToMap() error = %v", err)
			continue
		}

		if !reflect.DeepEqual(resultMap, expected[i]) {
			t.Errorf("TableToStructs() result %d = %v, want %v", i, resultMap, expected[i])
		}
	}
}