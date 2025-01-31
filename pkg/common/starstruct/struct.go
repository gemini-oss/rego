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

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ### StarStruct Options
// ---------------------------------------------------------------------

type pkgConfig struct {
	Sort     bool
	Generate bool
	Headers  *[]string
}

type Option func(*pkgConfig)

// WithSortFields tells the package to sort the fields of the struct
func WithSort() Option {
	return func(cfg *pkgConfig) {
		cfg.Sort = true
	}
}

// WithGenerateFields tells the package to generate fields dynamically
func WithGenerate() Option {
	return func(cfg *pkgConfig) {
		cfg.Generate = true
	}
}

// WithHeaders tells the package to use the provided headers
func WithHeaders(headers *[]string) Option {
	return func(cfg *pkgConfig) {
		cfg.Headers = headers
	}
}

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

// TableToStructs converts a [][]string into a slice of structs, with the first row as headers.
func TableToStructs(data [][]string) ([]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	headers := data[0]
	var results []interface{}

	// Create a dynamic struct type based on headers
	var fields []reflect.StructField
	c := cases.Title(language.English)
	for _, header := range headers {
		safeHeader := c.String(strings.ReplaceAll(header, " ", ""))
		safeHeader = ensureValidIdentifier(safeHeader) // Make sure it's a valid Go identifier
		fields = append(fields, reflect.StructField{
			Name: safeHeader,
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, header)),
		})
	}
	structType := reflect.StructOf(fields)

	// Populate the struct instances
	for _, row := range data[1:] {
		if len(row) != len(headers) {
			return nil, fmt.Errorf("data row does not match headers length")
		}
		instance := reflect.New(structType).Elem()
		for i, value := range row {
			instance.Field(i).SetString(value)
		}
		results = append(results, instance.Interface())
	}

	return results, nil
}

// ensureValidIdentifier makes sure the string is a valid Go identifier.
func ensureValidIdentifier(name string) string {
	if name == "" || !isLetter(rune(name[0])) {
		name = "Field" + name // Prefix to ensure it's a valid identifier
	}
	return name
}

// isLetter checks if the rune is a letter (unicode compliant)
func isLetter(ch rune) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
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
func FlattenStructFields(item interface{}, fields *[]string, opts ...Option) ([][]string, error) {

	// Default config
	cfg := &pkgConfig{
		Sort:     false,
		Generate: false,
		Headers:  nil,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	val, err := DerefPointers(reflect.ValueOf(item))
	if err != nil {
		return nil, err
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct or pointer to a struct, got %v", val.Kind())
	}

	// If no fields are provided, generate them dynamically
	if cfg.Generate {
		generatedFields, err := GenerateFieldNames("", val)
		if err != nil {
			return nil, err
		}
		*fields = append(*fields, *generatedFields...)
	}

	// Create a map to hold the fields and their values
	fieldMap := make(map[string]string)

	// Recursively parse the struct
	err = FlattenNestedStructs(item, "", &fieldMap)
	if err != nil {
		return nil, err
	}

	switch cfg.Generate {
	case false:
		// If fields were not generated, limit the map to only include the provided fields while keeping data intact
		newMap := make(map[string]string, len(*fields))
		for _, field := range *fields {
			newMap[field] = fieldMap[field]
		}
		fieldMap = newMap
	}

	// Convert the fieldMap to a slice and update fields with new subfields
	fieldSlice, err := mapToSliceAndUpdateFields(&fieldMap, fields, false)
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
		typ := val.Type()
		// Handle struct fields
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			jsonTag := getFirstTag(field.Tag.Get("json"))
			if jsonTag == "-" {
				continue // Ignore fields marked to be ignored
			}
			fieldKey := joinPrefixKey(prefix, jsonTag)

			// Recursively handle nested structs and inline structs if specified
			if shouldInlineStruct(field) {
				subFields, err := GenerateFieldNames(prefix, val.Field(i))
				if err != nil {
					return nil, err
				}
				fields = append(fields, *subFields...)
			} else if field.Type.Kind() == reflect.Struct ||
				(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
				subPrefix := fieldKey
				subFields, err := GenerateFieldNames(subPrefix, val.Field(i))
				if err != nil {
					return nil, err
				}
				fields = append(fields, *subFields...)
			} else if field.Type.Kind() == reflect.Map {
				subFields, err := generateMapFieldNames(fieldKey, val.Field(i))
				if err != nil {
					return nil, err
				}
				fields = append(fields, *subFields...)
			} else {
				fields = append(fields, fieldKey)
			}
		}

		// Handle embedded fields
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			if field.Anonymous {
				// Flatten embedded fields under the same prefix
				subFields, err := GenerateFieldNames(prefix, val.Field(i))
				if err != nil {
					return nil, err
				}
				fields = append(fields, *subFields...)
			}
		}

		return &fields, nil

	case reflect.Interface:
		if val.IsNil() {
			return &fields, nil
		}
		return GenerateFieldNames(prefix, val.Elem())

	default:
		return &[]string{prefix}, nil
	}
}

