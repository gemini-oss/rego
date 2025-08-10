# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `jamf` package provides a Go client library for interacting with Jamf Pro's REST APIs. It supports both the Classic API (XML-based) and the modern Jamf Pro API (JSON-based), enabling comprehensive management of Apple devices including Macs and iOS/iPadOS devices.

**Key Features:**
- Dual API support (Classic XML and Pro JSON)
- Method chaining for complex queries
- Section-based response optimization
- Concurrent pagination for large datasets
- Automatic token management
- Built-in caching with 5-minute TTL
- MDM command execution
- Configuration profile management

## Architecture

### Core Components

1. **Client** (`jamf.go`):
   - Main client struct with HTTP client, caching, and logging
   - Key methods: `NewClient()`, `do[T any]()`, `doConcurrent[T]()`
   - Token-based authentication with automatic renewal
   - Dual API support (Classic XML and Pro JSON)
   - Note: Client panics if required environment variables are missing

### Configuration

Unlike other services in ReGo, Jamf doesn't use a Config struct. Configuration is handled entirely through environment variables:

```go
// Client initialization only takes verbosity parameter
client := jamf.NewClient(log.INFO)

// Required environment variables must be set before initialization
// The client will panic if they are missing
```

2. **Entities** (`entities.go`):
   - Comprehensive struct definitions for all Jamf resources
   - `JamfAPIResponse` interface for pagination
   - Device types: Computer, MobileDevice, UserDevice
   - Configuration types: Profile, Group, Extension Attribute

3. **API Implementations**:
   - `devices.go`: Computer and mobile device operations
   - `admin.go`: User and privilege management
   - `management.go`: MDM operations (profile renewal, framework repair)
   - `classic_*.go`: Classic API implementations (users, groups, profiles)
   - `version.go`: Version information retrieval

### Key Design Patterns

- **Fluent Interface**: Chainable methods for complex queries
- **Generic Requests**: Type-safe API calls with `do[T any]()`
- **Concurrent Pagination**: Parallel page fetching with `doConcurrent()`
- **Dual API Support**: Automatic switching between Classic and Pro APIs
- **Token Management**: Automatic authentication token handling

## Development Tasks

### Running Tests
```bash
# Run tests for the jamf package
go test ./pkg/internal/tests/jamf/...

# Run with verbose output
go test -v ./pkg/internal/tests/jamf/...
```

### Usage Examples

#### Method Chaining Pattern
```go
// DeviceClient provides chainable methods
computers, err := client.Devices()
    .Sections([]string{jamf.Section.General, jamf.Section.Hardware})  // Select data sections
    .PageSize(100)                                                    // Items per page (default: 100)
    .Page(0)                                                          // Page number (0-based)
    .Sort([]string{"general.name:asc"})                               // Sort criteria
    .Filter("general.platform==\"Mac\"")                              // RSQL filter
    .ListAllComputers()                                               // Execute query

// Get specific device with minimal data
computer, err := client.Devices()
    .Sections([]string{jamf.Section.General})
    .GetComputerDetails("123")
```

#### Classic API Usage
```go
// Configuration profiles
profiles, err := client.Profiles().ListAllConfigurationProfiles()
profile, err := client.Profiles().GetConfigurationProfileDetails("42")

// User management
users, err := client.ClassicUsers().GetAllUsers()
user, err := client.ClassicUsers().GetUserByID("100")

// User groups
groups, err := client.UserGroups().GetAllUserGroups()
group, err := client.UserGroups().GetUserGroupByID("5")
```

#### MDM Commands
```go
// Renew MDM profile for multiple devices
response, err := client.RenewMDMProfile([]string{"udid1", "udid2", "udid3"})

// Repair management framework
result, err := client.RepairManagementFramework("device-id")
```

#### Filter Examples
```go
// Complex filter with multiple criteria
computers, err := client.Devices()
    .Filter("general.platform==\"Mac\" and hardware.model==\"MacBookPro18,1\"")
    .Sort([]string{"general.name:asc", "general.last_contact_time:desc"})
    .ListAllComputers()

// Filter mobile devices by OS version
devices, err := client.Devices()
    .Filter("operatingSystem.version>\"16.0\"")
    .ListAllMobileDevices()
```

### Common Operations

1. **Working with APIs**:
   - Pro API: Use `BuildURL()` with JSON payloads
   - Classic API: Use `BuildClassicURL()` with XML payloads
   - Content-Type is automatically set based on API type

2. **Adding New Endpoints**:
   - Define structs in `entities.go`
   - Implement methods in appropriate file
   - Use `do()` for single requests, `doConcurrent()` for paginated
   - Follow existing patterns for Classic vs Pro API

### Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `JSS_URL` | Jamf Pro server URL | `https://company.jamfcloud.com` | Yes |
| `JSS_USERNAME` | Username for authentication | `api-user` | Yes |
| `JSS_PASSWORD` | Password for authentication | `secure-password` | Yes |
| `REGO_ENCRYPTION_KEY` | Key for cache encryption | 32-byte key | Yes |

**Note:** The package uses `JSS_` prefix (Jamf Software Server) for legacy compatibility, not `JAMF_`.

### Available Sections

The Section constants optimize API responses by limiting returned data:

- `Section.General` - Basic device information
- `Section.Hardware` - Hardware specifications
- `Section.OperatingSystem` - OS details
- `Section.Security` - Security settings
- `Section.Storage` - Disk information
- `Section.Applications` - Installed applications
- `Section.UserAndLocation` - User assignment
- `Section.ConfigurationProfiles` - Installed profiles
- `Section.LocalUserAccounts` - Local users
- `Section.Certificates` - Installed certificates
- `Section.ExtensionAttributes` - Custom attributes
- `Section.GroupMemberships` - Group assignments
- `Section.DiskEncryption` - FileVault status
- `Section.Purchasing` - Purchase information
- `Section.Printers` - Configured printers
- `Section.Services` - Running services
- `Section.Attachments` - File attachments
- `Section.Plugins` - Browser plugins
- `Section.PackageReceipts` - Installed packages
- `Section.Fonts` - Installed fonts
- `Section.LicensedSoftware` - License compliance
- `Section.IBeacons` - iBeacon regions
- `Section.SoftwareUpdates` - Available updates
- `Section.ContentCaching` - Content cache status

## Important Notes

- **Authentication Flow**: Basic auth → Bearer token via `/api/v1/auth/token`
- **Token Management**: Stored in HTTP headers, no automatic refresh on expiry
- **API Format**: Classic API uses XML, Pro API uses JSON
- **Cache Duration**: Fixed 5-minute TTL for all operations
- **Concurrent Operations**: Uses `runtime.GOMAXPROCS(0)` for concurrency limit
- **Section Optimization**: Always specify sections to minimize payload size
- **Client Initialization**: Panics if required environment variables are missing
- **Error Handling**: Uses standard error wrapping without custom types

### Concurrent Operations

The `doConcurrent` function handles paginated requests in parallel:

- Automatically detects total pages from first request
- Limits concurrent requests to CPU core count
- Maintains result ordering despite parallel execution
- Used internally by `ListAllComputers()` and `ListAllMobileDevices()`
- Falls back to single request if only one page exists

### API Endpoints

Endpoints are organized as package variables:

**Pro API (JSON):**
- Base: `/api/v1`, `/api/v2`
- Devices: `ComputersInventory`, `ComputersInventoryDetail`, `MobileDev`
- Management: `ManagementFramework`, `RenewProfile`
- Admin: `AdminUsers`, `AdminUserPrivileges`

**Classic API (XML):**
- Base: `/JSSResource`
- Profiles: `ConfigurationProfiles`
- Users: `ClassicUsers`
- Groups: `ClassicUserGroups`

## Common Pitfalls

1. **API Selection**: Some operations only exist in Classic or Pro API
2. **Sections**: Always specify sections to avoid large response payloads
3. **Filters**: Pro API filters use RSQL syntax (e.g., `general.name=="value"`)
4. **Rate Limiting**: No built-in rate limiting - monitor server load
5. **XML Parsing**: Classic API responses may have inconsistent structure
6. **Token Expiry**: No automatic refresh - implement renewal logic
7. **Environment Variables**: Uses `JSS_` prefix, not `JAMF_`
8. **Client Panics**: Wrap initialization in error handling
9. **Cache Duration**: Fixed 5 minutes may not suit all use cases
10. **Page Size**: Default 100 may be too large for complex queries

## Limitations

1. **No Token Auto-Refresh**: Tokens must be manually renewed when expired
2. **No Built-in Retry**: Failed requests are not automatically retried
3. **Limited Error Types**: Uses standard errors without custom types
4. **Panic on Init**: Client panics instead of returning errors during initialization
5. **No Config Struct**: Cannot pass custom configuration (only verbosity)
6. **Fixed Cache Duration**: Hard-coded 5-minute cache TTL
7. **No Batch Operations**: Must handle batch operations manually
8. **Limited Classic API Coverage**: Only implements select Classic API endpoints

## Best Practices

1. **Always Specify Sections**: Minimize response size and improve performance
2. **Use Filters**: Leverage RSQL filters to reduce data transfer
3. **Handle Token Expiry**: Implement token refresh logic in long-running applications
4. **Cache Wisely**: 5-minute cache may be too short/long for your use case
5. **Page Size**: Adjust based on your network and Jamf server capabilities
6. **Concurrent Requests**: Monitor server load when using concurrent operations
7. **Error Handling**: Wrap client creation in error handling to avoid panics
8. **Section Selection**: Use only required sections to reduce memory usage
