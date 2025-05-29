/*
# Jamf - Users

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/classic_osxconfigurationprofiles.go
package jamf

import (
	"encoding/xml"
	"fmt"
	"time"
)

var (
	ClassicUsers = fmt.Sprintf("%s/users", "%s") // /users
)

// UserClient for chaining methods
type UserClient struct {
	client *Client
}

// Entry point for web-related operations
func (c *Client) Users() *UserClient {
	return &UserClient{
		client: c,
	}
}

/*
 * # List All Users
 * /users
 * - https://developer.jamf.com/jamf-pro/reference/findusers
 */
func (uc *UserClient) ListAllUsers() (*Users, error) {
	url := uc.client.BuildClassicURL(ClassicUsers)

	var cache Users
	if uc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	users, err := do[Users](uc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	uc.client.SetCache(url, users, 5*time.Minute)
	return &users, nil
}

/*
 * # Get User by ID
 * /users/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/findusersbyid
 */
func (uc *UserClient) GetUser(id string) (*User, error) {
	url := uc.client.BuildClassicURL(ClassicUsers, "id", id)

	var cache User
	if uc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	type user struct {
		XMLName xml.Name `xml:"user"`
		User
	}

	u, err := do[user](uc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	uc.client.SetCache(url, u, 5*time.Minute)
	return &u.User, nil
}

/*
 * # Get User Profile by Email
 * /users/email/{email}
 * - https://developer.jamf.com/jamf-pro/reference/findusersbyemailaddress
 */
func (uc *UserClient) GetUsersByEmail(email string) (*Users, error) {
	url := uc.client.BuildClassicURL(ClassicUsers, "email", email)

	var cache Users
	if uc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	user, err := do[Users](uc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	uc.client.SetCache(url, user, 5*time.Minute)
	return &user, nil
}

/*
 * # Create User
 * /users/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/createuserbyid
 */
func (uc *UserClient) CreateUser(user *User) error {
	url := uc.client.BuildClassicURL(ClassicUsers, "id", -1)

	userBody := struct {
		XMLName xml.Name `xml:"user"`
		*User
	}{
		User: user,
	}

	_, err := do[any](uc.client, "POST", url, nil, userBody)
	if err != nil {
		return err
	}

	return nil
}
