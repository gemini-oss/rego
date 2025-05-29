/*
# Google Workspace

This package initializes all the methods for functions which interact with the Google Workspace API:
https://developers.google.com/workspace

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/google.go
package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
	"golang.org/x/oauth2/google"
)

const (
	API_KEY         = "api_key"
	OAUTH_CLIENT    = "oauth_client"
	SERVICE_ACCOUNT = "service_account"
	BaseURL         = "https://www.googleapis.com"
	AdminBaseURL    = "https://admin.googleapis.com"
	ChromeBaseURL   = "https://chromepolicy.googleapis.com"
	OAuthURL        = "https://accounts.google.com/o/oauth2/auth"
	OAuthTokenURL   = "https://oauth2.googleapis.com/token"
	JWTTokenURL     = "https://oauth2.googleapis.com/token"
)

/*
 * Build a URL for the Google Workspace API
 * @param endpoint string
 * @param customer *Customer
 * @param parameters ...string
 * @return string
 */
func (c *Client) BuildURL(endpoint string, customer *Customer, parameters ...string) string {
	var url string
	if strings.Contains(endpoint, "/customer/%s") || strings.Contains(endpoint, "/customers/%s") {
		if customer == nil {
			customer = &Customer{}
		}
		url = fmt.Sprintf(endpoint, customer.String())
	} else {
		url = endpoint
	}

	// If the URL has multiple %s placeholders, ensure to use the appropriate parameters
	if strings.Contains(url, "%s") {
		args := make([]any, len(parameters))
		for i, param := range parameters {
			args[i] = param
		}
		url = fmt.Sprintf(url, args...)
		parameters = parameters[len(args):] // Remove used parameters from the slice
	}

	// Handle any remaining parameters
	for _, param := range parameters {
		if param != "" {
			if strings.HasPrefix(param, ":") {
				url = strings.TrimSuffix(url, "/") + param
			} else {
				url = fmt.Sprintf("%s/%s", url, param)
			}
		}
	}

	c.Log.Debug("url:", url)
	return url
}

/*
 * SetCache stores a Google API response in the cache
 */
func (c *Client) SetCache(key string, value any, duration time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		c.Log.Error("Error marshalling cache data:", err)
		return
	}
	c.Cache.Set(key, data, duration)
}

/*
 * GetCache retrieves a Google API response from the cache
 */
func (c *Client) GetCache(key string, target any) bool {
	data, found := c.Cache.Get(key)
	if !found {
		return false
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		c.Log.Error("Error unmarshalling cache data:", err)
		return false
	}
	return true
}

/*
 * # Generate JWT Client/Tokens for Google Workspace
 * @param auth AuthCredentials
 * @param Log *log.Logger
 * @return *Client
 * @return error
 * https://developers.google.com/identity/protocols/oauth2/service-account#jwt-auth
 */
func (c *Client) GenerateJWT(data []byte) (*requests.Client, error) {
	ctx := context.Background()

	c.Log.Println("Generating JWT Config")
	jwtConfig, err := google.JWTConfigFromJSON(data, c.Auth.Scopes...)
	jwtConfig.Subject = c.Auth.Subject
	if err != nil {
		c.Log.Printf("Unable to parse client secret file to config: %v", err)
	}
	c.Log.Printf("JWT Config Successfully Generated")

	c.Log.Println("Generating JWT Token")
	t, err := jwtConfig.TokenSource(ctx).Token()
	if err != nil {
		c.Log.Fatalf("Unable to generate token: %v", err)
	}
	if t == nil {
		c.Log.Fatal("Unable to generate token. Ensure you have the correct scopes set in your client secret file or are using the correct subject")
	}
	c.Log.Printf("Token Successfully Generated")

	c.Log.Println("Reconfiguring HTTP Client")
	type contextKey string
	jwtClient := jwtConfig.Client(context.WithValue(ctx, contextKey("token"), t))
	headers := requests.Headers{
		"Accept":        requests.JSON,
		"Content-Type":  requests.JSON,
		"Authorization": "Bearer " + t.AccessToken,
	}

	return requests.NewClient(jwtClient, headers, c.HTTP.RateLimiter), nil
}

