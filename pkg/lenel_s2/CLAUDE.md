# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `lenel_s2` package is a Go client library for interacting with the Lenel S2 NetBox API. Lenel S2 is a physical access control and security management system. This package provides a comprehensive wrapper around the XML-based NetBox API to manage people, credentials, access levels, events, and physical security configurations.

## Architecture

### Core Components

1. **Client** (`lenel_s2.go`):
   - Main client struct managing HTTP connections, session state, logging, and caching
   - Key methods: `NewClient()`, `Login()`, `Logout()`, `BuildURL()`, `BuildRequest()`
   - Generic request handlers: `do()`, `doPaginated()`, `doStream()`

2. **Entities** (`entities.go`):
   - Core data structures and XML marshaling/unmarshaling logic
   - Defines all API request/response types
   - Command constants organized by category (Actions, Configuration, Events, History, People, Portals, etc.)

3. **API Implementation Files**:
   - `netbox_people.go`: User/person management operations
   - `netbox_events.go`: Event history and streaming capabilities
   - `netbox_configuration.go`: System configuration (UDFs, access levels, etc.)

### Key Design Patterns

- **XML-based API**: All communication uses XML payloads with custom marshaling
- **Session Management**: Authentication creates a session ID used for all subsequent requests
- **Pagination Support**: Handles paginated responses via `PaginatedResponse` interface
- **Caching**: Built-in encrypted cache support for API responses
- **Streaming**: Real-time event streaming with context cancellation support

## Development Tasks

### Running Tests
Currently no test files exist in this package. When implementing tests:
```bash
# Run tests for the lenel_s2 package
go test ./pkg/lenel_s2/...

# Run with verbose output
go test -v ./pkg/lenel_s2/...
```

### Common Operations

1. **Adding New API Methods**:
   - Define request/response structs in `entities.go`
   - Add command constant to appropriate category in `NetboxCommands`
   - Implement method in relevant file (people, events, configuration, etc.)
   - Use appropriate request handler (`do`, `doPaginated`, or `doStream`)

2. **Working with XML**:
   - Custom `UnmarshalXML` methods handle complex response parsing
   - Pay attention to XML namespaces and attributes
   - Error responses have special handling in `Response[E].UnmarshalXML()`

3. **Pagination**:
   - Implement `PaginatedResponse` interface: `NextToken()`, `Append()`, `SetCommand()`
   - Use `doPaginated()` for endpoints that return paginated data

4. **Environment Variables**:
   - `S2_URL`: Base URL for the S2 system
   - `S2_USERNAME`: Authentication username
   - `S2_PASSWORD`: Authentication password
   - `S2_INSECURE_SKIP_VERIFY`: Set to `"true"` to disable TLS certificate verification (use with caution)
   - `REGO_ENCRYPTION_KEY`: Key for cache encryption

## Important Notes

- The API uses XML exclusively - ensure proper XML tags on all structs
- Session management is handled automatically by the client
- All timestamps should use the "tzoffset" date format
- Error handling includes both API-level errors and command-level failures
- The streaming API uses multipart responses with XML chunks separated by boundaries
- Cache keys follow the pattern: `{url}_{command}_{parameters}`

### TLS Certificate Handling

The client supports disabling TLS certificate verification for environments with self-signed certificates or certificate verification issues. This should **only be used** when:
- Connecting to systems with self-signed certificates in trusted environments
- Troubleshooting certificate chain issues
- Development/testing environments

**WARNING**: Disabling TLS verification removes protection against man-in-the-middle attacks. Only use this in secure, trusted networks.

#### Configuration Options

**Option 1: Environment Variable** (Recommended for deployment)
```bash
export S2_INSECURE_SKIP_VERIFY=true
client := lenel_s2.NewClient("", log.INFO)
```

**Option 2: Functional Option** (Recommended for programmatic control)
```go
client := lenel_s2.NewClient("", log.INFO, lenel_s2.WithInsecureSkipVerify())
```

**Option 3: Custom HTTP Client** (Advanced use cases)
```go
customClient := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    },
}
client := lenel_s2.NewClient("", log.INFO, lenel_s2.WithHTTPClient(customClient))
```

#### Backwards Compatibility

The new options are fully backwards compatible. Existing code continues to work without any changes:
```go
// Existing code - still works perfectly
client := lenel_s2.NewClient("", log.INFO)
```

