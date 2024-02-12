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
	"encoding/json"
	"fmt"
	"sync"
)

var (
	ComputersInventory       = fmt.Sprintf("%s/computers-inventory", V1)        // /api/v1/computers-inventory
	ComputersInventoryDetail = fmt.Sprintf("%s/computers-inventory-detail", V1) // /api/v1/computers-inventory-detail
	ComputerGroups           = fmt.Sprintf("%s/computer-groups", V1)            // /api/v1/computer-groups
	MobileDev                = fmt.Sprintf("%s/mobile-devices", V2)             // /api/v2/mobile-devices
)

/*
- Query parameters for Computer Details

  - Example:
    Fetch details from GENERAL and HARDWARE sections
    section=GENERAL&section=HARDWARE

    Fetch the second page of results with 50 items per page
    page=1&page-size=50

    Sort by the unique device identifier in descending order and then by name in ascending order
    sort=udid:desc,general.name:asc

    Filter results where the general name is "Orchard"
    filter=general.name=="Orchard"
*/
type DeviceQuery struct {
	Sections []string `json:"section,omitempty"`   // Sections of computer details to return. If not specified, the General section data is returned. Multiple sections can be specified, e.g., section=GENERAL&section=HARDWARE.
	Page     int      `json:"page,omitempty"`      // The pagination index (starting from 0) for the query results.
	PageSize int      `json:"page-size,omitempty"` // The number of records per page. Default is 100.
	Sort     []string `json:"sort,omitempty"`      // Sorting criteria in the format: property:asc/desc. Default sort is general.name:asc. Multiple criteria can be specified and separated by a comma.
	Filter   string   `json:"filter,omitempty"`    // RSQL query string used for filtering the computer inventory collection. The default filter is an empty query, returning all results for the requested page.
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
			"GENERAL",
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

/*
 * # Get Computer Devices
 * /api/v1/computers-inventory
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-computers-inventory
 */
func (c *Client) ListAllComputers() (*Computers, error) {
	allDevices := &Computers{}

	q := DeviceQuery{
		Page:     0,
		PageSize: 100,
	}

	url := c.BuildURL(ComputersInventory)
	res, body, err := c.HTTP.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &allDevices)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	// Use a buffered channel as a semaphore to limit concurrent requests.
	sem := make(chan struct{}, 10)

	// WaitGroup to ensure all go routines complete their tasks.
	var wg sync.WaitGroup

	totalPages := allDevices.TotalCount / q.PageSize
	if allDevices.TotalCount%q.PageSize > 0 {
		totalPages++
	}

	// Buffered channel to hold device pages result from each goroutine
	devicesCh := make(chan map[string]*Computers, totalPages)

	// Buffered channel to hold any errors that occur while getting device pages
	rolesErrCh := make(chan error)

	for next_page := true; next_page; next_page = (q.Page < totalPages) {

		wg.Add(1)

		q.Page++ // Increment page number
		c.Logger.Println("Page: ", q.Page)

		// Start a new goroutine to get the next device page
		go func(q DeviceQuery) {
			// Release one semaphore resource when the goroutine completes
			defer wg.Done()

			sem <- struct{}{} // acquire one semaphore resource
			page := &Computers{}

			res, body, err := c.HTTP.DoRequest("GET", url, q, nil)
			if err != nil {
				rolesErrCh <- err
				return
			}
			c.Logger.Println("Response Status:", res.Status)
			c.Logger.Debug("Response Body: ", string(body))

			err = json.Unmarshal(body, &page)
			if err != nil {
				rolesErrCh <- err
				return
			}

			newPage := make(map[string]*Computers)
			newPage[fmt.Sprint(q.Page)] = page
			devicesCh <- newPage
			<-sem // release one semaphore resource
		}(q) // Pass the query to the goroutine
	}

	// Wait for all goroutines to finish and close channels
	go func() {
		wg.Wait()
		close(devicesCh)
		close(rolesErrCh)
	}()

	// Collect devices from all pages
	for deviceRecords := range devicesCh {
		for _, results := range deviceRecords {
			allDevices.Results = append(allDevices.Results, results.Results...)
		}
	}

	// Check if there were any errors
	if len(rolesErrCh) > 0 {
		// Handle or return errors. For simplicity, only returning the first error here
		return nil, <-rolesErrCh
	}

	return allDevices, nil
}

/*
 * # Get Computer Details
 * /api/v1/computers-inventory-detail/{id}
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-computers-inventory-detail-id
 */
func (c *Client) GetComputerDetails(id string) (*Computer, error) {
	computer := &Computer{}

	url := c.BuildURL(ComputersInventoryDetail, id)
	res, body, err := c.HTTP.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &computer)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return computer, nil
}

/*
 * # Get Computer Groups
 * /api/v1/computer-groups
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-computers-inventory
 */
func (c *Client) ListAllComputerGroups() (*[]GroupMembership, error) {
	allGroups := &[]GroupMembership{}

	url := c.BuildURL(ComputerGroups)
	res, body, err := c.HTTP.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &allGroups)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return allGroups, nil
}

/*
 * # Get Mobile Devices
 * /api/v2/mobile-devices
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-mobile-devices
 */
func (c *Client) ListAllMobileDevices() (*MobileDevices, error) {
	allDevices := &MobileDevices{}

	q := DeviceQuery{
		Page:     0,
		PageSize: 100,
	}

	url := c.BuildURL(MobileDev)
	res, body, err := c.HTTP.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &allDevices)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	// Use a buffered channel as a semaphore to limit concurrent requests.
	sem := make(chan struct{}, 10)

	// WaitGroup to ensure all go routines complete their tasks.
	var wg sync.WaitGroup

	totalPages := allDevices.TotalCount / q.PageSize
	if allDevices.TotalCount%q.PageSize > 0 {
		totalPages++
	}

	// Buffered channel to hold device pages result from each goroutine
	devicesCh := make(chan map[string]*MobileDevices, totalPages)

	// Buffered channel to hold any errors that occur while getting device pages
	rolesErrCh := make(chan error)

	for next_page := true; next_page; next_page = (q.Page < totalPages) {

		wg.Add(1)

		q.Page++ // Increment page number
		c.Logger.Println("Page:", q.Page)

		// Start a new goroutine to get the next device page
		go func(q DeviceQuery) {
			// Release one semaphore resource when the goroutine completes
			defer wg.Done()

			sem <- struct{}{} // acquire one semaphore resource
			page := &MobileDevices{}

			res, body, err := c.HTTP.DoRequest("GET", url, q, nil)
			if err != nil {
				rolesErrCh <- err
				return
			}
			c.Logger.Println("Response Status:", res.Status)
			c.Logger.Debug("Response Body:", string(body))

			err = json.Unmarshal(body, &page)
			if err != nil {
				rolesErrCh <- err
				return
			}

			newPage := make(map[string]*MobileDevices)
			newPage[fmt.Sprint(q.Page)] = page
			devicesCh <- newPage
			<-sem // release one semaphore resource
		}(q) // Pass the query to the goroutine
	}

	// Wait for all goroutines to finish and close channels
	go func() {
		wg.Wait()
		close(devicesCh)
		close(rolesErrCh)
	}()

	// Collect devices from all pages
	for deviceRecords := range devicesCh {
		for _, results := range deviceRecords {
			allDevices.Results = append(allDevices.Results, results.Results...)
		}
	}

	// Check if there were any errors
	if len(rolesErrCh) > 0 {
		// Handle or return errors. For simplicity, only returning the first error here
		return nil, <-rolesErrCh
	}

	return allDevices, nil
}
