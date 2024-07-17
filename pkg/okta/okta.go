/*
# Okta

This package initializes all the methods for functions which interact with the Okta API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/okta.go
package okta

import (
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
	if !c.Cache.Enabled {
		return false
	}
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
func NewClient(verbosity int) *Client {
	log := log.NewLogger("{okta}", verbosity)

	org_name := config.GetEnv("OKTA_ORG_NAME") // {ORG_NAME}.okta.com
	//org_name := config.GetEnv("OKTA_SANDBOX_ORG_NAME")
	if len(org_name) == 0 {
		log.Fatal("OKTA_ORG_NAME is not set")
	}

	org_name = strings.TrimPrefix(org_name, "https://")
	org_name = strings.TrimPrefix(org_name, "http://")
	org_name = strings.TrimSuffix(org_name, ".okta.com")

	base := config.GetEnv("OKTA_BASE_URL") // {ORG_NAME}.{BASE_URL}
	//base := config.GetEnv("OKTA_SANDBOX_BASE_URL") // oktapreview.com
	if len(base) == 0 {
		log.Fatal("OKTA_BASE_URL is not set")
	}

	base = strings.Trim(base, "./")
	base = strings.TrimSuffix(base, ".com")

	token := config.GetEnv("OKTA_API_TOKEN")
	//token := config.GetEnv("OKTA_SANDBOX_API_TOKEN")
	if len(token) == 0 {
		log.Fatal("OKTA_API_TOKEN is not set")
	}
	BaseURL := fmt.Sprintf(BaseURL, org_name, base)

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

	res, body, err := c.HTTP.DoRequest(method, url, query, data)
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
		res, body, err := c.HTTP.DoRequest(method, url, query, data)
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
		res, body, err := c.HTTP.DoRequest(method, url, query, data)
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
