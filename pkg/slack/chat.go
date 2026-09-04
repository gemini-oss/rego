/*
# Slack - Chat

This package contains methods for chat operations with the Slack Web API:
https://api.slack.com/methods#chat

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/chat.go
package slack

import (
	"fmt"
	"time"

	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	// Chat endpoints (Special rate limits - 1 msg/sec per channel)
	// https://api.slack.com/methods#chat
	ChatPostMessage   = fmt.Sprintf("%s/chat.postMessage", BaseURL)
	ChatUpdate        = fmt.Sprintf("%s/chat.update", BaseURL)
	ChatDelete        = fmt.Sprintf("%s/chat.delete", BaseURL)
	ChatPostEphemeral = fmt.Sprintf("%s/chat.postEphemeral", BaseURL)
	ChatGetPermalink  = fmt.Sprintf("%s/chat.getPermalink", BaseURL)
)

// ChatClient for chaining methods
type ChatClient struct {
	*Client
}

// ChatResponse represents the response from chat methods
type ChatResponse struct {
	OK        bool   `json:"ok"`                   // Response status
	Channel   string `json:"channel,omitempty"`    // Channel ID where message was posted
	TS        string `json:"ts,omitempty"`         // Timestamp of the message
	MessageTS string `json:"message_ts,omitempty"` // Alias for ts in some responses
	Error     string `json:"error,omitempty"`      // Error message if not OK
	Warning   string `json:"warning,omitempty"`    // Warning message
}

// Chat returns the chat sub-client for messaging operations
func (c *Client) Chat() *ChatClient {
	if c.chatClient != nil {
		return c.chatClient
	}

	// Chat APIs have special rate limits (1 msg/sec per channel, several hundred/min workspace-wide)
	// Using Tier 3 (50/min) as a conservative default
	// https://api.slack.com/docs/rate-limits
	chatRL := ratelimit.NewRateLimiter(50, 1*time.Minute)
	chatRL.ResetHeaders = true
	chatRL.Log.Verbosity = c.Log.Verbosity

	chatHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), chatRL)
	chatHTTP.BodyType = requests.JSON

	chatClient := &Client{
		BaseURL:       c.BaseURL,
		HTTP:          chatHTTP,
		Log:           c.Log,
		Token:         c.Token,
		SigningSecret: c.SigningSecret,
		Cache:         c.Cache,
	}

	c.chatClient = &ChatClient{Client: chatClient}
	return c.chatClient
}

/*
 * # Post a message to a channel
 * chat.postMessage
 * - https://api.slack.com/methods/chat.postMessage
 */
func (c *ChatClient) PostMessage(message *SlackMessage) (*ChatResponse, error) {
	if message.Channel == "" {
		return nil, fmt.Errorf("channel is required")
	}

	resp, err := do[ChatResponse](c.Client, "POST", ChatPostMessage, nil, message)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, fmt.Errorf("failed to post message: %s", resp.Error)
	}

	return &resp, nil
}

/*
 * # Update a message in a channel
 * chat.update
 * - https://api.slack.com/methods/chat.update
 */
func (c *ChatClient) UpdateMessage(channel, ts, text string) (*ChatResponse, error) {
	if channel == "" {
		return nil, fmt.Errorf("channel is required")
	}
	if ts == "" {
		return nil, fmt.Errorf("message timestamp (ts) is required")
	}

	req := struct {
		Channel string `json:"channel"`
		TS      string `json:"ts"`
		Text    string `json:"text,omitempty"`
	}{
		Channel: channel,
		TS:      ts,
		Text:    text,
	}

	resp, err := do[ChatResponse](c.Client, "POST", ChatUpdate, nil, req)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, fmt.Errorf("failed to update message: %s", resp.Error)
	}

	return &resp, nil
}

/*
 * # Delete a message from a channel
 * chat.delete
 * - https://api.slack.com/methods/chat.delete
 */
func (c *ChatClient) DeleteMessage(channel, ts string) error {
	if channel == "" {
		return fmt.Errorf("channel is required")
	}
	if ts == "" {
		return fmt.Errorf("message timestamp (ts) is required")
	}

	req := struct {
		Channel string `json:"channel"`
		TS      string `json:"ts"`
	}{
		Channel: channel,
		TS:      ts,
	}

	resp, err := do[ChatResponse](c.Client, "POST", ChatDelete, nil, req)
	if err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("failed to delete message: %s", resp.Error)
	}

	return nil
}

/*
 * # Post an ephemeral message (visible only to a specific user)
 * chat.postEphemeral
 * - https://api.slack.com/methods/chat.postEphemeral
 */
func (c *ChatClient) PostEphemeral(channel, userID, text string) (*ChatResponse, error) {
	if channel == "" {
		return nil, fmt.Errorf("channel is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	req := struct {
		Channel string `json:"channel"`
		User    string `json:"user"`
		Text    string `json:"text,omitempty"`
	}{
		Channel: channel,
		User:    userID,
		Text:    text,
	}

	resp, err := do[ChatResponse](c.Client, "POST", ChatPostEphemeral, nil, req)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, fmt.Errorf("failed to post ephemeral message: %s", resp.Error)
	}

	return &resp, nil
}

/*
 * # Get a permalink for a message
 * chat.getPermalink
 * - https://api.slack.com/methods/chat.getPermalink
 */
func (c *ChatClient) GetPermalink(channel, messageTS string) (string, error) {
	if channel == "" {
		return "", fmt.Errorf("channel is required")
	}
	if messageTS == "" {
		return "", fmt.Errorf("message timestamp is required")
	}

	q := struct {
		Channel   string `url:"channel"`
		MessageTS string `url:"message_ts"`
	}{
		Channel:   channel,
		MessageTS: messageTS,
	}

	resp, err := do[struct {
		OK        bool   `json:"ok"`
		Permalink string `json:"permalink,omitempty"`
		Channel   string `json:"channel,omitempty"`
		Error     string `json:"error,omitempty"`
	}](c.Client, "GET", ChatGetPermalink, q, nil)
	if err != nil {
		return "", err
	}

	if !resp.OK {
		return "", fmt.Errorf("failed to get permalink: %s", resp.Error)
	}

	return resp.Permalink, nil
}
