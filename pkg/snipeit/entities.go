/*
# SnipeIT - Entities (Structs)

This package initializes all the structs for the SnipeIT API:
https://snipe-it.readme.io/reference/api-overview

:Copyright: (c) 2025 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/entities.go
package snipeit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/generics"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ### SnipeIT Client Structs
// ---------------------------------------------------------------------
type Client struct {
	BaseURL string           // BaseURL is the base URL for the SnipeIT API.
	HTTP    *requests.Client // HTTP client for the SnipeIT API.
	Log     *log.Logger      // Log is the logger for the SnipeIT API.
	Cache   *cache.Cache     // Cache for the SnipeIT API.
}

// PaginatedList is a generic structure representing a paginated response from SnipeIT with items of any type.
type PaginatedList[E any] struct {
	Total int   `json:"total,omitempty"` // The total number of items.
	Rows  *[]*E `json:"rows,omitempty"`  // An array of items.
}

func (pl PaginatedList[E]) TotalCount() int {
	return pl.Total
}

func (pl PaginatedList[E]) Append(elements *[]*E) {
	*pl.Rows = append(*pl.Rows, *elements...)
}

func (pl PaginatedList[E]) Elements() *[]*E {
	return pl.Rows
}

func (pl PaginatedList[E]) Map() map[interface{}]*E {
	result := make(map[interface{}]*E)
	for _, item := range *pl.Rows {
		switch entity := any(item).(type) {
		case *Hardware[any]:
			result[entity.Serial] = item
		case *User:
			result[entity.Email] = item
		default:
			// Fallback to the `ID` field if available
			value := reflect.ValueOf(item).Elem()
			if idField := value.FieldByName("ID"); idField.IsValid() {
				id := idField.Interface()
				result[id] = item
			}
		}
	}
	return result
}

// SnipeITResponse is an interface for Snipe-IT API responses
type SnipeITResponse[E any] struct {
	Status   string   `json:"status,omitempty"`   // Status of the response
	Messages Messages `json:"messages,omitempty"` // Messages associated with the response
	Error    string   `json:"error,omitempty"`    // Error associated with the response
	Payload  *E       `json:"payload,omitempty"`  // Payload of the response -- can be an object of any type
}

type Messages struct {
	StringValue string
	MapValue    map[string][]string
	IsString    bool
}

func (m *Messages) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		m.StringValue = str
		m.IsString = true
		return nil
	}

	// Try to unmarshal as a map
	var messageMap map[string][]string
	if err := json.Unmarshal(data, &messageMap); err == nil {
		m.MapValue = messageMap
		m.IsString = false
		return nil
	}

	return fmt.Errorf("unable to unmarshal message into string or map")
}

// SnipeIT Common fields
type SnipeIT struct {
	*Record `json:",inline"`

	// CreatedAt *Timestamp `json:"created_at,omitempty"` // Time when the item was created.
	// UpdatedAt *Timestamp `json:"updated_at,omitempty"` // Time when the item was last updated.
	// DeletedAt *Timestamp `json:"deleted_at,omitempty"` // Time when the item was deleted.

	AvailableActions *AvailableActions `json:"available_actions,omitempty"` // Available actions for the entity
}

// SnipeIT {GET} fields
type GET struct {
	CreatedAt *DateInfo `json:"created_at,omitempty"` // Time when the item was created.
	UpdatedAt *DateInfo `json:"updated_at,omitempty"` // Time when the item was last updated.
	DeletedAt *DateInfo `json:"deleted_at,omitempty"` // Time when the item was deleted.

	Category     *Record `json:"category,omitempty"`     // Category of the hardware item.
	Company      *Record `json:"company,omitempty"`      // {HARDWARE,USER} Company
	Department   *Record `json:"department,omitempty"`   // {USER} Department
	Depreciation *Record `json:"depreciation,omitempty"` // {MODEL} Depreciation
	Location     *Record `json:"location,omitempty"`     // {HARDWARE,MODEL,USER} Location of the entity
	Manager      *Record `json:"manager,omitempty"`      // {LOCATION,USER} Manager
	Manufacturer *Record `json:"manufacturer,omitempty"` // {ACCESSORY,HARDWARE} Manufacturer
	Parent       *Record `json:"parent,omitempty"`       // {Location} Parent of the location
	RTDLocation  *Record `json:"rtd_location,omitempty"` // {HARDWARE} RTD [Ready to Deploy] location
	Supplier     *Record `json:"supplier,omitempty"`     // Supplier of the hardware item.
}

// SnipeIT {POST, PUT, PATCH, DELETE} fields
type PPPD struct {
	CreatedAt      *string `json:"created_at,omitempty"`      // Time when the item was created.
	UpdatedAt      *string `json:"updated_at,omitempty"`      // Time when the item was last updated.
	DeletedAt      *string `json:"deleted_at,omitempty"`      // Time when the item was deleted.
	CategoryID     *uint32 `json:"category_id,omitempty"`     // {MODEL} Category ID
	CompanyID      *uint32 `json:"company_id,omitempty"`      // Company ID
	DepartmentID   *uint32 `json:"department_id,omitempty"`   // Department ID
	DepreciationID *uint32 `json:"depreciation_id,omitempty"` // {MODEL} Depreciation ID
	FieldsetID     *uint32 `json:"fieldset_id,omitempty"`     // {MODEL} Fieldset ID
	LocationID     *uint32 `json:"location_id,omitempty"`     // Location ID
	ManufacturerID *uint32 `json:"manufacturer_id,omitempty"` // {MODEL} Manufacturer ID
	ModelID        *uint32 `json:"model_id,omitempty"`        // Model ID
	ParentID       *uint32 `json:"parent_id,omitempty"`       // {Location} Parent ID
	RTDLocationID  *uint32 `json:"rtd_location_id,omitempty"` // RTD Location ID
	StatusID       *uint32 `json:"status_id,omitempty"`       // {HARDWARE} Status ID
	SupplierID     *uint32 `json:"supplier_id,omitempty"`     // {HARDWARE} Supplier ID
}

// QueryInterface defines methods for queries with pagination and filtering
type QueryInterface interface {
	Copy() QueryInterface
	GetLimit() int
	SetLimit(int)
	GetOffset() int
	SetOffset(int)
}

// END OF SNIPEIT CLIENT STRUCTS
//---------------------------------------------------------------------

// ### Assets
// -------------------------------------------------------------------------
// Source: https://snipe-it.readme.io/reference/hardware-list
type HardwareList = PaginatedList[Hardware[HardwareGET]]

type HardwareBase struct {
	*SnipeIT       `json:",inline"`
	AssetTag       string  `json:"asset_tag,omitempty"`       // Asset tag of the hardware item.
	Serial         string  `json:"serial,omitempty"`          // Serial number of the hardware item.
	OrderNumber    string  `json:"order_number,omitempty"`    // Order number of the hardware item.
	Notes          string  `json:"notes,omitempty"`           // Notes associated with the hardware item.
	WarrantyMonths *string `json:"warranty_months,omitempty"` // Warranty months of the hardware item.
	//Requestable    *bool      `json:"requestable,omitempty"`     // Whether the hardware item is requestable.
}

type HardwareGET struct {
	GET          `json:",inline"`
	Archived     string    `json:"archived,omitempty"`      // Whether the hardware item is archived (string on GET, bool on PPPD)
	AssignedTo   *User     `json:"assigned_to,omitempty"`   // User to whom the hardware item is assigned. (object on GET, string on POST)
	PurchaseCost string    `json:"purchase_cost,omitempty"` // Purchase cost of the hardware item. (string on GET, float on POST)
	PurchaseDate *DateInfo `json:"purchase_date,omitempty"` // Purchase date of the hardware item. (object on GET, string on PPPD)
}

type HardwarePOST struct {
	/*
		You can do a checkout on creation if you add one of the following fields: assigned_user, assigned_asset, or assigned_location. This should be a valid primary key of the user, asset or location you wish to checkout to.
	*/
	PPPD         `json:",inline"`
	Archived     bool     `json:"archived,omitempty"`      // Whether the hardware item is archived (string on GET, bool on PPPD)
	AssignedTo   *string  `json:"assigned_to,omitempty"`   // User to whom the hardware item is assigned. (object on GET, string on POST)
	BYOD         bool     `json:"byod,omitempty"`          // Whether the hardware item is BYOD (bool on POST, int32 on PUT,PATCH)
	PurchaseCost *float64 `json:"purchase_cost,omitempty"` // Purchase cost of the hardware item. (string on GET, float on POST,PATCH)
}

