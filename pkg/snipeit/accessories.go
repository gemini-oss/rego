/*
# SnipeIT - Accessories

This package initializes all the methods for functions which interact with the SnipeIT Accessories endpoints:
https://snipe-it.readme.io/reference/accessories

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/accessories.go
package snipeit

import (
	"time"
)

// AccessoriesClient for chaining methods
type AccessoryClient struct {
	*Client
}

// Entry point for accessories-related operations
func (c *Client) Accessories() *AccessoryClient {
	ac := &AccessoryClient{
		Client: c,
	}

	return ac
}

/*
 * Query Parameters for Accessories
 */
type AccessoryQuery struct {
	Limit       int    `url:"limit,omitempty"`        // Specify the number of results you wish to return. Defaults to 50.
	Offset      int    `url:"offset,omitempty"`       // Specify the number of results to skip before starting to return items. Defaults to 0.
	Search      string `url:"search,omitempty"`       // Search for an asset by asset tag, serial, or model number.
	OrderNumber string `url:"order_number,omitempty"` // Return only assets associated with the specified order number.
	Sort        string `url:"sort,omitempty"`         // Sort the results by the specified column. Defaults to id.
	Order       string `url:"order,omitempty"`        // Sort the results in the specified order. Defaults to asc.
	Expand      string `url:"expand,omitempty"`       // Expand the results to include full details of the associated model, category, and manufacturer.
}

// ### AccessoryQuery implements QueryInterface
// ---------------------------------------------------------------------
func (q *AccessoryQuery) Copy() QueryInterface {
	return &AccessoryQuery{
		Limit:       q.Limit,
		Offset:      q.Offset,
		Search:      q.Search,
		OrderNumber: q.OrderNumber,
		Sort:        q.Sort,
		Order:       q.Order,
		Expand:      q.Expand,
	}
}

func (q *AccessoryQuery) GetLimit() int {
	return q.Limit
}

func (q *AccessoryQuery) SetLimit(limit int) {
	q.Limit = limit
}

func (q *AccessoryQuery) GetOffset() int {
	return q.Offset
}

func (q *AccessoryQuery) SetOffset(offset int) {
	q.Offset = offset
}

// END OF QUERYINTERFACE METHODS
//---------------------------------------------------------------------

/*
 * # List all Accessories in Snipe-IT
 * /api/v1/accessories
 * - https://snipe-it.readme.io/reference/accessories
 */
func (c *AccessoryClient) GetAllAccessories() (*AccessoryList, error) {
	url := c.BuildURL(Accessories)
	q := AccessoryQuery{
		Limit: 50,
	}

	var cache AccessoryList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	accessories, err := doConcurrent[AccessoryList](c.Client, "GET", url, &q, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching hardware list: %v", err)
	}

	c.SetCache(url, accessories, 5*time.Minute)

	return accessories, nil
}
