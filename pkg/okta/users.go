/*
# Okta Users

This package contains all the methods to interact with the Okta Users API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/users.go
package okta

import (
	"time"
)

/*
 * Query Parameters for Users
 */
type UserQuery struct {
	Q         string // Searches the records for matching value
	After     string // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the `Link` response header.
	Limit     string // Default: 200. Specifies the number of results returned. Defaults to 10 if `q` is provided
	Filter    string // Filters users with a supported expression for a subset of properties
	Search    string // A SCIM filter expression for most properties. Okta recommends using this parameter for search for best performance
	SortBy    string // Specifies the attribute by which to sort the results. Valid values are `id`, `created`, `activated`, `status`, and `lastUpdated`. The default is `id`
	SoftOrder string // Sorting is done in ASCII sort order (that is, by ASCII character value), but isn't case sensitive
}

/*
 * # Get all users, regardless of status
 * /api/v1/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers
 */
func (c *Client) ListAllUsers() (*Users, error) {
	url := c.BuildURL(OktaUsers)

	var cache Users
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := &UserQuery{
		Limit:  `200`,
		Search: `status eq "STAGED" or status eq "PROVISIONED" or status eq "ACTIVE" or status eq "RECOVERY" or status eq "LOCKED_OUT" or status eq "PASSWORD_EXPIRED" or status eq "SUSPENDED" or status eq "DEPROVISIONED"`,
	}

	users, err := doPaginated[Users](c, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, users, 30*time.Minute)
	return users, nil
}

/*
 * # List all ACTIVE users
 * /api/v1/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers
 */
func (c *Client) ListActiveUsers() (*Users, error) {
	url := c.BuildURL(OktaUsers)

	var cache Users
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := &UserQuery{
		Limit:  `200`,
		Search: `status eq "ACTIVE"`,
	}

	users, err := doPaginated[Users](c, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, users, 30*time.Minute)
	return users, nil
}

/*
 * # Get a user by ID
 * /api/v1/users/{userId}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/getUser
 */
func (c *Client) GetUser(userID string) (*User, error) {
	url := c.BuildURL(OktaUsers, userID)

	var cache User
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	user, err := do[User](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, user, 5*time.Minute)
	return &user, nil
}

/*
 * # Update a user's properties by ID
 * /api/v1/users/{userId}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/updateUser
 */
func (c *Client) UpdateUser(userID string, u *User) (*User, error) {

	url := c.BuildURL(OktaUsers, userID)

	user, err := do[User](c, "POST", url, nil, &u)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

/*
 * # Get all Assigned Application Links for a User
 * /api/v1/users/{userId}/appLinks
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listAppLinks
 */
func (c *Client) GetUserAppLinks(userID string) (*AppLinks, error) {
	url := c.BuildURL(OktaUsers, userID, "appLinks")

	var cache AppLinks
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	appLinks, err := do[AppLinks](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, appLinks, 5*time.Minute)
	return &appLinks, nil
}

/*
 * # List all Groups for a User
 * /api/v1/users/{userId}/groups
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/updateUser
 */
func (c *Client) GetUserGroups(userID string) (*Groups, error) {
	url := c.BuildURL(OktaUsers, userID, "groups")

	var cache Groups
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	groups, err := do[Groups](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, groups, 5*time.Minute)
	return &groups, nil
}
