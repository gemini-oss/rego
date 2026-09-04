/*
# Slack - SCIM (Enterprise Grid - Organization-wide)

This package contains methods for SCIM API operations with the Slack SCIM API:
https://docs.slack.dev/reference/scim-api/

SCIM (System for Cross-domain Identity Management) provides organization-wide
user management for Enterprise Grid organizations. Unlike the Web API's
admin.users.list which is workspace-scoped, SCIM operates across all workspaces.

Requirements:
- User OAuth Access Token (not bot token) with `admin` scope
- App installed at organization level by an Org Owner
- Enterprise Grid or Business+ plan

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/scim.go
package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

const (
	// SCIMBaseURL is the base URL for Slack SCIM 2.0 API (separate from Web API)
	SCIMBaseURL = "https://api.slack.com/scim/v2"

	// SCIM User endpoints
	// https://docs.slack.dev/reference/scim-api/users
	SCIMUsers = "%s/Users" // https://docs.slack.dev/reference/scim-api/users
)

// BuildURL builds a URL for a given SCIM resource and identifiers.
func (c *SCIMClient) BuildURL(endpoint string, identifiers ...interface{}) string {
	url := fmt.Sprintf(endpoint, c.SCIMBaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%v", url, id)
	}
	return url
}

// SCIM returns the SCIM sub-client for organization-wide user management
// https://docs.slack.dev/reference/scim-api/
func (c *Client) SCIM() *SCIMClient {
	if c.scimClient != nil {
		return c.scimClient
	}

	// SCIM API rate limits: 300/min for GET users
	// https://docs.slack.dev/reference/scim-api/rate-limits
	scimRL := ratelimit.NewRateLimiter(300, 1*time.Minute)
	scimRL.ResetHeaders = true
	scimRL.Log.Verbosity = c.Log.Verbosity

	scimHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), scimRL)
	scimHTTP.BodyType = requests.JSON

	scimClient := &Client{
		BaseURL:       c.BaseURL,
		HTTP:          scimHTTP,
		Log:           c.Log,
		Token:         c.Token,
		SigningSecret: c.SigningSecret,
		Cache:         c.Cache,
	}

	c.scimClient = &SCIMClient{
		Client:      scimClient,
		SCIMBaseURL: SCIMBaseURL,
	}
	return c.scimClient
}

// doSCIM performs a request to the SCIM API
func doSCIM[T any](c *SCIMClient, method string, url string, query any, data any) (T, error) {
	var result T
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()

	c.Log.Debug("SCIM URL:", url)

	res, body, err := c.HTTP.DoRequest(ctx, method, url, query, data)
	if err != nil {
		if res == nil {
			return *new(T), fmt.Errorf("SCIM request failed: %w", err)
		}
		return *new(T), err
	}

	c.Log.Debug("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	// Check for SCIM error response
	if res.StatusCode >= 400 {
		var scimError SCIMError
		if err := json.Unmarshal(body, &scimError); err == nil && scimError.Detail != "" {
			return *new(T), &scimError
		}
		return *new(T), fmt.Errorf("SCIM API error: %s", res.Status)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return *new(T), fmt.Errorf("unmarshalling error: %w", err)
	}

	return result, nil
}

/*
 * Perform a concurrent generic request to the SCIM API
 * Similar to SnipeIT's doConcurrent pattern but uses startIndex (1-based) instead of offset
 */
func doConcurrentSCIM[T SCIMPaginatedResponse[E], E any](c *SCIMClient, method, url string, query SCIMQueryInterface, data any) (*T, error) {
	// Fetch the first page to initialize the response and pagination details
	results, err := doSCIM[T](c, method, url, query, data)
	if err != nil {
		return nil, err
	}

	// If all results fit in one page, return early
	if len(*results.Elements()) >= results.TotalCount() {
		return &results, nil
	}

	// Init concurrency control
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex

	// Initialize pagination parameters
	startIndex := query.GetStartIndex()
	count := query.GetCount()

	// Function to fetch each page concurrently
	fetchPage := func(pageStartIndex int) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()

		q := query.Copy()
		q.SetStartIndex(pageStartIndex)
		q.SetCount(count)

		page, err := doSCIM[T](c, method, url, q, data)
		if err != nil {
			c.Log.Error("Error fetching SCIM page:", err)
			return
		}

		resultsMutex.Lock()
		results.Append(page.Elements())
		resultsMutex.Unlock()
	}

	// Start fetching remaining pages
	// SCIM uses 1-based startIndex, so next page starts at startIndex + count
	for nextStartIndex := startIndex + count; nextStartIndex <= results.TotalCount(); nextStartIndex += count {
		wg.Add(1)
		go fetchPage(nextStartIndex)
	}
	wg.Wait()

	return &results, nil
}

