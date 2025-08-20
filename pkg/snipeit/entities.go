/*
# SnipeIT - Entities (Structs)

This package initializes all the structs for the SnipeIT API:
https://snipe-it.readme.io/reference/api-overview

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/entities.go
package snipeit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
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

func (pl PaginatedList[E]) Map() map[any]*E {
	result := make(map[any]*E)
	for _, item := range *pl.Rows {
		switch entity := any(item).(type) {
		case *Hardware[HardwareGET]:
			//log.Printf("Processing Hardware entity - Type: %T, Serial: %s", entity, entity.Serial)
			result[entity.Serial] = item
		case *User[UserGET]:
			//log.Printf("Processing User entity - Type: %T, Username: %s, Email: %v", entity, entity.Username, entity.Email)
			switch entity.Email {
			case nil:
				result[entity.Username] = item
			default:
				result[(*(*(*entity).UserBase).Email)] = item
			}
		case *License[LicenseGET]:
			//log.Printf("Processing License entity - Type: %T, Name: %s", entity, (*(*(*(*entity).LicenseBase).SnipeIT).Record).Name)
			result[(*(*(*(*entity).LicenseBase).SnipeIT).Record).Name] = item
		default:
			fmt.Printf("Processing unknown entity - Type: %T, Value: %+v", entity, entity)
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
	Company      *Record `json:"company,omitempty"`      // {HARDWARE,USER,LICENSE} Company
	Department   *Record `json:"department,omitempty"`   // {USER} Department
	Depreciation *Record `json:"depreciation,omitempty"` // {MODEL} Depreciation
	Location     *Record `json:"location,omitempty"`     // {HARDWARE,MODEL,USER} Location of the entity
	Manager      *Record `json:"manager,omitempty"`      // {LOCATION,USER} Manager
	Manufacturer *Record `json:"manufacturer,omitempty"` // {ACCESSORY,HARDWARE} Manufacturer
	Parent       *Record `json:"parent,omitempty"`       // {Location} Parent of the location
	RTDLocation  *Record `json:"rtd_location,omitempty"` // {HARDWARE} RTD [Ready to Deploy] location
	Supplier     *Record `json:"supplier,omitempty"`     // {HARDWARE,LICENSE} Supplier of the item.
}

// SnipeIT {POST, PUT, PATCH, DELETE} fields
type PPPD struct {
	CreatedAt      *string `json:"created_at,omitempty"`      // Time when the item was created.
	UpdatedAt      *string `json:"updated_at,omitempty"`      // Time when the item was last updated.
	DeletedAt      *string `json:"deleted_at,omitempty"`      // Time when the item was deleted.
	AssetID        *uint32 `json:"asset_id,omitempty"`        // {LICENSE_SEAT} Asset ID
	CategoryID     *uint32 `json:"category_id,omitempty"`     // {MODEL,LICENSES} Category ID
	CompanyID      *uint32 `json:"company_id,omitempty"`      // {HARDWARE,LICENSE} Company ID
	DepartmentID   *uint32 `json:"department_id,omitempty"`   // {LICENSE,USER} Department ID
	DepreciationID *uint32 `json:"depreciation_id,omitempty"` // {MODEL} Depreciation ID
	FieldsetID     *uint32 `json:"fieldset_id,omitempty"`     // {MODEL} Fieldset ID
	LocationID     *uint32 `json:"location_id,omitempty"`     // Location ID
	ManagerID      *uint32 `json:"manager_id,omitempty"`      // {USER} Manager ID
	ManufacturerID *uint32 `json:"manufacturer_id,omitempty"` // {LICENSE,MODEL} Manufacturer ID
	ModelID        *uint32 `json:"model_id,omitempty"`        // Model ID
	ParentID       *uint32 `json:"parent_id,omitempty"`       // {Location} Parent ID
	RTDLocationID  *uint32 `json:"rtd_location_id,omitempty"` // RTD Location ID
	StatusID       *uint32 `json:"status_id,omitempty"`       // {HARDWARE} Status ID
	SupplierID     *uint32 `json:"supplier_id,omitempty"`     // {HARDWARE,LICENSE} Supplier ID
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
	Archived     string         `json:"archived,omitempty"`      // Whether the hardware item is archived (string on GET, bool on PPPD)
	AssignedTo   *User[UserGET] `json:"assigned_to,omitempty"`   // User to whom the hardware item is assigned. (object on GET, string on POST)
	PurchaseCost string         `json:"purchase_cost,omitempty"` // Purchase cost of the hardware item. (string on GET, float on POST)
	PurchaseDate *DateInfo      `json:"purchase_date,omitempty"` // Purchase date of the hardware item. (object on GET, string on PPPD)
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
	WarrantyExpires *Timestamp                    `json:"warranty_expires,omitempty"`  // Warranty expiry date of the hardware item.
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
type CategoryList = PaginatedList[Category[CategoryGET]]

// Category represents an individual category.
// https://snipe-it.readme.io/reference/categories#sortable-columns
type CategoryBase struct {
	*SnipeIT          `json:",inline"`
	Type              string `json:"category_type,omitempty"`
	UseDefaultEULA    bool   `json:"use_default_eula,omitempty,omitzero"`
	RequireAcceptance bool   `json:"require_acceptance,omitempty,omitzero"`
	CheckinEmail      bool   `json:"checkin_email,omitempty,omitzero"`
}

type CategoryGET struct {
	GET              `json:",inline"`
	Image            *string `json:"image,omitempty"`
	HasEULA          bool    `json:"has_eula,omitempty,omitzero"`
	EULA             string  `json:"use_default_eula,omitempty"`
	AssetsCount      uint32  `json:"assets_count,omitempty"`
	AccessoriesCount uint32  `json:"accessories_count,omitempty"`
	ConsumablesCount uint32  `json:"consumables_count,omitempty"`
	ComponentsCount  uint32  `json:"components_count,omitempty"`
	LicensesCount    uint32  `json:"licenses_count,omitempty"`
	Notes            string  `json:"notes,omitempty"`
}

type CategoryPOST struct {
	PPPD `json:",inline"`
}

type Category[M any] struct {
	*CategoryBase `json:",inline"`
	Method        M `json:",inline"`
}

func (c Category[M]) MarshalJSON() ([]byte, error) {
	return generics.MarshalGeneric[Category[M], M](&c)
}

func (c *Category[M]) UnmarshalJSON(data []byte) error {
	cat, err := generics.UnmarshalGeneric[Category[M], M](data)
	if err != nil {
		return err
	}

	*c = *cat
	return nil
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

// ### Users
// -------------------------------------------------------------------------
type UserList = PaginatedList[User[UserGET]]

type UserBase struct {
	*SnipeIT              `json:",inline"`
	Avatar                string           `json:"avatar,omitempty"`                  // URL of the user's avatar
	Name                  string           `json:"name,omitempty"`                    // Full name of the user
	FirstName             string           `json:"first_name"`                        // First name of the user
	LastName              string           `json:"last_name,omitempty"`               // Last name of the user
	Username              string           `json:"username"`                          // Username of the user
	Remote                bool             `json:"remote,omitempty"`                  // Specifies if the user is remote
	Locale                *string          `json:"locale,omitempty"`                  // Locale of the user
	EmployeeNum           *string          `json:"employee_num,omitempty"`            // Employee number of the user
	JobTitle              *string          `json:"jobtitle,omitempty"`                // Job title of the user
	VIP                   BoolInt          `json:"vip,omitempty"`                     // Specifies if the user is a VIP
	Phone                 *string          `json:"phone,omitempty"`                   // Phone number of the user
	Website               *string          `json:"website,omitempty"`                 // Website of the user
	Address               *string          `json:"address,omitempty"`                 // Address of the user
	City                  *string          `json:"city,omitempty"`                    // City of the user
	State                 *string          `json:"state,omitempty"`                   // State of the user
	Country               *string          `json:"country,omitempty"`                 // Country of the user
	Zip                   *string          `json:"zip,omitempty"`                     // Zip code of the user
	Email                 *string          `json:"email,omitempty"`                   // Email of the user
	Notes                 *string          `json:"notes,omitempty"`                   // Notes associated with the user
	Permissions           map[string]uint8 `json:"permissions,omitempty"`             // Permissions of the user
	Activated             bool             `json:"activated"`                         // Specifies if the user is active or not
	AutoAssignLicenses    bool             `json:"autoassign_licenses,omitempty"`     // Specifies if the licenses are automatically assigned to the user
	LDAPImport            bool             `json:"ldap_import,omitempty"`             // Specifies if the user is imported from LDAP
	TwoFactorEnrolled     bool             `json:"two_factor_enrolled,omitempty"`     // Specifies if the user has enrolled for two factor authentication
	TwoFactorOptIn        bool             `json:"two_factor_optin,omitempty"`        // Specifies if the user has opted for two factor authentication
	LastLogin             *DateInfo        `json:"last_login,omitempty"`              // Last login time of the user
	AssetsCount           uint32           `json:"assets_count,omitempty"`            // Number of assets associated with the user
	ConsumablesCount      uint32           `json:"consumables_count,omitempty"`       // Count of consumables associated with the user
	ManagesUsersCount     uint32           `json:"manages_users_count,omitempty"`     // Number of users managed by the user
	ManagesLocationsCount uint32           `json:"manages_locations_count,omitempty"` // Number of locations managed by the user
	CreatedBy             *Record          `json:"created_by,omitempty"`              // Who created the user
}

type User[M any] struct {
	*UserBase `json:",inline"`
	Method    M `json:",inline"`
}

func (u User[M]) MarshalJSON() ([]byte, error) {
	return generics.MarshalGeneric[User[M], M](&u)
}

func (u *User[M]) UnmarshalJSON(data []byte) error {
	user, err := generics.UnmarshalGeneric[User[M], M](data)
	if err != nil {
		return err
	}

	*u = *user
	return nil
}

type UserGET struct {
	GET       `json:",inline"`
	StartDate *DateInfo `json:"start_date,omitempty"` // Start date of the user
	EndDate   *DateInfo `json:"end_date,omitempty"`   // End date of the user
}

type UserPOST struct {
	PPPD                 `json:",inline"`
	Password             string     `json:"password"`              // Password of the user
	PasswordConfirmation string     `json:"password_confirmation"` // Password confirmation of the user
	Groups               []uint32   `json:"groups,omitempty"`      // Groups associated with the user
	StartDate            *Timestamp `json:"start_date,omitempty"`  // Start date of the user
	EndDate              *Timestamp `json:"end_date,omitempty"`    // End date of the user
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

func (m Model[M]) MarshalJSON() ([]byte, error) {
	return generics.MarshalGeneric[Model[M], M](&m)
}

func (m *Model[M]) UnmarshalJSON(data []byte) error {
	model, err := generics.UnmarshalGeneric[Model[M], M](data)
	if err != nil {
		return err
	}

	*m = *model
	return nil
}

// END OF MODELS STRUCTS
//-------------------------------------------------------------------------

// ### Licenses
// -------------------------------------------------------------------------
// Source: https://snipe-it.readme.io/reference/licenses
type LicenseList = PaginatedList[License[LicenseGET]]

type LicenseBase struct {
	*SnipeIT       `json:",inline"`
	LicenseName    string     `json:"license_name,omitempty"`    // Owner/contact name
	LicenseEmail   string     `json:"license_email,omitempty"`   // Owner/contact email
	Maintained     bool       `json:"maintained,omitempty"`      // Whether the license is maintained
	Notes          string     `json:"notes,omitempty"`           // Notes associated with the license
	OrderNumber    string     `json:"order_number,omitempty"`    // Order number of the license
	PurchaseOrder  string     `json:"purchase_order,omitempty"`  // Purchase order of the license
	PurchaseDate   *Timestamp `json:"purchase_date,omitempty"`   // Purchase date of the license
	ExpirationDate *Timestamp `json:"expiration_date,omitempty"` // Expiration date of the license
}

type LicenseGET struct {
	GET          `json:",inline"`
	ProductKey   string `json:"product_key,omitempty"`      // Product key of the license
	PurchaseCost string `json:"purchase_cost,omitempty"`    // Purchase cost of the license
	Seats        int    `json:"seats,omitempty"`            // Number of seats for the license (int on GET, string on POST)
	FreeSeats    int    `json:"free_seats_count,omitempty"` // Number of free seats for the license
}

type LicensePOST struct {
	PPPD         `json:",inline"`
	Seats        string   `json:"seats,omitempty"`         // Number of seats for the license (int on GET, string on POST)
	ProductKey   string   `json:"serial,omitempty"`        // Product key of the license (but POSTs as serial)
	PurchaseCost *float64 `json:"purchase_cost,omitempty"` // Purchase cost of the license. (string on GET, float on POST,PATCH)
	Reassignable bool     `json:"reassignable,omitempty"`  // Whether the license is reassignable
}

type License[M any] struct {
	*LicenseBase `json:",inline"`
	Method       M `json:",inline"`
}

func (l License[M]) MarshalJSON() ([]byte, error) {
	return generics.MarshalGeneric[License[M], M](&l)
}

func (l *License[M]) UnmarshalJSON(data []byte) error {
	lic, err := generics.UnmarshalGeneric[License[M], M](data)
	if err != nil {
		return err
	}

	*l = *lic
	return nil
}

type LicenseBuilder func() *License[LicensePOST]

// NewLicense starts a fresh *License[LicensePOST] and hands the builder back.
// Philosophy is that we build licenses for PPPD methods but not for GET, which is
// why LicensePOST is hard-coded instead of using a Generic type
func NewLicense(name string) LicenseBuilder {
	lic := &License[LicensePOST]{
		LicenseBase: &LicenseBase{
			SnipeIT: &SnipeIT{
				Record: &Record{Name: name},
			},
		},
		Method: *new(LicensePOST),
	}
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) Build() *License[LicensePOST] { return b() }

func (b LicenseBuilder) LicenseEmail(v string) LicenseBuilder {
	lic := b()
	lic.LicenseEmail = v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) Maintained(v bool) LicenseBuilder {
	lic := b()
	lic.Maintained = v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) LicenseName(v string) LicenseBuilder {
	lic := b()
	lic.LicenseName = v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) Notes(v string) LicenseBuilder {
	lic := b()
	lic.Notes = v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) OrderNumber(v string) LicenseBuilder {
	lic := b()
	lic.OrderNumber = v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) PurchaseOrder(v string) LicenseBuilder {
	lic := b()
	lic.PurchaseOrder = v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) PurchaseDate(v string) LicenseBuilder {
	lic := b()
	lic.PurchaseDate.ParseDate(v)
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) ExpirationDate(v string) LicenseBuilder {
	lic := b()
	lic.ExpirationDate.ParseDate(v)
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) CategoryID(v uint32) LicenseBuilder {
	lic := b()
	lic.Method.PPPD.CategoryID = &v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) Seats(v string) LicenseBuilder {
	lic := b()
	lic.Method.Seats = v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) ProductKey(v string) LicenseBuilder {
	lic := b()
	lic.Method.ProductKey = v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) PurchaseCost(v float64) LicenseBuilder {
	lic := b()
	lic.Method.PurchaseCost = &v
	return func() *License[LicensePOST] { return lic }
}

func (b LicenseBuilder) Reassignable(v bool) LicenseBuilder {
	lic := b()
	lic.Method.Reassignable = v
	return func() *License[LicensePOST] { return lic }
}

// END OF LICENSE STRUCTS
// -------------------------------------------------------------------------

// ### License Seats
// -------------------------------------------------------------------------
// Source: https://snipe-it.readme.io/reference/licenses
type SeatList = PaginatedList[Seat[SeatGET]]

type SeatBase struct {
	*SnipeIT `json:",inline"`
}

type SeatGET struct {
	GET          `json:",inline"`
	LicenseID    int `json:"license_id,omitempty"`
	AssignedUser struct {
		*Record
		Department *Record `json:"department,omitempty"`
	} `json:"assigned_user,omitempty"`
	AssignedAsset   string `json:"assigned_asset,omitempty"`
	Reassignable    bool   `json:"reassignable,omitempty"`
	UserCanCheckout bool   `json:"user_can_checkout,omitempty"`
}

type SeatPOST struct {
	PPPD         `json:",inline"`
	AssignedTo   *uint32 `json:"assigned_to"`            // The User ID to assign the license to
	Notes        string  `json:"notes,omitempty"`        // Notes about the seat (Exists in the response, but not used for body)
	Reassignable bool    `json:"reassignable,omitempty"` // Whether the license is reassignable
}

/*
MarshalJSON → send `Notes` out as "note"
*/
func (s SeatPOST) MarshalJSON() ([]byte, error) {
	type alias SeatPOST
	return json.Marshal(&struct {
		alias
		Note string `json:"note,omitempty"`
	}{
		alias: alias(s),
		Note:  s.Notes,
	})
}

