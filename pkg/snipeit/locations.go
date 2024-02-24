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
	"encoding/json"
	"fmt"
)

/*
 * Query Parameters for Assets
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

/*
 * # List all Locations in Snipe-IT
 * /api/v1/locations
 * - https://snipe-it.readme.io/reference/locations
 */
func (c *Client) GetAllLocations() (*LocationList, error) {
	locations := &LocationList{}

	q := AssetQuery{
		Limit: 50,
	}

	url := fmt.Sprintf(Locations, c.BaseURL)
	res, body, err := c.HTTP.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Log.Debug("Response from GetAllAccessories: ", res.Status)
	c.Log.Debug("Body from GetAllAccessories: ", string(body))

	err = json.Unmarshal(body, &locations)
	if err != nil {
		return nil, err
	}

	return locations, nil
}
