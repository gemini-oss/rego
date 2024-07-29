/*
# Backupify

This package initializes all the methods for functions which interact with the Datto's Backupify WebUI:
https://www.backupify.com/

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/backupify/backupify.go
package backupify

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

const (
	backupifyBaseURL    = "https://%s.backupify.com/%s"
	customerServices    = "%s/customerServices"
	delete              = "%s/delete"
	download            = "%s/download"
	getActivities       = "%s/getActivities"
	restoreExportAction = "%s/restoreExportAction"
	serviceSnapshots    = "%s/serviceSnaps"
)

var (
	kilobyte    float64 = 1024
	megabyte    float64 = 1024 * kilobyte
	gigabyte    float64 = 1024 * megabyte
	terabyte    float64 = 1024 * gigabyte
	GoogleDrive AppType = "GoogleDrive"
	SharedDrive AppType = "GoogleTeamDrives"
	GoogleMail  AppType = "GoogleMail"
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...interface{}) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%v", url, id)
	}
	return url
}

// UseCache() enables caching for the next method call.
func (c *Client) UseCache() *Client {
	c.Cache.Enabled = true
	return c
}

/*
 * SetCache stores an Backupify response in the cache
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
 * GetCache retrieves an Backupify response from the cache
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

type ClientOption func(*Client)

/*
  - # Generate Backupify Client
  - @param logger *log.Logger
  - @return *Client
  - Example:

```go

	b := backupify.NewClient(log.DEBUG)

```
*/
func NewClient(verbosity int, opts ...ClientOption) *Client {
	log := log.NewLogger("{backupify}", verbosity)

	nodeURL := config.GetEnv("BACKUPIFY_NODE_URL")
	if len(nodeURL) == 0 {
		log.Fatal("BACKUPIFY_NODE_URL is not set")
	}

	customerID := config.GetEnv("BACKUPIFY_CUSTOMER_ID")
	if len(customerID) == 0 {
		log.Fatal("BACKUPIFY_CUSTOMER_ID is not set")
	}

	token := config.GetEnv("BACKUPIFY_EXPORT_TOKEN")
	if len(token) == 0 {
		log.Fatal("BACKUPIFY_EXPORT_TOKEN is not set")
	}

	phpSessID := config.GetEnv("BACKUPIFY_PHPSESSID")
	if len(phpSessID) == 0 {
		log.Fatal("BACKUPIFY_PHPSESSID is not set")
	}

	url := fmt.Sprintf(backupifyBaseURL, nodeURL, customerID)

	headers := requests.Headers{
		"Cookie":           "PHPSESSID=" + phpSessID,
		"Accept":           requests.All,
		"X-Requested-With": "XMLHttpRequest",
	}
	httpClient := requests.NewClient(nil, headers, nil)
	httpClient.BodyType = requests.FormURLEncoded

	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_backupify.gob", 1000000)
	if err != nil {
		panic(err)
	}

	c := &Client{
		BaseURL:     url,
		HTTP:        httpClient,
		Log:         log,
		Cache:       cache,
		exportToken: token,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

/*
 * Perform a generic request to the Backupify WebUI
 */
func do[T any](c *Client, method string, url string, query interface{}, data interface{}) (T, error) {
	var result T
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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
