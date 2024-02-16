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
	"encoding/json"
	"fmt"
)

var (
	V1_ChromeBaseURL    = fmt.Sprintf("%s/v1/customers", ChromeBaseURL)
	DevicePolicies      = fmt.Sprintf("%s/%s/policies", V1_ChromeBaseURL, "%s")
	DevicePolicySchemas = fmt.Sprintf("%s/%s/policySchemas", V1_ChromeBaseURL, "%s")
)

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
	PolicySchemaFilter string          `json:"policySchemaFilter,omitempty"` // https://developers.google.com/chrome/policy/reference/rest/v1/customers.policies/resolve#PolicyRequest
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
func (c *Client) ListAllChromeOS(customerId string) (*ChromeOSDevices, error) {
	c.Logger.Println("Getting all ChromeOS Devices...")
	allDevices := &ChromeOSDevices{}
	q := &DeviceQuery{
		MaxResults: 500,
	}

	var url string
	switch customerId {
	case "":
		url = fmt.Sprintf(DirectoryChromeOSDevices, "my_customer")
	default:
		url = fmt.Sprintf(DirectoryChromeOSDevices, customerId)
	}

	for next_page := true; next_page; next_page = (q.PageToken != "") {
		devices := &ChromeOSDevices{}

		res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)

		if err != nil {
			return nil, err
		}
		c.Logger.Println("Response Status:", res.Status)
		c.Logger.Debug("Response Body:", string(body))

		err = json.Unmarshal(body, &devices)
		if err != nil {
			return nil, err
		}

		allDevices.ChromeOSDevices = append(allDevices.ChromeOSDevices, devices.ChromeOSDevices...)
		q.PageToken = devices.NextPageToken
	}

	return allDevices, nil
}

/*
 * List all Provisioned ChromeOS Devices in the domain with pagination support
 * admin/directory/v1/customer/{customerId}/devices/chromeos
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices/list
 */
func (c *Client) ListAllProvisionedChromeOS(customerId string) (*ChromeOSDevices, error) {
	c.Logger.Println("Getting all ChromeOS Devices...")
	allDevices := &ChromeOSDevices{}
	q := &DeviceQuery{
		MaxResults: 500,
		Query:      "status:provisioned",
	}

	var url string
	switch customerId {
	case "":
		url = fmt.Sprintf(DirectoryChromeOSDevices, "my_customer")
	default:
		url = fmt.Sprintf(DirectoryChromeOSDevices, customerId)
	}

	for next_page := true; next_page; next_page = (q.PageToken != "") {
		devices := &ChromeOSDevices{}

		res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)

		if err != nil {
			return nil, err
		}
		c.Logger.Println("Response Status:", res.Status)
		c.Logger.Debug("Response Body:", string(body))

		err = json.Unmarshal(body, &devices)
		if err != nil {
			return nil, err
		}

		allDevices.ChromeOSDevices = append(allDevices.ChromeOSDevices, devices.ChromeOSDevices...)
		q.PageToken = devices.NextPageToken
	}

	return allDevices, nil
}

/*
 * Gets a list of policy schemas that match a specified filter value for a given customer
 * chromepolicy.googleapis.com/v1/{customerId}/policySchemas
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policySchemas/list
 */
func (c *Client) ListAllDevicePolicySchemas(customerId string) (*PolicySchemas, error) {
	c.Logger.Println("Getting all ChromeOS Device Policy Schemas...")
	allPolicySchemas := &PolicySchemas{}
	q := &PolicyQuery{
		PageSize: 1000,
	}

	var url string
	switch customerId {
	case "":
		url = fmt.Sprintf(DevicePolicySchemas, "my_customer")
	default:
		url = fmt.Sprintf(DevicePolicySchemas, customerId)
	}

	for next_page := true; next_page; next_page = (q.PageToken != "") {
		policySchemas := &PolicySchemas{}

		res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)

		if err != nil {
			return nil, err
		}
		c.Logger.Println("Response Status:", res.Status)
		c.Logger.Debug("Response Body:", string(body))

		err = json.Unmarshal(body, &policySchemas)
		if err != nil {
			return nil, err
		}

		allPolicySchemas.PolicySchemas = append(allPolicySchemas.PolicySchemas, policySchemas.PolicySchemas...)
		q.PageToken = policySchemas.NextPageToken
	}

	return allPolicySchemas, nil
}

/*
 * Gets the resolved policy values for a list of policies that match a search query.
 * chromepolicy.googleapis.com/v1/{customerId}/policies:resolve
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policies/resolve
 */
func (c *Client) ResolvePolicySchemas(customerId string, OU string) (*ResolvedPolicies, error) {
	c.Logger.Println("Getting all ChromeOS Device Policies...")
	allPolicies := &ResolvedPolicies{}
	req := &PolicyRequest{
		PolicySchemaFilter: "chrome.users.*",
		PolicyTargetKey: PolicyTargetKey{
			TargetResource: fmt.Sprintf("orgunits/%s", OU),
		},
		PageSize: 1000,
	}

	var url string
	switch customerId {
	case "":
		url = fmt.Sprintf("%s:resolve", fmt.Sprintf(DevicePolicies, "my_customer"))
	default:
		url = fmt.Sprintf("%s:resolve", fmt.Sprintf(DevicePolicies, customerId))
	}

	var nextPageToken string
	for next_page := true; next_page; next_page = (nextPageToken != "") {
		policies := &ResolvedPolicies{}

		res, body, err := c.HTTPClient.DoRequest("POST", url, nil, req)

		if err != nil {
			return nil, err
		}
		c.Logger.Println("Response Status:", res.Status)
		c.Logger.Debug("Response Body:", string(body))

		err = json.Unmarshal(body, &policies)
		if err != nil {
			return nil, err
		}

		allPolicies.ResolvedPolicies = append(allPolicies.ResolvedPolicies, policies.ResolvedPolicies...)
		nextPageToken = policies.NextPageToken
	}

	return allPolicies, nil
}
