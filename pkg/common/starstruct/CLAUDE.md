# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `starstruct` package provides sophisticated struct manipulation utilities for flattening nested structures, generating field names dynamically, converting between structs and maps, and intelligently merging fields. It's essential for data transformation, particularly for spreadsheet exports and API parameter building.

## Architecture

### Core Components

1. **Field Flattening** (`struct.go`):
   - `FlattenStructFields()`: Recursively flattens structs to 2D arrays
   - Handles nested structs, slices, maps, and pointers
   - Uses dot notation for paths (e.g., `address.city`)
   - Zero-padded indices for arrays (e.g., `tags.00`)

2. **Field Generation** (`struct.go`):
   - `GenerateFieldNames()`: Dynamically discovers struct fields
   - Respects JSON/URL/XML tags
   - Handles inline fields and nested structures
   - Returns sorted, unique field paths

3. **Field Merging** (`fields.go`):
   - `MergeFields()`: Intelligently merges field lists
   - Preserves baseline ordering
   - Handles numeric suffixes (sorts `field.0` before `field.10`)
   - Inserts new fields near related baseline fields

4. **Conversion Utilities** (`struct.go`):
   - `ToMap()`: Struct to map conversion with zero-value handling
   - `TableToStructs()`: 2D arrays to struct slices
   - `PrettyJSON()`: Pretty-printed JSON output

### Configuration Options

```go
WithSort()        // Sort fields alphabetically
WithGenerate()    // Generate fields dynamically
WithHeaders()     // Use specific field headers
WithExcludeNil()  // Skip nil pointer fields
```

## Development Tasks

### Common Operations

1. **Flattening Structs for Spreadsheets**:
   ```go
   // Generate headers from struct
   headers, _ := starstruct.GenerateFieldNames("", reflect.ValueOf(data))

   // Flatten with headers
   rows, _ := starstruct.FlattenStructFields(data,
       starstruct.WithHeaders(&headers))
   ```

2. **Building API Parameters**:
   ```go
   // Convert struct to map, excluding zero values
   params, _ := starstruct.ToMap(queryStruct, true)
   ```

3. **Dynamic Field Discovery**:
   ```go
   // Flatten with automatic field generation
   data, _ := starstruct.FlattenStructFields(unknown,
       starstruct.WithGenerate())
   ```

4. **Merging Field Lists**:
   ```go
   // Merge new fields into existing list
   finalHeaders := starstruct.MergeFields(baseHeaders, newHeaders)
   ```

### Field Path Format

- Nested fields: `parent.child.grandchild`
- Array elements: `array.00`, `array.01` (zero-padded)
- Map entries: `map.key1`, `map.key2` (sorted keys)
- Inline fields: Flattened at parent level

## Important Notes

- Uses reflection extensively - performance cost for complex structs
- Time.Time is treated as terminal (not flattened further)
- Respects `json:",inline"` tags for embedded structs
- Map keys are sorted for deterministic output
- Nil pointers can be included or excluded via options

## Usage Patterns

### Google Sheets Export
```go
// From google/sheets.go
headers, _ := ss.GenerateFieldNames("", reflect.ValueOf(data))
for _, item := range data {
    row, _ := ss.FlattenStructFields(item, ss.WithHeaders(headers))
    rows = append(rows, row)
}
```

### HTTP Query Parameters
```go
// From requests/requests.go
if query != nil {
    params, _ := ss.ToMap(query, false)
    q := u.Query()
    for k, v := range params {
        q.Set(k, fmt.Sprint(v))
    }
}
```

## Common Pitfalls

1. **Performance**: Reflection is slow for large/complex structs
2. **Circular References**: Will cause infinite recursion
3. **Interface Fields**: May not flatten as expected
4. **Private Fields**: Not accessible via reflection
5. **Tag Priority**: JSON tags take precedence over field names

## Best Practices

1. **Cache Generated Headers**: Don't regenerate for each item
2. **Use WithHeaders**: More efficient than WithGenerate for known structures
3. **Handle Errors**: Reflection operations can fail
4. **Test Edge Cases**: Nil pointers, empty slices, zero values
5. **Consider Performance**: Profile if processing large datasets
