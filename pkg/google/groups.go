/*
# Google Workspace - Groups

This package implements logic related to the `Groups` resource of the Google Admin SDK API:
https://developers.google.com/admin-sdk/directory/reference/rest/v1/groups

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/groups.go
package google

import (
	"fmt"
	"time"
)

// GroupsClient for chaining methods
type GroupsClient struct {
	client *Client
}

// Entry point for groups-related operations
func (c *Client) Groups() *GroupsClient {
	return &GroupsClient{
		client: c,
	}
}

/*
 * Query Parameters for Groups
 * Reference: https://developers.google.com/admin-sdk/directory/reference/rest/v1/groups/list#query-parameters
 */
type GroupsQuery struct {
	Customer   string    `url:"customer,omitempty"`   // The unique ID for the customer's Google Workspace account. In case of a multi-domain account, to fetch all groups for a customer, use this field instead of domain. You can also use the my_customer alias to represent your account's customerId. The customerId is also returned as part of the Users resource. You must provide either the customer or the domain parameter.
	Domain     string    `url:"domain,omitempty"`     // The domain name. Use this field to get groups from only one domain. To return all domains for a customer account, use the customer query parameter instead. Either the customer or the domain parameter must be provided.
	MaxResults int       `url:"maxResults,omitempty"` // Maximum number of results to return. Default: 100. Max 500. https://developers.google.com/admin-sdk/directory/v1/limits#api-limits-and-quotas
	OrderBy    OrderBy   `url:"orderBy,omitempty"`    // Property to use for sorting results.
	PageToken  string    `url:"pageToken,omitempty"`  // Token to specify next page in the list
	Query      string    `url:"query,omitempty"`      // Query string for searching user fields. For more information on constructing user queries, see [Search for Users](https://developers.google.com/admin-sdk/directory/v1/guides/search-users).
	SortOrder  SortOrder `url:"sortOrder,omitempty"`  // Whether to return results in ascending or descending order, ignoring case.
	UserKey    string    `url:"userKey,omitempty"`    // Email or immutable ID of the user if only those groups are to be listed, the given user is a member of. If it's an ID, it should match with the ID of the user object. Cannot be used with the customer parameter.
}

func (q *GroupsQuery) SetPageToken(token string) {
	q.PageToken = token
}

/*
 * Check if the GroupsQuery is empty
 */
func (g *GroupsQuery) IsEmpty() bool {
	return g.Customer == "" &&
		g.Domain == "" &&
		g.MaxResults == 0 &&
		g.OrderBy == "" &&
		g.PageToken == "" &&
		g.Query == "" &&
		g.SortOrder == ""
}

/*
 * Validate the query parameters for the Groups resource
 */
func (u *GroupsQuery) ValidateQuery() error {
	if u.IsEmpty() {
		u.Customer = "my_customer"
		u.MaxResults = 100
		return nil
	}

	if u.Customer != "" && u.Domain != "" {
		return fmt.Errorf("cannot specify both customer and domain")
	}

	if u.Customer == "" && u.Domain == "" {
		u.Customer = "my_customer"
	}

	if u.MaxResults > 500 {
		return fmt.Errorf("maxResults cannot exceed %d", 500)
	}

	if u.MaxResults == 0 {
		u.MaxResults = 100
	}

	return nil
}

/*
 * List all groups
 * /admin/directory/v1/groups
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/groups/list
 */
func (gc *GroupsClient) ListAllGroups() (*Groups, error) {
	url := DirectoryGroups

	var cache Groups
	if gc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	q := GroupsQuery{}

	err := q.ValidateQuery()
	if err != nil {
		return nil, err
	}
	q.MaxResults = 500

	groups, err := do[Groups](gc.client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	for groups.NextPageToken != "" {
		q = GroupsQuery{
			Customer:   q.Customer,
			Domain:     q.Domain,
			MaxResults: 500,
			PageToken:  groups.NextPageToken,
		}

		groupsPage, err := do[Groups](gc.client, "GET", url, q, nil)
		if err != nil {
			return nil, err
		}
		groups.Groups = append(groups.Groups, groupsPage.Groups...)
		groups.NextPageToken = groupsPage.NextPageToken
	}

	gc.client.SetCache(url, groups, 30*time.Minute)
	return &groups, nil
}

/*
 * Search for groups based on filter conditions
 * /admin/directory/v1/groups
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/groups/list
 */
func (gc *GroupsClient) SearchGroups(q *GroupsQuery) (*Groups, error) {
	err := q.ValidateQuery()
	if err != nil {
		return nil, err
	}

	url := DirectoryGroups

	groups, err := do[Groups](gc.client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	return &groups, nil
}

/*
 * Retrieves a Groups's Profile
 * /admin/directory/v1/groups/{groupKey}
 * @param {string} groupKey - Group's email address, group alias, or the unique group ID.
 * https://developers.google.com/admin-sdk/directory/v1/reference/groups/get
 */
func (gc *GroupsClient) GetGroup(groupKey string) (*Group, error) {
	url := gc.client.BuildURL(DirectoryGroups, nil, groupKey)

	group, err := do[Group](gc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

/*
 * Update a Groups's Profile
 * /admin/directory/v1/groups/{groupKey}
 * @param {string} groupKey - Group's email address, group alias, or the unique group ID.
 * @body {struct} Group - An entity representing a group in Google Workspace.
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/groups/update
 */
func (gc *GroupsClient) UpdateGroup(groupKey string, u *Group) (*Group, error) {
	url := gc.client.BuildURL(DirectoryGroups, nil, groupKey)

	group, err := do[Group](gc.client, "PUT", url, nil, u)
	if err != nil {
		return nil, err
	}

	return &group, nil
}