func (c *Client) ImpersonateUser(email string) error {
	// Update the JWT config to impersonate a new user
	c.JWT.Subject = email

	// Create a new token for the new user
	ctx := context.Background()
	t, err := c.JWT.TokenSource(ctx).Token()
	if err != nil {
		return fmt.Errorf("unable to generate token: %v", err)
	}

	// Create a new HTTP client with the new token
	type contextKey string
	jwtClient := c.JWT.Client(context.WithValue(ctx, contextKey("token"), t))

	// Update the headers to use the new token
	headers := requests.Headers{
		"Accept":        requests.JSON,
		"Content-Type":  requests.JSON,
		"Authorization": "Bearer " + t.AccessToken,
	}

	// Update the HTTP client of the client object
	c.HTTP = requests.NewClient(jwtClient, headers, nil)
	c.HTTP.BodyType = requests.JSON

	return nil
}

/*
  - # Generate Google Workspace Client
  - @param auth AuthCredentials
  - @param log *log.Logger
  - @return *Client
  - @return error
  - Example:

```go

	ac := google.AuthCredentials{
		CICD: true,
		Type: google.SERVICE_ACCOUNT,
		Scopes: []string{
			"Admin SDK API",
			"Google Drive API",
			"Google Sheets API",
		},
		Subject: "super.user@domain.com",
	}
	g, _ := google.NewClient(ac, log.DEBUG)

```

  - Example 2: (Some Scopes may not work with Subject)

```go

	ac := google.AuthCredentials{
		CICD: true,
		Type: google.SERVICE_ACCOUNT,
		Scopes: []string{
			"Chrome Policy API",
			"Chrome Management API",
		},
	}
	g, _ := google.NewClient(ac, log.DEBUG)

```

  - Example 3: Direct URLs

```go

	ac := google.AuthCredentials{
		CICD: true,
		Type: google.SERVICE_ACCOUNT,
		Scopes: []string{
			"https://www.googleapis.com/auth/admin.directory.user",
		},
	}
	g, _ := google.NewClient(ac, log.DEBUG)

```

  - Example 4: Transfer Customer context to a new client

```go

	ac := google.AuthCredentials{
		CICD: true,
		Type: google.SERVICE_ACCOUNT,
		Scopes: []string{
			"Admin SDK API",
		},
		Subject: "super.user@domain.com",
	}
	g, _ := google.NewClient(ac, log.DEBUG)

	ac.Scopes = []string{
		"Chrome Policy API",
		"Chrome Management API",
	}
	chrome, _ := google.NewClient(ac, log.DEBUG)
	chrome.Customer, _ = g.MyCustomer()

```
*/
func NewClient(ac AuthCredentials, verbosity int) (*Client, error) {
	log := log.NewLogger("{google}", verbosity)

	// Look into `Functional Options` patterns for a better way to handle this (and other clients while we're at it)
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_google.gob", 1000000)
	if err != nil {
		panic(err)
	}

	// https://developers.google.com/drive/api/guides/limits
	rl := ratelimit.NewRateLimiter(12000, 75*time.Second)
	rl.Log.Verbosity = verbosity

	c := &Client{
		Auth:    ac,
		BaseURL: BaseURL,
		Log:     log,
		Cache:   cache,
		HTTP:    requests.NewClient(nil, nil, rl),
	}

	log.Println("Initializing Google Client")
	headers := requests.Headers{
		"Accept":       requests.JSON,
		"Content-Type": requests.JSON,
	}

	log.Println("Loading Scopes")
	scopes := []string{}
	c.Auth.Scopes = DedupeScopes(c.Auth.Scopes)
	for service := range c.Auth.Scopes {
		s, err := LoadScopes(c.Auth.Scopes[service])
		if err != nil {
			return nil, err
		}
		switch {
		case strings.HasPrefix(c.Auth.Scopes[service], "https://www.googleapis.com/auth/"):
			scopes = append(scopes, c.Auth.Scopes[service])
		default:
			scopes = append(scopes, s...)
		}
	}
	c.Auth.Scopes = scopes
	log.Debugf("Scopes Loaded: %s\n", scopes)

	log.Println("Loading Credentials")
	switch c.Auth.CICD {
	case true:
		log.Println("Detected CICD Environment: Reading Credentials from Environment Variables")
		switch c.Auth.Type {
		case API_KEY:
			headers["Authorization"] = "Bearer " + config.GetEnv("GOOGLE_API_KEY")
			if len(headers["Authorization"]) <= 7 {
				return nil, fmt.Errorf("GOOGLE_API_KEY is not set")
			}
		case OAUTH_CLIENT:
			b64 := config.GetEnv("GOOGLE_OAUTH_CLIENT")
			if len(b64) == 0 {
				return nil, fmt.Errorf("GOOGLE_OAUTH_CLIENT is not set")
			}

			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				fmt.Println("decode error:", err)
				return nil, err
			}
			j := &GoogleConfig{}
			err = json.Unmarshal([]byte(decoded), &j)
			if err != nil {
				fmt.Println("unmarshal error:", err)
				return nil, err
			}
		case SERVICE_ACCOUNT:
			b64 := config.GetEnv("GOOGLE_SERVICE_ACCOUNT")
			if len(b64) == 0 {
				return nil, fmt.Errorf("GOOGLE_SERVICE_ACCOUNT is not set")
			}

			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				fmt.Println("decode error:", err)
				return nil, err
			}

			c.HTTP, err = c.GenerateJWT(decoded)
			if err != nil {
				return nil, err
			}
			c.HTTP.BodyType = requests.JSON

			return c, nil
		}
	case false:
		log.Println("Detected Local Environment: Reading Credentials from Arguments")
		switch c.Auth.Type {
		case API_KEY:
			headers["Authorization"] = "Bearer " + c.Auth.Credentials
			if len(headers["Authorization"]) <= 7 {
				return nil, fmt.Errorf("GOOGLE_API_KEY is not set")
			}
		case OAUTH_CLIENT:
			file, err := os.ReadFile(c.Auth.Credentials)
			if err != nil {
				log.Printf("Error opening file: %s\n", err)
			}
			oauth, err := google.ConfigFromJSON(file, c.Auth.Scopes...)
			if err != nil {
				log.Printf("Unable to parse client secret file to config: %v", err)
			}
			_ = oauth // Will return to this later
		case SERVICE_ACCOUNT:
			log.Println("Service Account Credentials Detected")

			log.Println("Loading Service Account Credentials from file")
			file, err := os.ReadFile(c.Auth.Credentials)
			if err != nil {
				log.Printf("Error opening file: %s\n", err)
			}

			log.Println("Generating JWT Client")
			c.HTTP, err = c.GenerateJWT(file)
			if err != nil {
				return nil, err
			}

			return c, nil
		}
	}

	return nil, nil
}

// GoogleAPIResponse is a generic interface for Google API responses involving pagination
type GoogleAPIResponse[T any] interface {
	Append(T) T
	PageToken() string
}

// GoogleQuery is an interface for Google API queries involving pagination
type GoogleQuery interface {
	// SetPageToken updates the query’s page token.
	SetPageToken(string)
}

/*
 * Perform a generic request to the Google API
 */
func do[T any](c *Client, method string, url string, query any, data any) (T, error) {
	var result T
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, body, err := c.HTTP.DoRequest(ctx, method, url, query, data)
	if err != nil {
		if requests.IsNonRetryableCode(res.StatusCode) {
			var googleError ErrorResponse
			err = json.Unmarshal(body, &googleError)
			if err != nil {
				return *new(T), fmt.Errorf("error unmarshalling API error response: %w", err)
			}
			return *new(T), googleError.Error
		}
	}

	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &result)
	if err != nil {
		return *new(T), fmt.Errorf("unmarshalling error: %w", err)
	}

	return result, nil
}

func doPaginated[T GoogleAPIResponse[T], Q GoogleQuery](c *Client, method string, url string, query Q, data any) (*T, error) {
	var r T
	results := r

	pageToken := ""

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