/*
UnmarshalJSON → accept either "notes" or "note" from server
*/
func (s *SeatPOST) UnmarshalJSON(data []byte) error {
	type alias SeatPOST
	aux := &struct {
		alias
		Note     *string         `json:"note"`
		Notes    *string         `json:"notes"`
		Assigned json.RawMessage `json:"assigned_to"`
	}{}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	*s = SeatPOST(aux.alias)

	// Handle "notes" field → accept "notes" or "note"
	// Standardize into "notes" for Rego
	switch {
	case aux.Notes != nil:
		s.Notes = *aux.Notes
	case aux.Note != nil:
		s.Notes = *aux.Note
	}

	// Handle `assigned_to`
	if len(aux.Assigned) > 0 && string(aux.Assigned) != "null" {
		// 1) try as JSON number
		var id uint32
		if err := json.Unmarshal(aux.Assigned, &id); err == nil {
			s.AssignedTo = &id
			return nil
		}

		// 2) try as quoted string
		var str string
		if err := json.Unmarshal(aux.Assigned, &str); err == nil {
			v, err := strconv.ParseUint(str, 10, 32)
			if err != nil {
				return fmt.Errorf("assigned_to: %q is not a valid uint32", str)
			}
			u := uint32(v)
			s.AssignedTo = &u
			return nil
		}

		return fmt.Errorf("assigned_to has unsupported JSON type: %s", aux.Assigned)
	}

	return nil
}