type HardwarePUTPATCH struct {
	PPPD         `json:",inline"`
	Archived     bool     `json:"archived,omitempty"`      // Whether the hardware item is archived (string on GET, bool on PPPD)
	AssignedTo   uint32   `json:"assigned_to,omitempty"`   // User ID to whom the hardware item is assigned. (object on GET, string on POST)
	BYOD         uint32   `json:"byod,omitempty"`          // Whether the hardware item is BYOD (bool on POST, int32 on PUT,PATCH)
	PurchaseCost *float64 `json:"purchase_cost,omitempty"` // Purchase cost of the hardware item. (string on GET, float on POST,PATCH)
}

// Hardware represents an individual hardware item.
// https://snipe-it.readme.io/reference/hardware-list#sortable-columns
type Hardware[M any] struct {
	*HardwareBase   `json:",inline"`
	Method          M                             `json:",inline"`
	Model           *Model[GET]                   `json:"model,omitempty"`             // Model of the hardware item.
	ModelNumber     string                        `json:"model_number,omitempty"`      // Model number of the hardware item.
	EOL             *Timestamp                    `json:"eol,omitempty"`               // End of life of a hardware item.
	AssetEOLDate    *Timestamp                    `json:"asset_eol_date,omitempty"`    // Asset end of life date of the hardware item.
	StatusLabel     *StatusLabel                  `json:"status_label,omitempty"`      // Status label of the hardware item.
	Image           string                        `json:"image,omitempty"`             // Image of the hardware item.
	QR              string                        `json:"qr,omitempty"`                // QR code of the hardware item.
	AltBarcode      string                        `json:"alt_barcode,omitempty"`       // Alternate barcode of the hardware item.
	WarrantyExpires string                        `json:"warranty_expires,omitempty"`  // Warranty expiry date of the hardware item.
	LastAuditDate   *string                       `json:"last_audit_date,omitempty"`   // Last audit date of the hardware item.
	NextAuditDate   *string                       `json:"next_audit_date,omitempty"`   // Next audit date of the hardware item.
	Age             string                        `json:"age,omitempty"`               // Age of the hardware item.
	LastCheckout    *Timestamp                    `json:"last_checkout,omitempty"`     // Time when the hardware item was last checked out.
	ExpectedCheckin *Timestamp                    `json:"expected_checkin,omitempty"`  // Expected check-in date of the hardware item.
	CheckinCounter  int                           `json:"checkin_counter,omitempty"`   // Check-in counter of the hardware item.
	CheckoutCounter int                           `json:"checkout_counter,omitempty"`  // Check-out counter of the hardware item.
	RequestsCounter int                           `json:"requests_counter,omitempty"`  // Request counter of the hardware item.
	UserCanCheckout bool                          `json:"user_can_checkout,omitempty"` // Whether the user can check-out the hardware item.
	CustomFields    *map[string]map[string]string `json:"custom_fields,omitempty"`     // Custom fields of a Snipe-IT asset (This will typically be the `DB Field` property in the WebUI)
}

