/*
# Google Workspace - Admin (Devices)

This package initializes all the methods for functions which interact with Devices from the Google Admin API:
https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices

:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/drive.go
package google

import (
	"fmt"
	"strings"
	"time"
)

var (
	V1_ChromeBaseURL    = fmt.Sprintf("%s/v1/customers", ChromeBaseURL)
	DevicePolicies      = fmt.Sprintf("%s/%s/policies", V1_ChromeBaseURL, "%s")
	DevicePolicySchemas = fmt.Sprintf("%s/%s/policySchemas", V1_ChromeBaseURL, "%s")
)

// DeviceClient for chaining methods
type DeviceClient struct {
	client *Client
	query  DeviceQuery
}

// Entry point for device-related operations
func (c *Client) Devices() *DeviceClient {
	return &DeviceClient{
		client: c,
		query: DeviceQuery{ // Default query parameters
			MaxResults: 500,
		},
	}
}

/*
 * Query Parameters for ChromeOS Devices
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices/list#query-parameters
 */
type DeviceQuery struct {
	IncludeChildOrgunits bool   `url:"includeChildOrgunits,omitempty"` // If true, return devices from all child org units as well as the specified org unit.
	MaxResults           int    `url:"maxResults,omitempty"`           // Maximum number of results to return. Default is 100
	OrderBy              string `url:"orderBy,omitempty"`              // Device property to use for sorting results. Should be one of the defined OrderBy enums.
	OrgUnitPath          string `url:"orgUnitPath,omitempty"`          // Full path of the organizational unit (minus the leading /) or its unique ID.
	PageToken            string `url:"pageToken,omitempty"`            // Token for requesting the next page of query results.
	Projection           string `url:"projection,omitempty"`           // Restrict information returned to a set of selected fields. Should be one of the defined Projection enums.
	Query                string `url:"query,omitempty"`                // https://developers.google.com/admin-sdk/directory/v1/list-query-operators
	SortOrder            string `url:"sortOrder,omitempty"`            // Whether to return results in ascending or descending order. Should be one of the defined SortOrder enums.
}

// ### Chainable DeviceClient Methods
// ---------------------------------------------------------------------
func (dc *DeviceClient) MaxResults(max int) *DeviceClient {
	dc.query.MaxResults = max
	return dc
}

func (dc *DeviceClient) PageToken(token string) *DeviceClient {
	dc.query.PageToken = token
	return dc
}

func (dc *DeviceClient) Query(query string) *DeviceClient {
	dc.query.Query = query
	return dc
}

// END OF CHAINABLE METHODS
//---------------------------------------------------------------------

/*
 * Query Parameters for Device Policy Schemas
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policySchemas/list#query-parameters
 */
type PolicyQuery struct {
	Filter    string `url:"filter,omitempty"`    // https://developers.google.com/chrome/policy/guides/list-policy-schemas#filter_syntax
	PageSize  int    `url:"pageSize,omitempty"`  // The maximum number of policy schemas to return, defaults to 100 and has a maximum of 1000.
	PageToken string `url:"pageToken,omitempty"` // Token for requesting the next page of query results.
}

/*
 * Request Parameters for Device Policies
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policies/resolve#PolicyRequest
 */
type PolicyRequest struct {
	PolicySchemaFilter string          `json:"policySchemaFilter,omitempty"` // https://developers.google.com/chrome/policy/guides/policy-schemas#policy_schema_names
	PolicyTargetKey    PolicyTargetKey `json:"policyTargetKey,omitempty"`    // https://developers.google.com/chrome/policy/reference/rest/v1/PolicyTargetKey
	PageSize           int             `json:"pageSize,omitempty"`           // The maximum number of resolved policies to return, defaults to 100 and has a maximum of 1000.
	PageToken          string          `json:"pageToken,omitempty"`          // Token for requesting the next page of query results.
}

type PolicyTargetKey struct {
	TargetResource       string   `json:"targetResource,omitempty"`       // The target resource name for the policy target key.
	AdditionalTargetKeys []string `json:"additionalTargetKeys,omitempty"` // The additional target keys for the policy target key.
}

/*
 * List all ChromeOS Devices in the domain with pagination support
 * admin/directory/v1/customer/{customerId}/devices/chromeos
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices/list
 */