type Seat[M any] struct {
	*SeatBase `json:",inline"`
	Method    M `json:",inline"`
}

func (s Seat[M]) MarshalJSON() ([]byte, error) {
	return generics.MarshalGeneric[Seat[M], M](&s)
}

func (s *Seat[M]) UnmarshalJSON(data []byte) error {
	seat, err := generics.UnmarshalGeneric[Seat[M], M](data)
	if err != nil {
		return err
	}

	*s = *seat
	return nil
}

// END OF LICENSE SEAT STRUCTS
// -------------------------------------------------------------------------

// ### Common Asset types
// -------------------------------------------------------------------------
// Record represents an id:name pairing for many types of records in Snipe-IT.
type Record struct {
	ID   uint32 `json:"id,omitzero"` // ID of the record {category, company, department, location, manufacturer, supplier, etc.}
	Name string `json:"name"`        // Name of the record {category, company, department, location, manufacturer, supplier, etc.}
}

const (
	snipeDate = "2006-01-02"
	snipeTime = "2006-01-02 15:04:05"
	iso8601   = "2006-01-02T15:04:05.000000Z"
)

// Timestamp is a time.Time but JSON marshals/unmarshals as a string in the format "2006-01-02 15:04:05"
type Timestamp struct {
	time.Time `json:",omitzero"`
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(ts.Time.Format(snipeTime))
}

