/*
# SnipeIT

This package initializes all the methods for functions which interact with the SnipeIT API:
https://snipe-it.readme.io/reference/api-overview

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/snipeit.go
package snipeit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	BaseURL = fmt.Sprintf("https://%s/api/v1", "%s") // https://snipe-it.readme.io/reference/api-overview
)

const (
	Assets           = "%s/hardware"             // https://snipe-it.readme.io/reference#hardware
	Fields           = "%s/fields"               // https://snipe-it.readme.io/reference/fields-1
	FieldSets        = "%s/fieldsets"            // https://snipe-it.readme.io/reference/fieldsets
	Companies        = "%s/companies"            // https://snipe-it.readme.io/reference#companies
	Locations        = "%s/locations"            // https://snipe-it.readme.io/reference#locations
	Accessories      = "%s/accessories"          // https://snipe-it.readme.io/reference#accessories
	Consumables      = "%s/consumables"          // https://snipe-it.readme.io/reference#consumables
	Components       = "%s/components"           // https://snipe-it.readme.io/reference#components
	Users            = "%s/users"                // https://snipe-it.readme.io/reference#users
	StatusLabels     = "%s/statuslabels"         // https://snipe-it.readme.io/reference#status-labels
	Models           = "%s/models"               // https://snipe-it.readme.io/reference#models
	Licenses         = "%s/licenses"             // https://snipe-it.readme.io/reference#licenses
	Categories       = "%s/categories"           // https://snipe-it.readme.io/reference#categories
	Manufacturers    = "%s/manufacturers"        // https://snipe-it.readme.io/reference#manufacturers
	Suppliers        = "%s/suppliers"            // https://snipe-it.readme.io/reference#suppliers
	AssetMaintenance = "%s/hardware/maintenance" // https://snipe-it.readme.io/reference#maintenances
	Departments      = "%s/departments"          // https://snipe-it.readme.io/reference#departments
	Groups           = "%s/groups"               // https://snipe-it.readme.io/reference#groups
	Settings         = "%s/settings"             // https://snipe-it.readme.io/reference#settings
	Reports          = "%s/reports"              // https://snipe-it.readme.io/reference#reports
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...interface{}) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%v", url, id)
	}
	c.Log.Debug("url:", url)
	return url
}

/*
 * SetCache stores a SnipeIT API response in the cache
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
 * GetCache retrieves a SnipeIT API response from the cache
 */
func (c *Client) GetCache(key string, target interface{}) bool {
	data, found := c.Cache.Get(key)
	if !found || !c.Cache.Enabled {
		return false
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		c.Log.Error("Error unmarshalling cache data:", err)
		return false
	}
	return true
}

func NewClient(verbosity int) *Client {
	log := log.NewLogger("{snipeit}", verbosity)

	url := config.GetEnv("SNIPEIT_URL")
	if len(url) == 0 {
		log.Fatal("SNIPEIT_URL is not set.")
	}

	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.Trim(url, "./")

	BaseURL = fmt.Sprintf(BaseURL, url)
	token := config.GetEnv("SNIPEIT_TOKEN")
	if len(token) == 0 {
		log.Fatal("SNIPEIT_TOKEN is not set.")
	}

	headers := requests.Headers{
		"Authorization": "Bearer " + token,
		"Accept":        requests.JSON,
		"Content-Type":  requests.JSON,
	}

	// https://snipe-it.readme.io/reference/api-throttling
	rl := ratelimit.NewRateLimiter(120, 1*time.Minute)

	httpClient := requests.NewClient(nil, headers, rl)
	httpClient.BodyType = requests.JSON

	// To Do: Look into `Functional Options` patterns for a better way to handle this
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_snipeit.gob", 1000000)
	if err != nil {
		panic(err)
	}

	return &Client{
		BaseURL: BaseURL,
		HTTP:    httpClient,
		Log:     log,
		Cache:   cache,
	}
}

// PaginatedResponse is an interface for SnipeIT API responses involving pagination
type PaginatedResponse[E any] interface {
	TotalCount() int
	Append(*[]*E)
	Elements() *[]*E
}

/*
 * Perform a generic request to the SnipeIT API
 */
func do[T any](c *Client, method string, url string, query interface{}, data interface{}) (T, error) {
	var result T
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
 * Perform a concurrent generic request to the SnipeIT API
 */
func doConcurrent[T PaginatedResponse[E], E any](c *Client, method, url string, query QueryInterface, data interface{}) (*T, error) {
	// Fetch the first page to initialize the response and pagination details.
	results, err := do[T](c, method, url, query, data)
	if err != nil {
		return nil, err
	}

	// Init concurrency control
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex

	// Initialize offset and limit based on the query interface.
	offset := query.GetOffset()
	limit := query.GetLimit()

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
