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

type Devices []Device

type Device struct {
	Created             string         `json:"created,omitempty"`             // The timestamp when the device was created.
	ID                  string         `json:"id,omitempty"`                  // The unique key for the device.
	LastUpdated         string         `json:"lastUpdated,omitempty"`         // The timestamp when the device was last updated.
	Links               *Link          `json:"_links,omitempty"`              // A set of key/value pairs that provide additional information about the device.
	Profile             *DeviceProfile `json:"profile,omitempty"`             // The device profile.
	ResourceAlternate   interface{}    `json:"resourceAlternateId,omitempty"` // The alternate ID of the device.
	ResourceDisplayName *DisplayName   `json:"resourceDisplayName,omitempty"` // The display name of the device.
	ResourceID          string         `json:"resourceId,omitempty"`          // The ID of the device.
	ResourceType        string         `json:"resourceType,omitempty"`        // The type of the device.
	Status              string         `json:"status,omitempty"`              // The status of the device.
}

type DeviceProfile struct {
	DisplayName           string `json:"displayName,omitempty"`           // The display name of the device.
	Manufacturer          string `json:"manufacturer,omitempty"`          // The manufacturer of the device.
	Model                 string `json:"model,omitempty"`                 // The model of the device.
	OSVersion             string `json:"osVersion,omitempty"`             // The OS version of the device.
	Platform              string `json:"platform,omitempty"`              // The platform of the device.
	Registered            bool   `json:"registered,omitempty"`            // Indicates whether the device is registered with Okta.
	SecureHardwarePresent bool   `json:"secureHardwarePresent,omitempty"` // Indicates whether the device has secure hardware.
	SerialNumber          string `json:"serialNumber,omitempty"`          // The serial number of the device.
	SID                   string `json:"sid,omitempty"`                   // The SID of the device.
	UDID                  string `json:"udid,omitempty"`                  // The UDID of the device.
}

type DisplayName struct {
	Value     string `json:"value"`     // The display name of the device.
	Sensitive bool   `json:"sensitive"` // Indicates whether the display name is sensitive.
}

type DeviceUsers []DeviceUser

type DeviceUser struct {
	Created          time.Time `json:"created,omitempty"`          // The timestamp when the device user was created.
	ManagementStatus string    `json:"managementStatus,omitempty"` // The management status of the device user.
	User             *User     `json:"user,omitempty"`             // The user assigned to the device.
}

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
	Limit  string // Default: 20. Max. 200. A limit on the number of objects to return
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
	_, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	deviceUsers := DeviceUsers{}
	err = json.Unmarshal(body, &deviceUsers)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling device: %w", err)
	}

	return &deviceUsers, nil
}
