/*
# Okta Devices

This package contains all the methods to interact with the Okta Devices API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/devices.go
package okta

import (
	"encoding/json"
	"fmt"
	"time"
)

/*
- Query parameters for Devices

  - Example:
    Devices that have a `status` of `ACTIVE`
    search=status eq "ACTIVE"

    Devices last updated after a specific timestamp
    search=lastUpdated gt "yyyy-MM-dd'T'HH:mm:ss.SSSZ"

    Devices with a specified `id`
    search=id eq "guo4a5u7JHHhjXrMK0g4"

    Devices that have a `displayName` of `Bob`
    search=profile.displayName eq "Bob"

    Devices that have an `platform` of `WINDOWS`
    search=profile.platform eq "WINDOWS"

    Devices whose `sid` starts with `S-1`
    search=profile.sid sw "S-1"
*/
type DeviceQuery struct {
	After  string `url:"after,omitempty"`  // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the `Link` response header.
	Limit  string `url:"limit,omitempty"`  // Default: 200. A limit on the number of objects to return
	Search string `url:"search,omitempty"` // A SCIM filter expression that filters the results. Searches include all Device profile properties and the Device `id``, `status``, and `lastUpdated`` properties.
	Expand string `url:"expand,omitempty"` // Lists associated users for the device in `_embedded` element
}

/*
 * # List All Devices
 * Lists all devices with pagination support.
 * /api/v1/devices
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices
 */
func (c *Client) ListAllDevices() (*Devices, error) {
	url := c.BuildURL(OktaDevices)

	if c.Cache.Enabled {
		if data, found := c.Cache.Get(url); found {
			var cache Devices
			if err := json.Unmarshal(data, &cache); err != nil {
				return nil, err
			}

			c.Log.Debug("Cached Body:", string(data))
			return &cache, nil
		}
	}

	allDevices := Devices{}
	q := DeviceQuery{}

	rawMessages, err := c.HTTP.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	for _, raw := range rawMessages {
		c.Log.Debug(string(raw))
		device := Device{}
		err := json.Unmarshal(raw, &device)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user: %w", err)
		}
		allDevices = append(allDevices, device)
	}

	if data, err := json.Marshal(allDevices); err == nil {
		c.Cache.Set(url, data, 30*time.Minute)
	}

	return &allDevices, nil
}

/*
 * # List Devices (Queried)
 * Query devices with pagination support.
 * /api/v1/devices
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices
 */
func (c *Client) ListDevices(q DeviceQuery) (*Devices, error) {
	url := c.BuildURL(OktaDevices)

	if c.Cache.Enabled {
		if data, found := c.Cache.Get(url); found {
			var cache Devices
			if err := json.Unmarshal(data, &cache); err != nil {
				return nil, err
			}

			c.Log.Debug("Cached Body:", string(data))
			return &cache, nil
		}
	}
	allDevices := Devices{}

	rawMessages, err := c.HTTP.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	for _, raw := range rawMessages {
		c.Log.Debugf("Raw: %s\n", raw)
		device := Device{}
		err := json.Unmarshal(raw, &device)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user: %w", err)
		}
		allDevices = append(allDevices, device)
	}

	if data, err := json.Marshal(allDevices); err == nil {
		c.Cache.Set(url, data, 30*time.Minute) // Set cache with an expiration
	}

	return &allDevices, nil
}

/*
 * # List all Users for a Device
 * /api/v1/devices/{deviceId}/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices
 */
func (c *Client) ListUsersForDevice(deviceID string) (*DeviceUsers, error) {

	// url := fmt.Sprintf("%s/devices/%s/users", c.BaseURL, deviceID)
	url := c.BuildURL(OktaDevices, deviceID, "users")
	res, body, err := c.HTTP.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	deviceUsers := DeviceUsers{}
	err = json.Unmarshal(body, &deviceUsers)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling device: %w", err)
	}

	return &deviceUsers, nil
}

/*
 * # List all non-mobile devices with Managed Status
 * /api/v1/devices
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices
 */
func (c *Client) ListManagedDevices() (*Devices, error) {
	managedDevices := Devices{}

	devices, err := c.ListDevices(
		DeviceQuery{
			Limit:  "50",
			Search: `status eq "ACTIVE" AND (profile.platform eq "macOS" OR profile.platform eq "WINDOWS")`,
			Expand: "user",
		},
	)
	if err != nil {
		return nil, err
	}

	for _, device := range *devices {
		if device.Profile.Registered {
			isManaged := false
			if device.Embedded != nil {
				for _, user := range *device.Embedded.DeviceUsers {
					if user.ManagementStatus == "MANAGED" {
						isManaged = true
					}
				}
			}
			// If the device has at least one managed user, append it to managedDevices
			if isManaged {
				managedDevices = append(managedDevices, device)
			}
		}
	}

	return &managedDevices, nil
}
