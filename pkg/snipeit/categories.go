/*
# SnipeIT - Categories

This package initializes all the methods for functions which interact with the SnipeIT Categories endpoints:
https://snipe-it.readme.io/reference/categories

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/categories.go
package snipeit

import (
	"time"
)

// CategoryClient for chaining methods
type CategoryClient struct {
	*Client
}

// Entry point for asset-related operations
func (c *Client) Categories() *CategoryClient {
	cc := &CategoryClient{
		Client: c,
	}

	return cc
}

/*
 * Query Parameters for Categories
 */
type CategoryQuery struct {
	Name              string `url:"string,omitempty"`                      // Specify the name of the category.
	Limit             int    `url:"limit,omitempty"`                       // Specify the number of results you wish to return. Defaults to 50.
	Offset            int    `url:"offset,omitempty"`                      // Specify the number of results to skip before starting to return items. Defaults to 0.
	Search            string `url:"search,omitempty"`                      // Search for an asset by asset tag, serial, or model number.
	Sort              string `url:"sort,omitempty"`                        // Sort the results by the specified column. Defaults to id.
	Order             string `url:"order,omitempty"`                       // Sort the results in the specified order. Defaults to asc.
	CategoryID        int    `url:"category_id,omitempty"`                 // Return only assets associated with the specified category ID.
	Type              string `url:"category_type,omitempty"`               // Type of category
	UseDefaultEULA    bool   `url:"use_default_eula,omitempty,omitzero"`   // If a category is using the default EULA
	RequireAcceptance bool   `url:"require_acceptance,omitempty,omitzero"` // If the category required acceptance of the EULA
	CheckinEmail      bool   `url:"checkin_email,omitempty,omitzero"`      // Email
}

// ### CategoryQuery implements QueryInterface
// ---------------------------------------------------------------------
func (q *CategoryQuery) Copy() QueryInterface {
	return &CategoryQuery{
		Name:              q.Name,
		Limit:             q.Limit,
		Offset:            q.Offset,
		Search:            q.Search,
		Sort:              q.Sort,
		Order:             q.Order,
		CategoryID:        q.CategoryID,
		Type:              q.Type,
		UseDefaultEULA:    q.UseDefaultEULA,
		RequireAcceptance: q.RequireAcceptance,
		CheckinEmail:      q.CheckinEmail,
	}
}

func (q *CategoryQuery) GetLimit() int {
	return q.Limit
}

func (q *CategoryQuery) SetLimit(limit int) {
	q.Limit = limit
}

func (q *CategoryQuery) GetOffset() int {
	return q.Offset
}

func (q *CategoryQuery) SetOffset(offset int) {
	q.Offset = offset
}

// END OF QUERYINTERFACE METHODS
//---------------------------------------------------------------------

/*
 * List all Categories in Snipe-IT
 * /api/v1/categories
 * https://snipe-it.readme.io/reference/categories-1
 */
func (c *CategoryClient) GetAllCategories() (*CategoryList, error) {
	url := c.BuildURL(Categories)

	q := CategoryQuery{
		Limit:  500,
		Offset: 0,
	}

	var cache CategoryList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	categories, err := doConcurrent[CategoryList](c.Client, "GET", url, &q, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching category list: %v", err)
	}

	c.SetCache(url, categories, 5*time.Minute)
	return categories, nil
}

/*
 * # Create a category in Snipe-IT
 * /api/v1/categories
 * https://snipe-it.readme.io/reference/categories-2
 */
func (c *CategoryClient) CreateCategory(p *Category[CategoryPOST]) (*Category[CategoryPOST], error) {
	url := c.BuildURL(Categories)

	category, err := do[SnipeITResponse[Category[CategoryPOST]]](c.Client, "POST", url, nil, p)
	if err != nil {
		c.Log.Fatalf("Error creating category: %v", err)
	}

	return category.Payload, nil
}