## Common Pitfalls

1. **XML Parsing**: The API returns different structures for success/failure - handled by custom unmarshalers
2. **Pagination**: Watch for "-1" or empty `NextKey`/`NextLogID` indicating no more pages
3. **Streaming**: Heartbeat messages have empty responses - these should be ignored
4. **Authentication**: Session IDs expire - handle re-authentication gracefully

## StreamEvents Implementation Plan

### Problem Analysis
The StreamEvents API returns a different XML structure than standard endpoints:
- Standard: `<NETBOX><RESPONSE><DETAILS>...</DETAILS></RESPONSE></NETBOX>`
- StreamEvents: `<NETBOX><RESPONSE command="StreamEvents"><EVENT>...</EVENT></RESPONSE></NETBOX>`

Key differences:
1. Response has `command` attribute
2. Data is in `EVENT` tag instead of `DETAILS`
3. Different event types have different fields

### Solution Design

#### 1. Create Event-Specific Response Types
```go
// NetboxEventResponse handles streaming event responses
type NetboxEventResponse[E any] struct {
    XMLName  xml.Name         `xml:"NETBOX"`
    ID       string           `xml:"sessionid,attr"`
    Response EventResponse[E] `xml:"RESPONSE"`
}

// EventResponse handles the RESPONSE element for streaming events
type EventResponse[E any] struct {
    Command  string `xml:"command,attr"`     // Command name (e.g., "StreamEvents")
    APIError int    `xml:"APIERROR"`         // API-level error codes
    Code     string `xml:"CODE"`             // SUCCESS or FAIL
    Event    *E     `xml:"EVENT"`            // Event data
    Error    string `xml:"ERRMSG,omitempty"` // Error message for failures
}
```

#### 2. Update Events Struct
Based on actual API responses, the Events struct should handle multiple event types:
```go
type Events struct {
    // Access control events
    PersonName   string `xml:"PERSONNAME,omitempty"`
    PortalName   string `xml:"PORTALNAME,omitempty"`
    EventName    string `xml:"EVTNAME,omitempty"`

    // System events
    ActivityID   string `xml:"ACTIVITYID,omitempty"`
    DescName     string `xml:"DESCNAME,omitempty"`
    CDT          string `xml:"CDT,omitempty"`
    PartName     string `xml:"PARTNAME,omitempty"`

    // Additional fields
    ACName       string `xml:"ACNAME,omitempty"`
    LoginAddress string `xml:"LOGINADDRESS,omitempty"`
    Detail       string `xml:"DETAIL,omitempty"`
}
```

#### 3. Update parseResponse in doStream
Modify to handle both response types:
```go
func parseResponse[T any](xmlData string, collected *[]T, processFunc func(T) bool, c *Client) error {
    // Try NetboxEventResponse first for streaming events
    var eventResult NetboxEventResponse[T]
    if err := xml.Unmarshal([]byte(xmlData), &eventResult); err == nil &&
       eventResult.Response.Command != "" && eventResult.Response.Event != nil {
        // Handle event response
        // ...
    } else {
        // Fall back to standard NetboxResponse
        var result NetboxResponse[T]
        // ...
    }
}
```

#### 4. Test Structure
```
pkg/internal/tests/lenel_s2/
├── lenel_s2_test.go          # Common test helpers
├── events_test.go            # StreamEvents tests
├── people_test.go            # People management tests
└── testdata/                 # Fixtures
```

### Implementation Steps

1. **Add new types to entities.go**
   - NetboxEventResponse[E]
   - EventResponse[E] with custom UnmarshalXML if needed

2. **Update Events struct**
   - Add all possible event fields with omitempty
   - Document which fields appear in which event types

3. **Modify doStream/parseResponse**
   - Try parsing as NetboxEventResponse first
   - Fall back to NetboxResponse for compatibility
   - Handle heartbeats (empty responses)

4. **Create comprehensive tests**
   - Mock multipart streaming server
   - Test different event types
   - Test error handling
   - Test context cancellation

5. **Update main.go**
   - Convert playground code to proper example
   - Reset to clean state for commit

6. **Document patterns**
   - Event type mappings
   - Field usage by event type
   - Best practices for event handling

### Event Types (from logs)
- **Access Events**: PersonName, PortalName (badge swipes)
- **System Events**: ActivityID, DescName, CDT, PartName
- **Login Events**: LoginAddress, PersonName
- **Other Events**: Will need to discover through testing

