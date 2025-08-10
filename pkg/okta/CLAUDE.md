# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `okta` package provides a comprehensive Go client library for the Okta API. It enables programmatic access to user management, group operations, application assignments, device management, IAM roles, and multi-factor authentication settings with built-in caching and rate limiting.

**Key Features:**
- User lifecycle management (create, activate, deactivate, delete)
- Group and group rule operations
- Application management and user assignments
- Device inventory with OData query support
- Multi-factor authentication enrollment and management
- IAM role reporting with concurrent processing
- Dynamic rate limiting based on Okta response headers
- Encrypted caching with configurable TTL
- Sandbox environment support

## Architecture

### Core Components

1. **Client** (`okta.go`):
   - Main client struct with HTTP client, caching, and logging
   - Key methods: `NewClient()`, `do[T any]()`, `doPaginated[T, S]()`
   - Functional options for configuration (sandbox support)
   - SSWS token authentication

2. **Entities** (`entities.go`):
   - Comprehensive struct definitions for all Okta resources
   - Custom JSON marshaling for complex types
   - Pagination interfaces: `Slice[T]`, `Struct[T]`
   - Validation methods for attributes

3. **Domain Clients**:
   - `users.go`: User lifecycle and management
   - `groups.go`: Group operations and membership
   - `applications.go`: App management and assignments
   - `devices.go`: Device inventory and queries
   - `roles.go`: IAM roles and permissions
   - `user_factors.go`: MFA enrollment and management
   - `attributes.go`: Attribute operations (partial implementation)

### Key Design Patterns

- **Method Chaining**: Domain-specific clients (`client.Users()`, `client.Groups()`)
- **Generic Pagination**: Two methods - `doPaginated` for slices, `doPaginatedStruct` for structs
- **Functional Options**: Clean configuration with `WithSandbox()`, `WithCustomToken()`
- **Concurrent Processing**: Parallel operations with semaphore pattern (role reports)
- **Caching Strategy**: Variable TTL - 5 min (default), 30 min (user lists), 60 min (role reports)
- **Interface-based Pagination**: `Slice[T]` and `Struct[T]` interfaces for flexible pagination
- **Dynamic Rate Limiting**: Uses response headers to adjust rate limits automatically

## Development Tasks

### Running Tests
```bash
# Run tests for the okta package
go test ./pkg/internal/tests/okta/...

# Run with verbose output
go test -v ./pkg/internal/tests/okta/...
```

### Usage Examples

#### Client Initialization
```go
// Production environment
client := okta.NewClient(log.INFO)

// Sandbox environment
client := okta.NewClient(log.DEBUG, okta.WithSandbox())

// Custom configuration
client := okta.NewClient(log.INFO,
    okta.WithCustomOrgName("MY_ORG_ENV"),
    okta.WithCustomToken("MY_TOKEN_ENV"),
)
```

#### User Operations
```go
// List all users with caching
users, err := client.Users().ListAllUsers()

// Search users
query := okta.UserQuery{
    Q: "satoshi",
    Filter: "status eq \"ACTIVE\"",
    Limit: 100,
}
users, err := client.Users().Query(&query).List()

// User lifecycle
created, err := client.Users().CreateUser(&okta.User{
    Profile: okta.UserProfile{
        FirstName: "Satoshi",
        LastName:  "Nakamoto",
        Email:     "satoshi.nakamoto@example.com",
        Login:     "satoshi.nakamoto@example.com",
    },
})
err = client.Users().ActivateUser(userID, sendEmail)
err = client.Users().DeactivateUser(userID, sendEmail)
err = client.Users().DeleteUser(userID, sendEmail)
```

#### Device Management
```go
// List all devices
devices, err := client.Devices().ListAll()

// Query devices with OData
devices, err := client.Devices().
    Query("status eq \"ACTIVE\" and platform eq \"Linux\"").
    ListAll()

// Get managed devices (custom method)
managed, err := client.Devices().ListManagedDevices()

// List users for a device
users, err := client.Devices().ListUsersForDevice(deviceID)
```

#### Group Operations
```go
// List all groups
groups, err := client.Groups().ListAllGroups()

// Get group with specific fields
group, err := client.Groups().
    Expand("stats,app").
    GetGroup(groupID)

// List group rules
rules, err := client.Groups().ListAllGroupRules()
```

#### Application Management
```go
// List all applications
apps, err := client.Applications().ListAllApplications()

// List users assigned to an app
users, err := client.Applications().ListAllApplicationUsers(appID)

// Query with parameters
query := okta.AppQuery{IncludeNonDeleted: true}
apps, err := client.Applications().Query(&query).List()
```

#### Multi-Factor Authentication
```go
// List enrolled factors
factors, err := client.UserFactors().ListEnrolledFactors(userID)

// List available factors for enrollment
available, err := client.UserFactors().ListSupportedFactors(userID)

// Enroll a factor
enrolled, err := client.UserFactors().EnrollFactor(userID, &okta.UserFactor{
    FactorType: "token:software:totp",
    Provider:   "GOOGLE",
})
```

#### Role Reporting (Concurrent)
```go
// Generate comprehensive role report
report, err := client.Roles().GenerateRoleReport()
// This uses concurrent processing to fetch all roles and their assigned users
```

### Common Operations

1. **Pagination Handling**:
   - Use `doPaginated()` for endpoints returning arrays
   - Use `doPaginatedStruct()` for endpoints returning objects with embedded arrays
   - Implement `Slice` interface for array types
   - Implement `Struct` interface for object types with array fields

