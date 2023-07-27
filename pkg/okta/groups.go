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
	"encoding/json"
	"fmt"
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
	c.Logger.Println("Getting all groups")
	allGroups := Groups{}

	q := GroupParameters{
		Limit: 10000,
	}

	url := c.BuildURL(OktaGroups)
	res, err := c.HTTPClient.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Printf("Received response from %s", url)

	for _, r := range res {
		group := &Group{}
		err := json.Unmarshal(r, &group)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling group: %w", err)
		}
		allGroups = append(allGroups, group)
	}

	return &allGroups, nil
}

/*
 * # Get Group by ID
 * /api/v1/groups/{groupId}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/#tag/Group/operation/getGroup
 */
func (c *Client) GetGroup(groupID string) (*Group, error) {
	c.Logger.Printf("Getting group with ID %s", groupID)
	group := &Group{}

	url := c.BuildURL(OktaGroups, groupID)
	res, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Printf("Received response from %s. Status: %s", url, res.Status)

	err = json.Unmarshal(body, &group)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling group: %w", err)
	}

	return group, nil
}

/*
 * # List All Group Rules
 * /api/v1/groups/rules
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/#tag/Group/operation/listGroupRules
 */
func (c *Client) ListAllGroupRules() (*GroupRules, error) {
	c.Logger.Println("Getting all group rules")
	allGroupRules := GroupRules{}

	q := GroupParameters{
		Limit: 50,
	}

	url := c.BuildURL(OktaGroupRules)
	res, err := c.HTTPClient.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Printf("Received response from %s", url)

	for _, r := range res {
		groupRule := &GroupRule{}
		err := json.Unmarshal(r, &groupRule)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling group rule: %w", err)
		}
		allGroupRules = append(allGroupRules, groupRule)
	}

	return &allGroupRules, nil
}