func (h *Hardware[M]) UnmarshalJSON(data []byte) error {
	hw, err := generics.UnmarshalGeneric[Hardware[M], M](data)
	if err != nil {
		return err
	}

	*h = *hw
	return nil
}

// END OF ASSETS STRUCTS
//-------------------------------------------------------------------------

// ### Accessories
// -------------------------------------------------------------------------
// Source: https://snipe-it.readme.io/reference/accessories
type AccessoryList = PaginatedList[Accessory]

// Accessory represents an individual accessory.
// https://snipe-it.readme.io/reference/accessories#sortable-columns
type Accessory struct {
	*SnipeIT        `json:",inline"`
	Category        *Record `json:"category,omitempty"`          // Name and ID of the accessory's category
	Image           string  `json:"image,omitempty"`             // URL of the accessory's image
	MinQty          int     `json:"min_qty,omitempty"`           // Minimum quantity of the accessory
	ModelNumber     string  `json:"model_number,omitempty"`      // Model number of the accessory
	Notes           string  `json:"notes,omitempty"`             // Notes about the accessory
	OrderNumber     string  `json:"order_number,omitempty"`      // Order number associated with the accessory
	PurchaseCost    string  `json:"purchase_cost,omitempty"`     // Purchase cost of the accessory
	PurchaseDate    string  `json:"purchase_date,omitempty"`     // Purchase date of the accessory
	Qty             int     `json:"qty,omitempty"`               // Quantity of the accessory
	RemainingQty    int     `json:"remaining_qty,omitempty"`     // Remaining quantity of the accessory
	UserCanCheckout bool    `json:"user_can_checkout,omitempty"` // If the user can checkout the accessory
}

