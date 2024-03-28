/*
# Okta Groups

This package contains all the methods to interact with the Okta Groups API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/#tag/Group

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/groups.go
package okta

import (
	"time"
)

/*
 * Query Parameters for Groups
 */
type GroupParameters struct {
	Q         string `json:"q,omitempty"`         // Searches the name property of groups for matching value.
	After     string `json:"after,omitempty"`     // Specifies the pagination cursor for the next page of groups.
	Expand    string `json:"expand,omitempty"`    // If specified, it causes additional metadata to be included in the response.
	Filter    string `json:"filter,omitempty"`    // Filter expression for groups.
	Limit     int32  `json:"limit,omitempty"`     // Default: (10000 for `Groups`) and (50 for Group Rules) . Specifies the number of group results in a page.
	Search    string `json:"search,omitempty"`    // Searches for groups with a supported filtering expression for all attributes except for _embedded, _links, and objectClass.
	SortBy    string `json:"sortBy,omitempty"`    // Specifies field to sort by and can be any single property (for search queries only).
	SortOrder string `json:"sortOrder,omitempty"` // Specifies sort order asc or desc (for search queries only). This parameter is ignored if sortBy is not present. Groups with the same value for the sortBy parameter are ordered by id.
}

/*
 * # Get All Groups
 * /api/v1/groups
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/#tag/Group/operation/listGroups
 */
func (c *Client) ListAllGroups() (*Groups, error) {
	c.Log.Println("Getting all groups")
	url := c.BuildURL(OktaGroups)

	q := GroupParameters{
		Limit: 10000,
	}

	groups, err := doPaginated[Groups](c, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, groups, 5*time.Minute)
	return groups, nil
}

/*
 * # Get Group by ID
 * /api/v1/groups/{groupId}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/#tag/Group/operation/getGroup
 */
func (c *Client) GetGroup(groupID string) (*Group, error) {
	c.Log.Printf("Getting group with ID %s", groupID)
	url := c.BuildURL(OktaGroups, groupID)

	var cache Group
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	group, err := do[Group](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, group, 5*time.Minute)
	return &group, nil
}

/*
 * # List All Group Rules
 * /api/v1/groups/rules
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/#tag/Group/operation/listGroupRules
 */
func (c *Client) ListAllGroupRules() (*GroupRules, error) {
	c.Log.Println("Getting all group rules")
	url := c.BuildURL(OktaGroupRules)

	var cache GroupRules
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := GroupParameters{
		Limit: 50,
	}

	groupRules, err := doPaginated[GroupRules](c, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, groupRules, 30*time.Minute)
	return groupRules, nil
}
