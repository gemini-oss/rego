/*
# Slack - Audit Logs (Enterprise Grid - Organization-wide)

This package contains methods for the Audit Logs API with Slack's Enterprise Grid:
https://docs.slack.dev/reference/audit-logs-api/

The Audit Logs API is a read-only API that provides visibility into actions taken
across an Enterprise Grid organization. Every event is composed of an actor taking
an action on an entity within a context.

Requirements:
- User OAuth Token (xoxp) with `auditlogs:read` scope
- Token owner must be an Enterprise Grid Organization Owner
- Enterprise Grid plan

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/audit.go
package slack

import (
	"fmt"
	"strings"
	"time"

	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	// Audit Logs endpoints (Tier 3 - 50/min, org-wide)
	// https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
	AuditLogs    = "%s/logs"    // GET https://api.slack.com/audit/v1/logs
	AuditSchemas = "%s/schemas" // GET https://api.slack.com/audit/v1/schemas
	AuditActions = "%s/actions" // GET https://api.slack.com/audit/v1/actions
)

// AuditClient for Audit Logs API operations (organization-wide audit trail)
type AuditClient struct {
	*Client
	AuditBaseURL string
}

// AuditQuery for GET /audit/v1/logs
// https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
// All filters are optional and combined with AND logic
type AuditQuery struct {
	Latest int64  `url:"latest,omitempty"` // Unix timestamp (inclusive) for most recent event
	Oldest int64  `url:"oldest,omitempty"` // Unix timestamp (inclusive) for oldest event
	Limit  int    `url:"limit,omitempty"`  // Results to return, maximum 9,999
	Action string `url:"action,omitempty"` // Action name(s), comma-separated (max 30)
	Actor  string `url:"actor,omitempty"`  // User ID who initiated the action
	Entity string `url:"entity,omitempty"` // Target entity ID (channel, workspace, file, etc.)
	Cursor string `url:"cursor,omitempty"` // Pagination cursor from response_metadata.next_cursor
}

func (q *AuditQuery) SetCursor(cursor string) {
	q.Cursor = cursor
}

// BuildURL builds a URL for a given Audit Logs resource and identifiers.
func (c *AuditClient) BuildURL(endpoint string, identifiers ...interface{}) string {
	url := fmt.Sprintf(endpoint, c.AuditBaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%v", url, id)
	}
	return url
}

// Audit returns the audit sub-client for Audit Logs API operations
// https://docs.slack.dev/reference/audit-logs-api/
func (c *Client) Audit() *AuditClient {
	if c.auditClient != nil {
		return c.auditClient
	}

	// Audit Logs API is Tier 3 (50/min), applied org-wide (not per-app)
	// https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
	auditRL := ratelimit.NewRateLimiter(50, 1*time.Minute)
	auditRL.ResetHeaders = true
	auditRL.Log.Verbosity = c.Log.Verbosity

	auditHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), auditRL)
	auditHTTP.BodyType = requests.JSON

	auditClient := &Client{
		BaseURL:       c.BaseURL,
		HTTP:          auditHTTP,
		Log:           c.Log,
		Token:         c.Token,
		SigningSecret: c.SigningSecret,
		Cache:         c.Cache,
	}

	c.auditClient = &AuditClient{
		Client:       auditClient,
		AuditBaseURL: AuditBaseURL,
	}
	return c.auditClient
}

/*
 * # Retrieve audit log events with optional filtering
 * GET /audit/v1/logs
 * - https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
 *
 * Returns a single page of audit events matching the query filters.
 * Use GetAllLogs for automatic pagination through all results.
 */