// END OF ACCESSORIES STRUCTS
//-------------------------------------------------------------------------

// ### Categories
// -------------------------------------------------------------------------
// Source: https://snipe-it.readme.io/reference/categories
type CategoryList = PaginatedList[Category]

// Category represents an individual category.
// https://snipe-it.readme.io/reference/categories#sortable-columns
type Category struct {
	*SnipeIT          `json:",inline"`
	Image             string `json:"image,omitempty"`
	CategoryType      string `json:"category_type,omitempty"`
	EULA              bool   `json:"eula,omitempty"`
	CheckinEmail      bool   `json:"checkin_email,omitempty"`
	RequireAcceptance bool   `json:"require_acceptance,omitempty"`
	AssetsCount       int64  `json:"assets_count,omitempty"`
	AccessoriesCount  int64  `json:"accessories_count,omitempty"`
	ConsumablesCount  int64  `json:"consumables_count,omitempty"`
	ComponentsCount   int64  `json:"components_count,omitempty"`
	LicensesCount     int64  `json:"licenses_count,omitempty"`
}

// END OF CATEGORY STRUCTS
//-------------------------------------------------------------------------

// ### Locations
// -------------------------------------------------------------------------
type LocationList = PaginatedList[Location]

type Location struct {
	*SnipeIT       `json:",inline"`
	Image          string      `json:"image,omitempty"`                 // The URL of the location's image.
	Address        string      `json:"address,omitempty"`               // The address of the location.
	Address2       string      `json:"address2,omitempty"`              // The second address line of the location.
	City           string      `json:"city,omitempty"`                  // The city of the location.
	State          string      `json:"state,omitempty"`                 // The state of the location.
	Country        string      `json:"country,omitempty"`               // The country of the location.
	Zip            string      `json:"zip,omitempty"`                   // The zip code of the location.
	AssetsAssigned int         `json:"assigned_assets_count,omitempty"` // The number of assets assigned to the location.
	Assets         int         `json:"assets_count,omitempty"`          // The number of assets at the location.
	RTDAssets      int         `json:"rtd_assets_count,omitempty"`      // The number of assets ready to deploy at the location.
	Users          int         `json:"users_count,omitempty"`           // The number of users at the location.
	Currency       string      `json:"currency,omitempty"`              // The currency of the location.
	LDAP           interface{} `json:"ldap_ou,omitempty"`               // The LDAP OU of the location.
	Children       []Location  `json:"children,omitempty"`              // The children of the location.
}

// END OF LOCATION STRUCTS
//-------------------------------------------------------------------------

// ### Maintenances
// -------------------------------------------------------------------------
type MaintenanceList = PaginatedList[Maintenance]

