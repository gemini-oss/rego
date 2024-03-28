/*
# Slack

This package initializes the methods for functions which handle a user methods with the Slack Web API:
https://api.slack.com/web

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/users.go
package slack

import (
	"encoding/json"
	"fmt"

	"github.com/gemini-oss/rego/pkg/common/requests"
)

// UserParameters represents the common parameters for the Slack 'users.list' and 'users.conversations' methods.
// See the [source](//https://api.slack.com/methods/users.list) and [source](//https://api.slack.com/methods/users.conversations) for more details.
type UserParameters struct {
	Cursor          string `json:"cursor,omitempty"`           // Paginate through collections by setting the cursor parameter to a next_cursor attribute returned by a previous request's response_metadata.
	ExcludeArchived bool   `json:"exclude_archived,omitempty"` // Set to true to exclude archived channels from the list.
	IncludeLocale   bool   `json:"include_locale,omitempty"`   // Set this to true to receive the locale for users.
	Limit           int    `json:"limit,omitempty"`            // The maximum number of items to return, up to 999.
	Pretty          bool   `json:"pretty,omitempty"`           // Make the response pretty JSON or not
	TeamID          string `json:"team_id,omitempty"`          // Encoded team id to list users in or conversations in, required if org token is used.
	Token           string `json:"token"`                      // Authentication token bearing required scopes.
	Types           string `json:"types,omitempty"`            // Mix and match channel types, e.g., "im,mpim".
	User            string `json:"user,omitempty"`             // Browse conversations by a specific user ID's membership.
}

// https://api.slack.com/methods/users.list
func (c *Client) ListUsers() (*Users, error) {
	users := &Users{}
	url := c.BuildURL("%s/users.list")

	c.HTTP.UpdateContentType(requests.FormURLEncoded)

	p := UserParameters{
		Token: c.Token,
	}

	res, body, err := c.HTTP.DoRequest("POST", url, nil, p)
	if err != nil {
		return nil, err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return users, nil
}

// https://api.slack.com/methods/users.conversations
func (c *Client) GetUserChannels(userID string) (*UserChannels, error) {
	user_channels := &UserChannels{}

	c.HTTP.UpdateContentType(requests.FormURLEncoded)
	url := c.BuildURL("%s/users.conversations")

	q := UserParameters{
		User:  userID,
		Types: "public_channel,private_channel",
	}

	res, body, err := c.HTTP.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &user_channels)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return user_channels, nil
}
