/*
# Slack - Admin (Enterprise Grid)

This package contains methods for Enterprise Grid admin operations with the Slack Web API:
https://api.slack.com/methods#admin

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/admin.go
package slack

import (
	"fmt"
	"time"

	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	// Admin User endpoints (Tier 2 - 20/min)
	// https://api.slack.com/methods#admin.users
	AdminUsersList       = fmt.Sprintf("%s/admin.users.list", BaseURL)
	AdminUsersAssign     = fmt.Sprintf("%s/admin.users.assign", BaseURL)
	AdminUsersInvite     = fmt.Sprintf("%s/admin.users.invite", BaseURL)
	AdminUsersRemove     = fmt.Sprintf("%s/admin.users.remove", BaseURL)
	AdminUsersSetAdmin   = fmt.Sprintf("%s/admin.users.setAdmin", BaseURL)
	AdminUsersSetOwner   = fmt.Sprintf("%s/admin.users.setOwner", BaseURL)
	AdminUsersSetRegular = fmt.Sprintf("%s/admin.users.setRegular", BaseURL)

	// Admin Session endpoints
	// https://api.slack.com/methods#admin.users.session
	AdminUsersSessionReset      = fmt.Sprintf("%s/admin.users.session.reset", BaseURL)
	AdminUsersSessionResetBulk  = fmt.Sprintf("%s/admin.users.session.resetBulk", BaseURL)
	AdminUsersSessionInvalidate = fmt.Sprintf("%s/admin.users.session.invalidate", BaseURL)
	AdminUsersSessionList       = fmt.Sprintf("%s/admin.users.session.list", BaseURL)

	// Admin Teams endpoints
	// https://api.slack.com/methods#admin.teams
	AdminTeams       = fmt.Sprintf("%s/admin.teams.list", BaseURL)
	AdminTeamsCreate = fmt.Sprintf("%s/admin.teams.create", BaseURL)
)

// AdminClient for Enterprise Grid admin operations
type AdminClient struct {
	*Client
}

// AdminUserQuery for admin.users.list
// https://api.slack.com/methods/admin.users.list
type AdminUserQuery struct {
	Cursor                           string `json:"cursor,omitempty" url:"cursor,omitempty"`                                                           // Pagination cursor
	IsActive                         *bool  `json:"is_active,omitempty" url:"is_active,omitempty"`                                                     // Filter by active status
	Limit                            int    `json:"limit,omitempty" url:"limit,omitempty"`                                                             // Max 1000, default 100
	TeamID                           string `json:"team_id,omitempty" url:"team_id,omitempty"`                                                         // Required: Workspace ID
	IncludeDeactivatedUserWorkspaces bool   `json:"include_deactivated_user_workspaces,omitempty" url:"include_deactivated_user_workspaces,omitempty"` // Include deactivated
}

func (q *AdminUserQuery) SetCursor(cursor string) {
	q.Cursor = cursor
}

// Functional options for SessionReset
func WithMobileOnly() SessionResetOption {
	return func(r *SessionReset) { r.MobileOnly = true }
}

func WithWebOnly() SessionResetOption {
	return func(r *SessionReset) { r.WebOnly = true }
}

// Admin returns the admin sub-client for Enterprise Grid operations
func (c *Client) Admin() *AdminClient {
	if c.adminClient != nil {
		return c.adminClient
	}

	// Admin APIs are Tier 2 (20/min)
	// https://api.slack.com/docs/rate-limits
	adminRL := ratelimit.NewRateLimiter(20, 1*time.Minute)
	adminRL.ResetHeaders = true
	adminRL.Log.Verbosity = c.Log.Verbosity

	adminHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), adminRL)
	adminHTTP.BodyType = requests.FormURLEncoded

	adminClient := &Client{
		BaseURL:       c.BaseURL,
		HTTP:          adminHTTP,
		Log:           c.Log,
		Token:         c.Token,
		SigningSecret: c.SigningSecret,
		Cache:         c.Cache,
	}

	c.adminClient = &AdminClient{Client: adminClient}
	return c.adminClient
}

/*
 * # List all users in an Enterprise Grid organization
 * admin.users.list
 * - https://api.slack.com/methods/admin.users.list
 */
func (c *AdminClient) ListAllUsers(teamID string) (*AdminUsers, error) {
	cacheKey := fmt.Sprintf("%s_%s", AdminUsersList, teamID)

	var cache AdminUsers
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	q := &AdminUserQuery{
		TeamID: teamID,
		Limit:  100,
	}

	// c.Client.HTTP.BodyType = requests.JSON
	users, err := doPaginated[AdminUsers](c.Client, "POST", AdminUsersList, &AdminUserQuery{}, q)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return users, nil
}

/*
 * # List active users only in an Enterprise Grid organization
 * admin.users.list
 * - https://api.slack.com/methods/admin.users.list
 */
func (c *AdminClient) ListActiveUsers(teamID string) (*AdminUsers, error) {
	cacheKey := fmt.Sprintf("%s_%s_active", AdminUsersList, teamID)

	var cache AdminUsers
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	isActive := true
	q := &AdminUserQuery{
		TeamID:   teamID,
		Limit:    200,
		IsActive: &isActive,
	}

	users, err := doPaginated[AdminUsers](c.Client, "POST", AdminUsersList, &AdminUserQuery{}, q)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return users, nil
}

/*
 * # Reset a single user's sessions
 * admin.users.session.reset
 * - https://api.slack.com/methods/admin.users.session.reset
 */
func (c *AdminClient) ResetUserSession(userID string, opts ...SessionResetOption) error {
	req := &SessionReset{UserID: userID}
	for _, opt := range opts {
		opt(req)
	}

	resp, err := do[SlackOK](c.Client, "POST", AdminUsersSessionReset, nil, req)
	if err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("session reset failed: %s", resp.Error)
	}
	return nil
}

/*
 * # Reset sessions for multiple users in bulk
 * admin.users.session.resetBulk
 * - https://api.slack.com/methods/admin.users.session.resetBulk
 */
func (c *AdminClient) ResetUserSessionsBulk(userIDs []string, opts ...SessionResetOption) error {
	req := &SessionReset{UserIDs: userIDs}
	for _, opt := range opts {
		opt(req)
	}

	resp, err := do[SlackOK](c.Client, "POST", AdminUsersSessionResetBulk, nil, req)
	if err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("bulk session reset failed: %s", resp.Error)
	}
	return nil
}
