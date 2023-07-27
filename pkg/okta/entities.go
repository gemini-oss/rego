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
	Accessibility Accessibility       `json:"accessibility,omitempty"` // The accessibility of the application.
	Created       time.Time           `json:"created,omitempty"`       // The timestamp when the application was created.
	Features      []string            `json:"features,omitempty"`      // The features of the application.
	ID            string              `json:"id,omitempty"`            // The ID of the application.
	Label         string              `json:"label,omitempty"`         // The label of the application.
	LastUpdated   time.Time           `json:"lastUpdated,omitempty"`   // The timestamp when the application was last updated.
	Licensing     Licensing           `json:"licensing,omitempty"`     // The licensing of the application.
	Profile       ApplicationProfile  `json:"profile,omitempty"`       // The profile of the application.
	SignOnMode    string              `json:"signOnMode,omitempty"`    // The sign-on mode of the application.
	Status        string              `json:"status,omitempty"`        // The status of the application.
	Visibility    Visibility          `json:"visibility,omitempty"`    // The visibility of the application.
	Embedded      ApplicationEmbedded `json:"_embedded,omitempty"`     // The users assigned to the application.
	Links         Links               `json:"_links,omitempty"`        // Links related to the application.
}

type Accessibility struct {
	ErrorRedirectURL  string `json:"errorRedirectUrl,omitempty"`  // The error redirect URL of the application.
	LoginRedirectURL  string `json:"loginRedirectUrl,omitempty"`  // The login redirect URL of the application.
	SelfService       bool   `json:"selfService,omitempty"`       // Indicates whether the application is self-service.
	LoginRedirectURL2 string `json:"loginRedirectUrl2,omitempty"` // The second login redirect URL of the application.
}

