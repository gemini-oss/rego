// pkg/common/struct/struct.go
package starstruct

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

/*
 * Print a struct as a JSON string
 */
func PrettyJSON(data interface{}) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(data)
	if err != nil {
		return "", err
	}
	return buffer.String(), err
}

/*
 * Convert a struct to a map[string]string
 */
func StructToMap(item interface{}) map[string]string {
	out := make(map[string]string)

	v := reflect.ValueOf(item)
	typeOfItem := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if field.IsZero() {
			continue
		}

		fieldStr := fmt.Sprintf("%v", field.Interface())

		// Get the JSON tag value
		jsonTag := typeOfItem.Field(i).Tag.Get("url")
		jsonTag = strings.Split(jsonTag, ",")[0] // exclude ",omitempty" if exists

		if jsonTag != "" && fieldStr != "" {
			// Use the JSON tag as the key
			out[jsonTag] = fieldStr
		} else if fieldStr != "" {
			// If JSON tag is not defined, use the field name as the key
			// We are only lowercasing the first letter of the field name
			fieldName := typeOfItem.Field(i).Name
			out[strings.ToLower(fieldName[:1])+fieldName[1:]] = fieldStr
		}
	}

	return out
}
