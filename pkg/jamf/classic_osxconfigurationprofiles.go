/*
# Jamf - Configuration Profiles

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
	"fmt"
	"time"

	"github.com/gemini-oss/rego/pkg/common/log"
)

var (
	ConfigurationProfiles = fmt.Sprintf("%s/osxconfigurationprofiles", "%s") // /osxconfigurationprofiles
)

// ProfilesClient for chaining methods
type ProfilesClient struct {
	baseClient *Client
	Log        *log.Logger
}

// Entry point for web-related operations
func (c *Client) Profiles() *ProfilesClient {
	return &ProfilesClient{
		baseClient: c,
	}
}

/*
 * # List All Configuration Profiles
 * /osxconfigurationprofiles
 * - https://developer.jamf.com/jamf-pro/reference/findosxconfigurationprofiles
 */
func (pc *ProfilesClient) ListAllConfigurationProfiles() (*OSXConfigurationProfiles, error) {
	url := pc.baseClient.BuildClassicURL(ConfigurationProfiles)

	var cache OSXConfigurationProfiles
	if pc.baseClient.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[OSXConfigurationProfiles](pc.baseClient, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	pc.baseClient.SetCache(url, osxCP, 5*time.Minute)
	return &osxCP, nil
}

/*
 * # Get Configuration Profile by ID
 * /osxconfigurationprofiles/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/findosxconfigurationprofiles
 */
func (pc *ProfilesClient) GetConfigurationProfileDetails(id string) (*OSXConfigurationProfile, error) {
	url := pc.baseClient.BuildClassicURL(ConfigurationProfiles, "id", id)

	var cache OSXConfigurationProfile
	if pc.baseClient.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[OSXConfigurationProfile](pc.baseClient, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	pc.baseClient.SetCache(url, osxCP, 5*time.Minute)
	return &osxCP, nil
}

/*
 * # Update Configuration Profile by ID
 * /osxconfigurationprofiles/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/updateosxconfigurationprofilebyid
 */
func (pc *ProfilesClient) UpdateConfigurationProfile(id string) (*OSXConfigurationProfile, error) {
	url := pc.baseClient.BuildClassicURL(ConfigurationProfiles, "id", id)

	var cache OSXConfigurationProfile
	if pc.baseClient.GetCache(url, &cache) {
		return &cache, nil
	}

	osxCP, err := do[*OSXConfigurationProfile](pc.baseClient, "PUT", url, nil, nil)
	if err != nil {
		return nil, err
	}

	pc.baseClient.SetCache(url, osxCP, 5*time.Minute)
	return osxCP, nil
}
