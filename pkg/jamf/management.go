/*
# Jamf - Management

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/devices.go
package jamf

import (
	"encoding/json"
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
	mr := &ManagementResponse{}

	url := c.BuildURL(RenewProfile)

	payload := map[string][]string{
		"udids": udids,
	}

	res, body, err := c.HTTP.DoRequest("POST", url, nil, payload)
	if err != nil {
		return nil, err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debugf(string(body))

	err = json.Unmarshal(body, &mr)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
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

	res, body, err := c.HTTP.DoRequest("POST", url, nil, nil)
	if err != nil {
		return "", err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debugf(string(body))

	return string(body), nil
}
