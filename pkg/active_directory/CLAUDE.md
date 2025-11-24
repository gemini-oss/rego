# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `active_directory` package provides a Go client library for interacting with Microsoft Active Directory via LDAP protocol. It offers comprehensive user and group management capabilities with built-in caching, TLS support, and efficient parallel processing.

**Key Features:**
- LDAP and LDAPS support with automatic TLS configuration
- Generic query interface with type-safe operations
- Built-in caching with 30-minute TTL
- Parallel LDAP entry unmarshaling using worker pools
- Pre-defined LDAP filters for common queries
- Support for nested group membership queries
- Automatic time format conversion (Windows FILETIME and LDAP GeneralizedTime)
- Comprehensive attribute constants to prevent typos
- Support for binary attributes (objectGUID, objectSID)

## Architecture

### Core Components

1. **Client** (`active_directory.go`):
   - Main client struct managing LDAP connections, logging, and caching
   - Key methods: `NewClient()`, `Connect()`, `Bind()`, `Close()`
   - Generic request handler: `do[T Slice[E], E any]()`
   - Parallel unmarshaling with worker pools (defaults to `runtime.NumCPU()`)
   - Automatic connection establishment during `NewClient()`

2. **Entities** (`entities.go`):
   - Core AD object types: `User`, `Group`, `Computer`, `OrganizationalUnit`
   - LDAP attribute constants as typed `Attribute` values
   - Time conversion utilities for Windows FILETIME and LDAP GeneralizedTime
   - User Account Control (UAC) flag constants
   - SAMAccountType enum values
   - Default and minimal attribute sets for performance optimization

3. **Query Builder** (`query.go`):
   - `LDAPQuery` struct with fluent interface for query construction
   - LDAP filter constants for common queries
   - LDAP matching rules for advanced queries (bitwise operations, chain matching)
   - Pre-defined filters: `FILTER_USER_ACTIVE`, `FILTER_USER_ADMIN`, `FILTER_USER_DISABLED`, etc.

4. **API Implementation**:
   - `users.go`: User operations (ListAllUsers, ActiveUsers, DisabledUsers, MemberOf, etc.)
   - `groups.go`: Group operations (ListAllGroups)

### Key Design Patterns

- **Generic Slice Interface**: `Slice[T any]` allows polymorphic handling of result sets
- **Caching**: 30-minute default cache duration for all queries (cache file: `rego_cache_active_directory.gob`)
- **Parallel Processing**: Configurable worker pools for LDAP entry unmarshaling
- **Type Safety**: Strongly typed `Attribute` constants prevent typos in LDAP queries
- **Fluent Interface**: Query builder pattern for complex LDAP queries
- **Binary Attribute Handling**: Automatic handling of objectGUID, objectSID as binary data
- **Paged Search**: Automatic paging with configurable size (default 1000)

## Development Tasks

### Running Tests
```bash
# Run tests for the active_directory package
go test ./pkg/internal/tests/active_directory/...

# Run with verbose output
go test -v ./pkg/internal/tests/active_directory/...
```

### Common Operations

1. **Client Initialization**:
   ```go
   import "github.com/gemini-oss/rego/pkg/active_directory"

   // Create client (reads config from environment)
   client := active_directory.NewClient(0) // 0 = default log level
   defer client.Close()
   ```

2. **User Operations**:
   ```go
   // List all users
   users, err := client.ListAllUsers()

   // Get active users only
   activeUsers, err := client.ActiveUsers()

   // Get users in a specific group (supports nested groups)
   groupMembers, err := client.MemberOf("CN=Admins,OU=Groups,DC=example,DC=com")

   // Get locked out users
   lockedUsers, err := client.LockedUsers()
   ```

3. **Using the Query Builder**:
   ```go
   query := active_directory.NewLDAPQuery(
       "DC=example,DC=com",
       "(&(objectClass=user)(department=Sales))",
       active_directory.DefaultUserAttributes,
   ).SetScope(ldap.ScopeWholeSubtree).
     SetPagingSize(500).
     SetSizeLimit(0)

   var users active_directory.Users
   err := client.Do(query, &users)
   ```

4. **Working with LDAP Attributes**:
   ```go
   // Use predefined constants
   attrs := []string{
       active_directory.SAMAccountName.String(),
       active_directory.Mail.String(),
       active_directory.DisplayName.String(),
   }

   // Use minimal attributes for performance
   query.SetAttributes(active_directory.MinimalUserAttributes)
   ```

5. **Adding New Queries**:
   - Define filter constant in `query.go`
   - Add method to appropriate file (users.go, groups.go, etc.)
   - Use generic `do()` method with proper type parameters
   - Follow caching pattern with meaningful cache keys

6. **Adding New AD Object Types**:
   - Define struct in `entities.go` with LDAP tags
   - Implement slice type (e.g., `Computers []Computer`)
   - Implement `Append()` method for the slice type
   - Add relevant attribute constants

### Environment Variables

