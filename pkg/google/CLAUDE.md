# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `google` package provides a comprehensive Go client library for Google Workspace APIs. It offers unified access to Admin SDK Directory, Drive, Sheets, Chrome Policy, and Reports APIs with built-in authentication, caching, rate limiting, and pagination support.

**Key Features:**
- Multi-authentication support (API Key, OAuth2, Service Account)
- Domain-wide delegation for service accounts
- Method chaining for complex queries (Devices API)
- Automatic pagination handling
- Per-service rate limiting
- Encrypted caching with configurable TTL
- Embedded API discovery and scope management
- Recursive Drive operations
- Batch operations support
- Chrome Policy management with schema discovery

## Architecture

### Core Components

1. **Client** (`google.go`):
   - Main client struct with authentication, HTTP client, logging, and caching
   - Key methods: `NewClient()`, `do[T any]()`, `doPaginated[T, Q]()`
   - JWT token generation for service accounts
   - Support for API Key, OAuth, and Service Account authentication

2. **Entities** (`entities.go`):
   - Comprehensive struct definitions for all Google API resources
   - Generic interfaces: `GoogleAPIResponse[T any]`, `GoogleQuery`
   - Resource types: User, Group, OrgUnit, ChromeOSDevice, File, Spreadsheet, Policy

3. **API Implementations**:
   - `admin.go`: Roles, organizational units, customer management
   - `users.go`: User CRUD operations and search
   - `groups.go`: Group management and membership
   - `devices.go`: Chrome OS device management with query chaining
   - `drive.go`: File/folder operations, recursive traversal
   - `sheets.go`: Spreadsheet creation, data manipulation, formatting
   - `permissions.go`: Drive permissions and ownership management
   - `chrome_policy.go`: Chrome policy management and cloning
   - `api.go`: API discovery and scope management

### Key Design Patterns

- **Method Chaining**: Fluent interface for complex queries (especially devices)
- **Generic Pagination**: Automatic page handling via `doPaginated[T, Q]()`
- **Multi-Authentication**: Flexible auth supporting different Google auth methods
- **Rate Limiting**: Per-API configurable limits with automatic throttling
- **Embedded Resources**: API definitions and scopes in JSON files
- **Generic Interfaces**: `GoogleAPIResponse[T]` for unified pagination
- **Builder Pattern**: Query builders for complex API parameters
- **CICD Mode**: Environment-based credential loading

### Authentication

```go
// Authentication Types
type AuthType string

const (
    API_KEY         AuthType = "API_KEY"         // Simple API key
    OAUTH_CLIENT    AuthType = "OAUTH_CLIENT"    // OAuth2 client (partial)
    SERVICE_ACCOUNT AuthType = "SERVICE_ACCOUNT" // Service account with JWT
)

// Service Account with Domain-Wide Delegation
ac := google.AuthCredentials{
    Type:    google.SERVICE_ACCOUNT,
    Scopes:  []string{"Admin SDK API", "Google Drive API", "Google Sheets API"},
    Subject: "admin@domain.com", // User to impersonate
}

// API Key Authentication (limited APIs)
ac := google.AuthCredentials{
    Type:   google.API_KEY,
    APIKey: "your-api-key",
}

// CICD Mode (credentials from environment)
config := google.Config{
    CICD: true, // Reads from GOOGLE_SERVICE_ACCOUNT env var
}
```

## Development Tasks

### Running Tests
```bash
# Run tests for the google package
go test ./pkg/internal/tests/google/...

# Run with verbose output
go test -v ./pkg/internal/tests/google/...
```

### Usage Examples

#### Client Initialization
```go
// Standard initialization
client, err := google.NewClient(&google.Config{
    CustomerID: "my_customer",
    Domain:     "example.com",
}, ac, log.INFO)

// With custom rate limits
client, err := google.NewClient(&google.Config{
    CustomerID: "my_customer",
    Domain:     "example.com",
    RateLimit:  100, // Override default
}, ac, log.INFO)

// User impersonation without recreating client
client.ImpersonateUser("another-admin@example.com")
```

