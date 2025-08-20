/*
# Google Workspace - Chrome Management (Browsers)

This package contains methods for Chrome Managed Browsers from the Google Chrome Management API:
https://support.google.com/chrome/a/answer/9681204

:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Justin Hernandez <justin.hernandez@gemini.com>
*/

// pkg/google/browsers.go
package google

import (
	"fmt"
	"time"
)

var (
	ChromeManagementBaseURL     = "https://chromemanagement.googleapis.com/v1/customers"
	ChromeBrowsersEndpoint      = fmt.Sprintf("%s/%%s/reports:countChromeBrowsersNeedingAttention", ChromeManagementBaseURL)
	DirectoryBaseURL            = "https://www.googleapis.com/admin/directory/v1.1beta1"
	AllChromeBrowsersEndpoint   = fmt.Sprintf("%s/customer/%%s/devices/chromebrowsers", DirectoryBaseURL)
	UpdateChromeBrowserEndpoint = fmt.Sprintf("%s/customer/%%s/devices/chromebrowsers/%%s", DirectoryBaseURL)
	MoveChromeBrowsersEndpoint  = fmt.Sprintf("%s/customer/%%s/devices/chromebrowsers/moveChromeBrowsersToOu", DirectoryBaseURL)
)

// ChromeBrowser represents a Chrome browser in the organization
type ChromeBrowser struct {
	DeviceId            string                `json:"deviceId,omitempty"`            // Device ID
	BrowserVersion      string                `json:"browserVersion,omitempty"`      // Chrome browser version
	Channel             string                `json:"channel,omitempty"`             // Chrome release channel (stable, beta, dev, canary)
	LastActivity        string                `json:"lastActivity,omitempty"`        // Last activity timestamp
	LastPolicySync      string                `json:"lastPolicySync,omitempty"`      // Last policy sync timestamp
	Machine             string                `json:"machine,omitempty"`             // Machine name
	OrgUnitPath         string                `json:"orgUnitPath,omitempty"`         // Organizational unit path
	OsVersion           string                `json:"osVersion,omitempty"`           // Operating system version
	UserEmail           string                `json:"userEmail,omitempty"`           // User email address
	PolicyCount         int64                 `json:"policyCount,omitempty"`         // Number of policies applied
	ExtensionCount      int64                 `json:"extensionCount,omitempty"`      // Number of extensions installed
	Issues              []ChromeBrowserIssue  `json:"issues,omitempty"`              // List of issues with the browser
	Policies            []ChromeBrowserPolicy `json:"policies,omitempty"`            // Applied policies
	InstalledExtensions []ChromeExtension     `json:"installedExtensions,omitempty"` // Installed extensions
}

// ChromeBrowserIssue represents an issue with a Chrome browser
type ChromeBrowserIssue struct {
	Type        string `json:"type,omitempty"`        // Issue type (e.g., "POLICY_VIOLATION", "OUTDATED_VERSION")
	Severity    string `json:"severity,omitempty"`    // Issue severity (e.g., "HIGH", "MEDIUM", "LOW")
	Description string `json:"description,omitempty"` // Human-readable description
	Resolution  string `json:"resolution,omitempty"`  // Suggested resolution
}

// ChromeBrowserPolicy represents a policy applied to a Chrome browser
type ChromeBrowserPolicy struct {
	PolicyName   string      `json:"policyName,omitempty"`   // Policy name
	PolicyValue  interface{} `json:"policyValue,omitempty"`  // Policy value
	PolicySource string      `json:"policySource,omitempty"` // Source of the policy
	IsCompliant  bool        `json:"isCompliant,omitempty"`  // Whether the policy is compliant
}

// ChromeExtension represents a Chrome extension
type ChromeExtension struct {
	ExtensionId string   `json:"extensionId,omitempty"` // Extension ID
	Name        string   `json:"name,omitempty"`        // Extension name
	Version     string   `json:"version,omitempty"`     // Extension version
	Enabled     bool     `json:"enabled,omitempty"`     // Whether the extension is enabled
	InstallType string   `json:"installType,omitempty"` // Installation type (e.g., "ADMIN", "NORMAL")
	Permissions []string `json:"permissions,omitempty"` // Extension permissions
}

