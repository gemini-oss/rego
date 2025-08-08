/*
# Google Workspace - Admin (Devices)

This package initializes all the methods for functions which interact with Devices from the Google Admin API:
https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/devices.go
package google

import (
	"time"
)

// DeviceClient for chaining methods
type DeviceClient struct {
	*Client
	DeviceQuery
}

// Entry point for device-related operations
func (c *Client) Devices() *DeviceClient {
	dc := &DeviceClient{
		Client: c,
		DeviceQuery: DeviceQuery{ // Default query parameters
			MaxResults: 500,
		},
	}

	// https://developers.google.com/admin-sdk/directory/v1/limits
	dc.HTTP.RateLimiter.Available = 2400
	dc.HTTP.RateLimiter.Limit = 2400
	dc.HTTP.RateLimiter.Interval = 1 * time.Minute
	dc.HTTP.RateLimiter.Log.Verbosity = c.Log.Verbosity

	return dc
}

/*
 * Query Parameters for ChromeOS Devices
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices/list#query-parameters
 */
type DeviceQuery struct {
	IncludeChildOrgunits bool   `url:"includeChildOrgunits,omitempty"` // If true, return devices from all child org units as well as the specified org unit.
	MaxResults           int    `url:"maxResults,omitempty"`           // Maximum number of results to return. Default is 100
	OrderBy              string `url:"orderBy,omitempty"`              // Device property to use for sorting results. Should be one of the defined OrderBy enums.
	OrgUnitPath          string `url:"orgUnitPath,omitempty"`          // Full path of the organizational unit (minus the leading /) or its unique ID.
	PageToken            string `url:"pageToken,omitempty"`            // Token for requesting the next page of query results.
	Projection           string `url:"projection,omitempty"`           // Restrict information returned to a set of selected fields. Should be one of the defined Projection enums.
	Query                string `url:"query,omitempty"`                // https://developers.google.com/admin-sdk/directory/v1/list-query-operators
	SortOrder            string `url:"sortOrder,omitempty"`            // Whether to return results in ascending or descending order. Should be one of the defined SortOrder enums.
}

func (q *DeviceQuery) SetPageToken(token string) {
	q.PageToken = token
}

// ### Chainable DeviceClient Methods
// ---------------------------------------------------------------------
func (c *DeviceClient) MaxResults(max int) *DeviceClient {
	c.DeviceQuery.MaxResults = max
	return c
}

func (c *DeviceClient) PageToken(token string) *DeviceClient {
	c.DeviceQuery.PageToken = token
	return c
}

func (c *DeviceClient) Query(query string) *DeviceClient {
	c.DeviceQuery.Query = query
	return c
}

// END OF CHAINABLE METHODS
//---------------------------------------------------------------------

/*
 * List all ChromeOS Devices in the domain with pagination support
 * admin/directory/v1/customer/{customerId}/devices/chromeos
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices/list
 */
func (c *DeviceClient) ListAllChromeOS(customer *Customer) (*ChromeOSDevices, error) {
	c.Log.Println("Getting all ChromeOS Devices...")

	url := c.BuildURL(DirectoryChromeOSDevices, customer)

	var cache ChromeOSDevices
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	devices, err := doPaginated[ChromeOSDevices, *DeviceQuery](c.Client, "GET", url, &c.DeviceQuery, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, devices, 5*time.Minute)
	return devices, nil
}

/*
 * List all Provisioned ChromeOS Devices in the domain with pagination support
 * admin/directory/v1/customer/{customerId}/devices/chromeos
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices/list
 */
func (c *DeviceClient) ListAllProvisionedChromeOS(customer *Customer) (*ChromeOSDevices, error) {
	c.Log.Println("Getting all ChromeOS Devices...")

	url := c.BuildURL(DirectoryChromeOSDevices, customer)

	var cache ChromeOSDevices
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	c.Query("status:provisioned")

	devices, err := doPaginated[ChromeOSDevices, *DeviceQuery](c.Client, "GET", url, &c.DeviceQuery, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, devices, 5*time.Minute)
	return devices, nil
}
