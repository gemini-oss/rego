/*
# Okta Attributes

This package contains all the methods to interact with the Okta Attributes API:
https://developer.okta.com/docs/api/openapi/asa/asa/tag/attributes/

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/attributes.go
package okta

// AttributesClient for chaining methods
type AttributesClient struct {
	*Client
}

// Entry point for application-related operations
func (c *Client) Attributes() *ApplicationsClient {
	ac := &ApplicationsClient{
		Client: c,
	}

	return ac
}

/*
 * # Update a User's Attribute
 * /api/v1/teams/{team_name}/users/{user_name}/attributes/{attribute_id}
 * - https://developer.okta.com/docs/api/openapi/asa/asa/tag/attributes/#tag/attributes/operation/UpdateUserAttribute
 */
