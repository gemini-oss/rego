/*
# SnipeIT - Locations

This package initializes all the methods for functions which interact with the SnipeIT Locations endpoints:
https://snipe-it.readme.io/reference/locations

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/accessories.go
package snipeit

import (
	"time"
)

// LocationClient for chaining methods
type LocationClient struct {
	*Client
}

// Entry point for locations-related operations
func (c *Client) Locations() *LocationClient {
	lc := &LocationClient{
		Client: c,
	}

	return lc
}

/*
 * Query Parameters for Locations
 */
type LocationQuery struct {
	Limit       int    `url:"limit,omitempty"`        // Specify the number of results you wish to return. Defaults to 50.
	Offset      int    `url:"offset,omitempty"`       // Specify the number of results to skip before starting to return items. Defaults to 0.
	Search      string `url:"search,omitempty"`       // Search for a location by name or address.
	OrderNumber string `url:"order_number,omitempty"` // Return only assets associated with the specified order number.
	Sort        string `url:"sort,omitempty"`         // Sort the results by the specified column. Defaults to id.
	Order       string `url:"order,omitempty"`        // Sort the results in the specified order. Defaults to asc.
	Expand      string `url:"expand,omitempty"`       // Expand the results to include full details of the associated model, category, and manufacturer.
}

// ### LocationQuery implements QueryInterface
// ---------------------------------------------------------------------
func (q *LocationQuery) Copy() QueryInterface {
	return &LocationQuery{
		Limit:       q.Limit,
		Offset:      q.Offset,
		Search:      q.Search,
		OrderNumber: q.OrderNumber,
		Sort:        q.Sort,
		Order:       q.Order,
	}
}

func (q *LocationQuery) GetLimit() int {
	return q.Limit
}

func (q *LocationQuery) SetLimit(limit int) {
	q.Limit = limit
}

func (q *LocationQuery) GetOffset() int {
	return q.Offset
}

func (q *LocationQuery) SetOffset(offset int) {
	q.Offset = offset
}

// END OF QUERYINTERFACE METHODS
//---------------------------------------------------------------------

/*
 * # List all Locations in Snipe-IT
 * /api/v1/locations
 * - https://snipe-it.readme.io/reference/locations
 */
func (c *LocationClient) GetAllLocations() (*LocationList, error) {
	url := c.BuildURL(Locations)
	q := LocationQuery{
		Limit: 50,
	}

	var cache LocationList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	locations, err := doConcurrent[LocationList](c.Client, "GET", url, &q, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching hardware list: %v", err)
	}

	c.SetCache(url, locations, 5*time.Minute)

	return locations, nil
}
