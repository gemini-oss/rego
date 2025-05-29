/*
# Jamf - Admin

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/admin.go
package jamf

import (
	"fmt"
	"time"
)

var (
	AdminUsers          = fmt.Sprintf("%s/computers-inventory", V1)        // /v2/computers-inventory
	AdminUserPrivileges = fmt.Sprintf("%s/computers-inventory-detail", V1) // /api/v2/users/{id}/privileges
)

// AdminClient for chaining methods
type AdminClient struct {
	client *Client
}

// Entry point for web-related operations
func (c *Client) Admin() *AdminClient {
	return &AdminClient{
		client: c,
	}
}

/*
 * # Get Users
 * /v2/users
 * - https://developer.jamf.com/jamf-pro/reference/getusers
 */
func (ac *AdminClient) ListAllUsers() (*JamfUsers, error) {
	url := ac.client.BuildURL(ComputersInventory)

	var cache JamfUsers
	if ac.client.GetCache(url, &cache) {
		return &cache, nil
	}

	users, err := do[JamfUsers](ac.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	ac.client.SetCache(url, users, 5*time.Minute)
	return &users, nil
}

/*
 * # Get User Privileges
 * /v2/users/{id}/privileges
 * - https://developer.jamf.com/jamf-pro/reference/getuserprivileges
 */
