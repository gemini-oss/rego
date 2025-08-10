# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Regolith (ReGo) is an open-source Go library that provides a unified abstraction layer for multiple REST APIs. Named from Greek rhegos (blanket) + lithos (rock), it acts as a foundational "blanket layer" over enterprise services. The project simplifies interactions with Google Workspace, Okta, Jamf, Active Directory, Slack, SnipeIT, Atlassian, LenelS2, Backupify, and more.

**Key Features:**
- Consistent API patterns across diverse services
- Built-in authentication management (OAuth2, JWT, API keys, LDAP)
- Automatic retry logic with exponential backoff
- Encrypted file-based caching with configurable TTL
- Rate limiting per service with configurable thresholds
- Generic pagination handling for all services
- Comprehensive error handling with context
- Progress tracking for long-running operations
- Method chaining for complex queries
- Concurrent operations support (where applicable)

## Common Development Commands

### Build & Run
```bash
make build                                      # Build the binary
go run main.go                                  # Run directly
```

### Testing
```bash
make test                                       # Run all tests
go test -v ./...                                # Run all tests with verbose output
go test -v ./pkg/google/...                     # Run tests for a specific package
go test -v -run TestFunctionName ./pkg/google/. # Run a specific test
go test -cover ./...                            # Run tests with coverage report
```

### Code Quality
```bash
make pretty                                     # Format code with gofmt
make update-copyright                           # Update copyright headers in all files
```

### Documentation
```bash
make docs                                       # Generate documentation using gomarkdoc
make server                                     # Start Hugo documentation server
./gen_hugo_index.sh                             # Regenerate Hugo structure and documentation
```

### Cleanup
```bash
make clean                                      # Clean build artifacts
make flush                                      # Clear cache and temporary files (rego_cache_*.gob)
```

## Architecture

### Design Philosophy

Rego follows these core design principles:

1. **Consistency Over Cleverness**: All services follow similar patterns even when the underlying APIs differ significantly
2. **Type Safety with Generics**: Heavy use of Go generics for type-safe operations while maintaining flexibility
3. **Environment-First Configuration**: Following Twelve-Factor App methodology for configuration
4. **Defensive Programming**: Required environment variables are checked at initialization, not at usage time
5. **Performance Optimization**: Built-in caching, rate limiting, and concurrent operations where appropriate
6. **Developer Experience**: Method chaining, intuitive naming, and consistent patterns across services
7. **Security by Default**: Mandatory encryption for cached data, secure credential handling
8. **Flexibility Through Composition**: Services can be composed together (see orchestrators package)
9. **Progressive Enhancement**: Services start with basic functionality and add advanced features as needed
10. **API-Specific Adaptations**: While maintaining consistency, each service adapts to its API's peculiarities

### Project Structure
```
rego/
├── main.go                          # Entry point (example usage)
├── go.mod                           # Go module definition
├── Makefile                         # Build and development commands
├── gen_hugo_index.sh                # Documentation generation script
├── pkg/
│   ├── common/                      # Shared utilities used by all services
│   │   ├── auth/                    # OAuth2 token management
│   │   ├── cache/                   # Encrypted file-based caching
│   │   ├── config/                  # Environment variable helpers
│   │   ├── crypt/                   # Encryption/decryption utilities
│   │   ├── errors/                  # Custom error types with context
│   │   ├── generics/                # Generic slice and map operations
│   │   ├── log/                     # Structured logging framework
│   │   ├── ratelimit/               # Rate limiting implementation
│   │   ├── requests/                # HTTP client wrapper with retry
│   │   ├── retry/                   # Exponential backoff retry logic
│   │   ├── server/                  # Server utilities and middleware
│   │   └── starstruct/              # Struct manipulation utilities
│   ├── [service_name]/              # Service-specific implementations
│   │   ├── CLAUDE.md                # Service-specific AI guidance
│   │   ├── [service_name].go        # Main client implementation
│   │   ├── entities.go              # Data structures and models
│   │   ├── admin.go                 # Admin operations (if applicable)
│   │   ├── users.go                 # User management (if applicable)
│   │   ├── devices.go               # Device management (if applicable)
│   │   └── *.json                   # Embedded configuration files
│   ├── internal/tests/              # Test files organized by service
│   └── orchestrators/               # Multi-service workflow examples
├── docs/                            # Hugo documentation site
└── .github/                         # GitHub Actions workflows
```

### Package Structure Details

#### Service Package Anatomy (`/pkg/[service_name]/`)
Every service package follows this consistent structure:
- **`[service_name].go`**: Main client implementation with `NewClient()` constructor
- **`entities.go`**: Struct definitions for API resources and responses
- **`admin.go`**: Administrative operations (user management, settings)
- **Domain-specific files**: `users.go`, `devices.go`, `groups.go`, etc.
- **`CLAUDE.md`**: Service-specific guidance and patterns
- **Embedded JSON files**: Configuration, schemas, or discovery documents

