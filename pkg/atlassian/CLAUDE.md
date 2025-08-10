# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `atlassian` package provides a comprehensive Go client library for interacting with Atlassian Cloud APIs. It supports multiple Atlassian services including Confluence, Jira Software, Jira Service Management, Cloud Admin, Organizations, and SCIM-based user provisioning.

**Implementation Status:**
- ✅ Client initialization and configuration
- ✅ Authentication setup
- ✅ Caching infrastructure
- ✅ Rate limiting
- ✅ User Provisioning: `ListAllUsers()`
- 🚧 Organizations API (structure only, no methods implemented)
- ❌ Jira Cloud operations
- ❌ Confluence Cloud operations
- ❌ Error response parsing
- ❌ Pagination handling

## Architecture

### Core Components

1. **Client** (`atlassian.go`):
   - Main client struct with HTTP client, caching, logging, and rate limiting
   - Key methods: `NewClient()`, `do[T any]()`, `BuildURL()`
   - Functional options pattern for configuration
   - Method chaining support via service-specific clients

2. **Entities** (`entities.go`):
   - Core data structures and configuration options
   - Pagination support via `AtlassianPage` struct
   - Client options: `WithSandbox()`, `WithCustomOrgID()`, `WithCustomBaseURL()`, `WithCustomToken()`

3. **Service Clients**:
   - `organizations.go`: Organizations REST API operations
   - `user_provisioning.go`: SCIM-based user and group provisioning

### Key Design Patterns

- **Functional Options**: Clean configuration using `WithXxx()` functions
- **Method Chaining**: Fluent interface (e.g., `client.Organizations().Users()`)
- **Generic Requests**: Type-safe API calls via `do[T any]()`
- **Environment-based Config**: Twelve-Factor app principles with env vars
- **Caching**: Optional response caching with encrypted storage
- **Service Separation**: Each Atlassian service has its own client struct

### Service Endpoints

The package provides pre-configured endpoints for various Atlassian services:

- **Cloud Admin**: `https://api.atlassian.com/admin`
- **Jira Cloud**: `https://{org}.{base}/rest/api/3`
- **Jira Software Cloud**: `https://{org}.{base}/rest/agile/latest`
- **Jira Service Management**: `https://{org}.{base}/rest/servicedeskapi`
- **Confluence Cloud**: `https://{org}.{base}/wiki/api/v2`
- **Forms**: `https://api.atlassian.com/jira/forms`

#### Organizations API Endpoints:
- Base: `/v1/orgs`
- Users: `/v1/orgs/{orgId}/users`
- Groups: `/v1/orgs/{orgId}/groups`
- Domains: `/v1/orgs/{orgId}/domains`
- Events: `/v1/orgs/{orgId}/events`
- Policies: `/v1/orgs/{orgId}/policies`
- Directory: `/v1/orgs/{orgId}/directory`

#### User Provisioning (SCIM) Endpoints:
- Base: `/scim`
- Users: `/scim/directory/{directoryId}/Users`
- Groups: `/scim/directory/{directoryId}/Groups`
- Schemas: `/scim/directory/{directoryId}/Schemas`

## Development Tasks

### Running Tests
```bash
# Run tests for the atlassian package
go test ./pkg/internal/tests/atlassian/...

# Run with verbose output
go test -v ./pkg/internal/tests/atlassian/...
```

### Usage Examples

#### Basic Client Initialization
```go
import "github.com/gemini-oss/rego/pkg/atlassian"

// Create client with default configuration
client := atlassian.NewClient(log.INFO)

// Create client with sandbox configuration
client := atlassian.NewClient(log.DEBUG, atlassian.WithSandbox())

// Create client with custom environment variable keys
client := atlassian.NewClient(log.INFO,
    atlassian.WithCustomOrgID("MY_CUSTOM_ORG_ENV"),
    atlassian.WithCustomToken("MY_CUSTOM_TOKEN_ENV"),
)
```

#### Organizations API Usage
```go
// Access Organizations client
orgClient := client.Organizations()

// Future methods would be called like:
// users, err := orgClient.ListUsers()
// groups, err := orgClient.ListGroups()
```

#### User Provisioning (SCIM) API Usage
```go
// Access User Provisioning client
upClient := client.UserProvisioning()

// List all users in a directory
users, err := upClient.ListAllUsers(directoryID)
if err != nil {
    log.Fatal("Failed to list users:", err)
}

// Note: Response is currently typed as 'any' - type assertion required
```

