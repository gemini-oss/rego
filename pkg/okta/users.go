/*
# Okta Users

This package contains all the methods to interact with the Okta Users API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/users/users.go
package okta

import (
	"encoding/json"
	"fmt"
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
 * Get all users, regardless of status
 * /api/v1/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers
 */
func (c *Client) ListAllUsers() (*Users, error) {
	c.Logger.Println("Getting all users")
	allUsers := Users{}

	q := UserQuery{
		Limit:  `200`,
		Search: `status eq "STAGED" or status eq "PROVISIONED" or status eq "ACTIVE" or status eq "RECOVERY" or status eq "LOCKED_OUT" or status eq "PASSWORD_EXPIRED" or status eq "SUSPENDED" or status eq "DEPROVISIONED"`,
	}

	// url := fmt.Sprintf("%s/users", c.BaseURL)
	url := c.BuildURL(OktaUsers)
	res, err := c.HTTPClient.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Printf("Received response from %s", url)

	for _, r := range res {
		user := User{}
		err := json.Unmarshal(r, &user)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user: %w", err)
		}
		allUsers = append(allUsers, &user)
	}

	c.Logger.Println("Successfully listed all users.")
	return &allUsers, nil
}

/*
 * List all ACTIVE users
 * /api/v1/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers
 */
func (c *Client) ListActiveUsers() (*Users, error) {

	allUsers := Users{}

	q := UserQuery{
		Limit:  `200`,
		Search: `status eq "ACTIVE"`,
	}

	// url := fmt.Sprintf("%s/users", c.BaseURL)
	url := c.BuildURL(OktaUsers)
	res, err := c.HTTPClient.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		user := User{}
		err := json.Unmarshal(r, &user)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user: %w", err)
		}
		allUsers = append(allUsers, &user)
	}

	return &allUsers, nil
}

/*
 * Get a user by ID
 * /api/v1/users/{userId}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/getUser
 */
func (c *Client) GetUser(userID string) (*User, error) {

	// url := fmt.Sprintf("%s/users/%s", c.BaseURL, userID)
	url := c.BuildURL(OktaUsers, userID)
	_, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return user, nil
}

/*
 * Get all Assigned Application Links for a User
 * /api/v1/users/{userId}/appLinks
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listAppLinks
 */