#### Common Package Details (`/pkg/common/`)
- **`auth/`**: OAuth2 flow management, token refresh, service account handling
- **`cache/`**: File-based cache with GZIP compression and AES encryption
- **`config/`**: Environment variable parsing with type conversion
- **`crypt/`**: AES-256-GCM encryption for sensitive data
- **`errors/`**: Context-aware error wrapping and custom error types
- **`generics/`**: JSON marshaling for polymorphic API responses
- **`log/`**: Structured logging with levels (Debug, Info, Warning, Error)
- **`ratelimit/`**: Token bucket algorithm with per-service limits
- **`requests/`**: HTTP client with automatic retry, headers, and content types
- **`retry/`**: Configurable retry policies with jitter
- **`server/`**: HTTP server utilities, middleware, health checks
- **`starstruct/`**: Dynamic struct field manipulation

### Key Design Patterns

#### 1. Client Pattern
Every service implements a consistent client pattern:
```go
// Standard client structure
type Client struct {
    BaseURL  string
    HTTP     *requests.Client
    Log      *log.Logger
    Cache    *cache.Cache
    Rate     *ratelimit.RateLimiter // optional
}

// Configuration structure
type Config struct {
    // Common fields
    EnableCache     bool
    CacheTTL        time.Duration
    RateLimit       int

    // Service-specific fields
    Domain          string
    APIKey          string
}

// Client initialization pattern
func NewClient(verbosity int, opts ...Option) *Client {
    log := log.NewLogger("service", verbosity)

    // Environment variable configuration
    apiKey := config.GetEnv("SERVICE_API_KEY", "")
    baseURL := config.GetEnv("SERVICE_URL", "https://api.service.com")

    // Check required configuration
    if apiKey == "" {
        log.Error("missing required environment variable SERVICE_API_KEY")
        return nil
    }

    // Initialize cache with encryption
    encryptionKey := config.GetEnv("REGO_ENCRYPTION_KEY", "")
    if encryptionKey == "" {
        log.Error("missing required REGO_ENCRYPTION_KEY")
        return nil
    }

    cache := cache.NewCache("rego_cache_service.gob", encryptionKey, 30*time.Minute, verbosity)

    // Configure HTTP client
    httpClient := requests.NewClient(
        requests.WithBaseURL(baseURL),
        requests.WithHeader("Authorization", "Bearer " + apiKey),
        requests.WithRateLimit(100),
    )

    return &Client{
        BaseURL: baseURL,
        HTTP:    httpClient,
        Log:     log,
        Cache:   cache,
    }
}
```

#### 2. Method Chaining (Fluent Interface)
Particularly used in services such as Jamf, Google, and SnipeIT:
```go
// Jamf example
devices, err := client.Devices()
    .Sections([]string{Section.General, Section.Hardware})
    .PageSize(100)
    .Sort("id:asc")
    .Filter("general.platform==\"Mac\"")
    .ListAllComputers()

// Google example
users, err := client.Users()
    .Domain("example.com")
    .Query("orgUnitPath=/Sales")
    .Projection("full")
    .List()

// SnipeIT example
lic := snipeit.NewLicense("Rego -- SnipeIT API").
    LicenseName("New License").
    LicenseEmail("satoshi.nakamoto@gemini.com").
    Maintained(true).
    Notes("This is a test license created with ReGo").
    OrderNumber("123456789").
    PurchaseOrder("123456789").
    PurchaseDate(time.Now().GoString()).
    ExpirationDate(time.Now().GoString()).
    CategoryID(1).
    Seats("100").
    ProductKey("qwerty-uiopas-dfghjk-lzxcvb-nm").
    PurchaseCost(1_000_000).
    Reassignable(true).
    Build()
```

#### 3. Generic Request Handling
All services implement a standardized generic `do` method:
```go
// Standard generic do method signature
func do[T any](c *Client, method string, url string, query any, data any) (T, error) {
    var result T
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()

	res, body, err := c.HTTP.DoRequest(ctx, method, url, query, data)
	if err != nil {
		return *new(T), err
	}

	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &result)
	if err != nil {
		return *new(T), fmt.Errorf("unmarshalling error: %w", err)
	}

    return result, nil
}

// Usage examples
var user User
user, err := do[User](client, "GET", "/users/123", nil, nil)

var users []User
users, err := do[[]User](client, "GET", "/users", query, nil)
```

#### 4. Generic Pagination
Services implement pagination using one of three standardized patterns:

**Pattern 1: Interface-based (Google style)**
```go
// Define response interface
type PaginatedResponse[T any] interface {
    GetItems() []T
    GetNextPageToken() string
}

// Implement pagination
func doPaginated[T PaginatedResponse[T], Q PageableQuery](c *Client, url string, query Q) (*T, error) {
	var r T
	results := r

	for {
		r, err := do[T](c, method, url, query, data)
		if err != nil {
			return nil, err
		}

		results = results.Append(r)

		pageToken = r.PageToken()
		if pageToken == "" {
			break
		}

		query.SetPageToken(pageToken)
	}

	return &results, nil
}
```

