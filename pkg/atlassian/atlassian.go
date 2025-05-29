/*
# Atlassian

This package initializes all the methods for functions which interact with the Atlassian APIs:

* Confluence Cloud
- https://developer.atlassian.com/cloud/confluence/rest/v1/intro/#about

* Jira Service Management Cloud
- https://developer.atlassian.com/cloud/jira/service-desk/rest/intro/#about

* Jira Software Cloud
- https://developer.atlassian.com/cloud/jira/software/rest/intro/#introduction

* Cloud Admin
- https://developer.atlassian.com/cloud/admin/rest-apis/

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/atlassian/atlassian.go
package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	BaseURL                    = fmt.Sprintf("https://%s/", "%s")
	CloudAdmin                 = fmt.Sprintf("%s/admin", CloudURL)           // https://developer.atlassian.com/cloud/admin/organization/rest/intro/#uri
	JiraCloud                  = fmt.Sprintf("%s/rest/api/3", "%s")          // https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#ad-hoc-api-calls
	JiraSoftwareCloud          = fmt.Sprintf("%s/rest/agile/latest", "%s")   // https://developer.atlassian.com/cloud/jira/software/rest/intro/#uri-structure
	JiraServiceManagementCloud = fmt.Sprintf("%s/rest/servicedeskapi", "%s") // https://developer.atlassian.com/cloud/jira/service-desk/rest/intro/#ad-hoc-api-calls
	Forms                      = fmt.Sprintf("%s/jira/forms", CloudURL)      // https://developer.atlassian.com/cloud/forms/rest/intro/#ad-hoc-api-calls
	ConfluenceCloud            = fmt.Sprintf(V2, "%s/wiki/api")              // https://developer.atlassian.com/cloud/confluence/rest/v2/intro/#using
)

const (
	CloudURL = "https://api.atlassian.com/"
	V1       = "%s/v1"
	V2       = "%s/v2"
	V3       = "%s/v3"
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%s", url, id)
	}
	return url
}

// UseCache() enables caching for the next method call.
func (c *Client) UseCache() *Client {
	c.Cache.Enabled = true
	return c
}

/*
 * SetCache stores an Atlassian API response in the cache
 */
func (c *Client) SetCache(key string, value interface{}, duration time.Duration) {
	// Convert value to a byte slice and cache it
	data, err := json.Marshal(value)
	if err != nil {
		c.Log.Error("Error marshalling cache data:", err)
		return
	}
	c.Cache.Set(key, data, duration)
}

/*
 * GetCache retrieves an Atlassian API response from the cache
 */
func (c *Client) GetCache(key string, target interface{}) bool {
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
  - # Generate Atlassian Client
  - @param logger *log.Logger
  - @return *Client
  - Example:

```go

	a := atlassian.NewClient(log.DEBUG)

```
*/
func NewClient(verbosity int, opts ...Option) *Client {
	log := log.NewLogger("{atlassian}", verbosity)

	// Default config reads from environment for toggling -- `Twelve-Factor manifesto` principle
	cfg := &clientConfig{
		useSandbox: config.GetEnv("ATLASSIAN_USE_SANDBOX") == "true", // Expecting false if not set
		orgID:      "ATLASSIAN_ORG_ID",
		baseURL:    "ATLASSIAN_BASE_URL",
		token:      "ATLASSIAN_API_TOKEN",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// If sandbox is toggled (either from environment or an override),
	// switch to sandbox environment variable names
	switch cfg.useSandbox {
	case true:
		cfg.orgID = "ATLASSIAN_SANDBOX_ORG_ID"
		cfg.baseURL = "ATLASSIAN_SANDBOX_BASE_URL"
		cfg.token = "ATLASSIAN_SANDBOX_API_TOKEN"
	}

	orgName := config.GetEnv(cfg.orgID) // {ORG_NAME}.{BASE_URL}
	if len(orgName) == 0 {
		log.Fatalf("%s is not set", cfg.orgID)
	}

	orgName = strings.TrimPrefix(orgName, "https://")
	orgName = strings.TrimPrefix(orgName, "http://")
	orgName = strings.TrimSuffix(orgName, "") // Something.atlassian.net

	base := config.GetEnv(cfg.baseURL) // {ORG_NAME}.{BASE_URL}
	if len(base) == 0 {
		log.Fatalf("%s is not set", cfg.baseURL)
	}

	base = strings.Trim(base, "./")
	base = strings.TrimSuffix(base, ".com")

	token := config.GetEnv(cfg.token)
	if len(token) == 0 {
		log.Fatalf("%s is not set", cfg.token)
	}
	BaseURL := fmt.Sprintf(BaseURL, orgName, base)

	headers := requests.Headers{
		"Authorization": "SSWS " + token,
		"Accept":        requests.JSON,
		"Content-Type":  requests.JSON,
	}
	httpClient := requests.NewClient(nil, headers, nil)
	httpClient.BodyType = requests.JSON

	// Look into `Functional Options` patterns for a better way to handle this (and other clients while we're at it)
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_atlassian.gob", 1000000)
	if err != nil {
		panic(err)
	}

	// https://developer.atlassian.com/cloud/admin/organization/rest/intro/#rate%20limits
	httpClient.RateLimiter = ratelimit.NewRateLimiter()
	httpClient.RateLimiter.ResetHeaders = true
	httpClient.RateLimiter.Log.Verbosity = verbosity

	return &Client{
		BaseURL: BaseURL,
		HTTP:    httpClient,
		Log:     log,
		Cache:   cache,
	}
}

/*
 * Perform a generic request to the Atlassian API
 */
func do[T any](c *Client, method string, url string, query interface{}, data interface{}) (T, error) {
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
