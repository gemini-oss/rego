/*
# Google Workspace - Chrome Policies

This package implements logic related to the `Policy` resources of the Chrome Policy API:
https://developers.google.com/chrome/policy/reference/rest

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/chrome_policy.go
package google

import (
	"fmt"
	"strings"
	"time"

	"github.com/gemini-oss/rego/pkg/common/starstruct"
)

var (
	V1_ChromeBaseURL    = fmt.Sprintf("%s/v1/customers", ChromeBaseURL)
	ChromePolicies      = fmt.Sprintf("%s/%s/policies", V1_ChromeBaseURL, "%s")
	ChromePolicySchemas = fmt.Sprintf("%s/%s/policySchemas", V1_ChromeBaseURL, "%s")
)

// Chrome Policy Client for chaining methods
type ChromePolicyClient struct {
	client *Client
}

// Entry point for Chrome Policy-related operations
func (c *Client) ChromePolicy() *ChromePolicyClient {
	return &ChromePolicyClient{
		client: c,
	}
}

/*
 * Query Parameters for Chrome Policy Schemas
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policySchemas/list#query-parameters
 */
type ChromePolicyQuery struct {
	Filter    string `url:"filter,omitempty"`    // https://developers.google.com/chrome/policy/guides/list-policy-schemas#filter_syntax
	PageSize  int    `url:"pageSize,omitempty"`  // The maximum number of policy schemas to return, defaults to 100 and has a maximum of 1000.
	PageToken string `url:"pageToken,omitempty"` // Token for requesting the next page of query results.
}

func (q *ChromePolicyQuery) SetPageToken(token string) {
	q.PageToken = token
}

/*
 * Check if the ChromePolicyQuery is empty
 */
func (cp *ChromePolicyQuery) IsEmpty() bool {
	return cp.Filter == "" &&
		cp.PageSize == 0 &&
		cp.PageToken == ""
}

/*
 * Validate the query parameters for the Chrome Policy resource
 */
func (cp *ChromePolicyQuery) ValidateQuery() error {
	if cp.IsEmpty() {
		return nil
	}

	return nil
}

/*
 * Request Parameters for Chrome Policies
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
 * A PolicyTarget is an interface that defines what Google Workspace resources can be used as a target for Chrome Policies.
 * It is used to resolve policies for a specific target resource.
 */
type PolicyTarget interface {
	Resource() string
	ThisID() string
	ThisName() string
}

func (ou *OrgUnit) Resource() string {
	return fmt.Sprintf("orgunits/%s", strings.TrimPrefix(ou.ID, "id:"))
}
func (ou *OrgUnit) ThisID() string   { return ou.ID }
func (ou *OrgUnit) ThisName() string { return ou.Name }

func (g *Group) Resource() string {
	return fmt.Sprintf("groups/%s", g.ID)
}
func (g *Group) ThisID() string   { return g.ID }
func (g *Group) ThisName() string { return g.Name }

/*
 * Gets a list of policy schemas that match a specified filter value for a given customer
 * chromepolicy.googleapis.com/v1/{customerId}/policySchemas
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policySchemas/list
 */
