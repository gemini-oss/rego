/*
# Jamf - User Groups

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/classic_usergroups.go
package jamf

import (
	"encoding/xml"
	"fmt"
	"time"
)

var (
	ClassicUserGroups = fmt.Sprintf("%s/usergroups", "%s") // /usergroups
)

// UserClient for chaining methods
type UserGroupsClient struct {
	client *Client
}

// Entry point for web-related operations
func (c *Client) UserGroups() *UserGroupsClient {
	return &UserGroupsClient{
		client: c,
	}
}

/*
 * # List All User Groups
 * /usergroups
 * - https://developer.jamf.com/jamf-pro/reference/findusergroups
 */
func (ugc *UserGroupsClient) ListAllUserGroups() (*UserGroups, error) {
	url := ugc.client.BuildClassicURL(ClassicUserGroups)

	var cache UserGroups
	if ugc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	userGroups, err := do[UserGroups](ugc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	ugc.client.SetCache(url, userGroups, 5*time.Minute)
	return &userGroups, nil
}

/*
 * # Get User Group by ID
 * /usergroups/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/findusergroupsbyid
 */
func (ugc *UserGroupsClient) GetUserGroup(id string) (*UserGroup, error) {
	url := ugc.client.BuildClassicURL(ClassicUserGroups, "id", id)

	var cache UserGroup
	if ugc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	userGroup, err := do[UserGroup](ugc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	ugc.client.SetCache(url, userGroup, 5*time.Minute)
	return &userGroup, nil
}

/*
 * # Get User Group by Name
 * /usergroups/name/{name}
 * - https://developer.jamf.com/jamf-pro/reference/findusergroupsbyname
 */
func (ugc *UserGroupsClient) GetUserGroupByName(name string) (*UserGroup, error) {
	url := ugc.client.BuildClassicURL(ClassicUserGroups, "name", name)

	var cache UserGroup
	if ugc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	userGroup, err := do[UserGroup](ugc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	ugc.client.SetCache(url, userGroup, 5*time.Minute)
	return &userGroup, nil
}

/*
 * # Create User Group
 * /usergroups/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/createusergroupsbyid
 */
func (ugc *UserGroupsClient) CreateUserGroup(userGroup *UserGroup) error {
	url := ugc.client.BuildClassicURL(ClassicUserGroups, "id", -1)

	userGroupBody := struct {
		XMLName xml.Name `xml:"user_group"`
		*UserGroup
	}{
		UserGroup: userGroup,
	}

	_, err := do[any](ugc.client, "POST", url, nil, userGroupBody)
	if err != nil {
		return err
	}

	return nil
}