func generateMapFieldNames(prefix string, val reflect.Value) (*[]string, error) {
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("generateMapFieldNames: expected a map, got %v", val.Kind())
	}

	fields := make([]string, 0)

	for _, key := range val.MapKeys() {
		keyStr := fmt.Sprint(key.Interface())
		fieldKey := joinPrefixKey(prefix, keyStr)

		value := val.MapIndex(key)
		switch value.Kind() {
		case reflect.Map, reflect.Struct:
			subFields, err := GenerateFieldNames(fieldKey, value)
			if err != nil {
				return nil, err
			}
			fields = append(fields, *subFields...)
		default:
			fields = append(fields, fieldKey)
		}
	}
	return &fields, nil
}

// shortTypeName strips away any type parameters in a generic type.
// E.g., "Model[github.com/gemini-oss/rego/pkg/snipeit.GET]" -> "Model"
func shortTypeName(t reflect.Type) string {
	// Reflect’s .Name() for a generic type can look like:
	//    "Model[github.com/gemini-oss/rego/pkg/snipeit.GET]"
	// so we truncate everything after (and including) the first '['.
	name := t.Name()
	if idx := strings.IndexRune(name, '['); idx != -1 {
		name = name[:idx]
	}
	return name
}

// DerefPointers takes a reflect.Value and recursively dereferences it if it's a pointer.
func DerefPointers(val reflect.Value) (reflect.Value, error) {
	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			// Return the relevant nil value for the current type, "" for string, 0 for int, etc.
			switch val.Kind() {
			case reflect.String:
				return reflect.ValueOf(""), nil
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return reflect.ValueOf(0), nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return reflect.ValueOf(uint(0)), nil
			case reflect.Float32, reflect.Float64:
				return reflect.ValueOf(float64(0)), nil
			case reflect.Bool:
				return reflect.ValueOf(false), nil
			default:
				//return reflect.Value{}, fmt.Errorf("DerefPointers: nil pointer/interface element")
			}
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
func FlattenNestedStructs(item interface{}, prefix string, fieldMap *map[string]string) error {
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

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		keyPrefix := prefix + getMapKey(field)

		switch fieldVal.Kind() {
		case reflect.Slice:
			if fieldVal.Len() == 0 {
				(*fieldMap)[keyPrefix] = "" // Handle empty slice
			} else {
				err := flattenSlice(fieldVal, keyPrefix, indexFormat, fieldMap)
				if err != nil {
					return err
				}
			}
		case reflect.Struct:
			// Check if the struct should be inlined
			if shouldInlineStruct(field) {
				err := FlattenNestedStructs(fieldVal.Interface(), prefix, fieldMap)
				if err != nil {
					return err
				}
			} else {
				// Recursively handle nested structs
				err := FlattenNestedStructs(fieldVal.Interface(), keyPrefix+".", fieldMap)
				if err != nil {
					return err
				}
			}
		case reflect.Interface:
			// Handle interface types (like the generic parameter)
			if !fieldVal.IsNil() {
				elem := fieldVal.Elem()
				if elem.Kind() == reflect.Struct { // Only call FlattenNestedStructs if the type is a struct
					err := FlattenNestedStructs(elem.Interface(), prefix, fieldMap)
					if err != nil {
						return err
					}
				}
			}
		case reflect.Map:
			if fieldVal.Len() == 0 {
				(*fieldMap)[keyPrefix] = "" // Handle empty map
			} else {
				err := flattenMap(fieldVal, keyPrefix, fieldMap)
				if err != nil {
					return err
				}
			}
		case reflect.Ptr:
			if fieldVal.IsNil() {
				(*fieldMap)[keyPrefix] = "<nil>"
			} else {
				err := FlattenNestedStructs(fieldVal.Elem().Interface(), keyPrefix+".", fieldMap)
				if err != nil {
					return err
				}
			}
		default:
			// Handle basic types
			if fieldVal.IsValid() {
				(*fieldMap)[keyPrefix] = fmt.Sprint(fieldVal.Interface())
			} else {
				(*fieldMap)[keyPrefix] = "<nil>"
			}
		}
	}

	return nil
}

// joinPrefixKey helps avoid trailing dots or double dots.
func joinPrefixKey(prefix, key string) string {
	switch {
	case prefix == "" && key == "":
		return ""
	case prefix == "":
		return key
	case key == "":
		return prefix
	default:
		return prefix + "." + key
	}
}

func shouldInlineStruct(field reflect.StructField) bool {
	tag := field.Tag.Get("json")
	return strings.Contains(tag, ",inline")
}

func flattenSlice(slice reflect.Value, keyPrefix, indexFormat string, fieldMap *map[string]string) error {
	for j := 0; j < slice.Len(); j++ {
		elem := slice.Index(j)
		elemKey := fmt.Sprintf("%s.%s", keyPrefix, fmt.Sprintf(indexFormat, j))
		if elem.Kind() == reflect.Struct {
			// Recursively handle struct elements in a slice
			err := FlattenNestedStructs(elem.Interface(), elemKey+".", fieldMap)
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

func flattenMap(m reflect.Value, keyPrefix string, fieldMap *map[string]string) error {
	for _, key := range m.MapKeys() {
		keyStr := fmt.Sprint(key.Interface())
		value := m.MapIndex(key)

		newKey := fmt.Sprintf("%s.%s", keyPrefix, keyStr)

		switch value.Kind() {
		case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
			err := FlattenNestedStructs(value.Interface(), newKey, fieldMap)
			if err != nil {
				return err
			}
		default:
			(*fieldMap)[newKey] = fmt.Sprint(value.Interface())
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
