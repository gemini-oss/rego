// pkg/common/generics/generics.go
package generics

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

// UnmarshalGeneric unmarshals JSON into a generic struct T that contains an inline generic field of type M.
// The generic field is automatically detected by scanning T for a field whose type is M (or a pointer to M).
// All keys that are not consumed by other fields are assumed to belong to this generic field.
func UnmarshalGeneric[T any, M any](data []byte) (*T, error) {
	// Unmarshal the JSON into a raw map.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// Create a new instance of T.
	result := new(T)
	val := reflect.ValueOf(result).Elem()
	typ := val.Type()

	// Determine the concrete type for M.
	genericType := reflect.TypeOf((*M)(nil)).Elem()

	// Automatically detect the generic field in T.
	genericFieldIndex := -1
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type

		// If the field's type is exactly genericType or is a pointer whose element is genericType, we found the generic field.
		if fieldType == genericType || (fieldType.Kind() == reflect.Ptr && fieldType.Elem() == genericType) {
			if genericFieldIndex != -1 {
				return nil, errors.New("multiple generic fields found in target type")
			}
			genericFieldIndex = i

			// Skip this field
			continue
		}
	}

	// Process each field of T (except the detected generic field).
	for i := 0; i < typ.NumField(); i++ {
		if i == genericFieldIndex {
			continue
		}

		fieldStruct := typ.Field(i)
		fieldVal := val.Field(i)

		// Check the JSON tag.
		tag := fieldStruct.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// If the tag contains "inline", unmarshal as an inline (embedded) field.
		if strings.Contains(tag, "inline") {
			if err := unmarshalInlineField(fieldVal, raw); err != nil {
				return nil, err
			}
			continue
		}

		// Determine the key to look up.
		parts := strings.Split(tag, ",")
		key := parts[0]
		if key == "" {
			key = fieldStruct.Name
		}

		// If a matching key exists in the raw JSON, unmarshal it directly.
		if rawVal, ok := raw[key]; ok {
			if err := json.Unmarshal(rawVal, fieldVal.Addr().Interface()); err != nil {
				return nil, err
			}
			// Remove the key from the raw map
			delete(raw, key)
		}
	}

	// Any leftover keys in raw are assumed to belong to the generic field.
	if genericFieldIndex != -1 {
		if len(raw) > 0 {
			// Marshal the remaining keys back to JSON.
			genericJSON, err := json.Marshal(raw)
			if err != nil {
				return nil, err
			}

			// Unmarshal into a variable of type M.
			var genericValue M
			if err := json.Unmarshal(genericJSON, &genericValue); err != nil {
				return nil, err
			}

			// Set the generic field.
			genericField := val.Field(genericFieldIndex)
			v := reflect.ValueOf(genericValue)

			// Try a direct assignment, or if the target is a pointer, assign its address.
			if v.Type().AssignableTo(genericField.Type()) {
				genericField.Set(v)
			} else if v.Addr().Type().AssignableTo(genericField.Type()) {
				genericField.Set(v.Addr())
			} else {
				return nil, errors.New("cannot assign generic value to generic field")
			}
		}
	} else {
		// If no generic field was found but there are leftover keys, we consider this an error.
		if len(raw) > 0 {
			return nil, errors.New("unprocessed keys remain but no generic field found in target type")
		}
	}

	return result, nil
}

// unmarshalInlineField processes an inline field (an embedded struct or pointer to struct).
// It iterates over the fields of the embedded struct and, if any of their JSON keys are present
// in the raw map, it collects them and unmarshals them into the inline field.
// It then deletes the processed keys from raw.
func unmarshalInlineField(field reflect.Value, raw map[string]json.RawMessage) error {
	// If field is a pointer and nil, allocate a new instance.
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		// Work with the underlying value.
		field = field.Elem()
	}

	// The field should now be a struct.
	if field.Kind() != reflect.Struct {
		return errors.New("inline field is not a struct")
	}

	// Collect keys that belong to the inline struct.
	inlineMap := make(map[string]json.RawMessage)
	collectInlineKeys(field.Type(), raw, inlineMap)

	// If we found any keys for the inline field, unmarshal them into the field.
	if len(inlineMap) > 0 {
		data, err := json.Marshal(inlineMap)
		if err != nil {
			return err
		}
		// Unmarshal into the inline struct.
		if err := json.Unmarshal(data, field.Addr().Interface()); err != nil {
			return err
		}
	}

	return nil
}

// collectInlineKeys recursively iterates over the struct type t.
// For any field marked with an inline tag, it recurses into that field's type;
// for other fields it collects the key from the raw JSON if present.
// Keys found are added to inlineMap, and removed from raw.
func collectInlineKeys(t reflect.Type, raw map[string]json.RawMessage, inlineMap map[string]json.RawMessage) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// Check if the field is itself inline.
		if strings.Contains(tag, "inline") {
			var fieldType reflect.Type
			if field.Type.Kind() == reflect.Ptr {
				fieldType = field.Type.Elem()
			} else {
				fieldType = field.Type
			}
			// Recursively collect keys from the embedded inline struct.
			collectInlineKeys(fieldType, raw, inlineMap)
			continue
		}

		// Otherwise, extract the JSON key.
		parts := strings.Split(tag, ",")
		key := parts[0]
		if key == "" {
			key = field.Name
		}

		// If the key is present in the raw JSON, add it to inlineMap and remove it from raw.
		if rawVal, ok := raw[key]; ok {
			inlineMap[key] = rawVal
			delete(raw, key)
		}
	}
}
