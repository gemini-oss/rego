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

/*
 * # Get the Jamf Version
 * /api/v1/jamf-pro-version
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-jamf-pro-version
 */
func (c *Client) GetJamfVersion() (string, error) {
	// url := fmt.Sprintf("%s/jamf-pro-version", c.BaseURL)

	url := c.BuildURL("%s/jamf-pro-version")

	_, body, err := c.HTTP.DoRequest("GET", url, nil, nil)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf("Jamf Version Response: %s", string(body))
	return string(body), nil
}
