/*
# Okta Applications

This package contains all the methods to interact with the Okta Applications API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/#tag/Application

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/applications.go
package okta

import (
	"encoding/json"
	"fmt"
)

/*
 * Query parameters for Applications
 */
type AppQuery struct {
	Q                 string // Searches the records for matching value
	After             string // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the `Link` response header.
	Limit             string // Default: -1. Specifies the number of results for a page
	Filter            string // Filters apps by `status`, `user.id`, `group.id` or `credentials.signing.kid`` expression
	Search            string // A SCIM filter expression that filters the results. Searches include all Device profile properties and the Device `id``, `status``, and `lastUpdated`` properties.
	Expand            string // Traverses users link relationship and optionally embeds Application User resource
	IncludeNonDeleted bool   // Default: false.
}

/*
 * # List All Applications
 * Lists all applications with pagination. A subset of apps can be returned that match a supported filter expression or query.
 * /api/v1/apps
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/#tag/Application/operation/listApplications
 */
func (c *Client) ListAllApplications() (*Applications, error) {

	allApps := Applications{}

	q := AppQuery{
		IncludeNonDeleted: false,
	}

	url := c.BuildURL(OktaApps)
	res, err := c.HTTPClient.PaginatedRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		app := Application{}
		err := json.Unmarshal(r, &app)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user: %w", err)
		}
		allApps = append(allApps, app)
	}

	return &allApps, nil
}
