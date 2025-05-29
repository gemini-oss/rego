/*
# Atlassian

This package initializes all the methods for functions which interact with the Organizations REST APIs:

* User Provisioning REST API (SCIM)
- https://developer.atlassian.com/cloud/admin/user-provisioning/rest/intro/#uri

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/atlassian/organizations.go
package atlassian

import (
	"fmt"
	"time"
)

var (
	UserProvisioning        = fmt.Sprintf("%s/scim", CloudAdmin)                             // https://developer.atlassian.com/cloud/admin/user-provisioning/rest/intro/#org-api-token-uri
	UserProvisioningUsers   = fmt.Sprintf("%s/directory/%s/Users", UserProvisioning, "%s")   // https://developer.atlassian.com/cloud/admin/user-provisioning/rest/api-group-users/#api-group-users
	UserProvisioningGroups  = fmt.Sprintf("%s/directory/%s/Groups", UserProvisioning, "%s")  // https://developer.atlassian.com/cloud/admin/user-provisioning/rest/api-group-groups/#api-group-groups
	UserProvisioningSchemas = fmt.Sprintf("%s/directory/%s/Schemas", UserProvisioning, "%s") // https://developer.atlassian.com/cloud/admin/user-provisioning/rest/api-group-schemas/#api-group-schemas
)

// UserProvisioningClient for chaining methods
type UserProvisioningClient struct {
	*Client
}

// Entry point for user-related operations
func (c *Client) UserProvisioning() *UserProvisioningClient {
	up := &UserProvisioningClient{
		Client: c,
	}

	return up
}

/*
 * # Get all users
 * /scim/directory/{directoryId}/Users
 * - https://developer.atlassian.com/cloud/admin/user-provisioning/rest/api-group-users/#api-scim-directory-directoryid-users-get
 */
func (c *UserProvisioningClient) ListAllUsers(directoryID string) (*any, error) {
	url := c.BuildURL(UserProvisioningUsers, directoryID)

	var cache any
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	users, err := do[any](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, users, 30*time.Minute)
	return &users, nil
}
