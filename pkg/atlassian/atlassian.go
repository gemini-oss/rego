/*
# Atlassian

This package initializes all the methods for functions which interact with the Atlassian APIs:

* Jira Cloud
- https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#about

* Confluence Cloud
- https://developer.atlassian.com/cloud/confluence/rest/v1/intro/#about

* Jira Service Management Cloud
- https://developer.atlassian.com/cloud/jira/service-desk/rest/intro/#about

* Jira Software Cloud
- https://developer.atlassian.com/cloud/jira/software/rest/intro/#introduction

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/atlassian/atlassian.go
package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

var (
	CloudAdmin = fmt.Sprintf("%s/admin/%s", AtlassianBase, "%s")
)

const (
	AtlassianBase = "https://api.atlassian.com"
	V1            = "v1"
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
