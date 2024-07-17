/*
# Okta User Factors

This package contains all the methods to interact with the Okta Users API for Factors:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/UserFactor/

:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/user_factors.go
package okta

import (
	"time"
)

// FactorsClient for chaining methods
type FactorsClient struct {
	*Client
}

// Entry point for user-related operations
func (c *Client) Factors() *FactorsClient {
	f := &FactorsClient{
		Client: c,
	}

	return f
}

/*
 * Query Parameters for User Factors
 */
type UserFactorQuery struct {
	Activate                 bool   `url:"activate,omitempty"`                // If true, the `sms`` Factor is immediately activated as part of the enrollment. An activation text message isn't sent to the device.
	RemoveRecoveryEnrollment bool   `url:"removeRevokedEnrollment,omitempty"` // If true, revoked factors are removed from the user's factors list.
	TemplateID               string `url:"templateId,omitempty"`              // ID of an existing custom SMS template. Only applicable for SMS factors.
	TokenLifetime            int    `url:"tokenLifetime,omitempty"`           // Default: 300. The number of seconds before the token expires. Defaults to 3600 (1 hour).
	UpdatePhone              bool   `url:"updatePhone,omitempty"`             // If true, indicates you are replacing the currently registered phone number for the specified user. This parameter is ignored if the existing phone number is used by an activated Factor.
}

/*
 * # List all Enrolled Factors for a User
 * /api/v1/users/{userId}/factors
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/UserFactor/#tag/UserFactor/operation/listFactors
 */
func (c *FactorsClient) ListAllEnrolledFactors(userID string) (*Factors, error) {
	url := c.BuildURL(OktaUsers, userID, "factors")

	var cache Factors
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	factors, err := do[Factors](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, factors, 5*time.Minute)
	return &factors, nil
}

/*
 * # Enroll a Factor for a User
 * /api/v1/users/{userId}/factors
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/UserFactor/#tag/UserFactor/operation/enrollFactor
 */
func (c *FactorsClient) EnrollFactor(userID string, factor Factor) (*Factor, error) {
	url := c.BuildURL(OktaUsers, userID, "factors")

	factor, err := do[Factor](c.Client, "POST", url, nil, &factor)
	if err != nil {
		return nil, err
	}

	return &factor, nil
}

/*
* # List all supported Factors that can be enrolled for a User
* /api/v1/users/{userId}/factors/catalog
* - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/UserFactor/#tag/UserFactor/operation/listSupportedFactors
 */
func (c *FactorsClient) ListSupportedFactors(userID string) (*Factors, error) {
	url := c.BuildURL(OktaUsers, userID, "factors", "catalog")

	var cache Factors
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	factors, err := do[Factors](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, factors, 5*time.Minute)
	return &factors, nil
}

/*
 * # Reset all Factors
 * /api/v1/users/{userId}/lifecycle/resetFactors
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/resetFactors
 */
func (c *FactorsClient) ResetFactors(userID string) error {
	url := c.BuildURL(OktaUsers, userID, "lifecycle", "resetFactors")

	_, err := do[Factors](c.Client, "POST", url, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
