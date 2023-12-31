<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# okta

```go
import "github.com/gemini-oss/rego/pkg/okta"
```

pkg/okta/applications.go

pkg/okta/devices.go

pkg/okta/okta.go

pkg/okta/roles.go

pkg/okta/users/users.go

## Index

- [Constants](<#constants>)
- [Variables](<#variables>)
- [type Accessibility](<#Accessibility>)
- [type AppQuery](<#AppQuery>)
- [type Application](<#Application>)
- [type ApplicationProfile](<#ApplicationProfile>)
- [type Applications](<#Applications>)
- [type Client](<#Client>)
  - [func NewClient\(verbosity int\) \*Client](<#NewClient>)
  - [func \(c \*Client\) BuildURL\(endpoint string, identifiers ...string\) string](<#Client.BuildURL>)
  - [func \(c \*Client\) GenerateRoleReport\(\) \(\[\]\*RoleReport, error\)](<#Client.GenerateRoleReport>)
  - [func \(c \*Client\) GetRole\(roleID string\) \(\*Role, error\)](<#Client.GetRole>)
  - [func \(c \*Client\) GetUser\(userID string\) \(\*User, error\)](<#Client.GetUser>)
  - [func \(c \*Client\) GetUserRoles\(userID string\) \(\*Roles, error\)](<#Client.GetUserRoles>)
  - [func \(c \*Client\) ListActiveUsers\(\) \(\*Users, error\)](<#Client.ListActiveUsers>)
  - [func \(c \*Client\) ListAllApplications\(\) \(\*Applications, error\)](<#Client.ListAllApplications>)
  - [func \(c \*Client\) ListAllDevices\(\) \(\*Devices, error\)](<#Client.ListAllDevices>)
  - [func \(c \*Client\) ListAllRoles\(\) \(\*Roles, error\)](<#Client.ListAllRoles>)
  - [func \(c \*Client\) ListAllUsers\(\) \(\*Users, error\)](<#Client.ListAllUsers>)
  - [func \(c \*Client\) ListAllUsersWithRoleAssignments\(\) \(\*Users, error\)](<#Client.ListAllUsersWithRoleAssignments>)
  - [func \(c \*Client\) ListUsersForDevice\(deviceID string\) \(\*DeviceUsers, error\)](<#Client.ListUsersForDevice>)
- [type Device](<#Device>)
- [type DeviceProfile](<#DeviceProfile>)
- [type DeviceQuery](<#DeviceQuery>)
- [type DeviceUser](<#DeviceUser>)
- [type DeviceUsers](<#DeviceUsers>)
- [type Devices](<#Devices>)
- [type DisplayName](<#DisplayName>)
- [type Embedded](<#Embedded>)
- [type Error](<#Error>)
- [type ErrorCause](<#ErrorCause>)
- [type Hints](<#Hints>)
- [type Licensing](<#Licensing>)
- [type Link](<#Link>)
- [type Links](<#Links>)
- [type PasswordCredentials](<#PasswordCredentials>)
- [type PasswordHash](<#PasswordHash>)
- [type PasswordHook](<#PasswordHook>)
- [type Permission](<#Permission>)
- [type Provider](<#Provider>)
- [type RecoveryQuestion](<#RecoveryQuestion>)
- [type Role](<#Role>)
- [type RoleReport](<#RoleReport>)
- [type Roles](<#Roles>)
- [type User](<#User>)
- [type UserCredentials](<#UserCredentials>)
- [type UserProfile](<#UserProfile>)
- [type UserQuery](<#UserQuery>)
- [type UserType](<#UserType>)
- [type Users](<#Users>)
- [type Visibility](<#Visibility>)


## Constants

<a name="OktaApps"></a>

```go
const (
    OktaApps    = "%s/apps"      // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/
    OktaDevices = "%s/devices"   // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/
    OktaUsers   = "%s/users"     // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/
    OktaIAM     = "%s/iam"       // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/
    OktaRoles   = "%s/iam/roles" // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/
)
```

## Variables

<a name="BaseURL"></a>

```go
var (
    BaseURL = fmt.Sprintf("https://%s.%s.com/api/v1", "%s", "%s") // https://developer.okta.com/docs/api/#versioning
)
```

<a name="Accessibility"></a>
## type Accessibility



```go
type Accessibility struct {
    ErrorRedirectURL  string `json:"errorRedirectUrl,omitempty"`
    LoginRedirectURL  string `json:"loginRedirectUrl,omitempty"`
    SelfService       bool   `json:"selfService,omitempty"`
    LoginRedirectURL2 string `json:"loginRedirectUrl2,omitempty"`
}
```

<a name="AppQuery"></a>
## type AppQuery

\* Query parameters for Applications

```go
type AppQuery struct {
    Q                 string // Searches the records for matching value
    After             string // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the `Link` response header.
    Limit             string // Default: -1. Specifies the number of results for a page
    Filter            string // Filters apps by `status`, `user.id`, `group.id` or `credentials.signing.kid`` expression
    Search            string // A SCIM filter expression that filters the results. Searches include all Device profile properties and the Device `id``, `status``, and `lastUpdated`` properties.
    Expand            string // Traverses users link relationship and optionally embeds Application User resource
    IncludeNonDeleted bool   // Default: false.
}
```

<a name="Application"></a>
## type Application



```go
type Application struct {
    Accessibility Accessibility      `json:"accessibility,omitempty"`
    Created       time.Time          `json:"created,omitempty"`
    Features      []string           `json:"features,omitempty"`
    ID            string             `json:"id,omitempty"`
    Label         string             `json:"label,omitempty"`
    LastUpdated   time.Time          `json:"lastUpdated,omitempty"`
    Licensing     Licensing          `json:"licensing,omitempty"`
    Profile       ApplicationProfile `json:"profile,omitempty"`
    SignOnMode    string             `json:"signOnMode,omitempty"`
    Status        string             `json:"status,omitempty"`
    Visibility    Visibility         `json:"visibility,omitempty"`
    Embedded      Embedded           `json:"_embedded,omitempty"`
    Links         Links              `json:"_links,omitempty"`
}
```

<a name="ApplicationProfile"></a>
## type ApplicationProfile



```go
type ApplicationProfile struct {
    Property1 map[string]interface{} `json:"property1,omitempty"`
    Property2 map[string]interface{} `json:"property2,omitempty"`
}
```

<a name="Applications"></a>
## type Applications



```go
type Applications []Application
```

<a name="Client"></a>
## type Client



```go
type Client struct {
    BaseURL    string           // BaseURL is the base URL for Okta API requests.
    HTTPClient *requests.Client // HTTPClient is the client used to make HTTP requests.
    Error      *Error           // Error is the error response from the last request made by the client.
    Logger     *log.Logger      // Logger is the logger used to log messages.
}
```

<a name="NewClient"></a>
### func NewClient

```go
func NewClient(verbosity int) *Client
```

NewClient returns a new Okta API client.

<a name="Client.BuildURL"></a>
### func \(\*Client\) BuildURL

```go
func (c *Client) BuildURL(endpoint string, identifiers ...string) string
```

BuildURL builds a URL for a given resource and identifiers.

<a name="Client.GenerateRoleReport"></a>
### func \(\*Client\) GenerateRoleReport

```go
func (c *Client) GenerateRoleReport() ([]*RoleReport, error)
```

\* \# Generate a report of all Okta roles and their users

<a name="Client.GetRole"></a>
### func \(\*Client\) GetRole

```go
func (c *Client) GetRole(roleID string) (*Role, error)
```

\* \# Retrieves a role by \`roleIdOrLabel\`

- /api/v1/iam/roles/\{roleIdOrLabel\}
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/#tag/Role/operation/getRole

<a name="Client.GetUser"></a>
### func \(\*Client\) GetUser

```go
func (c *Client) GetUser(userID string) (*User, error)
```

\* Get a user by ID

- /api/v1/users/\{userId\}
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/getUser

<a name="Client.GetUserRoles"></a>
### func \(\*Client\) GetUserRoles

```go
func (c *Client) GetUserRoles(userID string) (*Roles, error)
```

\* Lists all roles assigned to a user identified by \`userId“

- /api/v1/users/\{userId\}/roles
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/#tag/RoleAssignment/operation/listAssignedRolesForUser

<a name="Client.ListActiveUsers"></a>
### func \(\*Client\) ListActiveUsers

```go
func (c *Client) ListActiveUsers() (*Users, error)
```

\* List all ACTIVE users

- /api/v1/users
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers

<a name="Client.ListAllApplications"></a>
### func \(\*Client\) ListAllApplications

```go
func (c *Client) ListAllApplications() (*Applications, error)
```

\* Lists all applications with pagination. A subset of apps can be returned that match a supported filter expression or query.

- /api/v1/apps
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/#tag/Application/operation/listApplications

<a name="Client.ListAllDevices"></a>
### func \(\*Client\) ListAllDevices

```go
func (c *Client) ListAllDevices() (*Devices, error)
```

\* Lists all devices with pagination support.

- /api/v1/devices
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices

<a name="Client.ListAllRoles"></a>
### func \(\*Client\) ListAllRoles

```go
func (c *Client) ListAllRoles() (*Roles, error)
```

\* \# Lists all roles with pagination support.

- \- By default, only custom roles can be listed from this endpoint
- /api/v1/iam/roles
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/#tag/Role/operation/listRoles

<a name="Client.ListAllUsers"></a>
### func \(\*Client\) ListAllUsers

```go
func (c *Client) ListAllUsers() (*Users, error)
```

\* Get all users, regardless of status

- /api/v1/users
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers

<a name="Client.ListAllUsersWithRoleAssignments"></a>
### func \(\*Client\) ListAllUsersWithRoleAssignments

```go
func (c *Client) ListAllUsersWithRoleAssignments() (*Users, error)
```

\* \# Get all Users with Role Assignments

- /api/v1/iam/assignees/users
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/#tag/RoleAssignment/operation/listUsersWithRoleAssignments

<a name="Client.ListUsersForDevice"></a>
### func \(\*Client\) ListUsersForDevice

```go
func (c *Client) ListUsersForDevice(deviceID string) (*DeviceUsers, error)
```

\* Lists all Users for a Device

- /api/v1/devices/\{deviceId\}/users
- \- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device/operation/listDevices

<a name="Device"></a>
## type Device



```go
type Device struct {
    Created             string         `json:"created,omitempty"`             // The timestamp when the device was created.
    ID                  string         `json:"id,omitempty"`                  // The unique key for the device.
    LastUpdated         string         `json:"lastUpdated,omitempty"`         // The timestamp when the device was last updated.
    Links               *Link          `json:"_links,omitempty"`              // A set of key/value pairs that provide additional information about the device.
    Profile             *DeviceProfile `json:"profile,omitempty"`             // The device profile.
    ResourceAlternate   interface{}    `json:"resourceAlternateId,omitempty"` // The alternate ID of the device.
    ResourceDisplayName *DisplayName   `json:"resourceDisplayName,omitempty"` // The display name of the device.
    ResourceID          string         `json:"resourceId,omitempty"`          // The ID of the device.
    ResourceType        string         `json:"resourceType,omitempty"`        // The type of the device.
    Status              string         `json:"status,omitempty"`              // The status of the device.
}
```

<a name="DeviceProfile"></a>
## type DeviceProfile



```go
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
```

<a name="DeviceQuery"></a>
## type DeviceQuery

\- Query parameters for Devices

- Example: Devices that have a \`status\` of \`ACTIVE\` search=status eq "ACTIVE"
  
  Devices last updated after a specific timestamp search=lastUpdated gt "yyyy\-MM\-dd'T'HH:mm:ss.SSSZ"
  
  Devices with a specified \`id\` search=id eq "guo4a5u7JHHhjXrMK0g4"
  
  Devices that have a \`displayName\` of \`Bob\` search=profile.displayName eq "Bob"
  
  Devices that have an \`platform\` of \`WINDOWS\` search=profile.platform eq "WINDOWS"
  
  Devices whose \`sid\` starts with \`S\-1\` search=profile.sid sw "S\-1"

```go
type DeviceQuery struct {
    After  string // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the `Link` response header.
    Limit  string // Default: 20. Max. 200. A limit on the number of objects to return
    Search string // A SCIM filter expression that filters the results. Searches include all Device profile properties and the Device `id``, `status``, and `lastUpdated`` properties.
    Expand string // Lists associated users for the device in `_embedded` element
}
```

<a name="DeviceUser"></a>
## type DeviceUser



```go
type DeviceUser struct {
    Created          time.Time `json:"created,omitempty"`          // The timestamp when the device user was created.
    ManagementStatus string    `json:"managementStatus,omitempty"` // The management status of the device user.
    User             *User     `json:"user,omitempty"`             // The user assigned to the device.
}
```

<a name="DeviceUsers"></a>
## type DeviceUsers



```go
type DeviceUsers []DeviceUser
```

<a name="Devices"></a>
## type Devices



```go
type Devices []Device
```

<a name="DisplayName"></a>
## type DisplayName



```go
type DisplayName struct {
    Value     string `json:"value"`     // The display name of the device.
    Sensitive bool   `json:"sensitive"` // Indicates whether the display name is sensitive.
}
```

<a name="Embedded"></a>
## type Embedded



```go
type Embedded struct {
    Property1 map[string]interface{} `json:"property1,omitempty"` // Property1 is a map of string to interface.
    Property2 map[string]interface{} `json:"property2,omitempty"` // Property2 is a map of string to interface.
}
```

<a name="Error"></a>
## type Error



```go
type Error struct {
    ErrorCauses  []ErrorCause `json:"errorCauses,omitempty"`
    ErrorCode    string       `json:"errorCode,omitempty"`
    ErrorId      string       `json:"errorId,omitempty"`
    ErrorLink    string       `json:"errorLink,omitempty"`
    ErrorSummary string       `json:"errorSummary,omitempty"`
}
```

<a name="ErrorCause"></a>
## type ErrorCause



```go
type ErrorCause struct {
    ErrorSummary string `json:"errorSummary,omitempty"`
}
```

<a name="Hints"></a>
## type Hints



```go
type Hints struct {
    Allow []string `json:"allow,omitempty"` // Allow is a list of allowed methods.
}
```

<a name="Licensing"></a>
## type Licensing



```go
type Licensing struct {
    SeatCount int `json:"seatCount,omitempty"`
}
```

<a name="Link"></a>
## type Link



```go
type Link struct {
    Hints  Hints  `json:"hints,omitempty"`  // Hints is a list of hints for the link.
    Href   string `json:"href,omitempty"`   // Href is the URL for the link.
    Method string `json:"method,omitempty"` // Method is the HTTP method for the link.
    Type   string `json:"type,omitempty"`   // Type is the type of link.
}
```

<a name="Links"></a>
## type Links



```go
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
```

<a name="PasswordCredentials"></a>
## type PasswordCredentials



```go
type PasswordCredentials struct {
    Hook  *PasswordHook `json:"hook,omitempty"`
    Value string        `json:"value,omitempty"`
    Hash  *PasswordHash `json:"hash,omitempty"`
}
```

<a name="PasswordHash"></a>
## type PasswordHash



```go
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
```

<a name="PasswordHook"></a>
## type PasswordHook



```go
type PasswordHook struct {
    Type string `json:"type,omitempty"`
}
```

<a name="Permission"></a>
## type Permission



```go
type Permission struct {
    Created     time.Time `json:"created,omitempty"`
    Label       string    `json:"label,omitempty"`
    LastUpdated time.Time `json:"lastUpdated,omitempty"`
    Links       *Links    `json:"_links,omitempty"`
}
```

<a name="Provider"></a>
## type Provider



```go
type Provider struct {
    Name string `json:"name,omitempty"`
    Type string `json:"type,omitempty"`
}
```

<a name="RecoveryQuestion"></a>
## type RecoveryQuestion



```go
type RecoveryQuestion struct {
    Answer   string `json:"answer,omitempty"`
    Question string `json:"question,omitempty"`
}
```

<a name="Role"></a>
## type Role



```go
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
```

<a name="RoleReport"></a>
## type RoleReport



```go
type RoleReport struct {
    Role  *Role
    Users []*User
}
```

<a name="Roles"></a>
## type Roles



```go
type Roles struct {
    Roles []Role `json:"roles,omitempty"`
}
```

<a name="User"></a>
## type User



```go
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
    Embedded              *Embedded        `json:"_embedded,omitempty"`
    Links                 *Links           `json:"_links,omitempty"`
}
```

<a name="UserCredentials"></a>
## type UserCredentials



```go
type UserCredentials struct {
    Password         *PasswordCredentials `json:"password,omitempty"`
    Provider         *Provider            `json:"provider,omitempty"`
    RecoveryQuestion *RecoveryQuestion    `json:"recovery_question,omitempty"`
}
```

<a name="UserProfile"></a>
## type UserProfile



```go
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
```

<a name="UserQuery"></a>
## type UserQuery

\* Query Parameters for Users

```go
type UserQuery struct {
    Q         string // Searches the records for matching value
    After     string // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the `Link` response header.
    Limit     string // Default: 200. Specifies the number of results returned. Defaults to 10 if `q` is provided
    Filter    string // Filters users with a supported expression for a subset of properties
    Search    string // A SCIM filter expression for most properties. Okta recommends using this parameter for search for best performance
    SortBy    string // Specifies the attribute by which to sort the results. Valid values are `id`, `created`, `activated`, `status`, and `lastUpdated`. The default is `id`
    SoftOrder string // Sorting is done in ASCII sort order (that is, by ASCII character value), but isn't case sensitive
}
```

<a name="UserType"></a>
## type UserType



```go
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
```

<a name="Users"></a>
## type Users



```go
type Users []*User
```

<a name="Visibility"></a>
## type Visibility



```go
type Visibility struct {
    AppLinks          map[string]bool `json:"appLinks,omitempty"`
    AutoLaunch        bool            `json:"autoLaunch,omitempty"`
    AutoSubmitToolbar bool            `json:"autoSubmitToolbar,omitempty"`
    Hide              map[string]bool `json:"hide,omitempty"`
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