// ChromeBrowserQuery represents query parameters for Chrome Browser reports
type ChromeBrowserQuery struct {
	Filter      string `url:"filter,omitempty"`      // Filter expression
	PageSize    int    `url:"pageSize,omitempty"`    // Maximum number of results (default: 100, max: 1000)
	PageToken   string `url:"pageToken,omitempty"`   // Token for next page
	OrderBy     string `url:"orderBy,omitempty"`     // Sort order
	OrgUnitPath string `url:"orgUnitPath,omitempty"` // Organizational unit path
}

// Directory API Structs for All Chrome Browsers
// ChromeBrowserDevices represents the response from Directory API for all Chrome browsers
type ChromeBrowserDevices struct {
	Kind          string                 `json:"kind,omitempty"`          // "admin#directory#browserdevices"
	Browsers      []*ChromeBrowserDevice `json:"browsers,omitempty"`      // List of Chrome browser devices
	NextPageToken string                 `json:"nextPageToken,omitempty"` // Token for next page of results
}

func (c *ChromeBrowserDevices) Append(result *ChromeBrowserDevices) *ChromeBrowserDevices {
	c.Browsers = append(c.Browsers, result.Browsers...)
	return c
}

func (c *ChromeBrowserDevices) PageToken() string {
	return c.NextPageToken
}

// ChromeBrowserDevice represents a Chrome browser device from Directory API
type ChromeBrowserDevice struct {
	DeviceId                string   `json:"deviceId,omitempty"`                // Unique device identifier
	Kind                    string   `json:"kind,omitempty"`                    // Resource type identifier
	LastPolicyFetchTime     string   `json:"lastPolicyFetchTime,omitempty"`     // Last policy fetch timestamp
	OsPlatform              string   `json:"osPlatform,omitempty"`              // Operating system platform
	OsArchitecture          string   `json:"osArchitecture,omitempty"`          // OS architecture
	OsVersion               string   `json:"osVersion,omitempty"`               // Operating system version
	MachineName             string   `json:"machineName,omitempty"`             // Machine name
	LastRegistrationTime    string   `json:"lastRegistrationTime,omitempty"`    // Last registration timestamp
	ExtensionCount          string   `json:"extensionCount,omitempty"`          // Number of extensions (as string)
	PolicyCount             string   `json:"policyCount,omitempty"`             // Number of policies (as string)
	LastDeviceUser          string   `json:"lastDeviceUser,omitempty"`          // Last user to use the device
	LastActivityTime        string   `json:"lastActivityTime,omitempty"`        // Last activity timestamp
	BrowserVersions         []string `json:"browserVersions,omitempty"`         // Array of browser versions
	SerialNumber            string   `json:"serialNumber,omitempty"`            // Device serial number
	VirtualDeviceId         string   `json:"virtualDeviceId,omitempty"`         // Virtual device identifier
	OrgUnitPath             string   `json:"orgUnitPath,omitempty"`             // Organizational unit path
	AnnotatedLocation       string   `json:"annotatedLocation,omitempty"`       // Annotated location
	AnnotatedUser           string   `json:"annotatedUser,omitempty"`           // Annotated user
	AnnotatedAssetId        string   `json:"annotatedAssetId,omitempty"`        // Annotated asset ID
	AnnotatedNotes          string   `json:"annotatedNotes,omitempty"`          // Annotated notes
	ChromeSignedinUserEmail string   `json:"chromeSignedinUserEmail,omitempty"` // Annotated notes
}

