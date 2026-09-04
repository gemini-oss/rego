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

import (
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// AttributesClient for chaining methods
type AttributesClient struct {
	*Client
}

// Entry point for attribute-related operations
func (c *Client) Attributes() *AttributesClient {
	// Return cached client if it exists
	if c.attributesClient != nil {
		return c.attributesClient
	}

	// Shallow copy a new client with Attributes-specific rate limiter
	// https://developer.okta.com/docs/reference/rl-best-practices/
	attributesRL := ratelimit.NewRateLimiter()
	attributesRL.ResetHeaders = true
	attributesRL.Log.Verbosity = c.Log.Verbosity

	attributesHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), attributesRL)
	attributesHTTP.BodyType = c.HTTP.BodyType

	attributesClient := &Client{
		BaseURL: c.BaseURL,
		HTTP:    attributesHTTP,
		Error:   c.Error,
		Log:     c.Log,
		Cache:   c.Cache,
	}

	c.attributesClient = &AttributesClient{
		Client: attributesClient,
	}

	return c.attributesClient
}

/*
 * # Update a User's Attribute
 * /api/v1/teams/{team_name}/users/{user_name}/attributes/{attribute_id}
 * - https://developer.okta.com/docs/api/openapi/asa/asa/tag/attributes/#tag/attributes/operation/UpdateUserAttribute
 */
