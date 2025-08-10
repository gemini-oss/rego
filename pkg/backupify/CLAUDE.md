# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `backupify` package provides a Go client library for interacting with Datto's Backupify cloud-to-cloud backup service. It enables programmatic access to backup management, export operations, activity monitoring, and snapshot handling for SaaS applications, with current focus on Google Workspace services.

**Key Features:**
- Web UI automation (not a public API) for backup operations
- Export creation and download management
- Activity tracking for backup/restore/export operations
- User storage reporting with deduplication estimates
- Concurrent export downloads with semaphore control
- Support for Google Drive, Shared Drive, and Gmail

**Important Note:** This package interacts with Backupify's web interface endpoints, not a stable public API. Authentication and functionality may change without notice.

## Architecture

### Core Components

1. **Client** (`backupify.go`):
   - Main client struct with HTTP client, caching, and logging
   - Key methods: `NewClient()`, `do[T any]()`, `BuildURL()`
   - Generic request handling with JSON marshaling
   - Export token management for authenticated operations

2. **Entities** (`entities.go`):
   - Core data structures for API operations
   - `AppType` enum for service types (GoogleDrive, SharedDrive, GoogleMail)
   - User, Activity, Export, and Snapshot types
   - Helper methods like `Map()` for data transformation

3. **Service Clients**:
   - `users.go`: User management and storage reporting
   - `activities.go`: Backup/restore/export activity tracking
   - `exports.go`: Export creation and download management
   - `snapshots.go`: Snapshot date retrieval

### Configuration

```go
// Configuration structure
type Config struct {
    NodeURL     string        // Required: Backupify node URL (e.g., "node01")
    CustomerID  string        // Required: Your customer ID
    ExportToken string        // Required: Export authentication token
    PHPSessID   string        // Required: PHP session ID for web UI auth
    AppType     AppType       // Required: Type of service (GoogleDrive, SharedDrive, GoogleMail)
    EnableCache bool          // Optional: Enable caching (default: true)
    CacheTTL    time.Duration // Optional: Cache duration (varies by operation)
    Verbosity   int           // Optional: Log level (log.DEBUG, log.INFO, etc.)
}

// Client initialization
client := backupify.NewClient(log.INFO,
    backupify.WithAppType(backupify.GoogleDrive),
    backupify.WithCache(true, 2*time.Hour),
)
```

### Key Design Patterns

- **Method Chaining**: Fluent interface via service-specific clients
- **Generic Requests**: Type-safe API calls with `do[T any]()`
- **Concurrent Downloads**: Semaphore-controlled export downloads
- **Caching**: Encrypted file-based caching with configurable TTL
- **Pagination**: Automatic handling of paginated responses

## Development Tasks

### Running Tests
```bash
# Run tests for the backupify package
go test ./pkg/internal/tests/backupify/...

# Run with verbose output
go test -v ./pkg/internal/tests/backupify/...
```

### Usage Examples

#### Complete Export Workflow
```go
// 1. Initialize client and set app type
client := backupify.NewClient(log.INFO)
client.AppType = backupify.GoogleDrive

// 2. Get all users
users, err := client.Users().GetAllUsers()
if err != nil {
    return fmt.Errorf("failed to get users: %w", err)
}

// 3. Filter out zero-storage and deleted users
activeUsers := client.Users().filterUsersBySize(users, 0)

// 4. Export data for each user
exports, err := client.Exports().ExportUsers(activeUsers)
if err != nil {
    return fmt.Errorf("failed to export users: %w", err)
}

// 5. Wait for exports to process
time.Sleep(5 * time.Minute)

// 6. Check export status
activities, err := client.Activities().GetActivities()
if err != nil {
    return fmt.Errorf("failed to get activities: %w", err)
}

// 7. Download completed exports
reports := client.Exports().DownloadAvailableExports(activities)

// 8. Generate storage report
storageReport := client.Users().UserStorageReport(users)
for letter, counts := range storageReport {
    fmt.Printf("%s: %d users\n", letter, counts.Count)
}
```

#### Storage Analysis
```go
// Get user storage information
users, _ := client.Users().GetAllUsers()

// Convert storage values to bytes (using binary calculation)
client.Users().convertUserBytes(users, true)

// Filter users by storage size (e.g., > 1GB)
largeUsers := client.Users().filterUsersBySize(users, 1073741824)

// Generate storage report grouped by first letter
report := client.Users().UserStorageReport(users)
```

#### Error Handling
```go
// Handle missing snapshots
snapshots, err := client.Snapshots().GetSnapshotDates(user)
if err != nil {
    if strings.Contains(err.Error(), "no snapshots found") {
        // User has no snapshots, skip
        continue
    }
    return fmt.Errorf("failed to get snapshots: %w", err)
}

// Handle export failures gracefully
export, err := client.Exports().ExportUser(user)
if err != nil {
    log.Warning("Failed to export user %s: %v", user.Email, err)
    continue // Continue with next user
}
```

