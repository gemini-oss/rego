/*
# Jamf - Devices

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/devices.go
package jamf

import (
	"fmt"
	"time"
)

var (
	ComputersInventory       = fmt.Sprintf("%s/computers-inventory", V1)        // /api/v1/computers-inventory
	ComputersInventoryDetail = fmt.Sprintf("%s/computers-inventory-detail", V1) // /api/v1/computers-inventory-detail
	ComputerGroups           = fmt.Sprintf("%s/computer-groups", V1)            // /api/v1/computer-groups
	MobileDev                = fmt.Sprintf("%s/mobile-devices", V2)             // /api/v2/mobile-devices
)

// DeviceClient for chaining methods
type DeviceClient struct {
	client *Client
	query  DeviceQuery
}

// Entry point for device-related operations
func (c *Client) Devices() *DeviceClient {
	return &DeviceClient{
		client: c,
		query: DeviceQuery{ // Default query parameters
			Sections: []string{
				Section.General,
			},
			Page:     0,
			PageSize: 100,
		},
	}
}

/*
- Query parameters for Computer Details

  - Example:
    Fetch details from GENERAL and HARDWARE sections
    section=GENERAL&section=HARDWARE

    Fetch the second page of results with 50 items per page
    page=1&page-size=50

    Sort by the unique device identifier in descending order and then by name in ascending order
    sort=udid:desc,general.name:asc

    RSQL Filter results where the general name is "Orchard"
    filter=general.name=="Orchard"
*/
type DeviceQuery struct {
	Sections []string `url:"section,omitempty"`   // Sections of computer details to return. If not specified, the General section data is returned. Multiple sections can be specified, e.g., section=GENERAL&section=HARDWARE.
	Page     int      `url:"page,omitempty"`      // The pagination index (starting from 0) for the query results.
	PageSize int      `url:"page-size,omitempty"` // The number of records per page. Default is 100.
	Sort     []string `url:"sort,omitempty"`      // Sorting criteria in the format: property:asc/desc. Default sort is general.name:asc. Multiple criteria can be specified and separated by a comma.
	Filter   string   `url:"filter,omitempty"`    // RSQL query string used for filtering the computer inventory collection. The default filter is an empty query, returning all results for the requested page.
}

/*
 * Check if the DeviceQuery is empty
 */
func (d *DeviceQuery) IsEmpty() bool {
	return d.Sections == nil &&
		d.Page == 0 &&
		d.PageSize == 0 &&
		d.Sort == nil &&
		d.Filter == ""
}

/*
 * Validate the query parameters for the Files resource
 */
func (d *DeviceQuery) ValidateQuery() error {
	if d.Sections != nil {
		d.Sections = []string{
			Section.General,
		}
	}

	if d.Page < 0 {
		return fmt.Errorf("page must be greater than or equal to 0")
	}

	if d.PageSize < 0 {
		return fmt.Errorf("page size must be greater than or equal to 0")
	}

	if d.Sort != nil {
		d.Sort = []string{
			"general.name:asc",
		}
	}

	return nil
}

// ### Chainable DeviceClient Methods
// ---------------------------------------------------------------------
func (dc *DeviceClient) Sections(sections []string) *DeviceClient {
	dc.query.Sections = sections
	return dc
}

func (dc *DeviceClient) Page(page int) *DeviceClient {
	dc.query.Page = page
	return dc
}

func (dc *DeviceClient) PageSize(pageSize int) *DeviceClient {
	dc.query.PageSize = pageSize
	return dc
}

func (dc *DeviceClient) Sort(sort []string) *DeviceClient {
	dc.query.Sort = sort
	return dc
}

func (dc *DeviceClient) Filter(filter string) *DeviceClient {
	dc.query.Filter = filter
	return dc
}

// END OF CHAINABLE METHODS
//---------------------------------------------------------------------

/*
 * # Get Computer Devices
 * /api/v1/computers-inventory
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-computers-inventory
 */
func (dc *DeviceClient) ListAllComputers() (*Computers, error) {
	url := dc.client.BuildURL(ComputersInventory)

	var cache Computers
	if dc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	q := &DeviceQuery{
		Sections: []string{
			Section.General,
		},
		Page:     0,
		PageSize: 100,
	}

	computers, err := doConcurrent[Computers](dc.client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	dc.client.SetCache(url, computers, 5*time.Minute)
	return computers, nil
}

/*
 * # Get Computer Details
 * /api/v1/computers-inventory-detail/{id}
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-computers-inventory-detail-id
 */
func (dc *DeviceClient) GetComputerDetails(id string) (*Computer, error) {
	url := dc.client.BuildURL(ComputersInventoryDetail, id)

	var cache Computer
	if dc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	computer, err := do[*Computer](dc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	dc.client.SetCache(url, computer, 5*time.Minute)
	return computer, nil
}

/*
 * # Get Computer Groups
 * /api/v1/computer-groups
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-computer-groups
 */
func (dc *DeviceClient) ListAllComputerGroups() (*[]GroupMembership, error) {
	url := dc.client.BuildURL(ComputerGroups)

	var cache *[]GroupMembership
	if dc.client.GetCache(url, cache) {
		return cache, nil
	}

	groups, err := do[*[]GroupMembership](dc.client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

/*
 * # Get Mobile Devices
 * /api/v2/mobile-devices
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-mobile-devices
 */
func (dc *DeviceClient) ListAllMobileDevices() (*MobileDevices, error) {
	url := dc.client.BuildURL(MobileDev)

	var cache MobileDevices
	if dc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	q := &DeviceQuery{
		Page:     0,
		PageSize: 100,
	}

	md, err := doConcurrent[MobileDevices](dc.client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	dc.client.SetCache(url, md, 5*time.Minute)
	return md, nil
}