- `AD_LDAP_SERVER`: LDAP server hostname (required)
- `AD_PORT`: LDAP server port (optional, defaults to 389 for LDAP, 636 for LDAPS)
- `AD_BASE_DN`: Base distinguished name (optional, shows warning if not set)
- `AD_USERNAME`: Bind username for authentication (required)
- `AD_PASSWORD`: Bind password for authentication (required)
- `AD_LDAP_CA`: LDAP CA certificate in PEM format (optional, for custom CAs)
- `REGO_ENCRYPTION_KEY`: Key for cache encryption (required)

## Important Notes

- **TLS Configuration**:
  - Port 389: Uses StartTLS for encryption
  - Port 636: Native LDAPS connection
  - Custom CA certificates supported via `AD_LDAP_CA`
  - Note: Currently uses `InsecureSkipVerify: true` for TLS
- **Time Handling**:
  - Automatic conversion from Windows FILETIME (100-ns ticks since 1601-01-01)
  - Support for LDAP GeneralizedTime format
  - Zero time values handled gracefully
- **Attribute Handling**:
  - Single and multi-valued attributes supported
  - Binary attributes (objectGUID, objectSID) handled automatically
  - Use attribute constants to prevent typos
- **Performance**:
  - Worker count defaults to `runtime.NumCPU()` for parallel unmarshaling
  - Default search size limit: 1,000,000 entries
  - Paging size: 1000 entries per page
- **Cache Patterns**:
  - User queries: `rego_ad_{operation}_{baseDN}`
  - Group membership: `rego_memberof_{groupDN}`
  - General: `ad_{operation}_{parameters}`

## Available Methods

### User Operations
- `ListAllAdmins()` - Get all admin users (UAC flag check)
- `ListAllUsers()` - Get all user objects
- `ActiveUsers()` - Get enabled user accounts
- `DisabledUsers()` - Get disabled user accounts
- `LockedUsers()` - Get locked out users
- `PasswordNeverExpiresUsers()` - Get users with non-expiring passwords
- `MemberOf(groupDN)` - Get users in a specific group (supports nested groups via LDAP_MATCHING_RULE_IN_CHAIN)

### Group Operations
- `ListAllGroups()` - Get all group objects

### Generic Operations
- `do[T, E](query)` - Execute custom LDAP query with type safety

## LDAP Filter Constants

### User Filters
- `FILTER_USER_ACTIVE` - Active user accounts
- `FILTER_USER_ADMIN` - Admin users (UAC flag 0x00000002)
- `FILTER_USER_DISABLED` - Disabled accounts
- `FILTER_USER_LOCKED` - Locked out users
- `FILTER_USER_PASSWORD_NEVER_EXPIRES` - Non-expiring passwords
- `FILTER_USER_NESTED_GROUP` - Template for nested group queries

### LDAP Matching Rules
- `LDAP_MATCHING_RULE_BIT_AND` (1.2.840.113556.1.4.803) - Bitwise AND
- `LDAP_MATCHING_RULE_BIT_OR` (1.2.840.113556.1.4.804) - Bitwise OR
- `LDAP_MATCHING_RULE_IN_CHAIN` (1.2.840.113556.1.4.1941) - Nested group membership
- `LDAP_MATCHING_RULE_DN_WITH_DATA` (1.2.840.113556.1.4.2253) - DN with binary

## Common Pitfalls

1. **LDAP Escaping**: Always escape special characters in LDAP filters (TODO: escaping function not implemented)
2. **Time Formats**: AD uses various time formats - the package handles conversion automatically
3. **Large Result Sets**: Default limit is 1M entries; use paging for very large queries
4. **Attribute Names**: Always use the predefined `Attribute` constants to avoid typos
5. **Connection Lifecycle**: Remember to call `Close()` when done with the client
6. **Environment Variables**: The package uses different env var names than other services (e.g., `AD_LDAP_SERVER` not `AD_HOST`)
7. **Binary Attributes**: objectGUID and objectSID are returned as binary data - handle accordingly

## Troubleshooting

### Connection Issues
- Verify `AD_LDAP_SERVER` is set correctly (hostname only, no protocol)
- Check `AD_PORT` matches your AD configuration (389 for LDAP, 636 for LDAPS)
- Ensure `AD_USERNAME` and `AD_PASSWORD` are valid bind credentials
- For custom CAs, provide the certificate via `AD_LDAP_CA`

### Query Issues
- Enable debug logging to see LDAP queries: `NewClient(4)` for debug level
- Check `AD_BASE_DN` is set correctly for your domain
- Verify LDAP filter syntax using an LDAP browser first
- For nested groups, ensure you're using the `LDAP_MATCHING_RULE_IN_CHAIN`

### Performance Issues
- Use `MinimalUserAttributes` instead of `DefaultUserAttributes` when possible
- Implement specific filters rather than filtering in Go code
- Consider increasing page size for large result sets
- Cache is enabled by default with 30-minute TTL