func (c *AuditClient) GetLogs(query *AuditQuery) (*AuditLogsResponse, error) {
	url := c.BuildURL(AuditLogs)

	if query == nil {
		query = &AuditQuery{}
	}

	resp, err := do[AuditLogsResponse](c.Client, "GET", url, query, nil)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

/*
 * # Retrieve all audit log events with automatic cursor-based pagination
 * GET /audit/v1/logs
 * - https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
 *
 * Automatically paginates through all audit events matching the query filters.
 * Each page returns up to 9,999 events (default ordering: newest first).
 */
func (c *AuditClient) GetAllLogs(query *AuditQuery) (*AuditLogsResponse, error) {
	url := c.BuildURL(AuditLogs)

	if query == nil {
		query = &AuditQuery{}
	}

	// Default to max page size for efficiency
	if query.Limit == 0 {
		query.Limit = 9999
	}

	logs, err := doPaginated[AuditLogsResponse](c.Client, "GET", url, query, nil)
	if err != nil {
		return nil, err
	}

	c.Log.Println(fmt.Sprintf("Audit: Retrieved %d total log entries", len(logs.Entries)))
	return logs, nil
}

/*
 * # List available audit log schemas
 * GET /audit/v1/schemas
 * - https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
 *
 * Returns information about the object types available in the Audit Logs API.
 */
func (c *AuditClient) GetSchemas() (*AuditSchemasResponse, error) {
	url := c.BuildURL(AuditSchemas)
	cacheKey := "AuditSchemas"

	var cache AuditSchemasResponse
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	resp, err := do[AuditSchemasResponse](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, &resp, 24*time.Hour)
	return &resp, nil
}

/*
 * # List available audit log actions
 * GET /audit/v1/actions
 * - https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
 *
 * Returns all available actions grouped by category.
 */
func (c *AuditClient) GetActions() (*AuditActionsResponse, error) {
	url := c.BuildURL(AuditActions)
	cacheKey := "AuditActions"

	var cache AuditActionsResponse
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	resp, err := do[AuditActionsResponse](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, &resp, 24*time.Hour)
	return &resp, nil
}

/*
 * # Retrieve emoji change events within a time range
 * GET /audit/v1/logs?action=emoji_added,emoji_removed
 * - https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
 *
 * Returns all emoji_added and emoji_removed audit events between the given
 * Unix timestamps. Automatically paginates through all matching results.
 *
 * Example: Get emoji changes on February 5th, 2026:
 *
 *   start := time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC)
 *   end := time.Date(2026, time.February, 5, 23, 59, 59, 0, time.UTC)
 *   logs, err := client.Audit().GetEmojiChanges(start.Unix(), end.Unix())
 */
func (c *AuditClient) GetEmojiChanges(oldest, latest int64) (*AuditLogsResponse, error) {
	q := &AuditQuery{
		Action: "emoji_added,emoji_removed",
		Oldest: oldest,
		Latest: latest,
		Limit:  9999,
	}

	return c.GetAllLogs(q)
}

/*
 * # Search for a specific emoji by name in audit logs
 * GET /audit/v1/logs?action=emoji_added,emoji_removed
 * - https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
 *
 * Retrieves all emoji_added and emoji_removed events within the time range,
 * then filters client-side for entries matching the given emoji name.
 * The Audit Logs API does not support server-side filtering by details.name.
 *
 * The name should be the emoji shortcode without colons (e.g., "american_flag" not ":american_flag:").
 *
 * Example: Find who added/removed :american_flag: on February 5th, 2026:
 *
 *   start := time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC)
 *   end := time.Date(2026, time.February, 5, 23, 59, 59, 0, time.UTC)
 *   logs, err := client.Audit().FindEmojiByName("american_flag", start.Unix(), end.Unix())
 */
func (c *AuditClient) FindEmojiByName(name string, oldest, latest int64) (*AuditLogsResponse, error) {
	all, err := c.GetEmojiChanges(oldest, latest)
	if err != nil {
		return nil, err
	}

	var filtered []AuditEntry
	for _, entry := range all.Entries {
		if entry.Details != nil && entry.Details.Name == name {
			filtered = append(filtered, entry)
		}
	}

	c.Log.Println(fmt.Sprintf("Audit: Filtered %d/%d entries matching emoji %q", len(filtered), len(all.Entries), name))

	return &AuditLogsResponse{
		Entries: filtered,
	}, nil
}

/*
 * # Search audit log events by user agent substring
 * GET /audit/v1/logs
 * - https://docs.slack.dev/reference/audit-logs-api/methods-actions-reference
 *
 * Retrieves all audit events within the time range, then filters client-side
 * for entries where the context user agent contains the given pattern (case-insensitive).
 *
 * Example: Find all actions performed from an iPad on February 5th, 2026:
 *
 *   start := time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC)
 *   end := time.Date(2026, time.February, 5, 23, 59, 59, 0, time.UTC)
 *   logs, err := client.Audit().FindByUserAgent("iPad", start.Unix(), end.Unix())
 */
func (c *AuditClient) FindByUserAgent(pattern string, oldest, latest int64) (*AuditLogsResponse, error) {
	all, err := c.GetAllLogs(&AuditQuery{
		Oldest: oldest,
		Latest: latest,
	})
	if err != nil {
		return nil, err
	}

	lowerPattern := strings.ToLower(pattern)
	var filtered []AuditEntry
	for _, entry := range all.Entries {
		if strings.Contains(strings.ToLower(entry.Context.UA), lowerPattern) {
			filtered = append(filtered, entry)
		}
	}

	c.Log.Println(fmt.Sprintf("Audit: Filtered %d/%d entries matching user agent %q", len(filtered), len(all.Entries), pattern))

	return &AuditLogsResponse{
		Entries: filtered,
	}, nil
}