**Pattern 2: Slice-based Response Style**
```go
// For responses that are arrays
func doPaginated[T Slice[E], E any](c *Client, url string) ([]E, error) {
	var emptySlice T = make([]E, 0)
	results := PagedSlice[T, E]{
		Results:  &emptySlice,
		NextPage: &NextPage{},
	}

	for {
		res, body, err := c.HTTP.DoRequest(context.Background(), method, url, query, data)
		if err != nil {
			return nil, err
		}

		c.Log.Println("Response Status:", res.Status)
		c.Log.Debug("Response Body:", string(body))

		var page []E
		err = json.Unmarshal(body, &page)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling error: %w", err)
		}

		*results.Results = append(*results.Results, page...)

		url = results.NextPage(res.Header.Values("Link"))
		query = nil
		if url == "" {
			break
		}
	}

	return results.Results, nil
}
```

**Pattern 3: Concurrent Offset Style**
```go
// For APIs that support offsets for concurrent fetching
func doConcurrent[T any](c *Client, fetchFunc func(page int) ([]T, error), totalPages int) []T {
	// Fetch the first page to initialize the response and pagination details.
	results, err := do[T](c, method, url, query, data)
	if err != nil {
		return nil, err
	}

	// Init concurrency control
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex

	// Function to fetch each page concurrently.
	fetchPage := func(offset int) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()

		q := query.Copy()
		q.SetOffset(offset)
		q.SetLimit(limit)

		page, err := do[T](c, method, url, q, data)
		if err != nil {
			c.Log.Error("Error fetching page:", err)
			return
		}

		resultsMutex.Lock()
		results.Append(page.Elements())
		resultsMutex.Unlock()
	}

	// Start fetching remaining pages.
	for nextOffset := offset + limit; nextOffset < results.TotalCount(); nextOffset += limit {
		wg.Add(1)
		go fetchPage(nextOffset)
	}
	wg.Wait()

	return &results, nil
}
```

#### 5. Query Parameter Handling
Structured query types with validation:
```go
// Standard query interface
type Query interface {
    ValidateQuery() error
    IsEmpty() bool
}

// Pageable query interface
type PageableQuery interface {
    Query
    SetPageToken(token string)
    GetPageToken() string
}

// Implementation example
type UserQuery struct {
    Domain      string `url:"domain,omitempty"`
    MaxResults  int    `url:"maxResults,omitempty"`
    PageToken   string `url:"pageToken,omitempty"`
    OrderBy     string `url:"orderBy,omitempty"`
}

func (q *UserQuery) ValidateQuery() error {
    if q.MaxResults > 500 {
        return fmt.Errorf("maxResults cannot exceed 500")
    }
    return nil
}

func (q *UserQuery) IsEmpty() bool {
    return q.Domain == "" && q.MaxResults == 0 && q.OrderBy == ""
}

func (q *UserQuery) SetPageToken(token string) {
    q.PageToken = token
}
```

#### 6. URL Building Patterns
Consistent URL construction:
```go
// Standard URL builder
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
    url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
    for _, id := range identifiers {
        url = fmt.Sprintf("%s/%s", url, id)
    }
    return url
}

// Usage
url := c.BuildURL("/users", userID, "profile")
// Result: https://api.service.com/users/123/profile
```

#### 7. Embedded Resources
Services embed configuration files at compile time:
```go
//go:embed json/scopes.json
var scopesFS embed.FS

//go:embed json/endpoints.json
var endpointsFS embed.FS

// Load at initialization
func loadEmbeddedConfig() (*Config, error) {
    data, err := scopesFS.ReadFile("json/scopes.json")
    if err != nil {
        return nil, fmt.Errorf("failed to read embedded scopes: %w", err)
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse scopes: %w", err)
    }

    return &config, nil
}
```

#### 8. Error Handling
Consistent error handling with context:
```go
// Standard error types
type APIError struct {
    StatusCode int
    Message    string
    Details    map[string]interface{}
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

// Error wrapping pattern
func (c *Client) GetUser(userID string) (*User, error) {
    user, err := do[User](c, "GET", c.BuildURL("/users", userID), nil, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
    }
    return &user, nil
}

// Error checking
if apiErr, ok := err.(*APIError); ok {
    if apiErr.StatusCode == 404 {
        // Handle not found
    }
}
```

#### 9. Caching Strategy
Uniform caching implementation:
```go
// Cache key generation
func cacheKey(operation string, params ...string) string {
    key := fmt.Sprintf("%s:%s", operation, strings.Join(params, ":"))
    return key
}

// Cache-aware operations
func (c *Client) GetUserCached(userID string) (*User, error) {
    key := cacheKey("user", userID)

    // Try cache first
    var user User
    if c.Cache.Get(key, &user) {
        c.Log.Debug("cache hit for user %s", userID)
        return &user, nil
    }

    // Fetch from API
    fetchedUser, err := c.GetUser(userID)
    if err != nil {
        return nil, err
    }

    // Store in cache
    c.Cache.Set(key, fetchedUser, 30*time.Minute)
    return fetchedUser, nil
}

// Method chaining for cache control
func (c *Client) NoCache() *Client {
    c.useCache = false
    return c
}
```