### Notes
- CDATA wrapping indicates fields may contain special characters
- Heartbeats have no EVENT/DETAILS, just empty RESPONSE
- Command attribute helps identify response type
- This approach maintains backward compatibility

## Implementation Progress (2025-08-07)

### StreamEvents Solution Implemented
Successfully fixed the XML parsing issue by:
1. Created `NetboxEventResponse[E]` and `EventResponse[E]` types in entities.go
2. Implemented custom `UnmarshalXML` for EventResponse to handle EVENT tags
3. Renamed `parseResponse` to `parseStreamEvent` in lenel_s2.go
4. Updated Events struct with all fields from real API responses

### Test Infrastructure Created
Complete test suite in `/pkg/internal/tests/lenel_s2/`:
- Mock server with proper session handling (REGO-SESSION-357)
- Multipart streaming tests with `mockMultipartEvent` function
- Crypto-themed test data (Satoshi Nakamoto)
- Number patterns: 3, 5, 7 (357, 753, 573)

### Key Lessons Learned
1. **XML Parsing**: Must handle raw body string for person ID detection in mock server
2. **Context Handling**: StreamEvents must return collected events even on cancellation
3. **Mock Server**: Need `io.ReadAll` before XML unmarshal to preserve body
4. **Compiler Warnings**: Use `_` for unused params, `any` instead of `interface{}`
5. **Encryption Key**: Must be exactly 32 bytes (added `!` to make it work)

### Known Issues
- `AccessHistory.Append()` has nil pointer issue
- Several unimplemented methods: GetUDFLists, GetCardFormats, GetAPIVersion, ModifyPerson

### Test Patterns Established
```go
// Event filtering
StreamEventsBuilder().
    WithEventType(EventTypes.AccessGranted).
    FilterByPersonName("Anthony Dardano", "Satoshi Nakamoto").
    Build()

// Mock person detection
if strings.Contains(bodyStr, "SATOSHI_753") {
    // Return Satoshi-specific data
}
```

### Next Priority Tasks
1. Test with actual S2 system
2. Implement remaining event functionality from PDF
3. Add SIEM forwarding integration tests

## Testing Strategy

The test infrastructure for Lenel S2 includes:

1. **Mock Server**: A comprehensive mock server that simulates S2 API responses for unit testing
2. **Test Helpers**: Reusable setup functions for creating test clients and servers
3. **Fixture-Based Integration Tests**: Support for three test modes:
   - `fixture` (default): Uses mock server with predefined responses
   - `live`: Connects to real S2 system using environment variables
   - `record`: Connects to real S2 system and saves responses as fixtures
4. **Crypto-Themed Test Data**: All test data uses crypto-themed names and numbers (3, 5, 7):
   - Anthony Dardano (REGO Master)
   - Satoshi Nakamoto (Bitcoin Creator)
   - Portal/Reader IDs: 357, 753, 573

### Running Tests

```bash
# Run all S2 tests (default fixture mode)
go test ./pkg/internal/tests/lenel_s2/... -v

# Run tests against live S2 system
REGO_TEST_MODE=live S2_URL=https://your-s2-system.com go test ./pkg/internal/tests/lenel_s2/... -v

# Record fixtures from live system
REGO_TEST_MODE=record S2_URL=https://your-s2-system.com go test ./pkg/internal/tests/lenel_s2/... -v

# Run specific test
go test ./pkg/internal/tests/lenel_s2/events_test.go -v
```

### Integration Test Pattern

Tests automatically detect the test mode and use appropriate setup:

```go
func TestGetPerson(t *testing.T) {
    var client *lenel_s2.Client
    var cleanup func()

    // Check if we should use integration testing
    if mode := os.Getenv("REGO_TEST_MODE"); mode == "live" || mode == "record" {
        it := NewIntegrationTest(t, "people")
        if err := it.Setup(); err != nil {
            t.Skip("Skipping integration test:", err)
        }
        client = it.Client
        cleanup = it.Cleanup
    } else {
        // Use mock server for fixture mode (default)
        th := SetupTestServer(t)
        client = lenel_s2.NewClient(th.server.URL, log.INFO)
        cleanup = th.Cleanup
    }
    defer cleanup()

    // Test implementation...
}
```

Fixtures are saved as XML files in `testdata/fixtures/{test_name}/{command}.xml`