// DirectoryQuery represents query parameters for Directory API Chrome Browser requests
type DirectoryQuery struct {
	Projection  string `url:"projection,omitempty"`  // BASIC, FULL - amount of information to include
	MaxResults  int    `url:"maxResults,omitempty"`  // Maximum number of results (default: 100, max: 300)
	PageToken   string `url:"pageToken,omitempty"`   // Token for next page
	OrderBy     string `url:"orderBy,omitempty"`     // Property to sort by
	SortOrder   string `url:"sortOrder,omitempty"`   // ASCENDING or DESCENDING
	OrgUnitPath string `url:"orgUnitPath,omitempty"` // Organizational unit path
	Query       string `url:"query,omitempty"`       // Search query
}

func (q DirectoryQuery) SetPageToken(token string) {
	q.PageToken = token
}

// ChromeBrowserUpdateRequest represents the updatable fields for a Chrome browser device
type ChromeBrowserUpdateRequest struct {
	AnnotatedUser     string `json:"annotatedUser,omitempty"`     // Annotated user
	AnnotatedLocation string `json:"annotatedLocation,omitempty"` // Annotated location
	AnnotatedAssetId  string `json:"annotatedAssetId,omitempty"`  // Annotated asset ID
	AnnotatedNotes    string `json:"annotatedNotes,omitempty"`    // Annotated notes
	OrgUnitPath       string `json:"orgUnitPath,omitempty"`       // Organizational unit path
}

// BrowserClient for chaining methods
type BrowserClient struct {
	client         *Client
	query          ChromeBrowserQuery
	directoryQuery DirectoryQuery
}

type BrowserMoveOUPayload struct {
	ResourceIds []string `json:"resource_ids,omitempty"`  // List of unique device IDs of Chrome browser devices to move.
	OrgUnitPath string   `json:"org_unit_path,omitempty"` // Destination organization unit to move devices to.
}

// Entry point for browser-related operations
func (c *Client) Browsers() *BrowserClient {
	return &BrowserClient{
		client: c,
		query: ChromeBrowserQuery{
			PageSize: 100, // Default page size
			Filter:   "",
		},
		directoryQuery: DirectoryQuery{
			MaxResults: 100,     // Default page size
			Projection: "BASIC", // Default to full information
		},
	}
}

// ### Chainable BrowserClient Methods
// ---------------------------------------------------------------------
func (bc *BrowserClient) PageSize(size int) *BrowserClient {
	bc.query.PageSize = size
	return bc
}

func (bc *BrowserClient) PageToken(token string) *BrowserClient {
	bc.query.PageToken = token
	return bc
}

func (bc *BrowserClient) Filter(filter string) *BrowserClient {
	bc.query.Filter = filter
	return bc
}

func (bc *BrowserClient) OrderBy(orderBy string) *BrowserClient {
	bc.query.OrderBy = orderBy
	return bc
}

func (bc *BrowserClient) OrgUnitPath(path string) *BrowserClient {
	bc.query.OrgUnitPath = path
	return bc
}

// ### Directory API Chainable Methods
// ---------------------------------------------------------------------
func (bc *BrowserClient) MaxResults(max int) *BrowserClient {
	bc.directoryQuery.MaxResults = max
	return bc
}

func (bc *BrowserClient) Projection(proj string) *BrowserClient {
	bc.directoryQuery.Projection = proj
	return bc
}

func (bc *BrowserClient) DirectoryPageToken(token string) *BrowserClient {
	bc.directoryQuery.PageToken = token
	return bc
}

func (bc *BrowserClient) DirectoryOrderBy(orderBy string) *BrowserClient {
	bc.directoryQuery.OrderBy = orderBy
	return bc
}

func (bc *BrowserClient) SortOrder(order string) *BrowserClient {
	bc.directoryQuery.SortOrder = order
	return bc
}

func (bc *BrowserClient) DirectoryOrgUnitPath(path string) *BrowserClient {
	bc.directoryQuery.OrgUnitPath = path
	return bc
}

func (bc *BrowserClient) DirectoryQuery(query string) *BrowserClient {
	bc.directoryQuery.Query = query
	return bc
}