#### 10. Service-Specific Client Organization
Organize related operations into sub-clients:
```go
// Main client provides access to sub-clients
type Client struct {
    *baseClient
    users   *UserClient
    devices *DeviceClient
    groups  *GroupClient
}

func (c *Client) Users() *UserClient {
    if c.users == nil {
        c.users = &UserClient{client: c.baseClient}
    }
    return c.users
}

// Sub-client implementation
type UserClient struct {
    client *baseClient
    // Method chaining state
    filter   string
    pageSize int
}

func (uc *UserClient) Filter(filter string) *UserClient {
    uc.filter = filter
    return uc
}

func (uc *UserClient) PageSize(size int) *UserClient {
    uc.pageSize = size
    return uc
}

func (uc *UserClient) List() ([]User, error) {
    // Use configured options
    query := UserQuery{
        Filter:   uc.filter,
        PageSize: uc.pageSize,
    }
    return doPaginated[UsersResponse](uc.client, "/users", query)
}
```

### Environment Variables

#### Global Environment Variables
- **`REGO_ENCRYPTION_KEY`**: Required for cache encryption (32-byte key)

#### Service-Specific Environment Variables
Each service will have tailored environtal variables specific to their usage.

### Service Implementation Standards

#### Required Files for Each Service

1. **`{service}.go`** - Main client implementation
   - Must include `NewClient()` function
   - Standard client struct with BaseURL, HTTP, Log, Cache
   - Environment variable configuration
   - Cache initialization with encryption

2. **`entities.go`** - All data structures
   - Request/response structs
   - Query parameter structs
   - Error types
   - Constants and enums

3. **Domain-specific files** (as needed)
   - `users.go` - User management operations
   - `devices.go` - Device management operations
   - `groups.go` - Group management operations
   - `admin.go` - Administrative operations

4. **`CLAUDE.md`** - Service-specific AI guidance
   - Authentication details
   - Common operations examples
   - Service-specific patterns
   - Known limitations

5. **Embedded resources** (optional)
   - Configuration JSON files
   - Schema definitions
   - Discovery documents

#### Authentication Patterns

**API Key/Token Pattern:**
```go
// Environment variable
token := config.GetEnv("{SERVICE}_API_TOKEN", "")

// Header configuration
headers["Authorization"] = "Bearer " + token
// or for custom schemes
headers["Authorization"] = "SSWS " + token
headers["X-Api-Key"] = token
```

**OAuth2/JWT Pattern:**
```go
// Service account credentials
creds := config.GetEnv("{SERVICE}_CREDENTIALS", "")
jwt, err := auth.GenerateJWT(creds)

// Token refresh handling
if token.Expired() {
    token, err = auth.RefreshToken(refreshToken)
}
```

**Session-based Pattern:**
```go
// Initial authentication
session, err := auth.Login(username, password)

// Session maintenance
if session.IsExpired() {
    session, err = auth.Renew(session)
}
```

### Adding a New Service

#### Step 1: Create Package Structure
```bash
mkdir -p pkg/newservice
cd pkg/newservice
```

#### Step 2: Implement Core Files

**newservice.go:**
```go
package newservice

import (
    "github.com/gemini-oss/rego/pkg/common/config"
    "github.com/gemini-oss/rego/pkg/common/requests"
)

type Config struct {
    APIKey      string
    BaseURL     string
    EnableCache bool
    CacheTTL    time.Duration
}

type Client struct {
    config   *Config
    client   *requests.Client
}

func NewClient(verbosity int, opts ...Option) *Client {
    log := log.NewLogger("newservice", verbosity)

	// Default config reads from environment for toggling -- `Twelve-Factor manifesto` principle
	cfg := &clientConfig{
		useSandbox: config.GetEnv("NEWSERVICE_SANDBOX") == "true", // If service has multiple instances
		baseURL:    "NEWSERVICE_URL",
		token:      "NEWSERVICE_API_TOKEN",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// If sandbox is toggled (either from environment or an override),
	// switch to sandbox environment variable names
	switch cfg.useSandbox {
	case true:
		cfg.baseURL = "NEWSERVICE_URL"
		cfg.token = "NEWSERVICE_API_TOKEN"
	}

    // Environment variables and validation
	baseURL := config.GetEnv(cfg.baseURL)
	if len(baseURL) == 0 {
		log.Fatalf("%s is not set", cfg.baseURL)
	}

    token := config.GetEnv(cfg.token)
	if len(token) == 0 {
		log.Fatalf("%s is not set", cfg.token)
	}

	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

    // Initialize cache
	cache, err := cache.NewCache(encryptionKey, "rego_cache_newservice.gob", 1000000)
	if err != nil {
		panic(err)
	}

    // Configure HTTP client based on service's documentation
	headers := requests.Headers{
		"Authorization": "SSWS " + token,
		"Accept":        requests.JSON,
		"Content-Type":  requests.JSON,
	}
    httpClient := requests.NewClient(nil, headers, nil)
	httpClient.BodyType = requests.JSON

	// Link to documentation for service-specific ratelimiting for humans to understand why this is set
	httpClient.RateLimiter = ratelimit.NewRateLimiter()
	httpClient.RateLimiter.ResetHeaders = true
	httpClient.RateLimiter.Log.Verbosity = verbosity

    return &Client{
        BaseURL: baseURL,
        HTTP:    httpClient,
        Log:     log,
        Cache:   cache,
    }
}
```

