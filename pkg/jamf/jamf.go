/*
# Jamf

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/jamf.go
package jamf

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	BaseURL      = fmt.Sprintf("https://%s/api", "%s")         // https://developer.jamf.com/jamf-pro/reference/jamf-pro-api
	V1           = "%s/v1"                                     // https://developer.jamf.com/jamf-pro/reference/jamf-pro-api
	V1_AuthToken = fmt.Sprintf("%s/auth/token", V1)            // https://developer.jamf.com/jamf-pro/reference/post_v1-auth-token
	V2           = "%s/v2"                                     // https://developer.jamf.com/jamf-pro/reference/jamf-pro-api
	ClassicURL   = fmt.Sprintf("https://%s/JSSResource", "%s") // https://developer.jamf.com/jamf-pro/reference/classic-api
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%s", url, id)
	}
	c.Log.Debug("url:", url)
	return url
}

// UseCache() enables caching for the next method call.
func (c *Client) UseCache() *Client {
	c.Cache.Use = true
	return c
}

/*
 * # Create a new Jamf Token based on the credentials provided
 * /api/v1/auth/token
 * - https://developer.jamf.com/jamf-pro/reference/post_v1-auth-token
 */
func GetToken(baseURL string) (*JamfToken, error) {
	url := fmt.Sprintf(V1_AuthToken, baseURL)

	// Prepare the credentials for Basic Auth
	creds := Credentials{
		Username: config.GetEnv("JSS_USERNAME", ""),
		Password: config.GetEnv("JSS_PASSWORD", ""),
	}
	basicCreds := fmt.Sprintf("%s:%s", creds.Username, creds.Password)

	headers := requests.Headers{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(basicCreds)),
		"Content-Type":  requests.JSON,
	}

	hc := requests.NewClient(nil, headers, nil)
	_, body, err := hc.DoRequest("POST", url, nil, nil)
	if err != nil {
		return nil, err
	}

	token := &JamfToken{}
	err = json.Unmarshal(body, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

/*
 * Create a new Jamf Client
 */
func NewClient(verbosity int) *Client {

	url := config.GetEnv("JSS_URL", "https://yourserver.jamfcloud.com")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")

	BaseURL := fmt.Sprintf(BaseURL, url)
	ClassicURL := fmt.Sprintf(ClassicURL, url)

	token, err := GetToken(BaseURL)
	if err != nil {
		panic(err)
	}

	headers := requests.Headers{
		"Authorization":             fmt.Sprintf("Bearer %s", token.Token),
		"Accept":                    fmt.Sprintf("%s, %s;q=0.9", requests.JSON, requests.XML),
		"Cache-Control":             "no-store, no-cache, must-revalidate, max-age=0, post-check=0, pre-check=0",
		"Strict-Transport-Security": "max-age=31536000 ; includeSubDomains",
		"Content-Type":              requests.JSON,
	}

	// Look into `Functional Options` patterns for a better way to handle this (and othe clients while we're at it)
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY", "32-byte-long-encryption-key-1234"))
	cache, err := cache.NewCache(encryptionKey, "/tmp/rego_cache_jamf.json")
	if err != nil {
		panic(err)
	}

	return &Client{
		BaseURL:    BaseURL,
		ClassicURL: ClassicURL,
		HTTP:       requests.NewClient(nil, headers, nil),
		Log:        log.NewLogger("{jamf}", verbosity),
		Cache:      cache,
	}
}
