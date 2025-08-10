# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `errors` package provides a custom error type (`CustomError`) that extends Go's standard errors with numeric error codes. However, this package is currently **not actively used** in the Rego codebase, which predominantly uses standard Go error handling.

**Status: Unused/Legacy**
- No production code imports this package
- Empty test file indicates it was never fully implemented
- Rego services use their own error handling patterns

## Architecture

### Core Components

1. **CustomError Type**:
   ```go
   type CustomError struct {
       Code int    // Numeric error code
       Err  error  // Underlying error
   }
   ```

2. **Constructor**:
   - `New(code int, msg string) error`: Creates a new CustomError

3. **Error Interface**:
   - `Error() string`: Returns formatted string "code=%d, error=%v"

## Current Status

- **Defined but unused**: The package exists but isn't imported by any production code
- **Empty tests**: Test file exists but contains no tests
- **No error constants**: No predefined error codes

## Actual Error Handling Patterns in Rego

The Rego codebase uses several established patterns instead of this package:

### 1. **Standard Error Wrapping** (Most Common)
```go
// Always use %w to preserve error chain
return fmt.Errorf("failed to fetch user %s: %w", userID, err)

// Check wrapped errors
if errors.Is(err, ErrNotFound) {
    // Handle specific error
}
```

### 2. **Service-Specific Error Types**

**HTTP/Request Errors** (`requests.RequestError`):
```go
type RequestError struct {
    StatusCode  int    `json:"status_code"`
    Method      string `json:"method"`
    URL         string `json:"url"`
    Message     string `json:"message"`
    RawResponse string `json:"raw_response"`
}

// Usage
if reqErr, ok := err.(*requests.RequestError); ok {
    if reqErr.StatusCode == 404 {
        return nil, fmt.Errorf("resource not found")
    }
}
```

**API-Specific Errors**:
- `okta.Error` - Includes error causes array
- `slack.Error` - Maps error codes to descriptions
- `crypt.PassphraseError` - Lists validation issues
- `jamf.DataMigrationError` - Migration-specific errors

### 3. **Error Joining** (Go 1.20+)
```go
// Used in active_directory for multiple operations
errors.Join(err1, err2, err3)
```

### 4. **Fatal vs Recoverable Errors**
```go
// Fatal - for missing configuration
if token == "" {
    log.Fatal("API_TOKEN not set")
}

// Recoverable - return error for handling
if err != nil {
    return nil, fmt.Errorf("failed to connect: %w", err)
}
```

## Development Guidance

### If Using This Package

1. Define meaningful error codes:
   ```go
   const (
       ErrCodeValidation = 1001
       ErrCodeNetwork    = 2001
       ErrCodeAuth       = 3001
   )
   ```

2. Type assertion for code access:
   ```go
   if customErr, ok := err.(*errors.CustomError); ok {
       switch customErr.Code {
       case ErrCodeAuth:
           // Handle auth error
       }
   }
   ```

## Best Practices for Error Handling in Rego

### 1. **Always Wrap with Context**
```go
// Good - includes operation and ID
return fmt.Errorf("failed to get user %s: %w", userID, err)

// Bad - no context
return err
```

### 2. **Service-Specific Error Types**
When building a new service, consider:
```go
// Define service-specific error when needed
type ServiceError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

// Implement Error() interface
func (e *ServiceError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
```

### 3. **HTTP Error Handling Pattern**
```go
// From requests package
func handleHTTPError(resp *http.Response) error {
    return &RequestError{
        StatusCode: resp.StatusCode,
        Method:     resp.Request.Method,
        URL:        resp.Request.URL.String(),
        Message:    http.StatusText(resp.StatusCode),
    }
}
```

### 4. **Configuration Validation**
```go
// Early validation in constructors
func NewClient(config *Config) (*Client, error) {
    if config.APIKey == "" {
        return nil, errors.New("API key is required")
    }
    // Continue initialization
}
```

## Important Notes

- **Do Not Use This Package For Now Until It's Expanded Further**: It's unused and doesn't follow Rego patterns, but is intended for future error code standardization
- **Follow Established Patterns**: Use standard error wrapping with `fmt.Errorf`
- **Service-Specific Errors**: Define custom types only when needed
- **No External Dependencies**: Rego avoids third-party packages as much as possible

## Common Error Scenarios in Rego

### Authentication Failures
```go
// Pattern used across services
if resp.StatusCode == 401 {
    return fmt.Errorf("authentication failed: check credentials")
}
```

### Rate Limiting
```go
// From various services
if resp.StatusCode == 429 {
    return fmt.Errorf("rate limit exceeded")
}
```

### Not Found
```go
// Common pattern
if resp.StatusCode == 404 {
    return nil, nil // Return nil for optional resources
    // OR
    return fmt.Errorf("resource not found: %s", resourceID)
}
```

### Validation Errors
```go
// From crypt package
type PassphraseError struct {
    Issues []string
}

func (e *PassphraseError) Error() string {
    return fmt.Sprintf("passphrase validation failed: %s",
        strings.Join(e.Issues, ", "))
}
```

## Recommendation

**For New Development**:
- Use standard Go error handling with `fmt.Errorf` and `%w`
- Define service-specific error types only when API requires it
- Follow patterns from `requests`, `okta`, or `slack` packages
- Consider removing this package in future cleanup