**entities.go:**
```go
package newservice

// Query structures
type UserQuery struct {
    Domain     string `url:"domain,omitempty"` // "Plain-English" Description of the field, preferably from documentation
    MaxResults int    `url:"limit,omitempty"`
    PageToken  string `url:"page_token,omitempty"`
    Filter     string `url:"filter,omitempty"`
}

func (q *UserQuery) ValidateQuery() error {
    if q.MaxResults > 1000 {
        return fmt.Errorf("maxResults cannot exceed 1000")
    }
    return nil
}

func (q *UserQuery) SetPageToken(token string) {
    q.PageToken = token
}

// Entity structures
type User struct {
    ID        string    `json:"id"`     // "Plain-English" Description of the field, preferably from documentation
    Email     string    `json:"email"`  // "Plain-English" Description of the field, preferably from documentation
    Name      string    `json:"name"`   // "Plain-English" Description of the field, preferably from documentation
    Status    string    `json:"status"` // etc.
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Response structures
type UsersResponse struct {
    Users      []User `json:"users"`
    NextPage   string `json:"next_page,omitempty"`
    TotalCount int    `json:"total_count"`
}

func (r *UsersResponse) GetItems() []User {
    return r.Users
}

func (r *UsersResponse) GetNextPageToken() string {
    return r.NextPage
}

// Error types
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"status"`
}

func (e *APIError) Error() string {
    return fmt.Sprintf("newservice API error %d: %s", e.Status, e.Message)
}
```

#### Step 3: Implement Operations
```go
// users.go
package newservice

// Generic do method
func do[T any](c *Client, method, endpoint string, query, body interface{}) (T, error) {
    var result T
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()

	res, body, err := c.HTTP.DoRequest(ctx, method, url, query, data)
	if err != nil {
		return *new(T), err
	}

	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &result)
	if err != nil {
		return *new(T), fmt.Errorf("unmarshalling error: %w", err)
	}

    return result, nil
}

// Paginated operations
func doPaginated[T interface{ GetItems() []User; GetNextPageToken() string }](
    c *Client, endpoint string, query UserQuery,
) ([]User, error) {
	var r T
	results := r

	for {
		r, err := do[T](c, method, url, query, data)
		if err != nil {
			return nil, err
		}

		results = results.Append(r)

		pageToken = r.PageToken()
		if pageToken == "" {
			break
		}

		query.SetPageToken(pageToken)
	}

	return &results, nil
}

// User operations
func (c *Client) ListUsers(query UserQuery) ([]User, error) {
    if err := query.ValidateQuery(); err != nil {
        return nil, fmt.Errorf("invalid query: %w", err)
    }

    // Check cache
    cacheKey := fmt.Sprintf("users:%v", query)
    var cachedUsers []User
    if c.Cache.Get(cacheKey, &cachedUsers) {
        c.Log.Debug("returning cached users")
        return cachedUsers, nil
    }

    // Fetch from API
    users, err := doPaginated[UsersResponse](c, "/users", query)
    if err != nil {
        return nil, fmt.Errorf("failed to list users: %w", err)
    }

    // Cache results
    c.Cache.Set(cacheKey, users, 30*time.Minute)

    return users, nil
}

func (c *Client) GetUser(userID string) (*User, error) {
    user, err := do[User](c, "GET", fmt.Sprintf("/users/%s", userID), nil, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
    }
    return &user, nil
}

func (c *Client) CreateUser(user *User) (*User, error) {
    created, err := do[User](c, "POST", "/users", nil, user)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    return &created, nil
}

func (c *Client) UpdateUser(userID string, updates map[string]interface{}) (*User, error) {
    updated, err := do[User](c, "PATCH", fmt.Sprintf("/users/%s", userID), nil, updates)
    if err != nil {
        return nil, fmt.Errorf("failed to update user %s: %w", userID, err)
    }
    return &updated, nil
}

