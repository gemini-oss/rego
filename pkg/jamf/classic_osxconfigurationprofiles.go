/*
# Jamf - Configuration Profiles

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
	ConfigurationProfiles = fmt.Sprintf("%s/osxconfigurationprofiles", "%s") // /osxconfigurationprofiles
)

/*
 * # List All Configuration Profiles
 * /osxconfigurationprofiles
 * - https://developer.jamf.com/jamf-pro/reference/findosxconfigurationprofiles
 */
func (c *Client) ListAllConfigurationProfiles() (*OSXConfigurationProfiles, error) {
	url := c.BuildClassicURL(ConfigurationProfiles)

	var cache OSXConfigurationProfiles
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[OSXConfigurationProfiles](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, osxCP, 5*time.Minute)
	return &osxCP, nil
}

/*
 * # Get Configuration Profile by ID
 * /osxconfigurationprofiles/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/findosxconfigurationprofiles
 */
func (c *Client) GetConfigurationProfileDetails(id string) (*OSXConfigurationProfile, error) {
	url := c.BuildClassicURL(ConfigurationProfiles, "id", id)

	var cache OSXConfigurationProfile
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[OSXConfigurationProfile](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, osxCP, 5*time.Minute)
	return &osxCP, nil
}

/*
 * # Update Configuration Profile by ID
 * /osxconfigurationprofiles/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/updateosxconfigurationprofilebyid
 */
func (c *Client) UpdateConfigurationProfile(id string) (*OSXConfigurationProfile, error) {
	url := c.BuildClassicURL(ConfigurationProfiles, "id", id)

	var cache OSXConfigurationProfile
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[*OSXConfigurationProfile](c, "PUT", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, osxCP, 5*time.Minute)
	return osxCP, nil
}
