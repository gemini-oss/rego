/*
# Slack

This package initializes all the methods for functions which interact with the Slack Web API:
https://api.slack.com/web

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/slack.go
package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

const (
	BaseURL      = "https://slack.com/api"          // https://slack.com/api/METHOD_FAMILY.method?pretty=1
	AuditBaseURL = "https://api.slack.com/audit/v1" // AuditBaseURL is the base URL for Slack Audit Logs API (separate from Web API)
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...interface{}) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%v", url, id)
	}
	return url
}

/*
 * SetCache stores a Slack API response in the cache
 */
func (c *Client) SetCache(key string, value any, duration time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		c.Log.Error("Error marshalling cache data:", err)
		return
	}
	c.Cache.Set(key, data, duration)
}

/*
 * GetCache retrieves a Slack API response from the cache
 */
func (c *Client) GetCache(key string, target any) bool {
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

// clearServiceClients resets all cached service clients
func (c *Client) clearServiceClients() {
	c.usersClient = nil
	c.adminClient = nil
	c.chatClient = nil
	c.scimClient = nil
	c.auditClient = nil
}

// ResetRateLimiters clears all cached service clients, forcing fresh rate limiters
// Note: This does NOT affect the parent Client's HTTP rate limiter
func (c *Client) ResetRateLimiters() {
	c.clearServiceClients()
	c.Log.Println("Rate limiters reset - service clients will be recreated on next access")
}

/*
  - # Generate Slack Client
  - @param log *log.Logger
  - @return *Client
  - Example:

```go

	startServer1 := flag.Bool("startServer1", false, "Start server 1")
	flag.Parse()

	s := slack.NewClient(log.DEBUG)

	if *startServer1 {
		handlers := server.Handler{
			"/slack/events":           s.EventHandler,
			"/slack/command/userlist": s.CommandHandler,
		}
		go server.StartServer("127.0.0.1:8080", handlers)
	}

```
*/
func NewClient(verbosity int) *Client {
	log := log.NewLogger("{slack}", verbosity)

	token := config.GetEnv("SLACK_API_TOKEN")
	if len(token) == 0 {
		log.Fatal("SLACK_API_TOKEN is not set.")
	}

	signingSecret := config.GetEnv("SLACK_SIGNING_SECRET")
	if len(signingSecret) == 0 {
		log.Warning("SLACK_SIGNING_SECRET is not set.")
	}

	// Cache initialization
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_slack.gob", 1000000)
	if err != nil {
		panic(err)
	}

	headers := requests.Headers{
		"Authorization": "Bearer " + token,
		"Accept":        requests.JSON,
		"Content-Type":  fmt.Sprintf("%s; charset=utf-8", requests.JSON),
	}

	// Tier 3 rate limit (50/min) as default
	// https://api.slack.com/docs/rate-limits
	rl := ratelimit.NewRateLimiter(50, 1*time.Minute)
	rl.ResetHeaders = true
	rl.Log.Verbosity = verbosity

	httpClient := requests.NewClient(nil, headers, rl)
	httpClient.BodyType = requests.JSON

	return &Client{
		BaseURL:       BaseURL,
		HTTP:          httpClient,
		Log:           log,
		Token:         token,
		SigningSecret: signingSecret,
		Cache:         cache,
	}
}

// SlackAPIResponse is a generic interface for Slack API responses involving pagination
type SlackAPIResponse[T any] interface {
	Append(T) T
	NextCursor() string
}

// SlackQuery is an interface for Slack API queries involving pagination
type SlackQuery interface {
	// SetCursor updates the query's cursor for pagination.
	SetCursor(string)
}

/*
 * Perform a generic request to the Slack API
 */
func do[T any](c *Client, method string, url string, query any, data any) (T, error) {
	var result T
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()

	c.Log.Debug("URL:", url)

	res, body, err := c.HTTP.DoRequest(ctx, method, url, query, data)
	if err != nil {
		// Handle nil response (e.g., context timeout, network error)
		if res == nil {
			return *new(T), fmt.Errorf("request failed: %w", err)
		}
		return *new(T), err
	}

	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	// Check for Slack API error (ok: false) before unmarshalling to target type
	// Slack returns HTTP 200 even for API errors, so we need to inspect the response body
	var slackError SlackAPIError
	if err := json.Unmarshal(body, &slackError); err == nil && !slackError.OK && slackError.ErrorCode != "" {
		return *new(T), &slackError
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return *new(T), fmt.Errorf("unmarshalling error: %w", err)
	}

	return result, nil
}

/*
 * Perform a paginated request to the Slack API using cursor-based pagination
 * GET methods: parameters sent as URL query params
 * POST methods: parameters sent as form body (Slack admin endpoints)
 */
func doPaginated[T SlackAPIResponse[T], Q SlackQuery](c *Client, method string, url string, query Q, data any) (*T, error) {
	var r T
	results := r

	page := 1
	for {
		c.Log.Debug(fmt.Sprintf("Fetching page %d with query: %+v", page, query))

		r, err := do[T](c, method, url, query, data)
		if err != nil {
			return nil, err
		}

		results = results.Append(r)

		cursor := r.NextCursor()
		c.Log.Debug(fmt.Sprintf("Page %d complete, next_cursor: %q", page, cursor))

		if cursor == "" {
			break
		}

		query.SetCursor(cursor)
		page++
	}

	c.Log.Println(fmt.Sprintf("Pagination complete: fetched %d pages", page))
	return &results, nil
}
