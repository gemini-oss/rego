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
	Name     string `url:"name,omitempty"`     // Search for a location by name.
	Limit    int    `url:"limit,omitempty"`    // Specify the number of results you wish to return. Defaults to 50.
	Offset   int    `url:"offset,omitempty"`   // Specify the number of results to skip before starting to return items. Defaults to 0.
	Search   string `url:"search,omitempty"`   // Search for a location by name or address.
	Sort     string `url:"sort,omitempty"`     // Sort the results by the specified column. Defaults to id.
	Order    string `url:"order,omitempty"`    // Sort the results in the specified order. Defaults to asc.
	Address  string `url:"address,omitempty"`  // Search for a location by address.
	Address2 string `url:"address2,omitempty"` // Search for a location by address2.
	City     string `url:"city,omitempty"`     // Search for a location by city.
	State    string `url:"state,omitempty"`    // Search for a location by state.
	Country  string `url:"country,omitempty"`  // Search for a location by country.
	Expand   string `url:"expand,omitempty"`   // Expand the results to include full details of the associated model, category, and manufacturer.
}

// ### LocationQuery implements QueryInterface
// ---------------------------------------------------------------------
func (q *LocationQuery) Copy() QueryInterface {
	return &LocationQuery{
		Limit:  q.Limit,
		Offset: q.Offset,
		Search: q.Search,
		Sort:   q.Sort,
		Order:  q.Order,
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

/* Get Location Details by ID
 * /api/v1/locations/{id}
 * - https://snipe-it.readme.io/reference/locations-1
 */
func (c *LocationClient) GetLocation(id int) (*Location, error) {
	url := c.BuildURL(Locations, id)

	var cache Location
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	location, err := do[Location](c.Client, "GET", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching location: %v", err)
	}

	c.SetCache(url, location, 5*time.Minute)
	return &location, nil
}

/*
 * # Create a new Location in Snipe-IT
 * /api/v1/locations
 * - https://snipe-it.readme.io/reference/locations-2
 */
func (c *LocationClient) CreateLocation(loc *Location) (*Location, error) {
	url := c.BuildURL(Locations)

	location, err := do[Location](c.Client, "POST", url, nil, loc)
	if err != nil {
		c.Log.Fatalf("Error creating location: %v", err)
	}

	return &location, nil
}

/*
 * # Update a Location in Snipe-IT
 * /api/v1/locations/{id}
 * - https://snipe-it.readme.io/reference/locations-3
 */
func (c *LocationClient) UpdateLocation(id int, loc *Location) (*Location, error) {
	url := c.BuildURL(Locations, loc.ID)

	location, err := do[Location](c.Client, "PUT", url, nil, loc)
	if err != nil {
		c.Log.Fatalf("Error updating location: %v", err)
	}

	c.SetCache(url, location, 5*time.Minute)
	return &location, nil
}

/* Partially update a Location in Snipe-IT
 * /api/v1/locations/{locationId}
 * - https://snipe-it.readme.io/reference/locationsid
 */
func (c *LocationClient) PartialUpdateLocation(id int,loc *Location) (*Location, error) {
	url := c.BuildURL(Locations, id)

	location, err := do[Location](c.Client, "PATCH", url, nil, loc)
	if err != nil {
		c.Log.Fatalf("Error updating location: %v", err)
	}

	c.SetCache(url, location, 5*time.Minute)
	return &location, nil
}

/*
 * # Delete a Location in Snipe-IT
 * /api/v1/locations/{locationId}
 * - https://snipe-it.readme.io/reference/locationsid-2
 */
func (c *LocationClient) DeleteLocation(id int) (*Location, error) {
	url := c.BuildURL(Locations, id)

	location, err := do[Location](c.Client, "DELETE", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error deleting location: %v", err)
	}

	return &location, nil
}
