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
	ExcludeNil bool // If true, skip generating fields for nil pointer-structs
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

// WithExcludeNilStructs instructs the package to skip expanding fields in nil pointer-structs.
func WithExcludeNil() Option {
	return func(cfg *pkgConfig) {
		cfg.ExcludeNil = true
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
func FlattenStructFields(item interface{}, opts ...Option) ([][]string, error) {

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
	if cfg.Generate && (cfg.Headers == nil || len(*cfg.Headers) == 0) {
		cfg.Headers = &[]string{}
		generatedFields, err := GenerateFieldNames("", val)
		if err != nil {
			return nil, err
		}
		*cfg.Headers = append(*cfg.Headers, *generatedFields...)
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
		newMap := make(map[string]string, len(*cfg.Headers))
		for _, field := range *cfg.Headers {
			for key, value := range fieldMap {
				if key == field || strings.HasPrefix(key, field+".") {
					newMap[key] = value
				}
			}
		}
		fieldMap = newMap
	}

	// Convert the fieldMap to a slice and update fields with new subfields
	fieldSlice, err := mapToSliceAndUpdateFields(&fieldMap, cfg.Headers, cfg.Sort)
	if err != nil {
		return nil, err
	}

	return fieldSlice, nil
}

// GenerateFieldNames recursively generates field names from a struct, dereferencing pointers as needed, and returns a pointer to a slice of strings.
func GenerateFieldNames(prefix string, val reflect.Value, opts ...Option) (*[]string, error) {
	// Default config
	cfg := &pkgConfig{
		Sort:     false,
		Generate: false,
		Headers:  nil,
		ExcludeNil: false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	var err error
	if val, err = DerefPointers(val); err != nil {
		return nil, err
	}

	// Prepare the field names slice
	fields := make([]string, 0)

	switch val.Kind() {
	case reflect.Map:
		mapFields, err := generateMapFieldNames(prefix, val)
		if err != nil {
			return nil, err
		}
		fields = append(fields, *mapFields...)
		return &fields, nil

	case reflect.Slice, reflect.Array:
		if val.Len() == 0 {
			return nil, fmt.Errorf("GenerateFieldNames: empty slice or array")
		}

		var mergedFields []string
		// Use the first non-nil merge candidate as the baseline ordering.
		for i := 0; i < val.Len(); i++ {
			mergeCandidate := val.Index(i)
			if (mergeCandidate.Kind() == reflect.Ptr || mergeCandidate.Kind() == reflect.Interface) && mergeCandidate.IsNil() {
				continue
			}
			mergeCandidate, err = DerefPointers(mergeCandidate)
			if err != nil {
				return nil, err
			}
			subFieldsPtr, err := GenerateFieldNames(prefix, mergeCandidate, opts...)
			if err != nil {
				return nil, err
			}
			mergedFields = *subFieldsPtr
			break
		}
		// Now, for every candidate, merge in its field names into the baseline.
		for i := 0; i < val.Len(); i++ {
			mergeCandidate := val.Index(i)
			if (mergeCandidate.Kind() == reflect.Ptr || mergeCandidate.Kind() == reflect.Interface) && mergeCandidate.IsNil() {
				continue
			}
			mergeCandidate, err = DerefPointers(mergeCandidate)
			if err != nil {
				return nil, err
			}
			subFieldsPtr, err := GenerateFieldNames(prefix, mergeCandidate, opts...)
			if err != nil {
				return nil, err
			}
			mergedFields = mergeFields(mergedFields, *subFieldsPtr)
		}
		// If still empty, fall back to a zero value.
		if len(mergedFields) == 0 {
			mergeCandidate := reflect.Zero(val.Type().Elem())
			subFieldsPtr, err := GenerateFieldNames(prefix, mergeCandidate, opts...)
			if err != nil {
				return nil, err
			}
			mergedFields = *subFieldsPtr
		}
		return &mergedFields, nil
	case reflect.Struct:
		typ := val.Type()
		// Handle struct fields
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldVal := val.Field(i)
			jsonTag := getFirstTag(field.Tag.Get("json"))

			// If the type of the struct itself is time.Time and it's not an embedded field, add it to the fields
			switch {
			case field.Type.String() == "time.Time" && !field.Anonymous:
				fields = append(fields, jsonTag)
				continue
			}

			// Skip ignored field
			if jsonTag == "-" {
				continue
			}

			// Exclude nil pointer expansions, if set
			if cfg.ExcludeNil {
				if (fieldVal.Kind() == reflect.Ptr || fieldVal.Kind() == reflect.Interface) && fieldVal.IsNil() {
					continue
				}
			}

			fieldKey := joinPrefixKey(prefix, jsonTag)

			// Recursively handle nested structs and inline structs if specified
			if shouldInline(field) {
				subFields, err := GenerateFieldNames(prefix, val.Field(i), opts...)
				if err != nil {
					return nil, err
				}
				fields = append(fields, *subFields...)
			} else if field.Type.Kind() == reflect.Struct ||
				(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
				subPrefix := fieldKey
				subFields, err := GenerateFieldNames(subPrefix, val.Field(i), opts...)
				if err != nil {
					return nil, err
				}
				fields = append(fields, *subFields...)
			} else if field.Type.Kind() == reflect.Map ||
				(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Map) {
				subFields, err := generateMapFieldNames(fieldKey, val.Field(i))
				if err != nil {
					return nil, err
				}
				fields = append(fields, *subFields...)
			} else {
				fields = append(fields, fieldKey)
			}
		}

		return &fields, nil
	case reflect.Interface:
		if val.IsNil() {
			return &fields, nil
		}
		return GenerateFieldNames(prefix, val.Elem(), opts...)
	case reflect.Ptr:
		if val.IsNil() {
			return &fields, nil
		}
		return GenerateFieldNames(prefix, val.Elem(), opts...)
	case reflect.Invalid:
		// Even if the reflect is invalid, other items in a list may be valid
		// so we need to make sure we still return all fields of the struct
		return &[]string{prefix}, nil
	default:
		// Return an error if the input is not a struct or pointer to a struct.
		err = fmt.Errorf("GenerateFieldNames: unsupported input type: %v", val.Kind())
		return nil, err
	}
}

func generateMapFieldNames(prefix string, val reflect.Value) (*[]string, error) {
	var err error
	if val, err = DerefPointers(val); err != nil {
		return nil, err
	}

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

func mergeFields(baseline, candidate []string) []string {
	// Make a copy of the baseline.
	merged := make([]string, len(baseline))
	copy(merged, baseline)

	// Build a map for fast lookup of candidate fields.
	candMap := make(map[string]struct{}, len(candidate))
	for _, f := range candidate {
		candMap[f] = struct{}{}
	}

	// eplace any bare field if candidate has subfields.
	for i, baseField := range baseline {
		if strings.Contains(baseField, ".") {
			continue // skip non-bare fields
		}
		parent := baseField
		var subs []string
		for _, f := range candidate {
			if strings.HasPrefix(f, parent+".") {
				subs = append(subs, f)
			}
		}
		if len(subs) > 0 {
			// Replace the bare field at index i with candidate subfields.
			before := merged[:i]
			after := merged[i+1:]
			merged = append(append(before, subs...), after...)
		}
	}

	// Insert any candidate fields not already present.
	mergedMap := make(map[string]struct{}, len(merged))
	for _, field := range merged {
		mergedMap[field] = struct{}{}
	}
	for _, field := range candidate {
		if _, exists := mergedMap[field]; exists {
			continue
		}
		// If field is a subfield, try to insert it right after the last field with the same parent.
		if dot := strings.Index(field, "."); dot != -1 {
			parent := field[:dot]
			lastIndex := -1
			for j, m := range merged {
				if strings.HasPrefix(m, parent+".") {
					lastIndex = j
				}
			}
			if lastIndex != -1 {
				// Insert field after lastIndex.
				merged = append(merged[:lastIndex+1], append([]string{field}, merged[lastIndex+1:]...)...)
			} else {
				merged = append(merged, field)
			}
		} else {
			// For a bare field, simply append it.
			merged = append(merged, field)
		}
		// Update our lookup map.
		mergedMap[field] = struct{}{}
	}

	return merged
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
		// Handle map and slice types separately and temporarily here (will require refactor of package for more elefant solution)
		if val.Kind() == reflect.Map {
			return flattenMap(val, prefix, fieldMap)
		}
		if val.Kind() == reflect.Slice {
			return flattenSlice(val, prefix, fmt.Sprintf("%%0%dd", len(strconv.Itoa(10))), fieldMap)
		}
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

		keyPrefix := joinPrefixKey(prefix, getMapKey(field))

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
			// If the type of the struct itself is time.Time and it's not an embedded field, add it to the map
			switch {
			case field.Type.String() == "time.Time" && !field.Anonymous:
				(*fieldMap)[keyPrefix] = fmt.Sprint(fieldVal.Interface())
				continue
			}

			// Check if the struct should be inlined
			if shouldInline(field) {
				err := FlattenNestedStructs(fieldVal.Interface(), prefix, fieldMap)
				if err != nil {
					return err
				}
			} else {
				// Recursively handle nested structs
				err := FlattenNestedStructs(fieldVal.Interface(), keyPrefix, fieldMap)
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
				// Check if the map should be inlined
				if shouldInline(field) {
					err := flattenMap(fieldVal, prefix, fieldMap)
					if err != nil {
						return err
					}
				} else {
					err := flattenMap(fieldVal, keyPrefix, fieldMap)
					if err != nil {
						return err
					}
				}
			}
		case reflect.Ptr:
			if fieldVal.IsNil() {
				(*fieldMap)[keyPrefix] = "<nil>"
			} else {
                // Check the underlying type.
                underlying := fieldVal.Elem()
                switch underlying.Kind() {
                case reflect.Struct:
                    if shouldInline(field) {
                        err = FlattenNestedStructs(underlying.Interface(), prefix, fieldMap)
                    } else {
                        err = FlattenNestedStructs(underlying.Interface(), keyPrefix, fieldMap)
                    }
                case reflect.Map, reflect.Slice, reflect.Array:
                    err = FlattenNestedStructs(underlying.Interface(), keyPrefix, fieldMap)
                default:
                    (*fieldMap)[keyPrefix] = fmt.Sprint(underlying.Interface())
                }
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

// joinPrefixKey helps avoid trailing dots or double dots when generating field names.
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


/*
 * shouldInline reports whether the field should be embedded, making it appear as if it belongs to the parent struct.
 * It returns true if the field has the "inline" tag.

 * Example:
 * Field: profile.customAttributes `json:",inline"`
 * profile.customAttributes.key1 ==> profile.key1
 *
 * as opposed to:
 *
 * Field: profile.customAttributes `json:"customAttributes,omitempty"`
 * profile.customAttributes.key1 ==> profile.customAttributes.key1
 */
func shouldInline(field reflect.StructField) bool {
	tag := field.Tag.Get("json")
	return strings.Contains(tag, ",inline")
}

func flattenSlice(slice reflect.Value, keyPrefix, indexFormat string, fieldMap *map[string]string) error {
	for j := 0; j < slice.Len(); j++ {
		elem := slice.Index(j)
		elemKey := joinPrefixKey(keyPrefix, fmt.Sprintf(indexFormat, j))
		if elem.Kind() == reflect.Struct {
			// Recursively handle struct elements in a slice
			err := FlattenNestedStructs(elem.Interface(), elemKey, fieldMap)
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

func flattenMap(m reflect.Value, prefix string, fieldMap *map[string]string) error {
	for _, key := range m.MapKeys() {
		keyStr := fmt.Sprint(key.Interface())
		value := m.MapIndex(key)
		newKey := joinPrefixKey(prefix, keyStr)

		value, err := DerefPointers(value)
        if err != nil {
            return err
        }

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