func (ts *Timestamp) UnmarshalJSON(b []byte) error {
	d := &DateInfo{}
	// If successful and the Date field is not empty, use it
	if err := json.Unmarshal(b, d); err == nil && d.Date != "" {
		return ts.ParseDate(d.Date)
	}

	// If that fails, try to unmarshal directly into a string
	var dateStr string
	if err := json.Unmarshal(b, &dateStr); err != nil {
		return err
	}

	return ts.ParseDate(dateStr)
}

// parseDate attempts to parse a date string in ISO 8601 format first, then falls back to MySQL DATETIME format.
func (ts *Timestamp) ParseDate(dateStr string) error {
	// Try ISO 8601
	if t, err := time.Parse(iso8601, dateStr); err == nil {
		ts.Time = t
		return nil
	}
	// Try snipeTime
	if t, err := time.Parse(snipeTime, dateStr); err == nil {
		ts.Time = t
		return nil
	}
	// Try date-only fallback
	if t, err := time.Parse(snipeDate, dateStr); err == nil {
		ts.Time = t
		return nil
	}
	return fmt.Errorf("unable to parse date string: %s", dateStr)
}

// DateInfo represents a date and its formatted representation.
type DateInfo struct {
	Date      string `json:"datetime,omitempty"`  // The date in yyyy-mm-dd format.
	Formatted string `json:"formatted,omitempty"` // The formatted date.
}

