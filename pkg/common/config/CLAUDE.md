# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `config` package provides a simple, centralized abstraction for accessing environment variables throughout the Rego application. It follows the Twelve-Factor App principle of storing configuration in the environment.

**Key Features:**
- Minimal abstraction over os.LookupEnv()
- Type conversion utilities (string, int)
- Thread-safe environment variable access
- No caching - reads fresh values each time
- Zero dependencies on other packages

## Architecture

### Core Functions

1. **`GetEnv(key string) string`**:
   - Safely retrieves environment variables
   - Returns empty string if variable doesn't exist
   - Most commonly used function in the package

2. **`GetEnvAsInt(key string) int`**:
   - Retrieves environment variables as integers
   - Returns 0 if variable doesn't exist or isn't numeric
   - Currently unused but available for future use

### Design Philosophy

- **Minimal abstraction**: Thin wrapper over `os.LookupEnv()`
- **No defaults**: Returns empty/zero values, not defaults
- **Caller responsibility**: Services handle missing values
- **Consistent behavior**: All variables treated the same
- **Performance**: No caching, direct OS calls
- **Security**: Never logs sensitive values

## Development Tasks

### Common Usage Patterns

1. **Mandatory Configuration**:
   ```go
   // Pattern 1: Fatal for missing required vars
   token := config.GetEnv("API_TOKEN")
   if len(token) == 0 {
       log.Fatal("API_TOKEN is not set")
   }

   // Pattern 2: Return error for library functions
   if token == "" {
       return nil, fmt.Errorf("SNIPEIT_TOKEN not set")
   }
   ```

2. **Default Values**:
   ```go
   // Common pattern throughout Rego
   baseURL := config.GetEnv("CUSTOM_BASE_URL")
   if baseURL == "" {
       baseURL = "https://default.example.com"
   }

   // Ternary-like pattern
   port := config.GetEnv("AD_PORT")
   if port == "" {
       port = "389"
   }
   ```

3. **Boolean Flags**:
   ```go
   // Standard pattern for booleans
   useSandbox := config.GetEnv("USE_SANDBOX") == "true"

   // Case-insensitive variant (not in current code)
   // useSandbox := strings.ToLower(config.GetEnv("USE_SANDBOX")) == "true"
   ```

4. **Dynamic Variable Selection**:
   ```go
   // Okta's sandbox pattern
   cfg := struct{
       orgName string
       token   string
   }{
       orgName: "OKTA_ORG_NAME",
       token:   "OKTA_API_TOKEN",
   }

   if config.GetEnv("OKTA_USE_SANDBOX") == "true" {
       cfg.orgName = "OKTA_SANDBOX_ORG_NAME"
       cfg.token = "OKTA_SANDBOX_API_TOKEN"
   }

   actualOrgName := config.GetEnv(cfg.orgName)
   ```

5. **Integer Configuration**:
   ```go
   // Using GetEnvAsInt (currently unused in codebase)
   port := config.GetEnvAsInt("SERVER_PORT")
   if port == 0 {
       port = 8080 // default
   }
   ```

### Service Integration Examples

**Google Service Pattern**:
```go
// From google.go - CICD mode with base64 decoding
if cfg.CICD {
    if cfg.Credentials.Type == SERVICE_ACCOUNT {
        data, _ := base64.StdEncoding.DecodeString(
            config.GetEnv("GOOGLE_SERVICE_ACCOUNT"),
        )
    }
}
```

**Active Directory Pattern**:
```go
// From active_directory.go - with defaults and warnings
baseURL := config.GetEnv("AD_LDAP_SERVER")
if baseURL == "" {
    return nil, errors.New("AD_LDAP_SERVER not set")
}

port := config.GetEnv("AD_PORT")
if port == "" {
    port = "389" // default LDAP port
}
```

## Important Notes

- **Early Validation**: Check required vars in `NewClient()` constructors
- **Error Messages**: Be specific about which variable is missing
- **Documentation**: List all env vars in service CLAUDE.md files
- **Simplicity**: Package intentionally minimal - no validation/defaults
- **Performance**: Each call reads from OS (no caching)
- **Thread Safety**: `os.LookupEnv()` is thread-safe
- **Security**: Never log tokens or passwords

### Potential Future Enhancements

Based on usage patterns, these functions would be useful:

