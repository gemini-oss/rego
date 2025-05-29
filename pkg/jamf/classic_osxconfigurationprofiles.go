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

// ProfilesClient for chaining methods
type ProfilesClient struct {
	client *Client
}

// Entry point for web-related operations
func (c *Client) Profiles() *ProfilesClient {
	return &ProfilesClient{
		client: c,
	}
}

/*
 * # List All Configuration Profiles
 * /osxconfigurationprofiles
 * - https://developer.jamf.com/jamf-pro/reference/findosxconfigurationprofiles
 */
func (pc *ProfilesClient) ListAllConfigurationProfiles() (*OSXConfigurationProfiles, error) {
	url := pc.client.BuildClassicURL(ConfigurationProfiles)

	var cache OSXConfigurationProfiles
	if pc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[OSXConfigurationProfiles](pc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	pc.client.SetCache(url, osxCP, 5*time.Minute)
	return &osxCP, nil
}

/*
 * # Get Configuration Profile by ID
 * /osxconfigurationprofiles/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/findosxconfigurationprofiles
 */
func (pc *ProfilesClient) GetConfigurationProfileDetails(id string) (*OSXConfigurationProfile, error) {
	url := pc.client.BuildClassicURL(ConfigurationProfiles, "id", id)

	var cache OSXConfigurationProfile
	if pc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[OSXConfigurationProfile](pc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	pc.client.SetCache(url, osxCP, 5*time.Minute)
	return &osxCP, nil
}

/*
 * # Update Configuration Profile by ID
 * /osxconfigurationprofiles/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/updateosxconfigurationprofilebyid
 */
func (pc *ProfilesClient) UpdateConfigurationProfile(id string) (*OSXConfigurationProfile, error) {
	url := pc.client.BuildClassicURL(ConfigurationProfiles, "id", id)

	var cache OSXConfigurationProfile
	if pc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[*OSXConfigurationProfile](pc.client, "PUT", url, nil, nil)
	if err != nil {
		return nil, err
	}

	pc.client.SetCache(url, osxCP, 5*time.Minute)
	return osxCP, nil
}
