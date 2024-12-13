/*
# SnipeIT - Entities (Structs)

This package initializes all the structs for the SnipeIT API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/entities.go
package snipeit

import (
	"reflect"

	"github.com/gemini-oss/rego/pkg/common/cache"
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
		case *Hardware:
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
	Status   string `json:"status,omitempty"`   // Status of the response
	Messages string `json:"messages,omitempty"` // Messages associated with the response
	Error    string `json:"error,omitempty"`    // Error associated with the respons
	Payload  *E     `json:"payload,omitempty"`  // Payload of the response -- can be an object of any type
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
type HardwareList = PaginatedList[Hardware]

// Hardware represents an individual hardware item.
// https://snipe-it.readme.io/reference/hardware-list#sortable-columns
type Hardware struct {
	ID               int               `json:"id,omitempty"`                // ID of the hardware item.
	Name             string            `json:"name,omitempty"`              // Name of the hardware item.
	AssetTag         string            `json:"asset_tag,omitempty"`         // Asset tag of the hardware item.
	Serial           string            `json:"serial,omitempty"`            // Serial number of the hardware item.
	Model            *Record           `json:"model,omitempty"`             // Model of the hardware item.
	BYOD             bool              `json:"byod,omitempty"`              // Whether the hardware item is BYOD.
	ModelNumber      string            `json:"model_number,omitempty"`      // Model number of the hardware item.
	EOL              int               `json:"eol,omitempty"`               // End of life of the hardware item.
	AssetEOLDate     *DateInfo         `json:"asset_eol_date,omitempty"`    // Asset end of life date of the hardware item.
	StatusLabel      *StatusLabel      `json:"status_label,omitempty"`      // Status label of the hardware item.
	Category         *Record           `json:"category,omitempty"`          // Category of the hardware item.
	Manufacturer     *Record           `json:"manufacturer,omitempty"`      // Manufacturer of the hardware item.
	Supplier         *Record           `json:"supplier,omitempty"`          // Supplier of the hardware item.
	Notes            string            `json:"notes,omitempty"`             // Notes associated with the hardware item.
	OrderNumber      string            `json:"order_number,omitempty"`      // Order number of the hardware item.
	Company          *Record           `json:"company,omitempty"`           // Company of the hardware item.
	Location         *Record           `json:"location,omitempty"`          // Location of the hardware item.
	RTDLocation      *Record           `json:"rtd_location,omitempty"`      // RTD location of the hardware item.
	Image            string            `json:"image,omitempty"`             // Image of the hardware item.
	QR               string            `json:"qr,omitempty"`                // QR code of the hardware item.
	AltBarcode       string            `json:"alt_barcode,omitempty"`       // Alternate barcode of the hardware item.
	AssignedTo       *User             `json:"assigned_to,omitempty"`       // User to whom the hardware item is assigned.
	WarrantyMonths   string            `json:"warranty_months,omitempty"`   // Warranty months of the hardware item.
	WarrantyExpires  string            `json:"warranty_expires,omitempty"`  // Warranty expiry date of the hardware item.
	CreatedAt        *DateInfo         `json:"created_at,omitempty"`        // Time when the hardware item was created.
	UpdatedAt        *DateInfo         `json:"updated_at,omitempty"`        // Time when the hardware item was last updated.
	LastAuditDate    *DateInfo         `json:"last_audit_date,omitempty"`   // Last audit date of the hardware item.
	NextAuditDate    *DateInfo         `json:"next_audit_date,omitempty"`   // Next audit date of the hardware item.
	DeletedAt        string            `json:"deleted_at,omitempty"`        // Time when the hardware item was deleted.
	PurchaseDate     *DateInfo         `json:"purchase_date,omitempty"`     // Purchase date of the hardware item.
	Age              string            `json:"age,omitempty"`               // Age of the hardware item.
	LastCheckout     *DateInfo         `json:"last_checkout,omitempty"`     // Time when the hardware item was last checked out.
	ExpectedCheckin  *DateInfo         `json:"expected_checkin,omitempty"`  // Expected check-in date of the hardware item.
	PurchaseCost     string            `json:"purchase_cost,omitempty"`     // Purchase cost of the hardware item.
	CheckinCounter   int               `json:"checkin_counter,omitempty"`   // Check-in counter of the hardware item.
	CheckoutCounter  int               `json:"checkout_counter,omitempty"`  // Check-out counter of the hardware item.
	RequestsCounter  int               `json:"requests_counter,omitempty"`  // Request counter of the hardware item.
	UserCanCheckout  bool              `json:"user_can_checkout,omitempty"` // Whether the user can check-out the hardware item.
	CustomFields     *CustomFields     `json:"custom_fields,omitempty"`     // Custom fields of the hardware item.
	AvailableActions *AvailableActions `json:"available_actions,omitempty"` // Available actions for the hardware item.
	CustomAssetFields
}

type CustomAssetFields map[string]interface{} // Custom fields of a Snipe-IT asset (This will typically be the `DB Field` property in the WebUI)

// END OF ASSETS STRUCTS
//-------------------------------------------------------------------------

// ### Accessories
// -------------------------------------------------------------------------
// Source: https://snipe-it.readme.io/reference/accessories
type AccessoryList = PaginatedList[Accessory]

// Accessory represents an individual accessory.
// https://snipe-it.readme.io/reference/accessories#sortable-columns
type Accessory struct {
	AvailableActions *AvailableActions `json:"available_actions,omitempty"` // Actions that are available for the row
	Category         *Record           `json:"category,omitempty"`          // Name and ID of the accessory's category
	Company          string            `json:"company,omitempty"`           // Company associated with the accessory, if applicable.
	CreatedAt        *DateInfo         `json:"created_at,omitempty"`        // When the accessory was created
	ID               int               `json:"id,omitempty"`                // Asset ID
	Image            string            `json:"image,omitempty"`             // URL of the accessory's image
	Location         *Record           `json:"location,omitempty"`          // Name and ID of the accessory's location
	Manufacturer     *Record           `json:"manufacturer,omitempty"`      // Name and ID of the accessory's manufacturer
	MinQty           int               `json:"min_qty,omitempty"`           // Minimum quantity of the accessory
	ModelNumber      string            `json:"model_number,omitempty"`      // Model number of the accessory
	Name             string            `json:"name,omitempty"`              // Asset Name
	Notes            string            `json:"notes,omitempty"`             // Notes about the accessory
	OrderNumber      string            `json:"order_number,omitempty"`      // Order number associated with the accessory
	PurchaseCost     string            `json:"purchase_cost,omitempty"`     // Purchase cost of the accessory
	PurchaseDate     string            `json:"purchase_date,omitempty"`     // Purchase date of the accessory
	Qty              int               `json:"qty,omitempty"`               // Quantity of the accessory
	RemainingQty     int               `json:"remaining_qty,omitempty"`     // Remaining quantity of the accessory
	Supplier         *Record           `json:"supplier,omitempty"`          // Name and ID of the accessory's supplier
	UpdatedAt        *DateInfo         `json:"updated_at,omitempty"`        // When the accessory was updated
	UserCanCheckout  bool              `json:"user_can_checkout,omitempty"` // If the user can checkout the accessory
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
	ID                int64             `json:"id,omitempty"`
	Name              string            `json:"name,omitempty"`
	Image             string            `json:"image,omitempty"`
	CategoryType      string            `json:"category_type,omitempty"`
	EULA              bool              `json:"eula,omitempty"`
	CheckinEmail      bool              `json:"checkin_email,omitempty"`
	RequireAcceptance bool              `json:"require_acceptance,omitempty"`
	AssetsCount       int64             `json:"assets_count,omitempty"`
	AccessoriesCount  int64             `json:"accessories_count,omitempty"`
	ConsumablesCount  int64             `json:"consumables_count,omitempty"`
	ComponentsCount   int64             `json:"components_count,omitempty"`
	LicensesCount     int64             `json:"licenses_count,omitempty"`
	CreatedAt         *DateInfo         `json:"created_at,omitempty"`
	UpdatedAt         *DateInfo         `json:"updated_at,omitempty"`
	Actions           *AvailableActions `json:"available_actions,omitempty"`
}

// END OF CATEGORY STRUCTS
//-------------------------------------------------------------------------

// ### Locations
// -------------------------------------------------------------------------
type LocationList = PaginatedList[Location]

type Location struct {
	ID             int              `json:"id,omitempty"`                    // The ID of the location.
	Name           string           `json:"name,omitempty"`                  // The name of the location.
	Image          string           `json:"image,omitempty"`                 // The URL of the location's image.
	Address        string           `json:"address,omitempty"`               // The address of the location.
	Address2       string           `json:"address2,omitempty"`              // The second address line of the location.
	City           string           `json:"city,omitempty"`                  // The city of the location.
	State          string           `json:"state,omitempty"`                 // The state of the location.
	Country        string           `json:"country,omitempty"`               // The country of the location.
	Zip            string           `json:"zip,omitempty"`                   // The zip code of the location.
	AssetsAssigned int              `json:"assigned_assets_count,omitempty"` // The number of assets assigned to the location.
	Assets         int              `json:"assets_count,omitempty"`          // The number of assets at the location.
	RTDAssets      int              `json:"rtd_assets_count,omitempty"`      // The number of assets ready to deploy at the location.
	Users          int              `json:"users_count,omitempty"`           // The number of users at the location.
	Currency       string           `json:"currency,omitempty"`              // The currency of the location.
	LDAP           interface{}      `json:"ldap_ou,omitempty"`               // The LDAP OU of the location.
	CreatedAt      *DateInfo        `json:"created_at,omitempty"`            // The date the location was created.
	UpdatedAt      *DateInfo        `json:"updated_at,omitempty"`            // The date the location was updated.
	Parent         *Record          `json:"parent,omitempty"`                // The parent location of the location.
	ParentID       int              `json:"parent_id,omitempty"`             // The ID of the parent location.
	Manager        *Record          `json:"manager,omitempty"`               // The manager of the location.
	ManagerID      int              `json:"manager_id,omitempty"`            // The ID of the manager.
	Children       []Location       `json:"children,omitempty"`              // The children of the location.
	Actions        AvailableActions `json:"available_actions,omitempty"`     // The available actions on the location.
}

// END OF LOCATION STRUCTS
//-------------------------------------------------------------------------

// ### Asset Maintenances
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

// END OF Asset Maintenances STRUCTS
//-------------------------------------------------------------------------

// ### Users
// -------------------------------------------------------------------------
type UserList = PaginatedList[User]

type User struct {
	Activated          bool              `json:"activated"`                     // Specifies if the user is active or not
	Address            string            `json:"address,omitempty"`             // Address of the user
	AssetsCount        int64             `json:"assets_count,omitempty"`        // Number of assets associated with the user
	AutoassignLicenses bool              `json:"autoassign_licenses,omitempty"` // Specifies if the licenses are automatically assigned to the user
	Avatar             string            `json:"avatar,omitempty"`              // URL of the user's avatar
	AvailableActions   AvailableActions  `json:"available_actions,omitempty"`   // Available actions on the user profile
	City               string            `json:"city,omitempty"`                // City of the user
	Company            Record            `json:"company,omitempty"`             // Company associated with the user
	ConsumablesCount   int64             `json:"consumables_count,omitempty"`   // Count of consumables associated with the user
	Country            string            `json:"country,omitempty"`             // Country of the user
	CreatedAt          *DateInfo         `json:"created_at,omitempty"`          // Time when the user was created
	CreatedBy          *DateInfo         `json:"created_by,omitempty"`          // Who created the user
	Department         Record            `json:"department,omitempty"`          // Department of the user
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
	Location           *Record           `json:"location,omitempty"`            // Location of the user
	Manager            *Record           `json:"manager,omitempty"`             // Manager of the user
	Name               string            `json:"name,omitempty"`                // Full name of the user
	Notes              string            `json:"notes,omitempty"`               // Notes associated with the user
	Permissions        map[string]string `json:"permissions,omitempty"`         // Permissions of the user
	Phone              string            `json:"phone,omitempty"`               // Phone number of the user
	Remote             bool              `json:"remote,omitempty"`              // Specifies if the user is remote
	StartDate          *DateInfo         `json:"start_date,omitempty"`          // Start date of the user
	State              string            `json:"state,omitempty"`               // State of the user
	TwoFactorEnrolled  bool              `json:"two_factor_enrolled,omitempty"` // Specifies if the user has enrolled for two factor authentication
	TwoFactorOptin     bool              `json:"two_factor_optin,omitempty"`    // Specifies if the user has opted for two factor authentication
	UpdatedAt          *DateInfo         `json:"updated_at,omitempty"`          // Time when the user was last updated
	Username           string            `json:"username,omitempty"`            // Username of the user
	Vip                bool              `json:"vip,omitempty"`                 // Specifies if the user is a VIP
	Website            string            `json:"website,omitempty"`             // Website of the user
	Zip                string            `json:"zip,omitempty"`                 // Zip code of the user
}

// END OF USER STRUCTS
//-------------------------------------------------------------------------

// ### Common Asset types
// -------------------------------------------------------------------------
// Record represents an id:name pairing for many types of records in Snipe-IT.
type Record struct {
	ID   int64  `json:"id"`   // ID of the record {category, company, department, location, manufacturer, supplier, etc.}
	Name string `json:"name"` // Name of the record {category, company, department, location, manufacturer, supplier, etc.}
}

// DateInfo represents a date and its formatted representation.
type DateInfo struct {
	Date      string `json:"datetime,omitempty"`  // The date in yyyy-mm-dd format.
	Formatted string `json:"formatted,omitempty"` // The formatted date.
}

// StatusLabel represents the status label of a hardware item.
type StatusLabel struct {
	ID         int    `json:"id,omitempty"`          // ID of the status label.
	Name       string `json:"name,omitempty"`        // Name of thestatus label.
	StatusMeta string `json:"status_meta,omitempty"` // Meta status of the status label.
	StatusType string `json:"status_type,omitempty"` // Type of the status label.
}

// CustomFields represents the custom fields of a hardware item.
type CustomFields struct {
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