func (d *DateInfo) UnmarshalJSON(b []byte) error {
	var aux struct {
		DateTime  string `json:"datetime"`
		Date      string `json:"date"`
		Formatted string `json:"formatted"`
	}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	// Prefer "datetime" if present, else fall back to "date"
	if aux.DateTime != "" {
		d.Date = aux.DateTime
	} else {
		d.Date = aux.Date
	}

	d.Formatted = aux.Formatted
	return nil
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

// BoolInt encodes a bool as 0/1 and decodes either form (0/1 or true/false).
type BoolInt bool

func (b BoolInt) MarshalJSON() ([]byte, error) {
	// outbound payload: 1 for true, 0 for false
	if b {
		return json.Marshal(1)
	}
	return json.Marshal(0)
}

func (b *BoolInt) UnmarshalJSON(data []byte) error {
	// accept: 1, 0, "1", "0", true, false
	switch string(bytes.TrimSpace(data)) {
	case "1", `"1"`, "true":
		*b = true
	case "0", `"0"`, "false":
		*b = false
	case "null": // allow nullable field
		*b = false
	default:
		return fmt.Errorf("BoolInt: expected 0/1/true/false, got %s", data)
	}
	return nil
}

// END OF COMMON ASSET TYPES
//-------------------------------------------------------------------------

// ### Enums
// --------------------------------------------------------------------

const (
	CATEGORY_TYPE_ASSET      string = "Asset"
	CATEGORY_TYPE_ACCESSORY  string = "Accessory"
	CATEGORY_TYPE_CONSUMABLE string = "Consumable"
	CATEGORY_TYPE_COMPONENT  string = "Component"
	CATEGORY_TYPE_LICENSE    string = "License"
)