// GetAllChromeBrowsers retrieves all Chrome browser devices using Directory API
func (bc *BrowserClient) GetAllChromeBrowsers() (*ChromeBrowserDevices, error) {
	// Check if customer is set
	if bc.client.Customer == nil {
		_, err := bc.client.MyCustomer()
		if err != nil {
			return nil, fmt.Errorf("failed to get customer: %w", err)
		}
	}

	// Build URL with customer ID - use "my_customer" for Directory API
	customerID := "my_customer"
	if bc.client.Customer != nil {
		customerID = bc.client.Customer.String()
	}
	url := fmt.Sprintf(AllChromeBrowsersEndpoint, customerID)

	// Make API request
	result, err := do[ChromeBrowserDevices](bc.client, "GET", url, bc.directoryQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all Chrome browsers: %w", err)
	}

	return &result, nil
}

// GetAllChromeBrowsersPaginated retrieves all Chrome browser devices using pagination
func (bc *BrowserClient) GetAllChromeBrowsersPaginated() (*ChromeBrowserDevices, error) {
	// Check if customer is set
	if bc.client.Customer == nil {
		_, err := bc.client.MyCustomer()
		if err != nil {
			return nil, fmt.Errorf("failed to get customer: %w", err)
		}
	}

	// Build URL with customer ID
	customerID := "my_customer"
	if bc.client.Customer != nil {
		customerID = bc.client.Customer.String()
	}
	url := fmt.Sprintf(AllChromeBrowsersEndpoint, customerID)

	// Use paginated request
	result, err := doPaginated[*ChromeBrowserDevices, GoogleQuery](bc.client, "GET", url, bc.directoryQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all Chrome browsers (paginated): %w", err)
	}

	// Debug output to identify API response issues
	fmt.Printf("DEBUG: API Response - Kind: %s\n", (*result).Kind)
	fmt.Printf("DEBUG: Number of browsers found: %d\n", len((*result).Browsers))
	fmt.Printf("DEBUG: Next page token: %s\n", (*result).NextPageToken)

	return *result, nil
}

func (bc *BrowserClient) UpdateChromeBrowser(deviceId string, updates *ChromeBrowserUpdateRequest) (*ChromeBrowserDevice, error) {
	if bc.client.Customer == nil {
		_, err := bc.client.MyCustomer()
		if err != nil {
			return nil, fmt.Errorf("failed to get customer: %w", err)
		}
	}

	customerID := "my_customer"
	if bc.client.Customer != nil {
		customerID = bc.client.Customer.String()
	}
	url := fmt.Sprintf(UpdateChromeBrowserEndpoint, customerID, deviceId)

	bc.client.Log.Printf("Updating Chrome browser device %s...", deviceId)
	device, err := do[ChromeBrowserDevice](bc.client, "PUT", url, nil, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update Chrome browser device: %w", err)
	}

	return &device, nil
}

func (bc *BrowserClient) MoveChromeBrowserOU(rids []string, oup string) error {
	if bc.client.Customer == nil {
		_, err := bc.client.MyCustomer()
		if err != nil {
			return fmt.Errorf("failed to get customer: %w", err)
		}
	}

	customerID := "my_customer"
	if bc.client.Customer != nil {
		customerID = bc.client.Customer.String()
	}
	url := fmt.Sprintf(MoveChromeBrowsersEndpoint, customerID)

	data := struct {
		ResourceIDs []string `json:"resource_ids,omitempty"`  // List of unique device IDs of Chrome browser devices to move
		OrgUnitPath string   `json:"org_unit_path,omitempty"` // Destination organization unit to move devices to
	}{
		ResourceIDs: rids,
		OrgUnitPath: oup,
	}

	//bc.client.Log.Printf("Updating Chrome browser device %s...", deviceId)
	_, err := do[any](bc.client, "POST", url, nil, data)
	if err != nil {
		return fmt.Errorf("failed to update Chrome browser device: %w", err)
	}

	return nil
}

func (c *Client) MyCustomer() (*Customer, error) {
	url := fmt.Sprintf(DirectoryCustomers, "my_customer")

	var cache Customer
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	c.Log.Println("Getting ID of current client...")
	customer, err := do[Customer](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, customer, 1*time.Hour)
	return &customer, nil
}