#### Caching Usage
```go
// Enable cache for the next request
client.UseCache().UserProvisioning().ListAllUsers(directoryID)

// Manual cache operations
client.SetCache("custom-key", data, 1*time.Hour)
var cachedData MyType
if client.GetCache("custom-key", &cachedData) {
    // Use cached data
}
```

### Common Operations

1. **Adding New Service Endpoints**:
   - Create service-specific client struct
   - Add method to main Client returning the service client
   - Define endpoints as struct fields
   - Implement API methods using generic `do()` function

2. **Working with Pagination**:
   - Use `AtlassianPage` for paginated responses
   - Check `HasNextPage()` before fetching next page
   - Use `NextPage()` to get the next page URL

3. **Implementing Caching**:
   - Chain `UseCache()` before API calls
   - Use meaningful cache keys: `{service}_{operation}_{params}`
   - Set appropriate TTL with `SetCache()`

### Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `ATLASSIAN_ORG_ID` | Organization subdomain | `mycompany` | Yes |
| `ATLASSIAN_BASE_URL` | Base domain | `atlassian.net` | Yes |
| `ATLASSIAN_API_TOKEN` | API authentication token | `ATATT3xFfG...` | Yes |
| `ATLASSIAN_SANDBOX_ORG_ID` | Sandbox organization subdomain | `mycompany-sandbox` | If using sandbox |
| `ATLASSIAN_SANDBOX_BASE_URL` | Sandbox base domain | `atlassian.net` | If using sandbox |
| `ATLASSIAN_SANDBOX_API_TOKEN` | Sandbox API token | `ATATT3xFfG...` | If using sandbox |
| `ATLASSIAN_USE_SANDBOX` | Enable sandbox mode | `true` or `false` | No (default: false) |
| `REGO_ENCRYPTION_KEY` | Cache encryption key | 32-byte key | Yes |

### Authentication Details

The Atlassian client uses API token authentication with specific formatting requirements:

1. **API Token**: Obtained from [Atlassian account settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. **Organization ID**: The subdomain part of your Atlassian URL (e.g., "mycompany" from "mycompany.atlassian.net")
3. **Base URL**: The domain suffix (e.g., "atlassian.net")

The client constructs the full base URL as: `https://{orgID}.{baseURL}/`

## Important Notes

- Authentication uses "SSWS" token format in headers
- Default request timeout is 75 seconds
- Rate limiting is automatically handled according to [Atlassian's API limits](https://developer.atlassian.com/cloud/admin/organization/rest/intro/#rate-limits)
- JSON is the default content type for all requests
- Cache storage uses GOB format with encryption
- Sandbox mode is activated via `WithSandbox()` option
- Currently returns `any` type for API responses - type assertions required

### Rate Limiting

The client automatically handles rate limiting:
- Uses reset headers from API responses
- Automatically pauses requests when limits are approached
- Logging of rate limit status available at DEBUG level

### Error Handling

The package uses standard Go error handling patterns:

```go
users, err := client.UserProvisioning().ListAllUsers(directoryID)
if err != nil {
    // Handle error - could be network, authentication, or API error
    return fmt.Errorf("failed to list users: %w", err)
}
```

Note: The Error struct in entities.go is currently not implemented. Errors are returned as standard Go errors.

### Data Structures

Currently, the package uses generic `any` types for API responses. Future versions will implement proper type definitions for:
- User entities (SCIM format)
- Group entities
- Organization details
- Error responses
- Pagination metadata

For now, users must handle type assertions on returned data.

## Common Pitfalls

1. **Environment Variables**: Ensure correct env vars are set for sandbox/production
2. **Token Format**: API token must include "SSWS" prefix in headers (handled automatically)
3. **Pagination**: Always check for next page to avoid incomplete results
4. **Cache Keys**: Use unique, descriptive cache keys to prevent collisions
5. **Error Handling**: Check both HTTP status and unmarshaling errors
6. **Organization ID Format**: Don't include the full URL, just the subdomain
7. **Directory ID Required**: User Provisioning API requires a directory ID parameter
8. **Generic Types**: Some methods return `any` type - ensure proper type assertion
9. **Timeout**: Default timeout is 75 seconds - may need adjustment for large operations
10. **Cache Key Collisions**: Use descriptive cache keys including operation parameters

## Configuration Options

The client supports functional options for flexible configuration:

- `WithSandbox()`: Switch to sandbox environment variables
- `WithCustomOrgID(key)`: Use a custom environment variable for organization ID
- `WithCustomBaseURL(key)`: Use a custom environment variable for base URL
- `WithCustomToken(key)`: Use a custom environment variable for API token

These options allow for multiple Atlassian instances or non-standard environment variable naming.