func (c *Client) DeleteUser(userID string) error {
    _, err := do[interface{}](c, "DELETE", fmt.Sprintf("/users/%s", userID), nil, nil)
    if err != nil {
        return fmt.Errorf("failed to delete user %s: %w", userID, err)
    }
    return nil
}
```

#### Step 4: Implement Method Chaining (Optional)
For better developer experience:
```go
// User client for method chaining
type UserClient struct {
    client *Client
    query  UserQuery
}

func (c *Client) Users() *UserClient {
    return &UserClient{
        client: c,
        query:  UserQuery{},
    }
}

func (uc *UserClient) Domain(domain string) *UserClient {
    uc.query.Domain = domain
    return uc
}

func (uc *UserClient) Filter(filter string) *UserClient {
    uc.query.Filter = filter
    return uc
}

func (uc *UserClient) MaxResults(max int) *UserClient {
    uc.query.MaxResults = max
    return uc
}

func (uc *UserClient) List() ([]User, error) {
    return uc.client.ListUsers(uc.query)
}

// Usage:
// users, err := client.Users().Domain("example.com").Filter("status:active").List()
```

#### Step 5: Add Tests
Create test files in `/pkg/internal/tests/newservice/`:
```go
// newservice_test.go
package newservice_test

import (
    "testing"
    "net/http/httptest"
    "github.com/gemini-oss/rego/pkg/newservice"
)

func setupTestServer(t *testing.T) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/users":
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(mockUsersResponse)
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
}

func TestListUsers(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()

    // Set test environment
    t.Setenv("NEWSERVICE_URL", server.URL)
    t.Setenv("NEWSERVICE_API_KEY", "test-key")
    t.Setenv("REGO_ENCRYPTION_KEY", "test-encryption-key-32-bytes-long")

    client := newservice.NewClient(0)

    users, err := client.ListUsers(newservice.UserQuery{})
    assert.NoError(t, err)
    assert.Len(t, users, 2)
}
```

#### Step 6: Create Service-Specific CLAUDE.md
Create `/pkg/newservice/CLAUDE.md`:
```markdown
# NewService Package - AI Assistance Guide

This package provides integration with the NewService API.

## Authentication
- Uses Bearer token authentication
- Token set via NEWSERVICE_API_KEY environment variable

## Common Operations

### List Users
```go
users, err := client.ListUsers(newservice.UserQuery{
    Domain: "example.com",
    Filter: "status:active",
})
```

### Get Specific User
```go
user, err := client.GetUser("user-id-123")
```

## Service-Specific Patterns
- Rate limit: 100 requests per minute
- Pagination: Uses page tokens
- Cache TTL: 30 minutes default

## Known Limitations
- Maximum page size: 1000 items
- Filter syntax: field:value format only
```

### Common Patterns for API Operations

#### List Operations with Pagination
```go
// Pattern 1: Using generics.ProcessAllPagesOrdered
items, err := generics.ProcessAllPagesOrdered(
    func(params url.Values) (*ListResponse, error) {
        var resp ListResponse
        params.Set("limit", "100")
        err := client.Get("/api/items", &resp, params)
        return &resp, err
    },
    100, // items per page
)

// Pattern 2: Manual pagination
var allItems []Item
page := 1
for {
    resp, err := client.ListItems(page, 100)
    if err != nil {
        return nil, err
    }
    allItems = append(allItems, resp.Items...)
    if resp.NextPage == "" {
        break
    }
    page++
}
```

#### CRUD Operations
```go
// Create
newItem, err := client.CreateItem(ctx, &CreateItemRequest{
    Name: "Test Item",
    Type: "example",
})

// Read
item, err := client.GetItem(ctx, itemID)

// Update
updated, err := client.UpdateItem(ctx, itemID, &UpdateItemRequest{
    Name: "Updated Name",
})

// Delete
err := client.DeleteItem(ctx, itemID)
```

#### Batch Operations
```go
// Batch create with progress tracking
results, err := client.BatchCreate(items,
    progress.WithCallback(func(current, total int) {
        fmt.Printf("Progress: %d/%d\n", current, total)
    }),
)