### Available Methods

#### User Operations
- `GetAllUsers() (*Users, error)` - Retrieve all users
- `UserStorageReport(users *Users) map[string]UserCounts` - Generate storage report
- `convertUserBytes(users *Users, useBinary bool)` - Convert storage units to bytes
- `filterUsersBySize(users *Users, size float64) *Users` - Filter users by storage size

#### Activity Operations
- `GetActivities() (*Activities, error)` - Get all backup/export activities

#### Export Operations
- `ExportUsers(users *Users) ([]Exports, error)` - Bulk export users
- `ExportUser(user *User) (*Exports, error)` - Export single user
- `DownloadAvailableExports(activities *Activities) [][]string` - Download all ready exports
- `DownloadExport(activity *Item, export *Export) ([]string, error)` - Download specific export
- `DeleteExport(activity *Item, export *Export) error` - Delete an export

#### Snapshot Operations
- `GetSnapshotDates(user *User) (*Snapshots, error)` - Get backup snapshot dates

### Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `BACKUPIFY_NODE_URL` | Backupify node URL | `https://node01.backupify.com` | Yes |
| `BACKUPIFY_CUSTOMER_ID` | Your customer ID | `123456` | Yes |
| `BACKUPIFY_EXPORT_TOKEN` | Export authentication token | `abc123...` | Yes |
| `BACKUPIFY_PHPSESSID` | PHP session ID | `sess_xyz789...` | Yes |
| `REGO_ENCRYPTION_KEY` | Cache encryption key | 32-byte key | Yes |

### Authentication

**Obtaining Credentials:**
1. Log into Backupify web interface
2. Navigate to export page
3. Use browser developer tools to find:
   - `PHPSESSID` cookie value
   - Export token from API requests
   - Customer ID from URL
   - Node URL from current domain

**Session Management:**
- PHP sessions may expire - monitor for authentication errors
- Export tokens are typically long-lived
- Consider implementing session refresh logic for long-running operations

## Important Notes

- **API Type**: Web UI automation, not a stable public API - expect changes
- **Authentication**: Requires both export token AND PHP session ID
- **Timeouts**: Default 30 seconds for requests
- **Concurrency**: Export downloads limited to 5 concurrent operations (adjustable)
- **Storage Calculations**:
  - Automatic unit conversion (KB/MB/GB/TB to bytes)
  - Support for binary (1024) and decimal (1000) calculations
  - Includes deduplication estimates in storage values
- **Caching Strategy**:
  - Activities: 5 minutes (frequently changing)
  - Users: 6 hours (relatively static)
  - Snapshots: 24 hours (very static)
  - Cache keys: `{endpoint}_{appType}` or `{userID}_{userName}_{appType}`

### Download Organization

- Downloads saved to: `Backupify/{AppType}/{UserEmail}/`
- File naming: `{timestamp}_{filename}`
- CSV reports generated for bulk downloads
- Automatic directory creation

### Activity Mapping

The `Map()` method on ActivityDetail:
- Deduplicates activities by snapshot ID
- Groups related export/download activities
- Useful for tracking export progress

## Common Pitfalls

1. **App Type**: Must set `AppType` before operations - no default value
2. **Authentication**: Both tokens required; PHP sessions expire unexpectedly
3. **Export Status**: Exports are asynchronous - always check activities
4. **Rate Limiting**: No official limits but use reasonable delays (1s between requests)
5. **User Filtering**: Methods filter out zero-storage and deleted users
6. **Large Exports**: May timeout or fail - implement retry logic
7. **Snapshot Errors**: "No snapshots found" is common for new/inactive users
8. **Download Paths**: Ensure write permissions for download directories
9. **Memory Usage**: Large user lists can consume significant memory
10. **Error Messages**: Web UI errors may be HTML instead of JSON

## Troubleshooting

### Authentication Issues
```go
// Enable debug logging to see full requests/responses
client := backupify.NewClient(log.DEBUG)

// Test authentication with a simple request
activities, err := client.Activities().GetActivities()
if err != nil {
    // Check: Are environment variables set?
    // Check: Is PHP session still valid?
    // Check: Is export token correct?
    // Try: Get fresh credentials from web UI
}
```

### Export Failures
```go
// Add delays between export requests
for _, user := range users.Items {
    export, err := client.Exports().ExportUser(&user)
    if err != nil {
        log.Warning("Export failed for %s: %v", user.Email, err)
    }
    time.Sleep(1 * time.Second) // Rate limiting
}
```

### Performance Optimization
```go
// Use caching for repeated operations
client.UseCache = true
client.CacheTTL = 6 * time.Hour

// Filter users before processing
activeUsers := client.Users().filterUsersBySize(users, 1024) // > 1KB

// Adjust concurrent downloads based on needs
const maxConcurrentDownloads = 10 // Increase for faster downloads
```