/*
 * # List users in the organization with pagination control
 * GET /scim/v1/Users
 * - https://docs.slack.dev/reference/scim-api/users
 *
 * Returns users across all workspaces in the Enterprise Grid organization.
 * Use Count (max 1000) and StartIndex (1-based) for pagination.
 */
func (c *SCIMClient) ListUsers(query *SCIMQuery) (*SCIMUserList, error) {
	url := c.BuildURL(SCIMUsers)

	if query == nil {
		query = &SCIMQuery{}
	}

	// Default to max allowed page size
	if query.Count == 0 {
		query.Count = 1000
	}

	// StartIndex is 1-based, default to 1
	if query.StartIndex == 0 {
		query.StartIndex = 1
	}

	resp, err := doSCIM[SCIMUserList](c, "GET", url, query, nil)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

/*
 * # List all users in the organization with concurrent pagination
 * GET /scim/v1/Users
 * - https://docs.slack.dev/reference/scim-api/users
 *
 * Automatically paginates through all users in the organization using
 * concurrent requests for improved performance.
 */
func (c *SCIMClient) ListAllUsers() (*SCIMUserList, error) {
	url := c.BuildURL(SCIMUsers)
	cacheKey := "SCIMUsers"

	var cache SCIMUserList
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	query := &SCIMQuery{
		Count:      1000,
		StartIndex: 1,
	}

	results, err := doConcurrentSCIM[SCIMUserList](c, "GET", url, query, nil)
	if err != nil {
		return nil, err
	}

	c.Log.Println(fmt.Sprintf("SCIM: Pagination complete, fetched %d users", len(*results.Resources)))

	c.SetCache(cacheKey, results, 30*time.Minute)
	return results, nil
}

/*
 * # Get a single user by ID
 * GET /scim/v1/Users/{id}
 * - https://docs.slack.dev/reference/scim-api/users
 */
func (c *SCIMClient) GetUser(userID string) (*SCIMUser, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	url := c.BuildURL(SCIMUsers, userID)
	cacheKey := fmt.Sprintf("SCIMUser_%s", userID)

	var cache SCIMUser
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	user, err := doSCIM[SCIMUser](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, &user, 5*time.Minute)
	return &user, nil
}

/*
 * # Filter users with SCIM filter expressions
 * GET /scim/v1/Users?filter={expression}
 * - https://docs.slack.dev/reference/scim-api/users
 *
 * Filter expressions follow SCIM filter syntax:
 *   - userName Eq "john.doe"
 *   - active eq true
 *   - emails.value co "@example.com"
 */
func (c *SCIMClient) FilterUsers(filter string) (*SCIMUserList, error) {
	url := c.BuildURL(SCIMUsers)

	if filter == "" {
		return nil, fmt.Errorf("filter expression is required")
	}

	query := &SCIMQuery{
		Count:      1000,
		StartIndex: 1,
		Filter:     filter,
	}

	results, err := doConcurrentSCIM[SCIMUserList](c, "GET", url, query, nil)
	if err != nil {
		return nil, err
	}

	c.Log.Println(fmt.Sprintf("SCIM: Filter complete, found %d users matching '%s'",
		len(*results.Resources), filter))

	return results, nil
}

/*
 * # Look up a user by email address
 * GET /scim/v1/Users?filter=emails.value eq "{email}"
 * - https://docs.slack.dev/reference/scim-api/users
 *
 * Returns the user matching the given email address.
 * Returns nil if no user is found.
 */
func (c *SCIMClient) LookupByEmail(email string) (*SCIMUser, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	cacheKey := fmt.Sprintf("SCIMUserByEmail_%s", email)

	var cache SCIMUser
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	// SCIM filter syntax: emails.value eq "user@example.com"
	filter := fmt.Sprintf("emails.value eq \"%s\"", email)
	results, err := c.FilterUsers(filter)
	if err != nil {
		return nil, err
	}

	if results.Resources == nil || len(*results.Resources) == 0 {
		return nil, nil
	}

	user := (*results.Resources)[0]
	c.SetCache(cacheKey, user, 5*time.Minute)
	return user, nil
}
