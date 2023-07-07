/*
# Google Workspace

This package initializes all the methods for functions which interact with the Google Workspace API:
https://developers.google.com/workspace

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
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

	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
	"golang.org/x/oauth2/google"
)

const (
	API_KEY         = "api_key"
	OAUTH_CLIENT    = "oauth_client"
	SERVICE_ACCOUNT = "service_account"
	BaseURL         = "https://www.googleapis.com"
	AdminBaseURL    = "https://admin.googleapis.com"
	OAuthURL        = "https://accounts.google.com/o/oauth2/auth"
	OAuthTokenURL   = "https://oauth2.googleapis.com/token"
	JWTTokenURL     = "https://oauth2.googleapis.com/token"
)

/*
 * Build a URL for the Google Workspace API
 * @param endpoint string
 * @param identifiers ...string
 * @return string
 */
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%s", url, id)
	}
	return url
}

/*
 * # Generate JWT Client/Tokens for Google Workspace
 * @param auth AuthCredentials
 * @param logger *log.Logger
 * @return *Client
 * @return error
 * https://developers.google.com/identity/protocols/oauth2/service-account#jwt-auth
 */
func (c *Client) GenerateJWT(data []byte) (*requests.Client, error) {
	ctx := context.Background()

	c.Logger.Println("Generating JWT Config")
	jwtConfig, err := google.JWTConfigFromJSON(data, c.Auth.Scopes...)
	jwtConfig.Subject = c.Auth.Subject
	if err != nil {
		c.Logger.Printf("Unable to parse client secret file to config: %v", err)
	}
	c.Logger.Printf("JWT Config Successfully Generated")

	c.Logger.Println("Generating JWT Token")
	t, err := jwtConfig.TokenSource(ctx).Token()
	if err != nil {
		c.Logger.Printf("Unable to generate token: %v", err)
	}
	c.Logger.Printf("Token Successfully Generated")

	c.Logger.Println("Reconfiguring HTTP Client")
	type contextKey string
	jwtClient := jwtConfig.Client(context.WithValue(ctx, contextKey("token"), t))
	headers := requests.Headers{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + t.AccessToken,
	}

	return requests.NewClient(jwtClient, headers), nil
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
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + t.AccessToken,
	}

	// Update the HTTP client of the client object
	c.HTTPClient = requests.NewClient(jwtClient, headers)

	return nil
}

/*
 * # Generate Google Workspace Client
 * @param auth AuthCredentials
 * @param logger *log.Logger
 * @return *Client
 * @return error
 * Example:
```go
	ac := google.AuthCredentials{
		CICD: true,
		Type: google.SERVICE_ACCOUNT,
		Scopes: []string{
			"Admin SDK API",
			"Google Drive API",
			"Google Sheets API",
		},
	}
	g, _ := google.NewClient(ac, log.DEBUG)
```
 */
func NewClient(ac AuthCredentials, verbosity int) (*Client, error) {

	c := &Client{
		Auth:    ac,
		BaseURL: BaseURL,
		Logger:  log.NewLogger("{google}", verbosity),
	}

	c.Logger.Println("Initializing Google Client")
	headers := requests.Headers{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	// httpClient := requests.NewClient(nil, headers)

	c.Logger.Println("Loading Scopes")
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
	c.Logger.Printf("Scopes Loaded: %s\n", scopes)

	c.Logger.Println("Loading Credentials")
	switch c.Auth.CICD {
	case true:
		c.Logger.Println("Detected CICD Environment: Reading Credentials from Environment Variables")
		switch c.Auth.Type {
		case API_KEY:
			headers["Authorization"] = "Bearer " + config.GetEnv("GOOGLE_API_KEY", "GOOGLE_API_KEY")
		case OAUTH_CLIENT:
			b64 := config.GetEnv("GOOGLE_OAUTH_CLIENT", "GOOGLE_OAUTH_CLIENT")
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
			b64 := config.GetEnv("GOOGLE_SERVICE_ACCOUNT", "GOOGLE_SERVICE_ACCOUNT")
			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				fmt.Println("decode error:", err)
				return nil, err
			}

			c.HTTPClient, err = c.GenerateJWT(decoded)
			if err != nil {
				return nil, err
			}

			return c, nil
		}
	case false:
		c.Logger.Println("Detected Local Environment: Reading Credentials from Arguments")
		switch c.Auth.Type {
		case API_KEY:
			headers["Authorization"] = "Bearer " + c.Auth.Credentials
		case OAUTH_CLIENT:
			file, err := os.ReadFile(c.Auth.Credentials)
			if err != nil {
				c.Logger.Printf("Error opening file: %s\n", err)
			}
			oauth, err := google.ConfigFromJSON(file, c.Auth.Scopes...)
			if err != nil {
				c.Logger.Printf("Unable to parse client secret file to config: %v", err)
			}
			_ = oauth // Will return to this later
		case SERVICE_ACCOUNT:
			c.Logger.Println("Service Account Credentials Detected")

			c.Logger.Println("Loading Service Account Credentials from file")
			file, err := os.ReadFile(c.Auth.Credentials)
			if err != nil {
				c.Logger.Printf("Error opening file: %s\n", err)
			}

			c.Logger.Println("Generating JWT Client")
			c.HTTPClient, err = c.GenerateJWT(file)
			if err != nil {
				return nil, err
			}

			return c, nil
		}
	}
	return nil, nil
}