```go
// GetEnvWithDefault - Common pattern in services
func GetEnvWithDefault(key, defaultValue string) string {
    if value := GetEnv(key); value != "" {
        return value
    }
    return defaultValue
}

// GetEnvAsBool - Standardize boolean parsing
func GetEnvAsBool(key string) bool {
    return strings.ToLower(GetEnv(key)) == "true"
}

// MustGetEnv - Panic if not set (for required vars)
func MustGetEnv(key string) string {
    value := GetEnv(key)
    if value == "" {
        panic(fmt.Sprintf("%s not set", key))
    }
    return value
}
```

## Common Environment Variables in Rego

### Global Variables
| Variable | Description | Used By |
|----------|-------------|---------|
| `REGO_ENCRYPTION_KEY` | 32-byte cache encryption key | All services with caching |
| `REGO_LOG_LEVEL` | Logging verbosity | All services |
| `REGO_TEST_MODE` | Test mode (fixture/live/record) | Test infrastructure |

### Naming Conventions

**Standard Pattern**: `{SERVICE}_{VARIABLE}`
- `{SERVICE}_URL` - Base URL
- `{SERVICE}_API_TOKEN` - API authentication
- `{SERVICE}_USERNAME` - Username for auth
- `{SERVICE}_PASSWORD` - Password for auth

**Sandbox Pattern**: `{SERVICE}_SANDBOX_{VARIABLE}`
- Used with `{SERVICE}_USE_SANDBOX=true`
- Allows dual environment configuration

**Special Cases**:
- Google: `GOOGLE_CREDENTIALS`, `GOOGLE_IMPERSONATE_USER`
- Jamf: Uses `JSS_` prefix (legacy)
- Active Directory: Uses `AD_` prefix
- LenelS2: Uses `S2_` prefix

### Service-Specific Examples

```bash
# Google Workspace
GOOGLE_CREDENTIALS=/path/to/creds.json
GOOGLE_IMPERSONATE_USER=admin@example.com
GOOGLE_DOMAIN=example.com

# Okta with Sandbox
OKTA_ORG_NAME=mycompany
OKTA_API_TOKEN=00abc...
OKTA_USE_SANDBOX=true
OKTA_SANDBOX_ORG_NAME=mycompany-sandbox
OKTA_SANDBOX_API_TOKEN=00xyz...

# Active Directory
AD_LDAP_SERVER=dc.example.com
AD_PORT=636
AD_BASE_DN=DC=example,DC=com
AD_USERNAME=bind@example.com
AD_PASSWORD=secure-password
```

## Testing Considerations

### Unit Testing
```go
// Set test environment
os.Setenv("TEST_VAR", "value")
defer os.Unsetenv("TEST_VAR")

// Test with missing variable
os.Unsetenv("MISSING_VAR")
assert.Equal(t, "", config.GetEnv("MISSING_VAR"))

// Test integer parsing
os.Setenv("PORT", "8080")
assert.Equal(t, 8080, config.GetEnvAsInt("PORT"))

// Test invalid integer
os.Setenv("INVALID", "not-a-number")
assert.Equal(t, 0, config.GetEnvAsInt("INVALID"))
```

### Integration Testing
```go
// Save and restore environment
original := os.Environ()
defer func() {
    os.Clearenv()
    for _, env := range original {
        parts := strings.SplitN(env, "=", 2)
        os.Setenv(parts[0], parts[1])
    }
}()
```

### Docker/Container Testing
```yaml
# docker-compose.yml
services:
  app:
    environment:
      - REGO_ENCRYPTION_KEY=test-key-32-characters-long!!!!
      - SERVICE_URL=http://localhost:8080
    env_file:
      - .env.test
```

## Security Considerations

1. **Never Log Sensitive Values**:
   ```go
   // BAD
   log.Printf("Token: %s", config.GetEnv("API_TOKEN"))

   // GOOD
   log.Printf("Token is set: %v", config.GetEnv("API_TOKEN") != "")
   ```

2. **Validate Early**:
   ```go
   func NewClient() (*Client, error) {
       token := config.GetEnv("API_TOKEN")
       if token == "" {
           return nil, errors.New("API_TOKEN required")
       }
       // Continue initialization
   }
   ```

3. **Use Descriptive Names**:
   - Good: `OKTA_API_TOKEN`
   - Bad: `TOKEN`, `KEY`
