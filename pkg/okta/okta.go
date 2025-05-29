/*
# Okta

This package initializes all the methods for functions which interact with the Okta API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/okta.go
package okta

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
	BaseURL = fmt.Sprintf("https://%s.%s.com/api/v1", "%s", "%s") // https://developer.okta.com/docs/api/#versioning
)

const (
	OktaApps       = "%s/apps"         // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/
	OktaAttributes = "%s/attributes"   // https://developer.okta.com/docs/api/openapi/asa/asa/tag/attributes/
	OktaGroups     = "%s/groups"       // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/
	OktaGroupRules = "%s/groups/rules" // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/GroupRule/
	OktaDevices    = "%s/devices"      // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/
	OktaUsers      = "%s/users"        // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/
	OktaIAM        = "%s/iam"          // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/
	OktaRoles      = "%s/iam/roles"    // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/
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
 * SetCache stores an Okta API response in the cache
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
 * GetCache retrieves an Okta API response from the cache
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
  - # Generate Okta Client
  - @param logger *log.Logger
  - @return *Client
  - Example:

```go

	o := okta.NewClient(log.DEBUG)

```
*/
func NewClient(verbosity int, opts ...Option) *Client {
	log := log.NewLogger("{okta}", verbosity)

	// Default config reads from environment for toggling -- `Twelve-Factor manifesto` principle
	cfg := &clientConfig{
		useSandbox: config.GetEnv("OKTA_USE_SANDBOX") == "true", // Expecting false if not set
		orgName:    "OKTA_ORG_NAME",
		baseURL:    "OKTA_BASE_URL",
		token:      "OKTA_API_TOKEN",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// If sandbox is toggled (either from environment or an override),
	// switch to sandbox environment variable names
	switch cfg.useSandbox {
	case true:
		cfg.orgName = "OKTA_SANDBOX_ORG_NAME"
		cfg.baseURL = "OKTA_SANDBOX_BASE_URL"
		cfg.token = "OKTA_SANDBOX_API_TOKEN"
	}

	orgName := config.GetEnv(cfg.orgName) // {ORG_NAME}.{BASE_URL}
	if len(orgName) == 0 {
		log.Fatalf("%s is not set", cfg.orgName)
	}

	orgName = strings.TrimPrefix(orgName, "https://")
	orgName = strings.TrimPrefix(orgName, "http://")
	orgName = strings.TrimSuffix(orgName, ".okta.com")

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

	cache, err := cache.NewCache(encryptionKey, "rego_cache_okta.gob", 1000000)
	if err != nil {
		panic(err)
	}

	// https://developer.okta.com/docs/reference/rl-best-practices/
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
 * Perform a generic request to the Okta API
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

/*
 * Generically perform a paginated request to the Okta API for a slice
 */
func doPaginated[T Slice[E], E any](c *Client, method, url string, query interface{}, data interface{}) (*T, error) {
	var emptySlice T = make([]E, 0)
	results := PagedSlice[T, E]{
		Results:  &emptySlice,
		OktaPage: &OktaPage{},
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

/*
 * Generically perform a paginated request to the Okta API for a struct
 */
func doPaginatedStruct[T Struct[T]](c *Client, method, url string, query interface{}, data interface{}) (*T, error) {
	var t T
	results := PagedStruct[T]{
		Results:  t.Init(),
		OktaPage: &OktaPage{},
	}

	for {
		res, body, err := c.HTTP.DoRequest(context.Background(), method, url, query, data)
		if err != nil {
			return nil, err
		}

		c.Log.Println("Response Status:", res.Status)
		c.Log.Debug("Response Body:", string(body))

		var page T
		err = json.Unmarshal(body, &page)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling error: %w", err)
		}

		(*results.Results).Append(&page)

		url = results.NextPage(res.Header.Values("Link"))
		query = nil
		if url == "" {
			break
		}
	}

	return results.Results, nil
}
