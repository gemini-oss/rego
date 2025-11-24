# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `cache` package provides a secure, thread-safe caching solution with AES-GCM encryption, LRU eviction, time-based expiration, and optional disk persistence. It's designed for caching API responses and other sensitive data with automatic deduplication.

**Key Features:**
- AES-256-GCM encryption for all cached data
- GZIP compression before encryption
- SHA-256 based content deduplication
- LRU eviction with O(1) operations
- Sliding window TTL (1-minute extensions on access)
- Thread-safe with RWMutex
- GOB serialization for Go types
- File-based persistence with atomic writes

## Architecture

### Core Components

1. **Cache Structure**:
   - LRU tracking with access list and map
   - Encrypted data storage with SHA-256 deduplication
   - Thread-safe operations with sync.RWMutex
   - Flexible in-memory or disk-persistent modes

2. **Key Features**:
   - AES-GCM 256-bit encryption for all data
   - Gzip compression before encryption
   - Content-based deduplication
   - Sliding window expiration (1-minute extensions)
   - GOB serialization for Go types

### Constructor Options

```go
// Flexible constructor accepts various types:
cache, _ := cache.NewCache(
    encryptionKey,    // []byte: 32-128 chars required
    "cache_file.gob", // string: auto-prepends temp dir
    false,            // bool: true for in-memory only
    1000000,          // int: max items (default 1000)
)

// Common service patterns:
// Google: NewCache(key, "rego_cache_google.gob", false, 1000000)
// Okta:   NewCache(key, "rego_cache_okta.gob", false, 1000000)
// Jamf:   NewCache(key, "rego_cache_jamf.gob", false, 1000000)
```

### Integration with requests Package

Most services don't use cache directly but through `requests.Client`:
```go
// In requests.Config
config := &requests.Config{
    EnableCache: true,
    CacheTTL:    30 * time.Minute,
}

// Cache is created automatically with service-specific filename
```

## Development Tasks

### Common Operations

1. **Basic Usage**:
   ```go
   // Store data with TTL
   err := cache.Set("key", data, 5*time.Minute)
   if err != nil {
       // Handle serialization/encryption/disk errors
   }

   // Retrieve data
   if value, found := cache.Get("key"); found {
       var result MyType
       err := json.Unmarshal(value, &result)
   }
   // Note: Get returns (nil, false) for any error
   ```

2. **Cache Control**:
   ```go
   cache.Enabled = false     // Disable caching
   cache.Clear()            // Remove all entries
   cache.Delete("key")      // Remove specific entry
   ```

3. **Persistence Patterns**:
   - **In-memory**: Fast, no disk I/O, data lost on restart
   - **Disk mode**: Writes entire cache on each Set(), survives restarts
   - **File location**: `os.TempDir() + filename` (e.g., `/tmp/rego_cache_google.gob`)

4. **Service Cache Keys**:
   ```go
   // Common patterns from actual usage:
   // Google: "{endpoint}_{method}_{params}"
   // Okta:   Full URL including query params
   // Jamf:   "{url}_{method}_{body}"
   // AD:     "rego_ad_{operation}_{baseDN}"
   ```

## Important Notes

- **Encryption Key**: Must be 32-128 characters (uses `crypt.ValidPassphrase()`)
- **Serialization**: GOB format - avoid channels, functions, or unexported fields
- **Disk Performance**: Every Set() writes entire cache to disk (expensive for large caches)
- **TTL Behavior**:
  - Entries get 1-minute extension on each Get() access
  - Expired entries are lazily deleted on next access
  - No background cleanup goroutine
- **Deduplication**: SHA-256 hash identifies identical values to save space
- **File Permissions**: Cache files created with 0600 (owner read/write only)
- **LRU Implementation**: Custom implementation with O(1) access/update/eviction

### Thread Safety

- All operations protected by `sync.RWMutex`
- Safe for concurrent reads and writes
- LRU updates are atomic within lock
- No deadlock risks in current implementation

## Common Pitfalls

1. **Performance**:
   - Disk mode writes entire cache on each Set()
   - Large caches can cause significant I/O
   - Consider in-memory mode for high-frequency updates

2. **Serialization Issues**:
   - GOB can't handle: channels, functions, sync types
   - Unexported fields are ignored
   - Interface{} values need concrete types

3. **Key Collisions**:
   - Use service prefixes: `google_users_list`
   - Include parameters: `okta_user_123456`
   - Avoid generic keys: `data`, `result`

4. **Memory Management**:
   - Only count-based eviction (not size-based)
   - Default 1000 items may be too low
   - Monitor memory usage with large objects

5. **Error Handling**:
   ```go
   // Get() returns false for ALL errors:
   // - Key not found
   // - Entry expired
   // - Decryption failed
   // - Deserialization failed
   value, found := cache.Get("key")
   if !found {
       // Could be any of the above
   }
   ```

6. **Encryption Key Loss**:
   - Lost key = unrecoverable cache
   - Different keys = cache mismatch
   - Store key securely (e.g., environment variable)

## Usage Patterns in Rego

### Service-Specific Pattern Examples

| Service | Cache File | Default TTL | Key Pattern |
|---------|------------|-------------|-------------|
| Google | `rego_cache_google.gob` | 30 min | `{service_domain}_{function_name_derivative}_{params}` or URL with parameters  |
| Okta | `rego_cache_okta.gob` | 5-60 min | URL with parameters |
| Jamf | `rego_cache_jamf.gob` | 5 min | URL with parameters |
| Active Directory | `rego_cache_active_directory.gob` | 30 min | `rego_ad_{operation}` |
| SnipeIT | `rego_cache_snipeit.gob` | 5 min | URL with parameters |
| Backupify | `rego_cache_backupify_{appType}.gob` | 5m-24h | `{endpoint}_{appType}` |
| LenelS2 | `rego_cache_lenels2.gob` | 5 min | `{url}_{command}_{params}` |

### Common Use Cases

1. **API Response Caching**:
   - List operations (users, devices, groups)
   - Expensive queries with stable results
   - Rate limit mitigation

2. **Authentication Tokens**:
   - OAuth2 access tokens
   - Session IDs
   - JWT tokens (though often handled by auth library)

3. **Computed Results**:
   - Role permission matrices
   - Aggregated reports
   - Transformed data structures

## Troubleshooting

### Cache Inspection
```go
// Check if caching is working
log.Printf("Cache enabled: %v", cache.Enabled)
log.Printf("Cache size: %d", len(cache.data))
```

### Common Issues

1. **Permission Denied**: Check temp directory permissions
2. **Disk Full**: Monitor temp directory space
3. **Wrong Key**: Verify REGO_ENCRYPTION_KEY matches
4. **Stale Data**: Check TTL settings
5. **Memory Growth**: Monitor max items setting
