/*
# Jamf - Users

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/classic_osxconfigurationprofiles.go
package jamf

import (
	"fmt"
	"time"
)

var (
	ClassicUsers = fmt.Sprintf("%s/users", "%s") // /users
)

/*
 * # List All Users
 * /users
 * - https://developer.jamf.com/jamf-pro/reference/findusers
 */
func (c *Client) ListAllUsers() (*Users, error) {
	url := c.BuildClassicURL(ClassicUsers)

	var cache Users
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	users, err := do[Users](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, users, 5*time.Minute)
	return &users, nil
}

/*
 * # Get User by ID
 * /users/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/findusersbyid
 */
func (c *Client) GetUser(id string) (*Users, error) {
	url := c.BuildClassicURL(ClassicUsers, "id", id)

	return c.getUser(url)
}

/*
 * # Get User Profile by Email
 * /users/email/{email}
 * - https://developer.jamf.com/jamf-pro/reference/findusersbyemailaddress
 */
func (c *Client) GetUserByEmail(email string) (*Users, error) {
	url := c.BuildClassicURL(ClassicUsers, "email", email)

	return c.getUser(url)
}

func (c *Client) getUser(url string) (*Users, error) {

	var cache Users
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	user, err := do[Users](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, user, 5*time.Minute)
	return &user, nil
}
