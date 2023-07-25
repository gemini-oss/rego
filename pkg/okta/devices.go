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
	After  string // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the `Link` response header.
	Limit  string // Default: 200. A limit on the number of objects to return
	Search string // A SCIM filter expression that filters the results. Searches include all Device profile properties and the Device `id``, `status``, and `lastUpdated`` properties.
	Expand string // Lists associated users for the device in `_embedded` element
}

/*
 * Lists all devices with pagination support.
 * /api/v1/devices
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices
 */
func (c *Client) ListAllDevices() (*Devices, error) {
	allDevices := Devices{}

	q := DeviceQuery{}

	url := c.BuildURL(OktaDevices)
	rawMessages, err := c.HTTPClient.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	for _, raw := range rawMessages {
		c.Logger.Debug(string(raw))
		device := Device{}
		err := json.Unmarshal(raw, &device)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user: %w", err)
		}
		allDevices = append(allDevices, device)
	}

	return &allDevices, nil
}

/*
 * Query devices with pagination support.
 * /api/v1/devices
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices
 */
func (c *Client) ListDevices(q DeviceQuery) (*Devices, error) {
	allDevices := Devices{}

	url := c.BuildURL(OktaDevices)
	rawMessages, err := c.HTTPClient.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	for _, raw := range rawMessages {
		c.Logger.Debugf("Raw: %s\n", raw)
		device := Device{}
		err := json.Unmarshal(raw, &device)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user: %w", err)
		}
		allDevices = append(allDevices, device)
	}

	return &allDevices, nil
}

/*
 * Lists all Users for a Device
 * /api/v1/devices/{deviceId}/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices
 */
func (c *Client) ListUsersForDevice(deviceID string) (*DeviceUsers, error) {

	// url := fmt.Sprintf("%s/devices/%s/users", c.BaseURL, deviceID)
	url := c.BuildURL(OktaDevices, deviceID, "users")
	res, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	deviceUsers := DeviceUsers{}
	err = json.Unmarshal(body, &deviceUsers)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling device: %w", err)
	}

	return &deviceUsers, nil
}

/*
 * Lists all non-mobile devices with Managed Status
 * /api/v1/devices
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices
 */
func (c *Client) ListManagedDevices() (*Devices, error) {
	managedDevices := Devices{}

	devices, err := c.ListDevices(DeviceQuery{
		Limit:  "100",
		Search: `status eq "ACTIVE" AND (profile.platform eq "macOS" OR profile.platform eq "WINDOWS")`,
		Expand: "user",
	})
	if err != nil {
		return nil, err
	}

	for _, device := range *devices {
		if device.Profile.Registered {
			isManaged := false
			for _, user := range *device.Embedded.DeviceUsers {
				if user.ManagementStatus == "MANAGED" {
					isManaged = true
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
