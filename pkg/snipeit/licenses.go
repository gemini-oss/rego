/*
# SnipeIT - Licenses

This package initializes all the methods for functions which interact with the SnipeIT Licenses endpoints:
https://snipe-it.readme.io/reference/licenses

:Copyright: (c) 2025 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/licenses.go
package snipeit

import (
	"time"
)

// LicenseClient for chaining methods
type LicenseClient struct {
	*Client
}

// Entry point for asset-related operations
func (c *Client) Licenses() *LicenseClient {
	lc := &LicenseClient{
		Client: c,
	}

	return lc
}

/*
 * Query Parameters for Licenses
 */
type LicenseQuery struct {
	Name           string `url:"string,omitempty"`          // Specify the name of the license.
	ProductKey     string `url:"product_key,omitempty"`     // Specify the product key of the license.
	Limit          int    `url:"limit,omitempty"`           // Specify the number of results you wish to return. Defaults to 50.
	Offset         int    `url:"offset,omitempty"`          // Specify the number of results to skip before starting to return items. Defaults to 0.
	Search         string `url:"search,omitempty"`          // Search for an asset by asset tag, serial, or model number.
	OrderNumber    string `url:"order_number,omitempty"`    // Return only licenses associated with the specified order number.
	Sort           string `url:"sort,omitempty"`            // Sort the results by the specified column. Defaults to id.
	Order          string `url:"order,omitempty"`           // Sort the results in the specified order. Defaults to asc.
	Expand         string `url:"expand,omitempty"`          // Whether to include detailed information on categories, etc (true) or just the text name (false)
	PurchaseOrder  string `url:"purchase_order,omitempty"`  // Return only assets associated with the specified purchase order.
	LicenseName    string `url:"license_name,omitempty"`    // Name of the person on the license
	LicenseEmail   string `url:"license_email,omitempty"`   // Email address associated with license
	ManufacturerID int    `url:"manufacturer_id,omitempty"` // Return only assets associated with the specified manufacturer ID.
	SupplierID     int    `url:"supplier_id,omitempty"`     // Return only assets associated with the specified supplier ID.
	CategoryID     int    `url:"category_id,omitempty"`     // Return only assets associated with the specified category ID.
	DepreciationID int    `url:"depreciation_id,omitempty"` // Return only assets associated with the specified depreciation ID.
	Maintained     bool   `url:"maintained,omitempty"`      // True to return only maintained licenses
	Deleted        string `url:"deleted,omitempty"`         // Set to true to return deleted licenses
}

// ### LicenseQuery implements QueryInterface
// ---------------------------------------------------------------------
func (q *LicenseQuery) Copy() QueryInterface {
	return &LicenseQuery{
		Name:           q.Name,
		ProductKey:     q.ProductKey,
		Limit:          q.Limit,
		Offset:         q.Offset,
		Search:         q.Search,
		OrderNumber:    q.OrderNumber,
		Sort:           q.Sort,
		Order:          q.Order,
		Expand:         q.Expand,
		PurchaseOrder:  q.PurchaseOrder,
		LicenseName:    q.LicenseName,
		LicenseEmail:   q.LicenseEmail,
		ManufacturerID: q.ManufacturerID,
		SupplierID:     q.SupplierID,
		CategoryID:     q.CategoryID,
		DepreciationID: q.DepreciationID,
		Maintained:     q.Maintained,
		Deleted:        q.Deleted,
	}
}

func (q *LicenseQuery) GetLimit() int {
	return q.Limit
}

func (q *LicenseQuery) SetLimit(limit int) {
	q.Limit = limit
}

func (q *LicenseQuery) GetOffset() int {
	return q.Offset
}

func (q *LicenseQuery) SetOffset(offset int) {
	q.Offset = offset
}

// END OF QUERYINTERFACE METHODS
//---------------------------------------------------------------------

/*
 * List all Licenses in Snipe-IT
 * /api/v1/licenses
 * - https://snipe-it.readme.io/reference/licenses
 */
func (c *LicenseClient) GetAllLicenses() (*LicenseList, error) {
	url := c.BuildURL(Licenses)

	q := LicenseQuery{
		Limit:  500,
		Offset: 0,
	}

	var cache LicenseList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	licenses, err := doConcurrent[LicenseList](c.Client, "GET", url, &q, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching license list: %v", err)
	}

	c.SetCache(url, licenses, 5*time.Minute)
	return licenses, nil
}

/*
 * # Create an license in Snipe-IT
 * /api/v1/licenses
 * - https://snipe-it.readme.io/reference/testinput (Yes, this is the correct link)
 */
func (c *LicenseClient) CreateLicense(p *License[LicensePOST]) (*License[LicensePOST], error) {
	url := c.BuildURL(Licenses)

	license, err := do[SnipeITResponse[License[LicensePOST]]](c.Client, "POST", url, nil, p)
	if err != nil {
		c.Log.Fatalf("Error creating license: %v", err)
	}

	return license.Payload, nil
}
