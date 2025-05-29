/*
# SnipeIT - Assets

This package initializes all the methods for functions which interact with the SnipeIT Assets endpoints:
https://snipe-it.readme.io/reference/hardware-list

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/assets.go
package snipeit

import (
	"time"
)

// AssetClient for chaining methods
type AssetClient struct {
	*Client
}

// Entry point for asset-related operations
func (c *Client) Assets() *AssetClient {
	ac := &AssetClient{
		Client: c,
	}

	return ac
}

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

// ### AssetQuery implements QueryInterface
// ---------------------------------------------------------------------
func (q *AssetQuery) Copy() QueryInterface {
	return &AssetQuery{
		Limit:          q.Limit,
		Offset:         q.Offset,
		Search:         q.Search,
		OrderNumber:    q.OrderNumber,
		Sort:           q.Sort,
		Order:          q.Order,
		ModelID:        q.ModelID,
		CategoryID:     q.CategoryID,
		ManufacturerID: q.ManufacturerID,
		CompanyID:      q.CompanyID,
		LocationID:     q.LocationID,
		Status:         q.Status,
		StatusID:       q.StatusID,
	}
}

func (q *AssetQuery) GetLimit() int {
	return q.Limit
}

func (q *AssetQuery) SetLimit(limit int) {
	q.Limit = limit
}

func (q *AssetQuery) GetOffset() int {
	return q.Offset
}

func (q *AssetQuery) SetOffset(offset int) {
	q.Offset = offset
}

// END OF QUERYINTERFACE METHODS
//---------------------------------------------------------------------

/*
 * List all Hardware Assets in Snipe-IT
 * /api/v1/hardware
 * - https://snipe-it.readme.io/reference/hardware-list
 */
func (c *AssetClient) GetAllAssets() (*HardwareList, error) {
	url := c.BuildURL(Assets)

	q := AssetQuery{
		Limit:  500,
		Offset: 0,
	}

	var cache HardwareList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	assets, err := doConcurrent[HardwareList](c.Client, "GET", url, &q, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching hardware list: %v", err)
	}

	c.SetCache(url, assets, 5*time.Minute)
	return assets, nil
}

/*
 * Get Hardware Assets by Serial
 * /api/v1/hardware/byserial/{serial}
 * - https://snipe-it.readme.io/reference/hardware-by-serial
 */
func (c *AssetClient) GetAssetBySerial(serial string) (*Hardware[HardwareGET], error) {
	url := c.BuildURL(Assets, "byserial", serial)

	var cache Hardware[HardwareGET]
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	asset, err := do[HardwareList](c.Client, "GET", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching hardware asset: %v", err)
	}

	c.SetCache(url, asset, 5*time.Minute)
	return (*asset.Rows)[0], nil
}

/*
 * Get Hardware Assets by Tag
 * /api/v1/hardware/bytag/{tag}
 * - https://snipe-it.readme.io/reference/hardware-by-asset-tag
 */
func (c *AssetClient) GetAssetByTag(tag string) (*Hardware[HardwareGET], error) {
	url := c.BuildURL(Assets, "bytag", tag)

	var cache Hardware[HardwareGET]
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	asset, err := do[HardwareList](c.Client, "GET", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching hardware asset: %v", err)
	}

	c.SetCache(url, asset, 5*time.Minute)
	return (*asset.Rows)[0], nil
}

/*
 * # Create an asset in Snipe-IT
 * /api/v1/hardware
 * - https://snipe-it.readme.io/reference/hardware-create
 */
func (c *AssetClient) CreateAsset(p *Hardware[HardwarePOST]) (*Hardware[HardwarePOST], error) {
	url := c.BuildURL(Assets)

	hardware, err := do[SnipeITResponse[Hardware[HardwarePOST]]](c.Client, "POST", url, nil, p)
	if err != nil {
		c.Log.Fatalf("Error creating asset: %v", err)
	}

	return hardware.Payload, nil
}

/* Partially updates a specific asset in Snipe-IT
 * /api/v1/hardware/{id}
 * - https://snipe-it.readme.io/reference/hardware-partial-update
 */
func (c *AssetClient) PartialUpdateAsset(id uint32, h *Hardware[HardwarePUTPATCH]) (*Hardware[HardwarePUTPATCH], error) {
	url := c.BuildURL(Assets, id)

	hardware, err := do[SnipeITResponse[Hardware[HardwarePUTPATCH]]](c.Client, "PATCH", url, nil, h)
	if err != nil {
		c.Log.Fatalf("Error updating asset: %v", err)
	}

	c.SetCache(url, hardware, 5*time.Minute)
	return hardware.Payload, nil
}

/*
 * # Delete an asset in Snipe-IT
 * /api/v1/hardware/{id}
 * - https://snipe-it.readme.io/reference/hardware-delete
 */
func (c *AssetClient) DeleteAsset(id int64) (string, error) {
	url := c.BuildURL(Assets, id)

	hardware, err := do[SnipeITResponse[Hardware[PPPD]]](c.Client, "DELETE", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error deleting asset: %v", err)
	}

	switch hardware.Messages.StringValue {
	case "The asset was deleted successfully.", "Asset does not exist.":
		return hardware.Messages.StringValue, err
	default:
		return hardware.Status, err
	}
}
