/*
# Atlassian

This package initializes all the methods for functions which interact with the Organizations REST APIs:

* Organizations REST API
- https://developer.atlassian.com/cloud/admin/organization/rest/intro/#uri

:Copyright: (c) 2025 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/atlassian/organizations.go
package atlassian

import (
	"fmt"
)

var (
	Organizations           = fmt.Sprintf("%s/orgs", fmt.Sprintf(V1, CloudAdmin)) // https://developer.atlassian.com/cloud/admin/organization/rest/api-group-orgs/#api-group-orgs
	OrganizationsUsers      = fmt.Sprintf("%s/%s/users", Organizations, "%s")     // https://developer.atlassian.com/cloud/admin/organization/rest/api-group-users/#api-group-users
	OrganizationsGroups     = fmt.Sprintf("%s/%s/groups", Organizations, "%s")    // https://developer.atlassian.com/cloud/admin/organization/rest/api-group-groups/#api-group-groups
	OrganizationsDomains    = fmt.Sprintf("%s/%s/domains", Organizations, "%s")   // https://developer.atlassian.com/cloud/admin/organization/rest/api-group-domains/#api-group-domains
	OrganizationsEvents     = fmt.Sprintf("%s/%s/events", Organizations, "%s")    // https://developer.atlassian.com/cloud/admin/organization/rest/api-group-events/#api-group-events
	OrganizationsPolicies   = fmt.Sprintf("%s/%s/policies", Organizations, "%s")  // https://developer.atlassian.com/cloud/admin/organization/rest/api-group-policies/#api-group-policies
	OrganizationsDirectory  = fmt.Sprintf("%s/%s/directory", Organizations, "%s") // https://developer.atlassian.com/cloud/admin/organization/rest/api-group-directory/#api-group-directory
	OrganizationsWorkspaces = fmt.Sprintf("%s/%s/users", Organizations, "%s")     // https://developer.atlassian.com/cloud/admin/organization/rest/api-group-workspaces/#api-group-workspaces
)

// OrgClient for chaining methods
type OrganizationsClient struct {
	*Client
}

// Entry point for user-related operations
func (c *Client) Organizations() *OrganizationsClient {
	oc := &OrganizationsClient{
		Client: c,
	}

	return oc
}

/*
 * # Get all users, regardless of status
 * /api/v1/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers
 */