func (dc *DeviceClient) ListAllChromeOS(customer *Customer) (*ChromeOSDevices, error) {
	dc.client.Log.Println("Getting all ChromeOS Devices...")

	url := dc.client.BuildURL(DirectoryChromeOSDevices, customer)

	var cache ChromeOSDevices
	if dc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	devices, err := doPaginated[ChromeOSDevices](dc.client, "GET", url, dc.query, nil)
	if err != nil {
		return nil, err
	}

	dc.client.SetCache(url, devices, 5*time.Minute)
	return devices, nil
}

/*
 * List all Provisioned ChromeOS Devices in the domain with pagination support
 * admin/directory/v1/customer/{customerId}/devices/chromeos
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices/list
 */
func (dc *DeviceClient) ListAllProvisionedChromeOS(customer *Customer) (*ChromeOSDevices, error) {
	dc.client.Log.Println("Getting all ChromeOS Devices...")

	url := dc.client.BuildURL(DirectoryChromeOSDevices, customer)

	var cache ChromeOSDevices
	if dc.client.GetCache(url, &cache) {
		return &cache, nil
	}

	dc.Query("status:provisioned")

	devices, err := doPaginated[ChromeOSDevices](dc.client, "GET", url, dc.query, nil)
	if err != nil {
		return nil, err
	}

	dc.client.SetCache(url, devices, 5*time.Minute)
	return devices, nil
}

/*
 * Gets a list of policy schemas that match a specified filter value for a given customer
 * chromepolicy.googleapis.com/v1/{customerId}/policySchemas
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policySchemas/list
 */
func (c *Client) ListAllDevicePolicySchemas(customer *Customer) (*PolicySchemas, error) {
	c.Log.Println("Getting all ChromeOS Device Policy Schemas...")
	q := &PolicyQuery{
		PageSize: 1000,
	}

	url := c.BuildURL(DevicePolicySchemas, customer)

	var cache PolicySchemas
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	policySchemas, err := doPaginated[PolicySchemas](c, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, policySchemas, 60*time.Minute)
	return policySchemas, nil
}

/*
 * Gets the resolved policy values for a list of policies that match a search query.
 * chromepolicy.googleapis.com/v1/{customerId}/policies:resolve
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policies/resolve
 */
func (c *Client) ResolvePolicySchemas(customer *Customer, ou *OrgUnit) (*ResolvedPolicies, error) {
	c.Log.Println("Getting all ChromeOS Device Policies...")

	url := c.BuildURL(DevicePolicies, customer, ":resolve")
	cacheKey := fmt.Sprintf("%s_%s", url, ou.ID)

	var cache ResolvedPolicies
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}
	policies := new(ResolvedPolicies)
	policies.Init()

	req := &PolicyRequest{
		PolicyTargetKey: PolicyTargetKey{
			TargetResource: fmt.Sprintf("orgunits/%s", strings.TrimPrefix(ou.ID, "id:")),
		},
		PageSize: 1000,
	}

	req.PolicySchemaFilter = "chrome.users.*"
	userPolicies, err := doPaginated[ResolvedPolicies](c, "POST", url, nil, req)
	if err != nil {
		return nil, err
	}
	*policies.Users.ResolvedPolicies = append(*policies.Users.ResolvedPolicies, *userPolicies.ResolvedPolicies...)

	for _, policy := range *policies.Users.ResolvedPolicies {
		if strings.Contains(policy.SourceKey.TargetResource, strings.TrimPrefix(ou.ID, "id:")) {
			*policies.Users.Direct = append(*policies.Users.Direct, policy)
		} else {
			*policies.Users.Inherited = append(*policies.Users.Inherited, policy)
		}
	}

	req.PolicySchemaFilter = "chrome.devices.*"
	devicePolicies, err := doPaginated[ResolvedPolicies](c, "POST", url, nil, req)
	if err != nil {
		return nil, err
	}
	*policies.Devices.ResolvedPolicies = append(*policies.Devices.ResolvedPolicies, *devicePolicies.ResolvedPolicies...)

	for _, policy := range *policies.Devices.ResolvedPolicies {
		if strings.Contains(policy.SourceKey.TargetResource, strings.TrimPrefix(ou.ID, "id:")) {
			*policies.Devices.Direct = append(*policies.Devices.Direct, policy)
		} else {
			*policies.Devices.Inherited = append(*policies.Devices.Inherited, policy)
		}
	}

	c.SetCache(cacheKey, policies, 5*time.Minute)
	return policies, nil
}
