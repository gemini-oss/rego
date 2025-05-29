/*
# Jamf - Management

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/devices.go
package jamf

import (
	"fmt"
)

var (
	ManagementFramework = fmt.Sprintf("%s/jamf-management-framework", V1) // /api/v1/jamf-management-framework
	V1_MDM              = fmt.Sprintf("%s/mdm", V1)                       // /api/v1/mdm
	RenewProfile        = fmt.Sprintf("%s/renew-profile", V1_MDM)         // /api/v1/mdm/renew-profile
	V2_MDM              = fmt.Sprintf("%s/mdm", V2)                       // /api/v2/mdm
)

/*
 * # Renew MDM Profile
 * /api/v1/mdm/renew-profile
 * - https://developer.jamf.com/jamf-pro/reference/post_v1-mdm-renew-profile
 */
func (c *Client) RenewMDMProfile(udids []string) (*ManagementResponse, error) {
	url := c.BuildURL(RenewProfile)

	payload := map[string][]string{
		"udids": udids,
	}

	mr, err := do[*ManagementResponse](c, "POST", url, nil, payload)
	if err != nil {
		return nil, err
	}

	return mr, nil
}

/*
 * # Repair Jamf Management Framework
 * /api/v1/jamf-management-framework/redeploy/{id}
 * - https://developer.jamf.com/jamf-pro/reference/post_v1-jamf-management-framework-redeploy-id
 */
func (c *Client) RepairManagementFramework(id string) (string, error) {
	url := c.BuildURL(fmt.Sprintf("%s/redeploy/%s", ManagementFramework, id))

	mf, err := do[interface{}](c, "POST", url, nil, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", mf), nil
}