// Concurrent operations (Jamf pattern)
results := client.doConcurrent(func(item Item) error {
    return client.ProcessItem(item)
}, items, 10) // 10 concurrent workers
```

### Service-Specific Features

#### Google Workspace
- Multi-domain support with domain switching
- Service account impersonation
- Comprehensive Admin SDK coverage (users, groups, devices, org units)
- Drive API with recursive folder operations
- Sheets API with formatting and batch updates
- Chrome Policy management with proto-style schemas
- Directory API with custom schemas

#### Okta
- Comprehensive user lifecycle management
- Application assignment and provisioning
- Role and permission reporting
- Group management with rules
- Factor enrollment and management

#### Jamf
- Dual API support (Classic XML API and Pro JSON API)
- Section-based response optimization (fetch only needed data)
- MDM command execution
- Configuration profile management
- Smart group support
- Concurrent device operations

#### Active Directory
- LDAP query builder
- Group membership management
- User attribute updates
- Password management
- Organizational unit operations

#### Slack
- Event handler framework
- Slash command support
- Interactive message components
- File upload/download
- Channel and user management

### Testing Patterns

#### Test Structure
```go
// pkg/internal/tests/service/service_test.go
func TestServiceOperations(t *testing.T) {
    // Setup test server
    server := SetupTestServer(t)
    defer server.Close()

    // Create test client
    client := SetupTestClient(t, server.URL)

    // Test operations
    t.Run("ListUsers", func(t *testing.T) {
        users, err := client.ListUsers()
        assert.NoError(t, err)
        assert.Len(t, users, 2)
    })
}
```

#### Mock Server Setup
```go
func SetupTestServer(t *testing.T) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/api/users":
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(mockUsersResponse)
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
}
```

### Rate Limiting Configuration

Different services have different rate limits. Most of them are based on their documentation.

Configure during client initialization:
```go
client, _ := service.NewClient(&Config{
    RateLimit: 100, // requests per minute
})
```

### Caching Strategy

#### Cache Configuration
- Default TTL: 30 minutes (configurable per service)
- Cache files: `rego_cache_{service}.gob`
- Encryption: AES-256-GCM using REGO_ENCRYPTION_KEY
- Compression: GZIP before encryption

#### Cache Control
```go
// Disable cache for specific request
client.DisableCache().GetUser(userID)

// Custom cache TTL
client.SetCacheTTL(10 * time.Minute).ListUsers()

// Clear cache
make flush  # Command line
client.ClearCache() // Programmatically
```

### Orchestration Examples

The `pkg/orchestrators/` directory contains multi-service workflows:

```go
// Example: Sync Okta users to Google Workspace
func SyncOktaToGoogle(oktaClient *okta.Client, googleClient *google.Client) error {
    // Get all active Okta users
    oktaUsers, err := oktaClient.ListActiveUsers()

    // Create/update in Google
    for _, oktaUser := range oktaUsers {
        googleUser := mapOktaToGoogle(oktaUser)
        _, err := googleClient.CreateOrUpdateUser(googleUser)
    }
}

// Example: Generate cross-platform report
func GenerateComplianceReport(okta *okta.Client, jamf *jamf.Client, slack *slack.Client) {
    // Collect data from multiple sources
    // Generate report
    // Send to Slack
}
```

### Debugging and Troubleshooting

#### Enable Debug Logging
```bash
export REGO_LOG_LEVEL=DEBUG
```

#### Common Issues

1. **Authentication Failures**
   - Check environment variables are set
   - Verify credentials are valid
   - Check service-specific auth requirements

2. **Rate Limiting**
   - Monitor rate limit headers in responses
   - Adjust client rate limit configuration
   - Use caching to reduce API calls

3. **Cache Issues**
   - Ensure REGO_ENCRYPTION_KEY is set
   - Clear cache with `make flush`
   - Check file permissions on cache directory

4. **Timeout Errors**
   - Increase client timeout settings
   - Check network connectivity
   - Verify service endpoints are correct

### Development Best Practices

1. **Always use common packages** - Don't reinvent HTTP clients or retry logic
2. **Follow existing patterns** - Check similar services for implementation patterns
3. **Add comprehensive tests** - Test both success and error cases
4. **Document in CLAUDE.md** - Add service-specific guidance for AI assistance
5. **Use environment variables** - Never hardcode credentials
6. **Handle errors gracefully** - Wrap errors with context
7. **Implement proper pagination** - Use generics.ProcessAllPagesOrdered
8. **Cache appropriately** - Balance performance with data freshness
9. **Rate limit responsibly** - Respect API limits
10. **Log meaningful information** - Use structured logging

### Code Review Checklist

When reviewing or implementing a service package, ensure:

#### Structure
- [ ] Main client file named `{service}.go`
- [ ] All structs in `entities.go`
- [ ] Domain-specific operations in separate files
- [ ] Service-specific `CLAUDE.md` present

#### Client Implementation
- [ ] `NewClient()` function with standard signature
- [ ] Environment variable configuration
- [ ] Required `REGO_ENCRYPTION_KEY` check
- [ ] Cache initialization with service-specific file
- [ ] Proper rate limiting configuration
- [ ] Structured logging with service prefix

#### Request Handling
- [ ] Generic `do[T]()` method implementation
- [ ] Proper timeout configuration (0-75 seconds)
- [ ] Content-type aware response handling
- [ ] Comprehensive error handling with status codes

#### Pagination
- [ ] Appropriate pagination pattern for API style
- [ ] Query parameter validation
- [ ] Page token/URL handling
- [ ] Complete item aggregation

#### Operations
- [ ] CRUD operations follow standard patterns
- [ ] Consistent error wrapping with context
- [ ] Cache key generation for read operations
- [ ] Method chaining where it improves UX

#### Testing
- [ ] Mock server setup for unit tests
- [ ] Test coverage for success and error cases
- [ ] Environment variable mocking
- [ ] Cache behavior verification

### Automating Service Documentation

To generate or update service-specific CLAUDE.md files, use this Claude Code command:

```bash
# Generate CLAUDE.md for a specific service
claude-code "Please analyze the pkg/{service}/ directory and generate a comprehensive CLAUDE.md file for this service. The file should include:
1. Service overview and purpose
2. Authentication method and required environment variables
3. Common operations with code examples
4. Service-specific patterns or quirks
5. Rate limiting information
6. Cache TTL recommendations
7. Known limitations or API constraints
8. Testing considerations
9. Links to official API documentation

