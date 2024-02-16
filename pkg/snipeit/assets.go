/*
# SnipeIT - Assets

This package initializes all the methods for functions which interact with the SnipeIT Assets endpoints:
https://developer.okta.com/docs/api/

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/assets.go
package snipeit

import (
	"encoding/json"
	"fmt"
	"sync"
)

/*
 * Query Parameters for Assets
 */
type AssetQuery struct {
	Limit          int    `url:"limit,omitempty"`           // Specify the number of results you wish to return. Defaults to 50.
	Offset         int    `url:"offset,omitempty"`          // Specify the number of results to skip before starting to return items. Defaults to 0.
	Search         string `url:"search,omitempty"`          // Search for an asset by asset tag, serial, or model number.
	OrderNumber    string `url:"order_number,omitempty"`    // Return only assets associated with the specified order number.
	Sort           string `url:"sort,omitempty"`            // Sort the results by the specified column. Defaults to id.
	Order          string `url:"order,omitempty"`           // Sort the results in the specified order. Defaults to asc.
	ModelID        int    `url:"model_id,omitempty"`        // Return only assets associated with the specified model ID.
	CategoryID     int    `url:"category_id,omitempty"`     // Return only assets associated with the specified category ID.
	ManufacturerID int    `url:"manufacturer_id,omitempty"` // Return only assets associated with the specified manufacturer ID.
	CompanyID      int    `url:"company_id,omitempty"`      // Return only assets associated with the specified company ID.
	LocationID     int    `url:"location_id,omitempty"`     // Return only assets associated with the specified location ID.
	Status         string `url:"status,omitempty"`          // Optionally restrict asset results to one of these status types: RTD, Deployed, Undeployable, Deleted, Archived, Requestable
	StatusID       int    `url:"status_id,omitempty"`       // Return only assets associated with the specified status ID.
}

/*
 * List all Hardware Assets in Snipe-IT
 * /api/v1/hardware
 * - https://snipe-it.readme.io/reference/hardware
 */
func (c *Client) GetAllAssets() (*HardwareList, error) {
	assets := &HardwareList{}

	q := AssetQuery{
		Limit:  500,
		Offset: 0,
	}

	c.HTTPClient.RateLimiter.Start()

	url := fmt.Sprintf(Assets, c.BaseURL)
	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Debug(res.Status)
	c.Logger.Trace(string(body))

	err = json.Unmarshal(body, &assets)
	if err != nil {
		return nil, err
	}

	// Use a buffered channel as a semaphore to limit concurrent requests.
	sem := make(chan struct{}, 10)

	// WaitGroup to ensure all go routines complete their tasks.
	var wg sync.WaitGroup

	// Buffered channel to hold device pages result from each goroutine
	assetsCh := make(chan map[string]*HardwareList, assets.Total)

	// Buffered channel to hold any errors that occur while getting device pages
	rolesErrCh := make(chan error)

	for next_page := true; next_page; next_page = (q.Offset < assets.Total) {

		remainingAssets := assets.Total - q.Offset
		if remainingAssets < q.Limit {
			q.Limit = remainingAssets
		}

		wg.Add(1)

		// Start a new goroutine to get the next device page
		go func(q AssetQuery) {
			// Release one semaphore resource when the goroutine completes
			defer wg.Done()

			sem <- struct{}{} // acquire one semaphore resource
			page := &HardwareList{}

			res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
			if err != nil {
				rolesErrCh <- err
				return
			}

			c.Logger.Debug("Response Status:", res.Status)
			c.Logger.Trace("Response Body: ", string(body))

			err = json.Unmarshal(body, &page)
			if err != nil {
				rolesErrCh <- err
				return
			}

			newPage := make(map[string]*HardwareList)
			newPage[fmt.Sprint(q.Offset)] = page
			assetsCh <- newPage
			<-sem // release one semaphore resource
		}(q) // Pass the query to the goroutine

		q.Offset += q.Limit
	}

	// Wait for all goroutines to finish and close channels
	go func() {
		wg.Wait()
		close(assetsCh)
		close(rolesErrCh)
	}()

	// Collect devices from all pages
	for deviceRecords := range assetsCh {
		for _, results := range deviceRecords {
			assets.Rows = append(assets.Rows, results.Rows...)
		}
	}

	// Check if there were any errors
	if len(rolesErrCh) > 0 {
		// Handle or return errors. For simplicity, only returning the first error here
		return nil, <-rolesErrCh
	}

	c.HTTPClient.RateLimiter.Stop()

	return assets, nil
}