2. **Query Parameters**:
   ```go
   // UserQuery example
   query := okta.UserQuery{
       Q:         "satoshi",            // Search term
       After:     "cursor_value",       // Pagination cursor
       Limit:     200,                  // Results per page
       Filter:    "status eq \"ACTIVE\"",
       Search:    "profile.department eq \"Engineering\"",
       SortBy:    "profile.lastName",
       SortOrder: "asc",
   }
   ```

3. **Error Handling**:
   ```go
   users, err := client.Users().ListAllUsers()
   if err != nil {
       var oktaErr *okta.Error
       if errors.As(err, &oktaErr) {
           // Handle Okta-specific error
           for _, cause := range oktaErr.ErrorCauses {
               log.Printf("Error: %s", cause.ErrorSummary)
           }
       }
       return err
   }
   ```

### Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `OKTA_ORG_NAME` | Organization subdomain | `mycompany` | Yes |
| `OKTA_BASE_URL` | Base domain | `okta.com` | No (default: okta.com) |
| `OKTA_API_TOKEN` | API token (without SSWS prefix) | `00abc...` | Yes |
| `OKTA_SANDBOX_ORG_NAME` | Sandbox org subdomain | `mycompany-sandbox` | If using sandbox |
| `OKTA_SANDBOX_BASE_URL` | Sandbox base domain | `oktapreview.com` | If using sandbox |
| `OKTA_SANDBOX_API_TOKEN` | Sandbox API token | `00xyz...` | If using sandbox |
| `OKTA_USE_SANDBOX` | Enable sandbox mode | `true` | No |
| `REGO_ENCRYPTION_KEY` | 32-byte cache encryption key | `your-32-byte-key` | Yes |

**Important:** Do NOT include "SSWS " prefix in the API token environment variable. The client adds it automatically.

## Important Notes

- **Authentication**: SSWS tokens in Authorization header (prefix added automatically)
- **Rate Limiting**: Dynamic based on response headers (not fixed 1000/min)
- **Request Timeout**: 75 seconds for all operations
- **Cache File**: `rego_cache_okta.gob` with AES-256 encryption
- **Cache TTL**:
  - Default: 5 minutes
  - User lists: 30 minutes
  - Role reports: 60 minutes
- **Cache Keys**: Use full URL including query parameters
- **Pagination**: Uses Link headers with `rel="next"` parsing
- **OData Queries**: Supported for devices (not all fields are queryable)
- **Concurrent Processing**:
  - Uses `runtime.GOMAXPROCS(0)` for worker count
  - Semaphore pattern with buffered channels
  - Thread-safe map access with `sync.Mutex`

### API-Specific Details

#### Users API
- Supports lifecycle operations (create, activate, suspend, deactivate, delete)
- Search supports both `q` (query) and `search` parameters
- Filter syntax: `status eq "ACTIVE"`, `lastUpdated gt "2023-01-01T00:00:00.000Z"`

#### Devices API
- OData query support: `status`, `platform`, `managementStatus`
- Platform values: `ANDROID`, `IOS`, `MACOS`, `WINDOWS`
- Embedded user expansion with `expand=user`
- Custom managed devices method filters non-mobile devices

#### Groups API
- Supports group rules for dynamic membership
- Expand options: `stats`, `app`
- Group types: `OKTA_GROUP`, `APP_GROUP`, `BUILT_IN`

#### Applications API
- Sign-on modes vary by app type
- User assignment includes scope and credentials
- Supports inactive app queries with `IncludeNonDeleted`

#### Factors API
- Factor types: `push`, `sms`, `call`, `token:software:totp`, `token:hardware`
- Providers: `OKTA`, `GOOGLE`, `RSA`, `SYMANTEC`
- Enrollment requires activation for some factors

## Common Pitfalls

1. **Token Format**: Do NOT include "SSWS " in env var - client adds it automatically
2. **Pagination Types**:
   - Arrays: Use `doPaginated` with `Slice` interface
   - Objects with arrays: Use `doPaginatedStruct` with `Struct` interface
3. **Query Syntax**:
   - Devices: OData format (`status eq "ACTIVE"`)
   - Users: Mix of `q`, `filter`, and `search` parameters
4. **Rate Limits**: Don't assume 1000/min - check response headers
5. **Sandbox vs Production**: Set `OKTA_USE_SANDBOX=true` or use `WithSandbox()`
6. **User Lifecycle**:
   - Can't delete active users (must deactivate first)
   - Some operations need `sendEmail` parameter
7. **Device Queries**: Limited queryable fields - check Okta docs
8. **Error Arrays**: Okta returns `errorCauses` array, not single error
9. **Cache Invalidation**: No automatic invalidation on mutations
10. **Concurrent Limits**: Role reports can overwhelm API if too parallel

## Troubleshooting

### Authentication Issues
```go
// Enable debug logging
client := okta.NewClient(log.DEBUG)

// Check token format
if !strings.HasPrefix(os.Getenv("OKTA_API_TOKEN"), "SSWS") {
    // Good - token should NOT have prefix in env var
}
```

### Rate Limiting
```go
// The client automatically handles rate limiting
// To see rate limit headers in debug logs:
client := okta.NewClient(log.DEBUG)

// Rate limiter uses response headers:
// X-Rate-Limit-Limit
// X-Rate-Limit-Remaining
// X-Rate-Limit-Reset
```

### Query Debugging
```go
// For devices - use OData syntax
devices, err := client.Devices().
    Query("platform eq 'LINUX' and status eq 'ACTIVE'").
    ListAll()

// For users - use appropriate parameter
users, err := client.Users().
    Query(&okta.UserQuery{
        Q:      "satoshi",                          // Name/email search
        Filter: "status eq \"ACTIVE\"",             // Status filter
        Search: "profile.department eq \"IT\"",     // Profile search
    }).
    List()
```
