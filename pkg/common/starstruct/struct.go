// pkg/common/starstruct/struct.go
package starstruct

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
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

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

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
	val, err := DerefPointers(reflect.ValueOf(item))
	if err != nil {
		return nil, err
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct or pointer to a struct, got %v", val.Kind())
	}

	// If no fields are provided, generate them dynamically
	var sortFields bool
	if fields == nil || len(*fields) == 0 {
		fields, err = GenerateFieldNames("", val)
		if err != nil {
			return nil, err
		}
	} else {
		sortFields = false
	}

	// Create a map to hold the fields and their values
	fieldMap := make(map[string]string)

	// Recursively parse the struct
	err = flattenNestedStructs(item, "", &fieldMap)
	if err != nil {
		return nil, err
	}

	// Convert the fieldMap to a slice and update fields with new subfields
	fieldSlice, err := mapToSliceAndUpdateFields(&fieldMap, fields, sortFields)
	if err != nil {
		return nil, err
	}

	return fieldSlice, nil
}

// GenerateFieldNames recursively generates field names from a struct, dereferencing pointers as needed, and returns a pointer to a slice of strings.
func GenerateFieldNames(prefix string, val reflect.Value) (*[]string, error) {
	var err error
	if val, err = DerefPointers(val); err != nil {
		return nil, err
	}

	// Prepare the field names slice
	fields := make([]string, 0)

	switch val.Kind() {
	case reflect.Map:
		// First element only (assuming homogeneous types)
		if len(val.MapKeys()) == 0 {
			return nil, fmt.Errorf("GenerateFieldNames: empty map")
		}
		firstKey := val.MapKeys()[0]
		val = val.MapIndex(firstKey)
		return GenerateFieldNames(prefix, val)
	case reflect.Slice, reflect.Array:
		// First element only (assuming homogeneous types)
		if val.Len() == 0 {
			return nil, fmt.Errorf("GenerateFieldNames: empty slice or array")
		}
		return GenerateFieldNames(prefix, val.Index(0))
	case reflect.Struct:
		// Handle struct fields
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			jsonTag := getFirstTag(field.Tag.Get("json"))
			if jsonTag == "-" || jsonTag == "" {
				continue // Ignore fields without a JSON tag or marked to be ignored
			}
			fieldKey := prefix + jsonTag

			// Recursively handle nested structs
			if field.Type.Kind() == reflect.Struct {
				subFields, err := GenerateFieldNames(fieldKey+".", val.Field(i))
				if err != nil {
					return nil, err
				}
				fields = append(fields, *subFields...)
			} else {
				fields = append(fields, fieldKey)
			}
		}
		return &fields, nil
	default:
		return nil, fmt.Errorf("GenerateFieldNames: unsupported type %s", val.Kind())
	}
}

// DerefPointers takes a reflect.Value and recursively dereferences it if it's a pointer.
func DerefPointers(val reflect.Value) (reflect.Value, error) {
	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return reflect.Value{}, fmt.Errorf("GenerateFieldNames: nil pointer/interface element")
		}
		if val.Kind() == reflect.Pointer {
			val = val.Elem()
		}
		if val.Kind() == reflect.Interface {
			val = val.Elem()
		}
	}
	return val, nil
}

// flattenNestedStructs recursively navigates through a struct, parsing its fields and nested fields.
// It populates a map with field names as keys and their values as values.
func flattenNestedStructs(item interface{}, prefix string, fieldMap *map[string]string) error {
	val, err := DerefPointers(reflect.ValueOf(item))
	if err != nil {
		return err
	}

	typ := val.Type()

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct or pointer to a struct, got %v", val.Kind())
	}

	// Determine the max index length for zero-padding for proper sorting
	maxIndexLength := 0
	for i := 0; i < val.NumField(); i++ {
		if val.Field(i).Kind() == reflect.Slice {
			length := val.Field(i).Len()
			if length > maxIndexLength {
				maxIndexLength = length
			}
		}
	}
	// Ensure minimum width of 2 digits
	if maxIndexLength < 10 {
		maxIndexLength = 10
	}
	indexFormat := fmt.Sprintf("%%0%dd", len(strconv.Itoa(maxIndexLength)))

	// Iterate over the struct fields
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		keyPrefix := prefix + getMapKey(field)

		switch fieldVal.Kind() {
		case reflect.Slice:
			if fieldVal.Len() == 0 {
				(*fieldMap)[keyPrefix] = "" // Handle empty slice
			} else {
				flattenSlice(fieldVal, keyPrefix, indexFormat, fieldMap)
			}
		case reflect.Struct:
			// Recursively handle nested structs
			err := flattenNestedStructs(fieldVal.Interface(), keyPrefix+".", fieldMap)
			if err != nil {
				return err
			}
		default:
			// Handle basic types
			(*fieldMap)[keyPrefix] = fmt.Sprint(fieldVal.Interface())
		}
	}

	return nil
}

func flattenSlice(slice reflect.Value, keyPrefix, indexFormat string, fieldMap *map[string]string) error {
	for j := 0; j < slice.Len(); j++ {
		elem := slice.Index(j)
		elemKey := fmt.Sprintf("%s.%s", keyPrefix, fmt.Sprintf(indexFormat, j))
		if elem.Kind() == reflect.Struct {
			// Recursively handle struct elements in a slice
			err := flattenNestedStructs(elem.Interface(), elemKey+".", fieldMap)
			if err != nil {
				return err
			}
		} else {
			// Store basic slice elements
			(*fieldMap)[elemKey] = fmt.Sprint(elem.Interface())
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
func mapToSliceAndUpdateFields(fieldMap *map[string]string, fields *[]string, sortFields bool) ([][]string, error) {
	// Prepare slices for results and new fields
	var newFields []string
	fieldSlice := make([][]string, 0)

	// Store all keys that start with the same top-level prefix in the fieldMap
	prefixMap := make(map[string][]string)
	for key := range *fieldMap {
		prefix := strings.Split(key, ".")[0]
		prefixMap[prefix] = append(prefixMap[prefix], key)
	}

	// Use the ordered fields to determine output order
	for _, field := range *fields {
		prefix := strings.Split(field, ".")[0]
		if keys, found := prefixMap[prefix]; found {
			for _, key := range keys {
				if strings.HasPrefix(key, prefix) {
					newFields = append(newFields, key)
					value := (*fieldMap)[key]
					fieldSlice = append(fieldSlice, []string{key, value})
				}
			}
			delete(prefixMap, prefix) // Prevent reprocessing same prefix
		}
	}

	// Sort the newFields array if needed
	if sortFields {
		sort.Strings(newFields)
	}

	// Update the fields pointer
	*fields = newFields
	return fieldSlice, nil
}
