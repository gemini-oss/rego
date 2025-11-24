# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `auth` package provides minimal authentication abstractions for OAuth2 and JWT configurations. It serves as a thin wrapper around Google's OAuth2 library types, offering consistent type aliases used across the Rego project for API authentication.

**Key Features:**
- Type aliases for OAuth2 and JWT configurations
- No implementation logic - relies on underlying `golang.org/x/oauth2` library
- Automatic token refresh handled by OAuth2 library
- Context-aware authentication
- Support for service account impersonation

## Architecture

### Core Components

1. **Type Aliases**:
   - `JWTConfig`: Alias for `*jwt.Config` from golang.org/x/oauth2/jwt
   - `OAuthConfig`: Alias for `*oauth2.Config` from golang.org/x/oauth2

### Usage Pattern

This package doesn't implement authentication logic directly. Instead, it provides common types that consuming packages use to implement their specific authentication needs.

```go
// Service Account JWT example (used by Google package)
jwtConfig := auth.JWTConfig(&jwt.Config{
    Email:      "service-account@project.iam.gserviceaccount.com",
    PrivateKey: privateKey,
    Scopes:     []string{"https://www.googleapis.com/auth/admin.directory.user"},
    Subject:    "admin@example.com", // For domain-wide delegation
})

// Get token source with automatic refresh
tokenSource := jwtConfig.TokenSource(context.Background())
token, err := tokenSource.Token()

// OAuth2 example
oauthConfig := auth.OAuthConfig(&oauth2.Config{
    ClientID:     "client-id",
    ClientSecret: "client-secret",
    Scopes:       []string{"https://www.googleapis.com/auth/drive"},
    Endpoint:     google.Endpoint,
})
```

## Development Tasks

### Current Usage

- **Google Package**: Primary consumer, implements OAuth2, JWT, and API key authentication
- Other services implement their own authentication without using this package

### Integration with Other Packages

1. **Google Package Integration**:
   ```go
   // From google.go
   jwtConfig, err := google.JWTConfigFromJSON(credentialsJSON, scopes...)
   jwtConfig.Subject = impersonateUser // Set impersonation
   httpClient := jwtConfig.Client(ctx) // HTTP client with auth
   ```

2. **Requests Package Integration**:
   - Auth package provides token source
   - Requests package adds Bearer token to Authorization header
   - Automatic token refresh happens transparently

### Adding Authentication Support

When implementing authentication in a new service:
1. Consider if OAuth2/JWT patterns fit your needs
2. Use these type aliases for consistency if applicable
3. Otherwise, implement service-specific authentication as needed

### Token Management

**Automatic Token Refresh**:
- The underlying OAuth2 library handles token expiration
- `TokenSource()` returns a source that automatically refreshes
- No manual refresh logic needed in consuming code

**Context Support**:
- All token operations support context for cancellation
- Use `context.WithTimeout()` for token fetch timeouts

## Important Notes

- **Abstraction Layer**: This package provides types, not implementation
- **Automatic Token Refresh**: Handled by the underlying OAuth2 library
- **Thread-Safe**: Token sources are safe for concurrent use
- **Context-Aware**: All operations support Go contexts
- **Service Account Features**:
  - Domain-wide delegation via `Subject` field
  - Automatic JWT generation and signing
  - Token caching and refresh
- **Most Rego Services**: Use simpler token-based auth (Bearer tokens)

### Environment Variables

While this package doesn't read environment variables directly, common patterns include:

| Service | Variable | Usage |
|---------|----------|-------|
| Google | `GOOGLE_CREDENTIALS` | Service account JSON |
| Google | `GOOGLE_IMPERSONATE_USER` | User for delegation |
| Google | `GOOGLE_API_KEY` | Simple API key auth |

## Common Authentication Patterns in Rego

| Pattern | Services | Implementation |
|---------|----------|----------------|
| **Bearer Token** | Okta, Slack, SnipeIT, LenelS2, Atlassian, Backupify | `Authorization: Bearer {token}` |
| **Basic Auth → JWT** | Jamf | Exchange credentials for JWT token |
| **OAuth2/JWT** | Google | Via this package with auto-refresh |
| **API Key** | Various | Simple header or query parameter |
| **LDAP Bind** | Active Directory | Direct LDAP authentication |
| **Session-Based** | LenelS2 | Login creates session ID |

### Error Handling

Common authentication errors:
```go
// Token refresh failure
token, err := tokenSource.Token()
if err != nil {
    // Could be: invalid credentials, network error, expired refresh token
    return fmt.Errorf("failed to get token: %w", err)
}

// Insufficient scopes
// Check error response from API calls for scope-related errors
```

## Future Considerations

1. **Common Authentication Interface**:
   ```go
   type TokenProvider interface {
       Token(ctx context.Context) (string, error)
       Type() string // "Bearer", "Basic", etc.
   }
   ```

2. **Unified Token Caching**:
   - Integration with cache package
   - Persistent token storage
   - Cross-service token sharing

3. **Additional Auth Methods**:
   - mTLS support
   - SAML assertions
   - OAuth2 PKCE flow
   - API key rotation

4. **Helper Functions**:
   - Token validation
   - Scope checking
   - Expiry handling
   - Refresh token management