type Maintenance struct {
	ID               int64             `json:"id,omitempty"`                     // ID of the maintenance entry
	Asset            *Hardware         `json:"asset,omitempty"`                  // Asset object on maintenance
	Model            *Record           `json:"model,omitempty"`                  // Model of the hardware item on maintenance
	StatusLabel      *StatusLabel      `json:"status_label,omitempty"`           // StatusLabel of the hardware item on maintenance
	Company          *Record           `json:"company,omitempty"`                // Company of the hardware item on maintenance
	Title            string            `json:"title,omitempty"`                  // Title of the maintenance entry
	Location         *Record           `json:"location,omitempty"`               // Location of the hardware item.
	RTDLocation      *Record           `json:"rtd_location,omitempty"`           // RTD location of the hardware item.
	Notes            string            `json:"notes,omitempty"`                  // Notes on the maintenance entry
	Supplier         *Record           `json:"supplier,omitempty"`               // Supplier responsible for Maintenance
	Cost             string            `json:"cost,omitempty"`                   // Cost of performing the Maintenance
	Type             string            `json:"asset_maintenance_type,omitempty"` // Type of maintenance being performed on hardware
	StartDate        *DateInfo         `json:"start_date,omitempty"`             // Date that the maintenance started
	Time             int64             `json:"asset_maintenance_time,omitempty"` // Time the asset spent in maintenance, in days
	CompletionDate   *DateInfo         `json:"completion_date,omitempty"`        // Date that the maintenance was completed
	UserID           *Record           `json:"user_id,omitempty"`                // Record of the user that created the maintenance entry
	CreatedAt        *DateInfo         `json:"created_at,omitempty"`             // Date that the maintenance entry was created
	UpdatedAt        *DateInfo         `json:"updated_at,omitempty"`             // Last date that the maintenance entry was updated
	Warranty         int64             `json:"is_warranty,omitempty"`            // Is the maintenance entry part of the Warranty
	AvailableActions *AvailableActions `json:"available_actions,oimitempty"`     // AvailableActions on the maintenance entry
}

// END OF MAINTENANCES STRUCTS
//-------------------------------------------------------------------------

// ### Users
// -------------------------------------------------------------------------
type UserList = PaginatedList[User]

type User struct {
	*SnipeIT           `json:",inline"`
	Activated          bool              `json:"activated"`                     // Specifies if the user is active or not
	Address            string            `json:"address,omitempty"`             // Address of the user
	AssetsCount        int64             `json:"assets_count,omitempty"`        // Number of assets associated with the user
	AutoassignLicenses bool              `json:"autoassign_licenses,omitempty"` // Specifies if the licenses are automatically assigned to the user
	Avatar             string            `json:"avatar,omitempty"`              // URL of the user's avatar
	City               string            `json:"city,omitempty"`                // City of the user
	ConsumablesCount   int64             `json:"consumables_count,omitempty"`   // Count of consumables associated with the user
	Country            string            `json:"country,omitempty"`             // Country of the user
	CreatedBy          *DateInfo         `json:"created_by,omitempty"`          // Who created the user
	Email              string            `json:"email,omitempty"`               // Email of the user
	EmployeeNum        string            `json:"employee_num,omitempty"`        // Employee number of the user
	EndDate            *DateInfo         `json:"end_date,omitempty"`            // End date of the user
	FirstName          string            `json:"first_name,omitempty"`          // First name of the user
	Groups             interface{}       `json:"groups,omitempty"`              // Groups that the user belongs to
	ID                 int64             `json:"id,omitempty"`                  // ID of the user
	Jobtitle           string            `json:"jobtitle,omitempty"`            // Job title of the user
	LastLogin          string            `json:"last_login,omitempty"`          // Last login time of the user
	LastName           string            `json:"last_name,omitempty"`           // Last name of the user
	LdapImport         bool              `json:"ldap_import,omitempty"`         // Specifies if the user is imported from LDAP
	Locale             string            `json:"locale,omitempty"`              // Locale of the user
	Name               string            `json:"name,omitempty"`                // Full name of the user
	Notes              string            `json:"notes,omitempty"`               // Notes associated with the user
	Permissions        map[string]string `json:"permissions,omitempty"`         // Permissions of the user
	Phone              string            `json:"phone,omitempty"`               // Phone number of the user
	Remote             bool              `json:"remote,omitempty"`              // Specifies if the user is remote
	StartDate          *DateInfo         `json:"start_date,omitempty"`          // Start date of the user
	State              string            `json:"state,omitempty"`               // State of the user
	TwoFactorEnrolled  bool              `json:"two_factor_enrolled,omitempty"` // Specifies if the user has enrolled for two factor authentication
	TwoFactorOptin     bool              `json:"two_factor_optin,omitempty"`    // Specifies if the user has opted for two factor authentication
	Username           string            `json:"username,omitempty"`            // Username of the user
	Vip                bool              `json:"vip,omitempty"`                 // Specifies if the user is a VIP
	Website            string            `json:"website,omitempty"`             // Website of the user
	Zip                string            `json:"zip,omitempty"`                 // Zip code of the user
}

