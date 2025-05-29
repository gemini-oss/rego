/*
# Lenel S2

This package initializes all the methods for functions which interact with the Lenel S2 API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/lenel_s2/lenel_s2.go
package lenel_s2

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

const (
	NetBoxAPI = "%s/nbws/goforms/nbapi"
)

var (
	BaseURL = fmt.Sprintf("http://%s", "%s") // https://
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%s", url, id)
	}
	return url
}

// BuildRequest creates a consistent XML request struct.
func (c *Client) BuildRequest(name string, params interface{}) NetboxCommand {
	return NetboxCommand{
		SessionID: c.Session.ID,
		Command: Command{
			Name:       name,
			Num:        "1",
			DateFormat: "tzoffset",
			Params:     params,
		},
	}
}

// UseCache() enables caching for the next method call.
func (c *Client) UseCache() *Client {
	c.Cache.Enabled = true
	return c
}

/*
 * SetCache stores an S2 API response in the cache
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
 * GetCache retrieves an S2 API response from the cache
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
 * # Create a new S2 SessionID based on the credentials provided
 * <COMMAND name="Login" num="1" dateformat="tzoffset">
 */
func Login(baseURL string) (*NetboxResponse[any], error) {
	url := fmt.Sprintf(NetBoxAPI, baseURL)

	// Prepare the credentials for Basic Auth
	creds := Credentials{
		Username: config.GetEnv("S2_USERNAME"),
		Password: config.GetEnv("S2_PASSWORD"),
	}
	if len(creds.Username) == 0 || len(creds.Password) == 0 {
		return nil, fmt.Errorf("S2_USERNAME or S2_PASSWORD is not set")
	}

	headers := requests.Headers{
		"Content-Type": requests.XML,
	}

	payload := NetboxCommand{
		Command: Command{
			Name:       "Login",
			Num:        "1",
			DateFormat: "tzoffset",
			Params: struct {
				Username string `xml:"USERNAME"`
				Password string `xml:"PASSWORD"`
			}{
				Username: creds.Username,
				Password: creds.Password,
			},
		},
	}

	hc := requests.NewClient(nil, headers, nil)
	hc.BodyType = requests.XML
	_, body, err := hc.DoRequest(context.Background(), "POST", url, nil, payload)
	if err != nil {
		return nil, err
	}

	session := &NetboxResponse[any]{}
	err = xml.Unmarshal(body, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Logout ends the current session
func (c *Client) Logout() error {
	payload := NetboxCommand{
		SessionID: c.Session.ID,
		Command: Command{
			Name:       "Logout",
			Num:        "1",
			DateFormat: "tzoffset",
		},
	}
	_ = payload

	_, err := do[struct{}](c, "POST", "", payload)
	return err
}

/*
  - # Generate S2 Client
  - @param logger *log.Logger
  - @return *Client
  - Example:

```go

	s := s2.NewClient(log.DEBUG)

```
*/
func NewClient(baseURL string, verbosity int) *Client {
	log := log.NewLogger("{lenel_s2}", verbosity)

	url := config.GetEnv("S2_URL")
	if len(url) == 0 {
		log.Fatal("S2_URL is not set")
	}
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")
	url = fmt.Sprintf(BaseURL, url)

	session, _ := Login(url)
	if session == nil {
		log.Fatal("Failed to retrieve session ID")
	}

	headers := requests.Headers{
		//"Cookie":       fmt.Sprintf(".sessionId=%s", session.ID),
		"Content-Type": requests.XML,
	}

	httpClient := requests.NewClient(nil, headers, nil)
	httpClient.BodyType = requests.XML

	// Look into `Functional Options` patterns for a better way to handle this (and other clients while we're at it)
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_lenel_s2.gob", 1000000)
	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		BaseURL: url,
		Session: session,
		HTTP:    httpClient,
		Log:     log,
		Cache:   cache,
	}
}

// PaginatedResponse defines methods needed for handling pagination.
type PaginatedResponse interface {
	// NextToken returns the token to be used for the next request.
	NextToken() string
	// Append merges a new page of results with the existing result set.
	Append(resp PaginatedResponse) PaginatedResponse

	SetCommand(nb *Client, cmd NetboxCommand) NetboxCommand
}

func do[T any](c *Client, method string, url string, payload NetboxCommand) (T, error) {
	var result NetboxResponse[T]
	var empty T

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, body, err := c.HTTP.DoRequest(ctx, method, url, nil, payload)
	if err != nil {
		return empty, err
	}

	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	if err := xml.Unmarshal(body, &result); err != nil {
		return empty, fmt.Errorf("XML unmarshalling error: %w", err)
	}

	if result.Response.Code != "SUCCESS" {
		return empty, fmt.Errorf("error: %s", result.Response.Error)
	}

	return *result.Response.Details, nil
}

// doPaginated runs a paginated API request that returns a type implementing PaginatedResponse.
func doPaginated[T PaginatedResponse](c *Client, method string, url string, payload NetboxCommand) (*T, error) {
	var r T
	results := r

	pageToken := ""

	for {
		r, err := do[T](c, method, url, payload)
		if err != nil {
			return nil, err
		}

		// Merge the current page with previous pages.
		results = results.Append(r).(T)

		// Retrieve the token for the next page.
		pageToken = r.NextToken()
		if pageToken == "" {
			break
		}

		// Set the page token into the query for the next request.
		payload = r.SetCommand(c, payload)
	}

	return &results, nil
}
