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
	"time"
)

type Applications []Application

type Application struct {
	Accessibility Accessibility      `json:"accessibility,omitempty"`
	Created       time.Time          `json:"created,omitempty"`
	Features      []string           `json:"features,omitempty"`
	ID            string             `json:"id,omitempty"`
	Label         string             `json:"label,omitempty"`
	LastUpdated   time.Time          `json:"lastUpdated,omitempty"`
	Licensing     Licensing          `json:"licensing,omitempty"`
	Profile       ApplicationProfile `json:"profile,omitempty"`
	SignOnMode    string             `json:"signOnMode,omitempty"`
	Status        string             `json:"status,omitempty"`
	Visibility    Visibility         `json:"visibility,omitempty"`
	Embedded      Embedded           `json:"_embedded,omitempty"`
	Links         Links              `json:"_links,omitempty"`
}

type Accessibility struct {
	ErrorRedirectURL  string `json:"errorRedirectUrl,omitempty"`
	LoginRedirectURL  string `json:"loginRedirectUrl,omitempty"`
	SelfService       bool   `json:"selfService,omitempty"`
	LoginRedirectURL2 string `json:"loginRedirectUrl2,omitempty"`
}

type Licensing struct {
	SeatCount int `json:"seatCount,omitempty"`
}

type ApplicationProfile struct {
	Property1 map[string]interface{} `json:"property1,omitempty"`
	Property2 map[string]interface{} `json:"property2,omitempty"`
}

type Visibility struct {
	AppLinks          map[string]bool `json:"appLinks,omitempty"`
	AutoLaunch        bool            `json:"autoLaunch,omitempty"`
	AutoSubmitToolbar bool            `json:"autoSubmitToolbar,omitempty"`
	Hide              map[string]bool `json:"hide,omitempty"`
}

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
