/*
# Slack - Users

This package contains methods for user operations with the Slack Web API:
https://api.slack.com/methods#users

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/users.go
package slack

import (
	"fmt"
	"time"

	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	// User endpoints (Tier 3 - 50/min)
	// https://api.slack.com/methods#users
	SlackUsers              = fmt.Sprintf("%s/users.list", BaseURL)
	SlackUsersInfo          = fmt.Sprintf("%s/users.info", BaseURL)
	SlackUsersLookupByEmail = fmt.Sprintf("%s/users.lookupByEmail", BaseURL)
	SlackUsersConversations = fmt.Sprintf("%s/users.conversations", BaseURL)
	SlackUsersSetPresence   = fmt.Sprintf("%s/users.setPresence", BaseURL)
)

// UsersClient for chaining methods
type UsersClient struct {
	*Client
}

// UserQuery for users.list and users.info
// https://api.slack.com/methods/users.list
type UserQuery struct {
	Cursor        string `url:"cursor,omitempty"`         // Paginate through collections by setting the cursor parameter to a next_cursor attribute returned by a previous request's response_metadata
	IncludeLocale bool   `url:"include_locale,omitempty"` // Set this to true to receive the locale for users
	Limit         int    `url:"limit,omitempty"`          // The maximum number of items to return, up to 999 (recommended: no more than 200)
	TeamID        string `url:"team_id,omitempty"`        // Encoded team id to list users in, required if org token is used
}

func (q *UserQuery) SetCursor(cursor string) {
	q.Cursor = cursor
}

// ConversationQuery for users.conversations
// https://api.slack.com/methods/users.conversations
type ConversationQuery struct {
	Cursor          string `url:"cursor,omitempty"`           // Paginate through collections by setting the cursor parameter to a next_cursor attribute returned by a previous request's response_metadata
	ExcludeArchived bool   `url:"exclude_archived,omitempty"` // Set to true to exclude archived channels from the list
	Limit           int    `url:"limit,omitempty"`            // The maximum number of items to return, up to 999 (recommended: no more than 200)
	TeamID          string `url:"team_id,omitempty"`          // Encoded team id to list conversations in, required if org token is used
	Types           string `url:"types,omitempty"`            // Mix and match channel types by providing a comma-separated list (e.g., "im,mpim,public_channel,private_channel")
	User            string `url:"user,omitempty"`             // Browse conversations by a specific user ID's membership
}

func (q *ConversationQuery) SetCursor(cursor string) {
	q.Cursor = cursor
}

// Users returns the users sub-client for user operations
func (c *Client) Users() *UsersClient {
	if c.usersClient != nil {
		return c.usersClient
	}

	// Users APIs are Tier 3 (50/min)
	// https://api.slack.com/docs/rate-limits
	usersRL := ratelimit.NewRateLimiter(50, 1*time.Minute)
	usersRL.ResetHeaders = true
	usersRL.Log.Verbosity = c.Log.Verbosity

	usersHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), usersRL)
	usersHTTP.BodyType = requests.FormURLEncoded // users.list uses form encoding

	usersClient := &Client{
		BaseURL:       c.BaseURL,
		HTTP:          usersHTTP,
		Log:           c.Log,
		Token:         c.Token,
		SigningSecret: c.SigningSecret,
		Cache:         c.Cache,
	}

	c.usersClient = &UsersClient{Client: usersClient}
	return c.usersClient
}

/*
 * # List all users in a Slack workspace
 * users.list
 * - https://api.slack.com/methods/users.list
 */
func (c *UsersClient) ListAllUsers() (*Users, error) {
	var cache Users
	if c.GetCache(SlackUsers, &cache) {
		return &cache, nil
	}

	q := &UserQuery{
		Limit: 200,
	}

	users, err := doPaginated[Users](c.Client, "GET", SlackUsers, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(SlackUsers, users, 30*time.Minute)
	return users, nil
}

/*
 * # Get information about a user
 * users.info
 * - https://api.slack.com/methods/users.info
 */
func (c *UsersClient) GetUser(userID string) (*Member, error) {
	cacheKey := fmt.Sprintf("%s_%s", SlackUsersInfo, userID)

	var cache Member
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	q := struct {
		User          string `url:"user"`
		IncludeLocale bool   `url:"include_locale,omitempty"`
	}{
		User: userID,
	}

	resp, err := do[UserResponse](c.Client, "GET", SlackUsersInfo, q, nil)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, fmt.Errorf("failed to get user: %s", resp.Error)
	}

	c.SetCache(cacheKey, resp.User, 5*time.Minute)
	return &resp.User, nil
}

/*
 * # Lookup user by email
 * users.lookupByEmail
 * - https://api.slack.com/methods/users.lookupByEmail
 */
func (c *UsersClient) LookupByEmail(email string) (*Member, error) {
	cacheKey := fmt.Sprintf("%s_%s", SlackUsersLookupByEmail, email)

	var cache Member
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	q := struct {
		Email string `url:"email"`
	}{
		Email: email,
	}

	resp, err := do[UserResponse](c.Client, "GET", SlackUsersLookupByEmail, q, nil)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, fmt.Errorf("failed to lookup user by email: %s", resp.Error)
	}

	c.SetCache(cacheKey, resp.User, 5*time.Minute)
	return &resp.User, nil
}

/*
 * # Get channels for a user
 * users.conversations
 * - https://api.slack.com/methods/users.conversations
 */
func (c *UsersClient) GetUserChannels(userID string) (*UserChannels, error) {
	cacheKey := fmt.Sprintf("%s_%s", SlackUsersConversations, userID)

	var cache UserChannels
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	q := &ConversationQuery{
		User:  userID,
		Types: "public_channel,private_channel",
		Limit: 200,
	}

	channels, err := doPaginated[UserChannels](c.Client, "GET", SlackUsersConversations, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, channels, 5*time.Minute)
	return channels, nil
}

/*
 * # List all conversations for a user with custom options
 * users.conversations
 * - https://api.slack.com/methods/users.conversations
 */
func (c *UsersClient) ListUserConversations(query *ConversationQuery) (*UserChannels, error) {
	if query == nil {
		query = &ConversationQuery{
			Limit: 200,
		}
	}

	channels, err := doPaginated[UserChannels](c.Client, "GET", SlackUsersConversations, query, nil)
	if err != nil {
		return nil, err
	}

	return channels, nil
}
