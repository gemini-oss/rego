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

	"github.com/gemini-oss/rego/pkg/common/log"
)

var (
	ClassicUserGroups = fmt.Sprintf("%s/usergroups", "%s") // /usergroups
)

// UserClient for chaining methods
type UserGroupsClient struct {
	baseClient *Client
	Log        *log.Logger
}

// Entry point for web-related operations
func (c *Client) UserGroups() *UserGroupsClient {
	return &UserGroupsClient{
		baseClient: c,
		Log:        c.Log,
	}
}

/*
 * # List All User Groups
 * /usergroups
 * - https://developer.jamf.com/jamf-pro/reference/findusergroups
 */
func (ugc *UserGroupsClient) ListAllUserGroups() (*UserGroups, error) {
	url := ugc.baseClient.BuildClassicURL(ClassicUserGroups)

	var cache UserGroups
	if ugc.baseClient.GetCache(url, &cache) {
		return &cache, nil
	}

	userGroups, err := do[UserGroups](ugc.baseClient, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	ugc.baseClient.SetCache(url, userGroups, 5*time.Minute)
	return &userGroups, nil
}

/*
 * # Get User Group by ID
 * /usergroups/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/findusergroupsbyid
 */
func (ugc *UserGroupsClient) GetUserGroup(id string) (*UserGroup, error) {
	url := ugc.baseClient.BuildClassicURL(ClassicUserGroups, "id", id)

	var cache UserGroup
	if ugc.baseClient.GetCache(url, &cache) {
		return &cache, nil
	}

	userGroup, err := do[UserGroup](ugc.baseClient, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	ugc.baseClient.SetCache(url, userGroup, 5*time.Minute)
	return &userGroup, nil
}

/*
 * # Get User Group by Name
 * /usergroups/name/{name}
 * - https://developer.jamf.com/jamf-pro/reference/findusergroupsbyname
 */
func (ugc *UserGroupsClient) GetUserGroupByName(name string) (*UserGroup, error) {
	url := ugc.baseClient.BuildClassicURL(ClassicUserGroups, "name", name)

	var cache UserGroup
	if ugc.baseClient.GetCache(url, &cache) {
		return &cache, nil
	}

	userGroup, err := do[UserGroup](ugc.baseClient, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	ugc.baseClient.SetCache(url, userGroup, 5*time.Minute)
	return &userGroup, nil
}

/*
 * # Create User Group
 * /usergroups/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/createusergroupsbyid
 */
func (ugc *UserGroupsClient) CreateUserGroup(userGroup *UserGroup) error {
	url := ugc.baseClient.BuildClassicURL(ClassicUserGroups, "id", -1)

	userGroupBody := struct {
		XMLName xml.Name `xml:"user_group"`
		*UserGroup
	}{
		UserGroup: userGroup,
	}

	_, err := do[any](ugc.baseClient, "POST", url, nil, userGroupBody)
	if err != nil {
		return err
	}

	return nil
}

/*
 * # Update User Group by ID
 * /usergroups/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/updateusergroupsbyid
 */
func (ugc *UserGroupsClient) UpdateUserGroup(id int, users *Users) error {
	url := ugc.baseClient.BuildClassicURL(ClassicUserGroups, "id", id)

	var userGroupBody struct {
		XMLName       xml.Name `xml:"user_group"`
		UserAdditions []int    `xml:"user_additions>user>id"`
	}

	for _, user := range *users.List {
		userGroupBody.UserAdditions = append(userGroupBody.UserAdditions, user.ID)
	}
	if len(userGroupBody.UserAdditions) == 0 {
		return fmt.Errorf("no users to add to user group")
	}

	_, err := do[any](ugc.baseClient, "POST", url, nil, userGroupBody)
	if err != nil {
		return err
	}

	return nil
}