Analyze the following files:
- {service}.go for client structure and initialization
- entities.go for data models and query parameters
- Any domain-specific files (users.go, devices.go, etc.)
- Test files for usage examples

Follow the pattern established in other service CLAUDE.md files like pkg/google/CLAUDE.md or pkg/okta/CLAUDE.md"
```

Example for a new service:
```bash
claude-code "Please analyze pkg/newservice/ and create a CLAUDE.md file following Rego's established patterns"
```

To update all service CLAUDE.md files at once:
```bash
# Update all service documentation
for service in $(ls -d pkg/*/ | grep -v common | grep -v internal | grep -v orchestrators); do
    service_name=$(basename $service)
    claude-code "Please analyze $service and update or create the CLAUDE.md file to match current implementation. Ensure it follows the standard format used by other services in Rego."
done
```

### Service-Specific CLAUDE.md Template

```markdown
# {Service} Package - AI Assistance Guide

This package provides integration with the {Service} API.

## Overview
Brief description of what this service does and its primary use cases.

## Authentication
- Authentication method: {Bearer Token/OAuth2/Session/etc.}
- Required environment variables:
  - `{SERVICE}_API_KEY`: API authentication token
  - `{SERVICE}_URL`: Base URL (optional, defaults to: {url})
  - Additional service-specific variables...

## Common Operations

### List {Resources}
```go
// Basic listing
items, err := client.List{Resources}({Service}Query{})

// With filtering
items, err := client.List{Resources}({Service}Query{
    Filter: "status:active",
    MaxResults: 100,
})

// Using method chaining (if applicable)
items, err := client.{Resources}().
    Filter("status:active").
    MaxResults(100).
    List()
```

### Get Specific {Resource}
```go
item, err := client.Get{Resource}("id-123")
```

### Create {Resource}
```go
newItem, err := client.Create{Resource}(&{Resource}{
    Name: "Example",
    // ... other fields
})
```

### Update {Resource}
```go
updated, err := client.Update{Resource}("id-123", map[string]interface{}{
    "name": "Updated Name",
})
```

### Delete {Resource}
```go
err := client.Delete{Resource}("id-123")
```

## Service-Specific Patterns

### Rate Limiting
- Rate limit: {X} requests per {time period}
- Implementation: {Token bucket/Fixed window/etc.}

### Pagination
- Style: {Page token/Offset/Link header}
- Maximum page size: {number}
- Default page size: {number}

### Caching
- Default TTL: {duration}
- Cache key pattern: `{service}:{operation}:{parameters}`
- Disable cache: `client.NoCache().Operation()`

### Error Handling
Describe any service-specific error types or error handling patterns.

## Known Limitations
- List any API constraints
- Maximum request sizes
- Specific error conditions
- Unsupported features

## Best Practices
1. Service-specific recommendations
2. Optimal batch sizes
3. Retry strategies
4. Cache invalidation patterns

## Related Documentation
- [Official API Documentation]({url})
- [Service Console]({url})
- [API Reference]({url})

## Testing Considerations
- Mock server patterns
- Common test scenarios
- Environment setup

## Advanced Usage

### Batch Operations
If the service supports batch operations, show examples.

### Webhook/Event Handling
If the service supports webhooks or events, show setup.

### Custom Field Handling
If the service has complex field mappings or custom schemas.
```

### Maintaining Service Standards

When updating existing services or reviewing PRs:

1. **Run the documentation check**:
   ```bash
   claude-code "Please review pkg/{service}/ and verify it follows all Rego design patterns. Check for:
   - Proper client initialization with NewClient(verbosity int)
   - Generic do[T]() method implementation
   - Consistent error handling
   - Proper caching with encryption
   - Query parameter validation
   - Appropriate pagination pattern
   Report any deviations from the standard patterns."
   ```

2. **Generate migration guide** for pattern updates:
   ```bash
   claude-code "Please analyze pkg/{service}/ and create a migration guide to update it to the latest Rego patterns, specifically:
   - Convert to generic do[T]() if using old pattern
   - Update client initialization to match standard
   - Implement proper query validation
   - Add method chaining where beneficial"
   ```

3. **Validate environment variables**:
   ```bash
   claude-code "Please scan all service packages and create a comprehensive list of all environment variables used, ensuring they follow the naming convention: {SERVICE}_{PROPERTY}"
   ```
