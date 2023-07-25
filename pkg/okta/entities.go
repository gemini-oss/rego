/*
# Okta - Entities [Structs]

This package contains many structs for handling responses from the Okta API:

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/entities.go
package okta

import (
	"time"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ### Okta Client Structs
// ---------------------------------------------------------------------
type Client struct {
	BaseURL    string           // BaseURL is the base URL for Okta API requests.
	HTTPClient *requests.Client // HTTPClient is the client used to make HTTP requests.
	Error      *Error           // Error is the error response from the last request made by the client.
	Logger     *log.Logger      // Logger is the logger used to log messages.
}

type Error struct {
	ErrorCauses  []ErrorCause `json:"errorCauses,omitempty"`
	ErrorCode    string       `json:"errorCode,omitempty"`
	ErrorId      string       `json:"errorId,omitempty"`
	ErrorLink    string       `json:"errorLink,omitempty"`
	ErrorSummary string       `json:"errorSummary,omitempty"`
}

type ErrorCause struct {
	ErrorSummary string `json:"errorSummary,omitempty"`
}

type Links struct {
	AccessPolicy           Link   `json:"accessPolicy,omitempty"`           // AccessPolicy is a link to the access policy.
	Activate               Link   `json:"activate,omitempty"`               // Activate is a link to activate the user.
	ChangePassword         Link   `json:"changePassword,omitempty"`         // ChangePassword is a link to change the user's password.
	ChangeRecoveryQuestion Link   `json:"changeRecoveryQuestion,omitempty"` // ChangeRecoveryQuestion is a link to change the user's recovery question.
	Deactivate             Link   `json:"deactivate,omitempty"`             // Deactivate is a link to deactivate the user.
	ExpirePassword         Link   `json:"expirePassword,omitempty"`         // ExpirePassword is a link to expire the user's password.
	ForgotPassword         Link   `json:"forgotPassword,omitempty"`         // ForgotPassword is a link to reset the user's password.
	Groups                 Link   `json:"groups,omitempty"`                 // Groups is a link to the user's groups.
	Logo                   []Link `json:"logo,omitempty"`                   // Logo is a list of links to the logo.
	Metadata               Link   `json:"metadata,omitempty"`               // Metadata is a link to the user's metadata.
	ResetFactors           Link   `json:"resetFactors,omitempty"`           // ResetFactors is a link to reset the user's factors.
	ResetPassword          Link   `json:"resetPassword,omitempty"`          // ResetPassword is a link to reset the user's password.
	Schema                 Link   `json:"schema,omitempty"`                 // Schema is a link to the user's schema.
	Self                   Link   `json:"self,omitempty"`                   // Self is a link to the user.
	Suspend                Link   `json:"suspend,omitempty"`                // Suspend is a link to suspend the user.
	Users                  Link   `json:"users,omitempty"`                  // Users is a link to the user's users.
}

type Link struct {
	Hints  Hints  `json:"hints,omitempty"`  // Hints is a list of hints for the link.
	Href   string `json:"href,omitempty"`   // Href is the URL for the link.
	Method string `json:"method,omitempty"` // Method is the HTTP method for the link.
	Type   string `json:"type,omitempty"`   // Type is the type of link.
}

type Hints struct {
	Allow []string `json:"allow,omitempty"` // Allow is a list of allowed methods.
}

// END OF OKTA CLIENT STRUCTS
//---------------------------------------------------------------------

// ### Okta Application Structs
// ---------------------------------------------------------------------
type Applications []Application

type Application struct {
	Accessibility Accessibility       `json:"accessibility,omitempty"`
	Created       time.Time           `json:"created,omitempty"`
	Features      []string            `json:"features,omitempty"`
	ID            string              `json:"id,omitempty"`
	Label         string              `json:"label,omitempty"`
	LastUpdated   time.Time           `json:"lastUpdated,omitempty"`
	Licensing     Licensing           `json:"licensing,omitempty"`
	Profile       ApplicationProfile  `json:"profile,omitempty"`
	SignOnMode    string              `json:"signOnMode,omitempty"`
	Status        string              `json:"status,omitempty"`
	Visibility    Visibility          `json:"visibility,omitempty"`
	Embedded      ApplicationEmbedded `json:"_embedded,omitempty"`
	Links         Links               `json:"_links,omitempty"`
}

type Accessibility struct {
	ErrorRedirectURL  string `json:"errorRedirectUrl,omitempty"`
	LoginRedirectURL  string `json:"loginRedirectUrl,omitempty"`
	SelfService       bool   `json:"selfService,omitempty"`
	LoginRedirectURL2 string `json:"loginRedirectUrl2,omitempty"`
}

type Licensing struct {
	SeatCount int `json:"seatCount,omitempty"`
}

type ApplicationProfile struct {
	Property1 map[string]interface{} `json:"property1,omitempty"`
	Property2 map[string]interface{} `json:"property2,omitempty"`
}

type Visibility struct {
	AppLinks          map[string]bool `json:"appLinks,omitempty"`
	AutoLaunch        bool            `json:"autoLaunch,omitempty"`
	AutoSubmitToolbar bool            `json:"autoSubmitToolbar,omitempty"`
	Hide              map[string]bool `json:"hide,omitempty"`
}

type ApplicationEmbedded struct {
	Users *Users `json:"users,omitempty"`
}

// END OF OKTA APPLICATION STRUCTS
//---------------------------------------------------------------------

// ### Okta Device Structs
// ---------------------------------------------------------------------
type Devices []Device

type Device struct {
	Created             string          `json:"created,omitempty"`             // The timestamp when the device was created.
	ID                  string          `json:"id,omitempty"`                  // The unique key for the device.
	LastUpdated         string          `json:"lastUpdated,omitempty"`         // The timestamp when the device was last updated.
	Links               *Link           `json:"_links,omitempty"`              // A set of key/value pairs that provide additional information about the device.
	Profile             *DeviceProfile  `json:"profile,omitempty"`             // The device profile.
	ResourceAlternate   interface{}     `json:"resourceAlternateId,omitempty"` // The alternate ID of the device.
	ResourceDisplayName *DisplayName    `json:"resourceDisplayName,omitempty"` // The display name of the device.
	ResourceID          string          `json:"resourceId,omitempty"`          // The ID of the device.
	ResourceType        string          `json:"resourceType,omitempty"`        // The type of the device.
	Status              string          `json:"status,omitempty"`              // The status of the device.
	Embedded            *DeviceEmbedded `json:"_embedded,omitempty"`           // The users assigned to the device.
}

type DeviceProfile struct {
	DisplayName           string `json:"displayName,omitempty"`           // The display name of the device.
	Manufacturer          string `json:"manufacturer,omitempty"`          // The manufacturer of the device.
	Model                 string `json:"model,omitempty"`                 // The model of the device.
	OSVersion             string `json:"osVersion,omitempty"`             // The OS version of the device.
	Platform              string `json:"platform,omitempty"`              // The platform of the device.
	Registered            bool   `json:"registered,omitempty"`            // Indicates whether the device is registered with Okta.
	SecureHardwarePresent bool   `json:"secureHardwarePresent,omitempty"` // Indicates whether the device has secure hardware.
	SerialNumber          string `json:"serialNumber,omitempty"`          // The serial number of the device.
	SID                   string `json:"sid,omitempty"`                   // The SID of the device.
	UDID                  string `json:"udid,omitempty"`                  // The UDID of the device.
}

type DisplayName struct {
	Value     string `json:"value"`     // The display name of the device.
	Sensitive bool   `json:"sensitive"` // Indicates whether the display name is sensitive.
}

type DeviceUsers []DeviceUser

type DeviceUser struct {
	Created          time.Time `json:"created,omitempty"`          // The timestamp when the device user was created.
	ManagementStatus string    `json:"managementStatus,omitempty"` // The management status of the device user.
	User             *User     `json:"user,omitempty"`             // The user assigned to the device.
}

type DeviceEmbedded struct {
	DeviceUsers *DeviceUsers `json:"users,omitempty"`
}

// END OF OKTA DEVICE STRUCTS
//---------------------------------------------------------------------

// ### Okta Roles Structs
// ---------------------------------------------------------------------
type Roles struct {
	Roles []Role `json:"roles,omitempty"`
}

type Role struct {
	AssignmentType string    `json:"assignmentType,omitempty"`
	Created        time.Time `json:"created,omitempty"`
	Description    string    `json:"description,omitempty"`
	ID             string    `json:"id,omitempty"`
	Label          string    `json:"label,omitempty"`
	LastUpdated    time.Time `json:"lastUpdated,omitempty"`
	Links          *Links    `json:"_links,omitempty"`
	Status         string    `json:"status,omitempty"`
	Type           string    `json:"type,omitempty"`
}

type Permission struct {
	Created     time.Time `json:"created,omitempty"`
	Label       string    `json:"label,omitempty"`
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
	Links       *Links    `json:"_links,omitempty"`
}

type RoleReport struct {
	Role  *Role
	Users []*User
}

// END OF OKTA ROLES STRUCTS
//---------------------------------------------------------------------

// ### Google Users Structs
// ---------------------------------------------------------------------
type Users []*User

type User struct {
	Activated             time.Time        `json:"activated,omitempty"`
	Created               time.Time        `json:"created,omitempty"`
	Credentials           *UserCredentials `json:"credentials,omitempty"`
	ID                    string           `json:"id,omitempty"`
	LastLogin             time.Time        `json:"lastLogin,omitempty"`
	LastUpdated           time.Time        `json:"lastUpdated,omitempty"`
	PasswordChanged       time.Time        `json:"passwordChanged,omitempty"`
	Profile               *UserProfile     `json:"profile,omitempty"`
	Status                string           `json:"status,omitempty"`
	StatusChanged         time.Time        `json:"statusChanged,omitempty"`
	TransitioningToStatus string           `json:"transitioningToStatus,omitempty"`
	Type                  *UserType        `json:"type,omitempty"`
	Embedded              *UserEmbedded    `json:"_embedded,omitempty"`
	Links                 *Links           `json:"_links,omitempty"`
}

type UserCredentials struct {
	Password         *PasswordCredentials `json:"password,omitempty"`
	Provider         *Provider            `json:"provider,omitempty"`
	RecoveryQuestion *RecoveryQuestion    `json:"recovery_question,omitempty"`
}

type PasswordCredentials struct {
	Hook  *PasswordHook `json:"hook,omitempty"`
	Value string        `json:"value,omitempty"`
	Hash  *PasswordHash `json:"hash,omitempty"`
}

type PasswordHash struct {
	Algorithm       string `json:"algorithm,omitempty"`
	DigestAlgorithm string `json:"digestAlgorithm,omitempty"`
	IterationCount  int    `json:"iterationCount,omitempty"`
	KeySize         int    `json:"keySize,omitempty"`
	Salt            string `json:"salt,omitempty"`
	SaltOrder       string `json:"saltOrder,omitempty"`
	Value           string `json:"value,omitempty"`
	WorkFactor      int    `json:"workFactor,omitempty"`
}

type PasswordHook struct {
	Type string `json:"type,omitempty"`
}

type Provider struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type RecoveryQuestion struct {
	Answer   string `json:"answer,omitempty"`
	Question string `json:"question,omitempty"`
}

type UserProfile struct {
	City              string      `json:"city,omitempty"`
	CostCenter        string      `json:"costCenter,omitempty"`
	CountryCode       string      `json:"countryCode,omitempty"`
	Department        string      `json:"department,omitempty"`
	DisplayName       string      `json:"displayName,omitempty"`
	Division          string      `json:"division,omitempty"`
	Email             string      `json:"email,omitempty"`
	EmployeeNumber    string      `json:"employeeNumber,omitempty"`
	FirstName         string      `json:"firstName,omitempty"`
	HonorificPrefix   string      `json:"honorificPrefix,omitempty"`
	HonorificSuffix   string      `json:"honorificSuffix,omitempty"`
	LastName          string      `json:"lastName,omitempty"`
	Locale            string      `json:"locale,omitempty"`
	Login             string      `json:"login,omitempty"`
	Manager           string      `json:"manager,omitempty"`
	ManagerId         string      `json:"managerId,omitempty"`
	MiddleName        string      `json:"middleName,omitempty"`
	MobilePhone       string      `json:"mobilePhone,omitempty"`
	NickName          string      `json:"nickName,omitempty"`
	Organization      string      `json:"organization,omitempty"`
	PostalAddress     string      `json:"postalAddress,omitempty"`
	PreferredLanguage string      `json:"preferredLanguage,omitempty"`
	PrimaryPhone      string      `json:"primaryPhone,omitempty"`
	ProfileUrl        string      `json:"profileUrl,omitempty"`
	Property1         interface{} `json:"property1,omitempty"`
	Property2         interface{} `json:"property2,omitempty"`
	SecondEmail       string      `json:"secondEmail,omitempty"`
	State             string      `json:"state,omitempty"`
	StreetAddress     string      `json:"streetAddress,omitempty"`
	Timezone          string      `json:"timezone,omitempty"`
	Title             string      `json:"title,omitempty"`
	UserType          string      `json:"userType,omitempty"`
	ZipCode           string      `json:"zipCode,omitempty"`
}

type UserType struct {
	Created       time.Time `json:"created,omitempty"`
	CreatedBy     string    `json:"createdBy,omitempty"`
	Default       bool      `json:"default,omitempty"`
	Description   string    `json:"description,omitempty"`
	DisplayName   string    `json:"displayName,omitempty"`
	ID            string    `json:"id,omitempty"`
	LastUpdated   time.Time `json:"lastUpdated,omitempty"`
	LastUpdatedBy string    `json:"lastUpdatedBy,omitempty"`
	Name          string    `json:"name,omitempty"`
	Links         *Links    `json:"_links,omitempty"`
}

type UserEmbedded interface{}

// END OF OKTA USERS STRUCTS
//---------------------------------------------------------------------
