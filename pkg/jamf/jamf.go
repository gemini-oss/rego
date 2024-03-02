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
	"sync"
	"time"

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

/*
 * SetCache stores a Jamf API response in the cache
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
 * GetCache retrieves a Jamf API response from the cache
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
	cache, err := cache.NewCache(encryptionKey, "/tmp/rego_cache_jamf.gob")
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

// JamfResult is an interface for Jamf API responses involving pagination
type JamfAPIResponse interface {
	Total() int
	Append(interface{})
}

/*
 * Perform a generic request to the Jamf API
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
 * Perform a concurrent generic request to the Jamf API
 */
func doConcurrent[T JamfAPIResponse](c *Client, method string, url string, q *DeviceQuery, data interface{}) (*T, error) {
	// Do initial request to get the total number of items
	firstPage, err := do[T](c, method, url, q, data)
	if err != nil {
		return nil, err
	}

	// If there's only one page, return the result
	totalPages := calculateTotalPages(firstPage.Total(), q.PageSize)
	if totalPages <= 1 {
		return &firstPage, nil
	}

	// Create a channel to collect results from each goroutine
	resultsCh := make(chan *T, totalPages)
	errCh := make(chan error, totalPages)

	// Use a buffered channel as a semaphore to limit concurrent requests.
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	// Start goroutines for each page
	for i := (q.Page + 1); i < totalPages; i++ {
		wg.Add(1)
		go func(p int) {
			// Release one semaphore resource when the goroutine completes
			defer wg.Done()

			sem <- struct{}{} // acquire one semaphore resource

			// Create a new query with the current page
			q := *q
			c.Log.Println("Query:", q)
			q.Page = p

			result, err := do[T](c, method, url, q, data)
			if err != nil {
				errCh <- err
				return
			}

			resultsCh <- &result
			<-sem // release one semaphore resource
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultsCh)
	close(errCh)

	// Check for errors
	if len(errCh) > 0 {
		return nil, <-errCh
	}

	// Combine results from all pages
	results := firstPage
	for result := range resultsCh {
		results.Append(result)
	}

	return &results, nil
}

/*
 * Calculate the total number of pages
 * based on the total number of items and the page size
 */
func calculateTotalPages(totalItems, pageSize int) int {
	totalPages := totalItems / pageSize
	if totalItems%pageSize > 0 {
		totalPages++
	}
	return totalPages
}
