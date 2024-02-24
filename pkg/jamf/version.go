/*
# Jamf - Version

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/version.go
package jamf

import (
	"fmt"
)

var (
	JamfProVersion = fmt.Sprintf("%s/jamf-pro-version", V1) // /api/v1/jamf-pro-version
)

/*
 * # Get the Jamf Version
 * /api/v1/jamf-pro-version
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-jamf-pro-version
 */
func (c *Client) GetJamfVersion() (string, error) {
	url := c.BuildURL(JamfProVersion)

	res, body, err := c.HTTP.DoRequest("GET", url, nil, nil)
	if err != nil {
		return "", err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debugf("Jamf Version Response: %s", string(body))
	return string(body), nil
}
