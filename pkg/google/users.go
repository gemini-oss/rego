/*
# Google Workspace - Admin

This package implements logic related to the `Users` resource of the Google Admin SDK API:
https://developers.google.com/admin-sdk/directory/reference/rest/v1/users

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/users.go
package google

import (
	"encoding/json"
	"fmt"
)

/*
 * Retrieves a User's Profile
 * /admin/directory/v1/users/{userKey}
 * https://developers.google.com/admin-sdk/directory/v1/reference/users/get
 */
func (c *Client) GetUser(userKey string) (*User, error) {
	url := fmt.Sprintf(DirectoryUsers+"/%s", userKey)
	c.Logger.Debug("url:", url)

	res, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status: ", res.Status)
	c.Logger.Debug("Response Body: ", string(body))

	user := &User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