// END OF USER STRUCTS
//-------------------------------------------------------------------------

// ### Models
// -------------------------------------------------------------------------
// Source: https://snipe-it.readme.io/reference/models
type ModelList = PaginatedList[Model[GET]]

type ModelBase struct {
	*SnipeIT    `json:",inline"`
	ModelNumber string `json:"model_number,omitempty"` // Model number of the hardware item.
	Image       string `json:"image,omitempty"`        // Image of the hardware model.
	//Requestable          bool     `json:"requestable,omitempty"`            // Whether the hardware model is requestable.
	Notes                string   `json:"notes,omitempty"`                  // Notes of the hardware model.
	MinAmt               *float64 `json:"min_amt,omitempty"`                // Minimum amount of the hardware model.
	EOL                  int      `json:"eol,omitempty"`                    // End of life of the hardware model.
	DeprecatedMACAddress any      `json:"deprecated_mac_address,omitempty"` // Deprecated MAC address of the hardware model. (string on GET, number on PATCH)
}

type Model[M any] struct {
	*ModelBase `json:",inline"`
	Method     M `json:",inline"`
}

// END OF MODELS STRUCTS
//-------------------------------------------------------------------------

// ### Common Asset types
// -------------------------------------------------------------------------
// Record represents an id:name pairing for many types of records in Snipe-IT.
type Record struct {
	ID   uint32 `json:"id"`   // ID of the record {category, company, department, location, manufacturer, supplier, etc.}
	Name string `json:"name"` // Name of the record {category, company, department, location, manufacturer, supplier, etc.}
}

const (
	snipeTime = "2006-01-02 15:04:05"
	iso8601   = "2006-01-02T15:04:05.000000Z"
)

// Timestamp is a time.Time but JSON marshals/unmarshals as a string in the format "2006-01-02 15:04:05"
type Timestamp struct {
	time.Time
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(ts.Time.Format(snipeTime))
}

func (ts *Timestamp) UnmarshalJSON(b []byte) error {
	d := &DateInfo{}
	// If successful and the Date field is not empty, use it
	if err := json.Unmarshal(b, d); err == nil && d.Date != "" {
		return ts.parseDate(d.Date)
	}

	// If that fails, try to unmarshal directly into a string
	var dateStr string
	if err := json.Unmarshal(b, &dateStr); err != nil {
		return err
	}

	return ts.parseDate(dateStr)
}

// parseDate attempts to parse a date string in ISO 8601 format first, then falls back to MySQL DATETIME format.
func (ts *Timestamp) parseDate(dateStr string) error {
	t, err := time.Parse(iso8601, dateStr)
	if err != nil {
		t, err = time.Parse(snipeTime, dateStr)
		if err != nil {
			return err
		}
	}

	ts.Time = t
	return nil
}

// DateInfo represents a date and its formatted representation.
type DateInfo struct {
	Date      string `json:"datetime,omitempty"`  // The date in yyyy-mm-dd format.
	Formatted string `json:"formatted,omitempty"` // The formatted date.
}

// StatusLabel represents the status label of a hardware item.
type StatusLabel struct {
	*Record    `json:",inline"`
	StatusMeta string `json:"status_meta,omitempty"` // Meta status of the status label.
	StatusType string `json:"status_type,omitempty"` // Type of the status label.
}

// AvailableActions represents the available actions for a hardware item.
type AvailableActions struct {
	Checkin  bool `json:"checkin,omitempty"`  // Whether check-in action is available.
	Checkout bool `json:"checkout,omitempty"` // Whether check-out action is available.
	Clone    bool `json:"clone,omitempty"`    // Whether clone action is available.
	Delete   bool `json:"delete,omitempty"`   // Whether delete action is available.
	Restore  bool `json:"restore,omitempty"`  // Whether restore action is available.
	Update   bool `json:"update,omitempty"`   // Whether update action is available.
}

// END OF COMMON ASSET TYPES
//-------------------------------------------------------------------------
