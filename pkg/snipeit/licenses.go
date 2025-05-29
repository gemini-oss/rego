/*
# SnipeIT - Licenses

This package initializes all the methods for functions which interact with the SnipeIT Licenses endpoints:
https://snipe-it.readme.io/reference/licenses

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/licenses.go
package snipeit

import (
	"fmt"
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
 * # Get an license in Snipe-IT by ID
 * /api/v1/licenses/{id}
 * - https://snipe-it.readme.io/reference/licensesid
 */
func (c *LicenseClient) GetLicense(id uint32) (*License[LicenseGET], error) {
	url := c.BuildURL(Licenses, id)

	var cache License[LicenseGET]
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	license, err := do[License[LicenseGET]](c.Client, "GET", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error getting license: %v", err)
	}

	c.SetCache(url, license, 5*time.Minute)
	return &license, nil
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

/*
 * # Return all unique Seat ID's for a particular license (by ID)
 * /api/v1/licenses/{id}/seats
 * - https://snipe-it.readme.io/reference/licensesidseats
 */
func (c *LicenseClient) Seats(id uint32) (*SeatList, error) {
	url := c.BuildURL(Licenses, id, "seats")

	var cache SeatList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	seats, err := do[SeatList](c.Client, "GET", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching seat list: %v", err)
	}

	c.SetCache(url, seats, 5*time.Minute)
	return &seats, nil
}

/*
 * # Return a Seat for a particular license (by ID)
 * /api/v1/licenses/{id}/seats/{seat_id}
 * - https://snipe-it.readme.io/reference/licensesidseatsseat_id
 */
func (c *LicenseClient) Seat(licenseID, seatID uint32) (*Seat[SeatGET], error) {
	url := c.BuildURL(Licenses, licenseID, "seats")

	var cache Seat[SeatGET]
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	seat, err := do[Seat[SeatGET]](c.Client, "GET", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error creating license: %v", err)
	}

	c.SetCache(url, seat, 5*time.Minute)
	return &seat, nil
}

// CheckoutFunc is a function type for checking out a license's seat
type licenseCheckoutBuilder struct {
	client    *LicenseClient
	licenseID uint32
	seatID    uint32
	userID    *uint32
	assetID   *uint32
}

/*
 * # Checkout a Seat for a particular license (by ID)
 * /api/v1/licenses/{id}/seats/{seat_id}
 * - https://snipe-it.readme.io/reference/licensesidseatsseat_id-2
 */
func (c *LicenseClient) Checkout(licenseID, seatID uint32) *licenseCheckoutBuilder {
	return &licenseCheckoutBuilder{
		client:    c,
		licenseID: licenseID,
		seatID:    seatID,
	}
}

func (b *licenseCheckoutBuilder) ToUser(userID uint32) (*Seat[SeatPOST], error) {
	b.userID = &userID
	return b.commit()
}

func (b *licenseCheckoutBuilder) ToAsset(assetID uint32) (*Seat[SeatPOST], error) {
	b.assetID = &assetID
	return b.commit()
}

func (b *licenseCheckoutBuilder) commit() (*Seat[SeatPOST], error) {
	if (b.userID != nil && b.assetID != nil) || (b.userID == nil && b.assetID == nil) {
		// Valid for check-in (both nil), or must only use one
		if !(b.userID == nil && b.assetID == nil) {
			return nil, fmt.Errorf("must assign either userID or assetID, not both")
		}
	}

	url := b.client.BuildURL(Licenses, b.licenseID, "seats", b.seatID)
	body := SeatPOST{
		PPPD: PPPD{
			AssetID: b.assetID,
		},
		AssignedTo: b.userID,
	}

	seat, err := do[SnipeITResponse[Seat[SeatPOST]]](b.client.Client, "PATCH", url, nil, body)
	if err != nil {
		return nil, fmt.Errorf("error updating license seat: %w", err)
	}

	b.client.SetCache(url, seat, 5*time.Minute)
	return seat.Payload, nil
}

/*
 * # Checkout a Seat for a particular license (by ID)
 * /api/v1/licenses/{id}/seats/{seat_id}
 * - https://snipe-it.readme.io/reference/licensesidseatsseat_id
 */
func (c *LicenseClient) Checkin(licenseID, seatID uint32) (*Seat[SeatPOST], error) {
	url := c.BuildURL(Licenses, licenseID, "seats", seatID)

	body := SeatPOST{
		PPPD: PPPD{
			AssetID: nil,
		},
		AssignedTo: nil,
	}

	seat, err := do[SnipeITResponse[Seat[SeatPOST]]](c.Client, "PATCH", url, nil, body)
	if err != nil {
		c.Log.Fatalf("Error checking in license: %v", err)
	}

	c.SetCache(url, seat, 5*time.Minute)
	return seat.Payload, nil
}
