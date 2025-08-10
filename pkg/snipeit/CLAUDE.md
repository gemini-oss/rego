# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `snipeit` package provides a comprehensive Go client library for the Snipe-IT API, an open-source IT asset management system. It offers type-safe operations for managing hardware assets, licenses, users, locations, categories, and accessories with built-in caching and concurrent data fetching.

**Key Features:**
- Type-safe generic operations for different HTTP methods
- Concurrent pagination with automatic semaphore control
- Builder pattern for complex operations (license checkout)
- Custom types for Snipe-IT's unique data formats
- Comprehensive asset management (hardware, licenses, accessories)
- Encrypted caching with 5-minute TTL
- Rate limiting (120 requests/minute)
- Support for all major Snipe-IT resource types

## Architecture

### Core Components

1. **Client** (`snipeit.go`):
   - Main client struct with HTTP client, caching, and logging
   - Key methods: `NewClient()`, `do[T any]()`, `doConcurrent[T any]()`
   - Bearer token authentication
   - Rate limiting (120 requests/minute)

2. **Entities** (`entities.go`):
   - Generic types: `PaginatedList[E]`, `SnipeITResponse[E]`
   - Resource types with method-specific generics: `Hardware[M]`, `User[M]`, `License[M]`
   - Custom types: `Timestamp`, `BoolInt`, `Record`
   - Query interface implementations

3. **Domain Clients**:
   - `assets.go`: Hardware asset CRUD and search
   - `users.go`: User management with filtering
   - `licenses.go`: License and seat management with builder pattern
   - `categories.go`: Category CRUD operations
   - `accessories.go`: Accessory management
   - `locations.go`: Location management

### Available API Endpoints

The package defines constants for all Snipe-IT endpoints:
- **Core Resources**: Hardware, Users, Licenses, Categories, Accessories, Locations
- **Extended Resources**: Companies, Consumables, Components, StatusLabels, Models
- **System Resources**: Manufacturers, Suppliers, Departments, Groups, Settings
- **Specialized**: Fields, FieldSets, AssetMaintenance, Reports

### Key Design Patterns

- **Method Chaining**: Domain-specific clients (`client.Assets()`, `client.Users()`)
- **Generic Methods**: Type-safe handling of different HTTP methods (GET/POST/PATCH)
- **Builder Pattern**: Complex object creation (`LicenseBuilder`, checkout operations)
- **Concurrent Pagination**: Parallel fetching with `doConcurrent()`
- **Query Interface**: Flexible query parameter handling

## Development Tasks

### Running Tests
```bash
# Run tests for the snipeit package
go test ./pkg/internal/tests/snipeit/...

# Run with verbose output
go test -v ./pkg/internal/tests/snipeit/...
```

### Usage Examples

#### Asset Management
```go
// Get all assets with pagination handled automatically
assets, err := client.Assets().GetAllAssets()

// Search by serial number
asset, err := client.Assets().GetAssetBySerial("SN12345")

// Search by asset tag
asset, err := client.Assets().GetAssetByTag("ASSET001")

// Get specific asset
asset, err := client.Assets().GetAsset(assetID)

// Delete asset
err := client.Assets().DeleteAsset(assetID)

// Note: Query chaining shown in docs is not implemented
// Use specific query structs instead:
query := &AssetQuery{
    Status: "Ready to Deploy",
    ModelID: 1,
    Limit: 500,
}
```

#### License Management
```go
// Get all licenses
licenses, err := client.Licenses().GetAllLicenses()

// Fluent checkout pattern
err := client.Licenses().
    Checkout(licenseID, seatID).
    ToUser(userID)

// License builder pattern
builder := NewLicenseBuilder().
    WithName("Office 365").
    WithSeats(100).
    WithCompanyID(1)
license := builder.Build()
```

#### Location Management
```go
// CRUD operations
locations, err := client.Locations().GetAllLocations()
location, err := client.Locations().GetLocation(locationID)
created, err := client.Locations().CreateLocation(&Location{Name: "Main Office"})
updated, err := client.Locations().UpdateLocation(locationID, &Location{Name: "HQ"})
err := client.Locations().DeleteLocation(locationID)
```

#### User Management
```go
// Get all users
users, err := client.Users().GetAllUsers()

// Get specific user
user, err := client.Users().GetUser(userID)
```

#### Category Management
```go
// Get all categories
categories, err := client.Categories().GetAllCategories()

// Create category with type
category := &Category{
    Name: "Laptops",
    CategoryType: CATEGORY_TYPE_ASSET, // Use constants
}
created, err := client.Categories().CreateCategory(category)
```

### Common Operations

1. **Generic Type Usage**:
   - Use `[generics.M]` for mutations (POST/PATCH)
   - Use `[generics.S]` for selections (GET)
   - Handles API field type inconsistencies

2. **Query Structures**:
   Each resource has its own query struct:
   ```go
   type AssetQuery struct {
       Limit      int
       Offset     int
       Search     string
       OrderSort  string
       ModelID    int
       CategoryID int
       Status     string
   }
   ```

3. **Pagination Handling**:
   - Automatic with `GetAll*` methods
   - Uses `doConcurrent` for parallel fetching
   - Implements `QueryInterface` for consistency

### Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `SNIPEIT_URL` | Snipe-IT instance URL | `https://snipeit.company.com` | Yes |
| `SNIPEIT_TOKEN` | API authentication token | `eyJ0eXAiOi...` | Yes |
| `REGO_ENCRYPTION_KEY` | Cache encryption key | 32-byte key | Yes |

### Category Type Constants

```go
const (
    CATEGORY_TYPE_ASSET      = "Asset"
    CATEGORY_TYPE_ACCESSORY  = "Accessory"
    CATEGORY_TYPE_CONSUMABLE = "Consumable"
    CATEGORY_TYPE_COMPONENT  = "Component"
    CATEGORY_TYPE_LICENSE    = "License"
)
```

## Important Notes

- **Authentication**: Bearer token in Authorization header
- **Rate Limiting**: 120 requests per minute (enforced by Snipe-IT)
- **Concurrent Operations**: Uses semaphore pattern for pagination
- **Cache**:
  - File: `rego_cache_snipeit.gob`
  - TTL: 5 minutes for GET requests
  - Key pattern: `{url}_{method}_{params}`
- **URL Processing**: Strips protocol and trailing slashes from base URL
- **Field Type Inconsistencies**: API returns different types for GET vs POST/PATCH
- **Custom Types**:
  - `BoolInt`: Handles 0/1 as boolean
  - `Timestamp`: Custom time parsing
  - `DateInfo`: Handles date objects with formatted strings
  - `Record`: Wrapper for generic responses
  - `Messages`: Can be string or map

### Generic Type System

The package uses generics to handle field type differences:
```go
// GET operations - fields may be objects
type Hardware[M generics.M | generics.S] struct {
    WarrantyExpires DateInfo `json:"warranty_expires"`
    CustomFields    map[string]interface{} `json:"custom_fields"`
}

// POST/PATCH operations - fields are often strings/ints
type HardwareUpdate struct {
    WarrantyExpires string `json:"warranty_expires"`
}
```

## Common Pitfalls

1. **Generic Types**: Must specify method type ([M]) when creating/updating resources
2. **Field Types**: Same field may be string in GET but int/object in POST
3. **Timestamps**: Use custom `Timestamp` type for date/time fields
4. **Boolean Values**: API uses 0/1 integers - use `BoolInt` type
5. **Pagination**: Total count may be inaccurate - `doConcurrent` handles this
6. **Query Parameters**: Each resource has its own query struct (not generic)
7. **Date Fields**: Can be strings or objects - use `DateInfo` type
8. **Custom Fields**: Complex nested objects, not simple maps
9. **Method Not Found**: Query chaining shown in docs not actually implemented
10. **Limited CRUD**: Not all resources have full CRUD operations

## API Limitations

1. **Inconsistent Field Types**: GET vs POST/PATCH return different types
2. **Incomplete Pagination Info**: Total count sometimes wrong
3. **Limited Query Options**: Not all fields are searchable
4. **No Batch Operations**: Must process items individually
5. **No Webhooks**: Must poll for changes

## Testing Best Practices

### Lessons Learned

1. **Use Real API Responses**: Mock JSON data often misses field type inconsistencies that exist in real API responses. Always test with actual data from the SnipeIT API.

2. **Common Field Type Issues**:
   - `warranty_expires`: Returns as object `{"date": "2028-06-10", "formatted": "06/10/2028"}` not a string
   - `custom_fields`: Returns complex nested objects, not simple `map[string]string`
   - Date fields: Can be either strings or objects with `date` and `formatted` fields
   - Boolean fields: API returns 0/1 integers, not true/false

3. **Error Reporting**: The generic unmarshalling functions now include field names in error messages to make debugging easier. Errors show:
   - Field name that failed
   - Expected type vs actual type
   - Offset in JSON where error occurred

### Recommended Testing Approach

1. **Integration Tests with Live API**:
   ```go
   // Test with real API endpoint (requires SNIPEIT_URL and SNIPEIT_TOKEN)
   func TestLiveAPIResponse(t *testing.T) {
       client := setupRealClient(t)
       assets, err := client.Assets().GetAllAssets()
       // Validate response structure matches our structs
   }
   ```

2. **Response Fixture Tests**:
   ```go
   // Save real API responses as test fixtures
   // This catches regressions when struct definitions change
   func TestRealResponseFixtures(t *testing.T) {
       data, _ := os.ReadFile("testdata/hardware_response.json")
       var list HardwareList
       err := json.Unmarshal(data, &list)
       // Should unmarshal without errors
   }
   ```

3. **Field Validation Tests**:
   ```go
   // Test each problematic field individually
   func TestProblematicFields(t *testing.T) {
       // Test warranty_expires as object
       // Test custom_fields structure
       // Test date fields in both formats
   }
   ```

### Infrastructure for Prevention

1. **Save API Responses**: Store real API responses in `testdata/` directory
2. **Regular Validation**: Run tests against these fixtures in CI/CD
3. **Field Type Documentation**: Document expected vs actual types for each field
4. **Enhanced Error Messages**: Already implemented in `UnmarshalGeneric` and custom unmarshallers
