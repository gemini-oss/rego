// pkg/common/starstruct/struct.go
package starstruct

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
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

func ToMap(item interface{}, includeZeroValues bool) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(item)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	if v.Kind() == reflect.Map {
		return mapFromMap(v), nil
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", v.Kind())
	}

	typeOfItem := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if !includeZeroValues && field.IsZero() {
			continue
		}

		key := getMapKey(typeOfItem.Field(i))
		if key == "" {
			key = camelKey(typeOfItem.Field(i).Name)
		}

		var value interface{}
		switch field.Kind() {
		case reflect.Struct:
			nestedMap, err := ToMap(field.Interface(), includeZeroValues)
			if err != nil {
				return nil, err
			}
			value = nestedMap
		case reflect.Slice, reflect.Array:
			sliceValues, err := sliceToInterface(field, includeZeroValues)
			if err != nil {
				return nil, err
			}
			value = sliceValues
		default:
			value = field.Interface()
		}

		out[key] = value
	}

	return out, nil
}

func sliceToInterface(v reflect.Value, includeZeroValues bool) ([]interface{}, error) {
	var result []interface{}
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Struct {
			nestedMap, err := ToMap(elem.Interface(), includeZeroValues)
			if err != nil {
				return nil, err
			}
			result = append(result, nestedMap)
		} else {
			result = append(result, elem.Interface())
		}
	}
	return result, nil
}

func camelKey(s string) string {
	if len(s) == 0 {
		return s
	}

	firstChar := s[0]
	if firstChar >= 'A' && firstChar <= 'Z' {
		// ASCII, convert in place
		return string(firstChar+32) + s[1:]
	}

	// Non-ASCII, use rune conversion
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func mapFromMap(v reflect.Value) map[string]interface{} {
	out := make(map[string]interface{})
	for _, key := range v.MapKeys() {
		out[fmt.Sprintf("%v", key.Interface())] = v.MapIndex(key).Interface()
	}
	return out
}

// FlattenStructFields parses a struct and its nested fields, if any, to a flat slice.
// It also updates the input fields with any new subfields found.
func FlattenStructFields(item interface{}, fields *[]string) ([][]string, error) {
	// Check if the passed item is a struct or a pointer to a struct.
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct or pointer to a struct, got %v", val.Kind())
	}

	// Create a map to hold the fields and their values
	fieldMap := make(map[string]string)

	// Recursively parse the struct
	err := flattenNestedStructs(item, "", fieldMap)
	if err != nil {
		return nil, err
	}

	// Convert the fieldMap to a slice and update fields with new subfields
	fieldSlice, err := mapToSliceAndUpdateFields(fieldMap, fields)
	if err != nil {
		return nil, err
	}

	return fieldSlice, nil
}

// flattenNestedStructs recursively navigates through a struct, parsing its fields and nested fields.
// It populates a map with field names as keys and their values as values.
func flattenNestedStructs(obj interface{}, prefix string, fieldMap map[string]string) error {
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)

	// Check if the passed obj is a pointer and dereference it if it is
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	// Check if the dereferenced obj is a struct
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct or pointer to a struct, got %v", val.Kind())
	}

	// Iterate over all fields of the struct
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Handle different kinds of fields (struct, slice, others)
		switch fieldVal.Kind() {
		case reflect.Struct:
			// Recursively parse struct fields
			jsonTag := getFirstTag(field.Tag.Get("json"))
			err := flattenNestedStructs(fieldVal.Interface(), prefix+jsonTag+".", fieldMap)
			if err != nil {
				return err
			}

		case reflect.Slice:
			// Recursively parse slice elements if they are struct; add directly to map if not
			jsonTag := getFirstTag(field.Tag.Get("json"))
			for j := 0; j < fieldVal.Len(); j++ {
				elem := fieldVal.Index(j)
				if elem.Kind() == reflect.Struct {
					err := flattenNestedStructs(elem.Interface(), prefix+jsonTag+"."+strconv.Itoa(j)+".", fieldMap)
					if err != nil {
						return err
					}
				} else {
					key := prefix + jsonTag + "." + strconv.Itoa(j)
					fieldMap[key] = fmt.Sprint(elem.Interface())
				}
			}

		default:
			// Parse non-struct and non-slice fields
			fieldMap[prefix+getMapKey(field)] = fmt.Sprint(fieldVal.Interface())
		}
	}

	return nil
}

// getFirstTag extracts the first tag from a tag string.
func getFirstTag(tag string) string {
	return strings.Split(tag, ",")[0]
}

// getMapKey determines the key to be used in the fieldMap.
func getMapKey(field reflect.StructField) string {
	jsonTag := getFirstTag(field.Tag.Get("json"))
	urlTag := getFirstTag(field.Tag.Get("url"))
	xmlTag := getFirstTag(field.Tag.Get("xml"))
	mapKey := field.Name

	if jsonTag != "" && jsonTag != "-" {
		mapKey = jsonTag
	} else if urlTag != "" && urlTag != "-" {
		mapKey = urlTag
	} else if xmlTag != "" && xmlTag != "-" {
		mapKey = xmlTag
	} else {
		mapKey = camelKey(mapKey) // Convert to camel case if no tag is found
	}

	return mapKey
}

// mapToSliceAndUpdateFields converts a map to a two-dimensional slice and updates the fields with new subfields.
func mapToSliceAndUpdateFields(fieldMap map[string]string, fields *[]string) ([][]string, error) {
	// Create a map to quickly check if a field already exists
	existingFields := make(map[string]bool)
	for _, field := range *fields {
		existingFields[field] = true
	}

	newFields := make([]string, 0)
	fieldSlice := make([][]string, 0)

	for _, field := range *fields {
		found := false
		for key, value := range fieldMap {
			if strings.HasPrefix(key, field+".") {
				// Ensure it's a sub-field
				fieldSlice = append(fieldSlice, []string{key, value})

				// Add the key to the newFields if it doesn't already exist
				if _, exists := existingFields[key]; !exists {
					newFields = append(newFields, key)
					existingFields[key] = true
				}

				found = true
			}
		}

		// If no sub-fields were found for this field, keep the original field
		if !found {
			value, ok := fieldMap[field]
			if !ok {
				return nil, fmt.Errorf("field %s not found in struct", field)
			}

			newFields = append(newFields, field)
			fieldSlice = append(fieldSlice, []string{field, value})
		}
	}

	*fields = newFields
	return fieldSlice, nil
}