#### Device Management with Method Chaining
```go
// Complex device query with all parameters
devices, err := client.Devices()
    .MaxResults(500)
    .PageToken(nextPageToken)
    .Query("status:provisioned")
    .OrderBy("status")
    .OrgUnitPath("/Students")
    .IncludeChildOrgunits(true)
    .Projection("FULL")
    .SortOrder("ASCENDING")
    .ListAllChromeOS("my_customer")

// Simple query
allDevices, err := client.Devices().ListAllChromeOS("my_customer")
```

#### User Operations
```go
// List all users
users, err := client.ListAllUsers("my_customer")

// Search users
results, err := client.SearchUsers("name:John*", "my_customer")

// Create user
newUser := &google.User{
    PrimaryEmail: "satoshi.nakamoto@example.com",
    Name: &google.UserName{
        GivenName:  "Satoshi",
        FamilyName: "Nakamoto",
    },
    Password: "SecurePassword123!",
}
created, err := client.InsertUser(newUser, "my_customer")
```

#### Drive Operations
```go
// List files in folder
files, err := client.ListDriveFiles(&google.ListDriveFilesRequest{
    FolderID: "folder-id",
    Query:    "mimeType='application/pdf'",
})

// Recursive folder operations
allFiles, err := client.ListDriveFiles(&google.ListDriveFilesRequest{
    FolderID: "root",
    Depth:    5, // Traverse 5 levels deep
})

// Build file path
path, err := client.BuildFilePath("file-id")
```

#### Sheets Operations
```go
// Create spreadsheet from struct slice
type SalesData struct {
    Date     string  `json:"date"`
    Product  string  `json:"product"`
    Quantity int     `json:"quantity"`
    Revenue  float64 `json:"revenue"`
}

data := []SalesData{...}
spreadsheetID, err := client.CreateSpreadsheet(
    "Sales Report",
    "Q1 Data",
    data,
)

// Update values
valueRange := client.GenerateValueRange(data)
err = client.UpdateSheetValues(spreadsheetID, "A1", valueRange)
```

#### Chrome Policy Management
```go
// Get policy schemas
schemas, err := client.PolicySchemas("my_customer")

// Resolve policies for org unit
policies, err := client.PolicyResolve(
    "my_customer",
    &google.PolicyResolveRequest{
        PolicySchemaFilter: "chrome.users.*",
        PolicyTargetKey: &google.PolicyTargetKey{
            TargetResource: "orgunits/03ph8a2z1xdnme9",
        },
    },
)

// Clone policies between org units
err = client.ClonePolicies(
    "my_customer",
    "source-orgunit-id",
    "target-orgunit-id",
    "chrome.users.*",
)
```

#### Cache Management
```go
// Manual cache operations
client.SetCache("custom-key", data, 1*time.Hour)

var cachedData MyType
if client.GetCache("custom-key", &cachedData) {
    // Use cached data
}

// Clear all cache
client.ClearCache()
```

### Common Operations

1. **Working with Pagination**:
   - Use `doPaginated()` for automatic page handling
   - Implement `GoogleAPIResponse` interface for custom types
   - Set `MaxResults` to control page size

2. **Error Handling**:
   ```go
   users, err := client.ListAllUsers("my_customer")
   if err != nil {
       var googleErr *google.ErrorResponse
       if errors.As(err, &googleErr) {
           // Handle Google-specific error
           log.Printf("Error code: %d, Message: %s",
               googleErr.Error.Code,
               googleErr.Error.Message)
       }
       return err
   }
   ```

3. **Adding New API Endpoints**:
   - Define request/response types in `entities.go`
   - Add methods to appropriate file
   - Use generic `do()` or `doPaginated()` methods
   - Update rate limits in `NewClient()` if needed

### Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `GOOGLE_CREDENTIALS` | Path to credentials file or inline JSON | `/path/to/creds.json` | Yes |
| `GOOGLE_IMPERSONATE_USER` | Email to impersonate (service accounts) | `admin@example.com` | For Admin SDK |
| `GOOGLE_DOMAIN` | Default domain for operations | `example.com` | No |
| `GOOGLE_API_KEY` | API key (CICD mode) | `AIza...` | If using API key |
| `GOOGLE_OAUTH_CLIENT` | OAuth client (CICD mode, base64) | `eyJ0eXBlIj...` | If using OAuth |
| `GOOGLE_SERVICE_ACCOUNT` | Service account (CICD mode, base64) | `eyJ0eXBlIj...` | If using SA |
| `REGO_ENCRYPTION_KEY` | Cache encryption key | 32-byte key | Yes |

### Rate Limiting

Different Google APIs have different rate limits:

| API | Requests/Minute | Configuration |
|-----|-----------------|---------------|
| Admin SDK | 2400 | Set on admin operations |
| Drive API | 12000 | Set on drive operations |
| Sheets API | 60 | Set on sheets operations |
| Chrome Policy | Default | Uses client default |

Rate limits are automatically applied based on the operation type.

## Important Notes

- **Service Account Setup**: Requires domain-wide delegation for Admin SDK operations
- **Scope Management**: Use friendly names from `LoadScopes()` (e.g., "Admin SDK API" not URLs)
- **Cache Duration**: 30 minutes for most operations (configurable per request)
- **Embedded Resources**:
  - `google_directory.json`: Directory API endpoints
  - `google_endpoints.json`: All API endpoint definitions
  - `google_scopes.json`: Scope URL mappings
- **Multi-Domain**: Switch domains using client configuration
- **URL Building**: `BuildURL()` handles dynamic parameter substitution
- **Policy Schemas**: Cached after first retrieval for performance

### Special Features

#### Drive API
- Recursive folder traversal with depth control
- Path building from file ID
- Batch permission updates
- Ownership transfer capabilities

#### Sheets API
- Automatic struct-to-sheet conversion
- Value range generation and validation
- Batch update support
- A1 notation handling

#### Chrome Policy API
- Schema discovery and caching
- Policy cloning between OrgUnits
- Support for user and device policies
- PolicyTarget interface for flexible targeting

#### Device API
- Full method chaining support
- Complex query building
- Projection control for response size
- Child OrgUnit inclusion

## Common Pitfalls

1. **Scopes**: Use friendly names from `api.go` (e.g., "Admin SDK API" not the URL)
2. **Impersonation**: Service accounts must impersonate a user with admin privileges
3. **Rate Limits**: Different for each API - monitor usage to avoid 429 errors
4. **Pagination**: Some APIs limit max results per page (e.g., 500 for devices)
5. **Drive Paths**: Use `BuildFilePath()` for accurate path construction
6. **Spreadsheet Ranges**: Use A1 notation, quote sheet names with spaces
7. **Customer ID**: Use "my_customer" for your primary domain
8. **Auth Errors**: Check scopes match requested operations
9. **Policy Schemas**: First retrieval can be slow due to discovery
10. **CICD Mode**: Ensure env vars contain base64-encoded credentials

## Troubleshooting

### Authentication Issues
```go
// Enable debug logging
client, _ := google.NewClient(config, ac, log.DEBUG)

// Check JWT generation
jwt, err := client.GenerateJWT()
if err != nil {
    // Verify service account key format
    // Check impersonation email
    // Ensure scopes are correct
}
```

### Scope Errors
```go
// List available scopes
scopes := google.LoadScopes()
for name, url := range scopes {
    fmt.Printf("%s: %s\n", name, url)
}

// Ensure your AuthCredentials includes required scopes
ac.Scopes = []string{
    "Admin SDK API",
    "Google Drive API",
    "Google Sheets API",
}
```

### Rate Limit Handling
```go
// Monitor rate limit headers in responses
// The client automatically handles rate limiting
// For custom limits:
client.AdminRateLimit = 1000 // Reduce if hitting limits
```