func (c *ChromePolicyClient) ListAllChromePolicySchemas(customer *Customer) (*PolicySchemas, error) {
	c.client.Log.Println("Getting all Chrome Policy Schemas...")
	q := &ChromePolicyQuery{
		PageSize: 1000,
	}

	url := c.client.BuildURL(ChromePolicySchemas, customer)

	var cache PolicySchemas
	if c.client.GetCache(url, &cache) {
		return &cache, nil
	}

	policySchemas, err := doPaginated[PolicySchemas, *ChromePolicyQuery](c.client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.client.SetCache(url, policySchemas, 60*time.Minute)
	return policySchemas, nil
}

/*
 * Gets the resolved policy values for a list of policies that match a search query.
 * chromepolicy.googleapis.com/v1/{customerId}/policies:resolve
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policies/resolve
 */
func (c *ChromePolicyClient) ResolvePolicySchemas(customer *Customer, target PolicyTarget) (*ResolvedPolicies, error) {
	c.client.Log.Println("Getting all ChromeOS Device Policies...")

	url := c.client.BuildURL(ChromePolicies, customer, ":resolve")
	cacheKey := fmt.Sprintf("%s_%s", url, target.ThisID())

	var cache ResolvedPolicies
	if c.client.GetCache(cacheKey, &cache) {
		return &cache, nil
	}
	policies := new(ResolvedPolicies)
	policies.Init()

	req := &PolicyRequest{
		PolicyTargetKey: PolicyTargetKey{
			TargetResource: target.Resource(),
		},
		PageSize: 1000,
	}

	req.PolicySchemaFilter = "chrome.users.*"
	userPolicies, err := doPaginated[ResolvedPolicies, *ChromePolicyQuery](c.client, "POST", url, nil, req)
	if err != nil {
		return nil, err
	}
	*policies.Users.ResolvedPolicies = append(*policies.Users.ResolvedPolicies, *userPolicies.ResolvedPolicies...)

	for _, policy := range *policies.Users.ResolvedPolicies {
		if strings.Contains(policy.SourceKey.TargetResource, target.Resource()) {
			*policies.Users.Direct = append(*policies.Users.Direct, policy)
		} else {
			*policies.Users.Inherited = append(*policies.Users.Inherited, policy)
		}
	}

	req.PolicySchemaFilter = "chrome.devices.*"
	devicePolicies, err := doPaginated[ResolvedPolicies, *ChromePolicyQuery](c.client, "POST", url, nil, req)
	if err != nil {
		return nil, err
	}
	*policies.Devices.ResolvedPolicies = append(*policies.Devices.ResolvedPolicies, *devicePolicies.ResolvedPolicies...)

	for _, policy := range *policies.Devices.ResolvedPolicies {
		if strings.Contains(policy.SourceKey.TargetResource, target.Resource()) {
			*policies.Devices.Direct = append(*policies.Devices.Direct, policy)
		} else {
			*policies.Devices.Inherited = append(*policies.Devices.Inherited, policy)
		}
	}

	c.client.SetCache(cacheKey, policies, 5*time.Minute)
	return policies, nil
}

/*
 * Gets the resolved policy values for a list of policies that match a search query.
 * chromepolicy.googleapis.com/v1/{customerId}/policies:resolve
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policies/resolve
 */
func (c *ChromePolicyClient) Report(adminClient *Client, target PolicyTarget) (*ResolvedPolicies, error) {
	c.client.Log.Println("Saving Chrome Policy Report to Spreadsheet")

	// Create a ValueRange to hold the report data
	vr := &ValueRange{}
	headers := &[]string{"source", "target", "policyType", "policyApplication", "policySchema", "value"}
	vr.Values = append(vr.Values, *headers)

	customer, _ := adminClient.Admin().MyCustomer()
	sourcePolicies, err := c.ResolvePolicySchemas(customer, target)
	if err != nil {
		return nil, err
	}

	// directUserPolicies
	for _, policy := range *sourcePolicies.Users.Direct {
		pvv, _ := starstruct.PrettyJSON(policy.Value.Value)
		vr.Values = append(vr.Values, []string{policy.SourceKey.TargetResource, policy.TargetKey.TargetResource, "User & Browser Settings", "Locally Applied", policy.Value.PolicySchema, pvv})
	}
	// inheritedUserPolicies
	for _, policy := range *sourcePolicies.Users.Inherited {
		pvv, _ := starstruct.PrettyJSON(policy.Value.Value)
		vr.Values = append(vr.Values, []string{policy.SourceKey.TargetResource, policy.TargetKey.TargetResource, "User & Browser Settings", "Inherited", policy.Value.PolicySchema, pvv})
	}
	// directDevicePolicies
	for _, policy := range *sourcePolicies.Devices.Direct {
		pvv, _ := starstruct.PrettyJSON(policy.Value.Value)
		vr.Values = append(vr.Values, []string{policy.SourceKey.TargetResource, policy.TargetKey.TargetResource, "Device Settings", "Locally Applied", policy.Value.PolicySchema, pvv})
	}
	// inheritedDevicePolicies
	for _, policy := range *sourcePolicies.Devices.Inherited {
		pvv, _ := starstruct.PrettyJSON(policy.Value.Value)
		vr.Values = append(vr.Values, []string{policy.SourceKey.TargetResource, policy.TargetKey.TargetResource, "Device Settings", "Inherited", policy.Value.PolicySchema, pvv})
	}

	newSpreadsheet := &Spreadsheet{
		Properties: &SpreadsheetProperties{
			Title: fmt.Sprintf("{Google} Chrome Policy Report [%s] %s", target.ThisName(), time.Now().Format("2006-01-02")),
		},
		Sheets: []Sheet{
			{
				Properties: &SheetProperties{
					Title: target.ThisID(),
				},
			},
		},
	}
	sheet, err := adminClient.Sheets().CreateSpreadsheet(newSpreadsheet)
	if err != nil {
		return nil, err
	}

	err = adminClient.Sheets().SaveToSheet(vr.Values, sheet.SpreadsheetID, sheet.Sheets[0].Properties.Title, nil)
	if err != nil {
		// Do something
	}
	return nil, nil
}