type Licensing struct {
	SeatCount int `json:"seatCount,omitempty"` // The seat count of the application.
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

// AppLink represents an app link object.
type AppLink struct {
	AppAssignmentID  string `json:"appAssignmentId,omitempty"`  // The ID of the app assignment.
	AppInstanceID    string `json:"appInstanceId,omitempty"`    // The ID of the app instance.
	AppName          string `json:"appName,omitempty"`          // The name of the app.
	CredentialsSetup bool   `json:"credentialsSetup,omitempty"` // Indicates whether credentials are set up.
	Hidden           bool   `json:"hidden,omitempty"`           // Indicates whether the app link is hidden.
	ID               string `json:"id,omitempty"`               // The ID of the app link.
	Label            string `json:"label,omitempty"`            // The label of the app link.
	LinkURL          string `json:"linkUrl,omitempty"`          // The URL of the app link.
	LogoURL          string `json:"logoUrl,omitempty"`          // The URL of the logo for the app link.
	SortOrder        int    `json:"sortOrder,omitempty"`        // The sort order of the app link.
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
	AssignmentType string    `json:"assignmentType,omitempty"` // The assignment type of the role.
	Created        time.Time `json:"created,omitempty"`        // The timestamp when the role was created.
	Description    string    `json:"description,omitempty"`    // The description of the role.
	ID             string    `json:"id,omitempty"`             // The ID of the role.
	Label          string    `json:"label,omitempty"`          // The label of the role.
	LastUpdated    time.Time `json:"lastUpdated,omitempty"`    // The timestamp when the role was last updated.
	Links          *Links    `json:"_links,omitempty"`         // Links related to the role.
	Status         string    `json:"status,omitempty"`         // The status of the role.
	Type           string    `json:"type,omitempty"`           // The type of the role.
}

type Permission struct {
	Created     time.Time `json:"created,omitempty"`     // The timestamp when the permission was created.
	Label       string    `json:"label,omitempty"`       // The label of the permission.
	LastUpdated time.Time `json:"lastUpdated,omitempty"` // The timestamp when the permission was last updated.
	Links       *Links    `json:"_links,omitempty"`      // Links related to the permission.
}

type RoleReport struct {
	Role  *Role   // The role.
	Users []*User // The users assigned to the role.
}

// END OF OKTA ROLES STRUCTS
//---------------------------------------------------------------------

// ### Okta Users Structs
// ---------------------------------------------------------------------
type Users []*User

type User struct {
	Activated             time.Time        `json:"activated,omitempty"`             // The timestamp when the user was activated.
	Created               time.Time        `json:"created,omitempty"`               // The timestamp when the user was created.
	Credentials           *UserCredentials `json:"credentials,omitempty"`           // The user's credentials.
	ID                    string           `json:"id,omitempty"`                    // The ID of the user.
	LastLogin             time.Time        `json:"lastLogin,omitempty"`             // The timestamp when the user last logged in.
	LastUpdated           time.Time        `json:"lastUpdated,omitempty"`           // The timestamp when the user was last updated.
	PasswordChanged       time.Time        `json:"passwordChanged,omitempty"`       // The timestamp when the user's password was last changed.
	Profile               *UserProfile     `json:"profile,omitempty"`               // The user's profile.
	Status                string           `json:"status,omitempty"`                // The status of the user.
	StatusChanged         time.Time        `json:"statusChanged,omitempty"`         // The timestamp when the user's status was last changed.
	TransitioningToStatus string           `json:"transitioningToStatus,omitempty"` // The status that the user is transitioning to.
	Type                  *UserType        `json:"type,omitempty"`                  // The type of the user.
	Embedded              *UserEmbedded    `json:"_embedded,omitempty"`             // Embedded properties, to be revisited.
	Links                 *Links           `json:"_links,omitempty"`                // Links related to the user.
}

type UserCredentials struct {
	Password         *PasswordCredentials `json:"password,omitempty"`          // The user's password credentials.
	Provider         *Provider            `json:"provider,omitempty"`          // The user's provider credentials.
	RecoveryQuestion *RecoveryQuestion    `json:"recovery_question,omitempty"` // The user's recovery question credentials.
}

type PasswordCredentials struct {
	Hook  *PasswordHook `json:"hook,omitempty"`  // The password hook.
	Value string        `json:"value,omitempty"` // The password value.
	Hash  *PasswordHash `json:"hash,omitempty"`  // The password hash.
}

type PasswordHash struct {
	Algorithm       string `json:"algorithm,omitempty"`       // The algorithm used to hash the password.
	DigestAlgorithm string `json:"digestAlgorithm,omitempty"` // The digest algorithm used to hash the password.
	IterationCount  int    `json:"iterationCount,omitempty"`  // The iteration count used to hash the password.
	KeySize         int    `json:"keySize,omitempty"`         // The key size used to hash the password.
	Salt            string `json:"salt,omitempty"`            // The salt used to hash the password.
	SaltOrder       string `json:"saltOrder,omitempty"`       // The salt order used to hash the password.
	Value           string `json:"value,omitempty"`           // The password hash value.
	WorkFactor      int    `json:"workFactor,omitempty"`      // The work factor used to hash the password.
}

type PasswordHook struct {
	Type string `json:"type,omitempty"` // The type of the password hook.
}

type Provider struct {
	Name string `json:"name,omitempty"` // The name of the provider.
	Type string `json:"type,omitempty"` // The type of the provider. Enum: "ACTIVE_DIRECTORY" "FEDERATION" "IMPORT" "LDAP" "OKTA" "SOCIAL"
}

type RecoveryQuestion struct {
	Answer   string `json:"answer,omitempty"`   // The answer to the user's recovery question.
	Question string `json:"question,omitempty"` // The user's recovery question.
}

type UserProfile struct {
	Aliases           []string `json:"emailAliases,omitempty"`      // Custom Property: The email aliases of the user.
	City              string   `json:"city,omitempty"`              // The city of the user's address. Maximum length is 128 characters.
	CostCenter        string   `json:"costCenter,omitempty"`        // The cost center of the user.
	CountryCode       string   `json:"countryCode,omitempty"`       // The country code of the user's address. [ISO 3166-1 alpha-2 country code](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2) specification. Limit: <= 2 characters.
	Department        string   `json:"department,omitempty"`        // The department of the user.
	DisplayName       string   `json:"displayName,omitempty"`       // The display name of the user.
	Division          string   `json:"division,omitempty"`          // The division of the user.
	Email             string   `json:"email,omitempty"`             // The primary email address of the user, used as the login name and is always required for `create` requests. Must be unique. Limit: [5 - 100] characters.
	EmployeeNumber    string   `json:"employeeNumber,omitempty"`    // The employee number of the user.
	FirstName         string   `json:"firstName,omitempty"`         // The first name of the user. Limit: [1 .. 50] characters.
	HonorificPrefix   string   `json:"honorificPrefix,omitempty"`   // The honorific prefix of the user's name.
	HonorificSuffix   string   `json:"honorificSuffix,omitempty"`   // The honorific suffix of the user's name.
	LastName          string   `json:"lastName,omitempty"`          // The last name of the user. Limit: [1 .. 50] characters.
	Locale            string   `json:"locale,omitempty"`            // The locale of the user. Specified according to [IETF BCP 47 language tag](https://datatracker.ietf.org/doc/html/rfc5646). Example: `en-US`.
	Login             string   `json:"login,omitempty"`             // The login name of the user.
	Manager           string   `json:"manager,omitempty"`           // The manager of the user.
	ManagerId         string   `json:"managerId,omitempty"`         // The ID of the user's manager.
	MiddleName        string   `json:"middleName,omitempty"`        // The middle name of the user.
	MobilePhone       string   `json:"mobilePhone,omitempty"`       // The mobile phone number of the user. Maximum length is 100 characters.
	NickName          string   `json:"nickName,omitempty"`          // The nickname of the user.
	Organization      string   `json:"organization,omitempty"`      // The organization of the user.
	PostalAddress     string   `json:"postalAddress,omitempty"`     // The postal address of the user. Limit: <= 4096 characters.
	PreferredLanguage string   `json:"preferredLanguage,omitempty"` // The preferred language of the user.
	PrimaryPhone      string   `json:"primaryPhone,omitempty"`      // The primary phone number of the user.
	ProfileUrl        string   `json:"profileUrl,omitempty"`        // The profile URL of the user.
	SecondEmail       string   `json:"secondEmail,omitempty"`       // The secondary email address of the user. Limit: [5 - 100] characters.
	State             string   `json:"state,omitempty"`             // The state of the user's address. Limit: <= 128 characters.
	StreetAddress     string   `json:"streetAddress,omitempty"`     // The street address of the user. Limit: <= 1024 characters.
	Timezone          string   `json:"timezone,omitempty"`          // The time zone of the user.
	Title             string   `json:"title,omitempty"`             // The title of the user.
	UserType          string   `json:"userType,omitempty"`          // The type of the user.
	ZipCode           string   `json:"zipCode,omitempty"`           // The zip code of the user's address. Limit: <= 12 characters.
}

type UserType struct {
	Created       time.Time `json:"created,omitempty"`       // The timestamp when the user type was created.
	CreatedBy     string    `json:"createdBy,omitempty"`     // The ID of the user who created the user type.
	Default       bool      `json:"default,omitempty"`       // Indicates whether the user type is the default.
	Description   string    `json:"description,omitempty"`   // The description of the user type.
	DisplayName   string    `json:"displayName,omitempty"`   // The display name of the user type.
	ID            string    `json:"id,omitempty"`            // The ID of the user type.
	LastUpdated   time.Time `json:"lastUpdated,omitempty"`   // The timestamp when the user type was last updated.
	LastUpdatedBy string    `json:"lastUpdatedBy,omitempty"` // The ID of the user who last updated the user type.
	Name          string    `json:"name,omitempty"`          // The name of the user type.
	Links         *Links    `json:"_links,omitempty"`        // Links related to the user type.
}

type UserEmbedded interface{}

// END OF OKTA USERS STRUCTS
//---------------------------------------------------------------------

// ### Okta Group Structs
// ---------------------------------------------------------------------
type Groups []*Group

// Group represents a user group object.
type Group struct {
	Created               time.Time     `json:"created,omitempty"`               // The creation time of the user group.
	ID                    string        `json:"id,omitempty"`                    // The ID of the user group.
	LastMembershipUpdated time.Time     `json:"lastMembershipUpdated,omitempty"` // The last time the membership of the user group was updated.
	LastUpdated           time.Time     `json:"lastUpdated,omitempty"`           // The last time the user group was updated.
	ObjectClass           []string      `json:"objectClass,omitempty"`           // Array of object classes.
	Profile               GroupProfile  `json:"profile,omitempty"`               // The profile of the user group.
	Type                  string        `json:"type,omitempty"`                  // The type of the user group.
	Embedded              GroupEmbedded `json:"_embedded,omitempty"`             // Embedded properties, to be revisited.
	Links                 Links         `json:"_links,omitempty"`                // Links related to the user group.
}

type GroupProfile struct {
	Description string `json:"description,omitempty"` // The description of the user group.
	Name        string `json:"name,omitempty"`        // The name of the user group.
}

type GroupEmbedded interface{}

type GroupRules []*GroupRule

type GroupRule struct {
	Actions     GroupActions `json:"actions,omitempty"`     // Defines the actions to be taken when the rule is triggered.
	Conditions  Conditions   `json:"conditions,omitempty"`  // Defines the conditions that would trigger the rule.
	Created     string       `json:"created,omitempty"`     // Date and time when the rule was created.
	ID          string       `json:"id,omitempty"`          // ID of the rule.
	LastUpdated string       `json:"lastUpdated,omitempty"` // Date and time when the rule was last updated.
	Name        string       `json:"name,omitempty"`        // Name of the rule.
	Status      string       `json:"status,omitempty"`      // Status of the rule.
	Type        string       `json:"type,omitempty"`        // Type of the rule.
}

type GroupActions struct {
	AssignUserToGroups GroupRuleGroupAssignment `json:"assignUserToGroups,omitempty"` // Group assignments for the action.
}

type GroupRuleGroupAssignment struct {
	GroupIDs []string `json:"groupIds,omitempty"` // IDs of the groups involved in the assignment.
}

type Conditions struct {
	Expression GroupExpression `json:"expression,omitempty"` // Expression for the condition.
	People     PeopleCondition `json:"people,omitempty"`     // People involved in the condition.
}

type GroupExpression struct {
	Type  string `json:"type,omitempty"`  // Type of the expression.
	Value string `json:"value,omitempty"` // Value of the expression.
}

type PeopleCondition struct {
	Groups GroupCondition `json:"groups,omitempty"` // Groups involved in the people condition.
	Users  GroupCondition `json:"users,omitempty"`  // Users involved in the people condition.
}

type GroupCondition struct {
	Exclude []string `json:"exclude,omitempty"` // Excluded from the condition.
	Include []string `json:"include,omitempty"` // Included in the condition.
}

// END OF OKTA Group STRUCTS
//---------------------------------------------------------------------
