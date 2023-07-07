/*
# Okta

This package initializes all the methods for functions which interact with the Okta API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/okta.go
package okta

import (
	"fmt"
	"strings"

	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	BaseURL = fmt.Sprintf("https://%s.%s.com/api/v1", "%s", "%s") // https://developer.okta.com/docs/api/#versioning
)

const (
	OktaApps    = "%s/apps"      // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/
	OktaDevices = "%s/devices"   // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/
	OktaUsers   = "%s/users"     // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/
	OktaIAM     = "%s/iam"       // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/
	OktaRoles   = "%s/iam/roles" // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/
)

type Client struct {
	BaseURL    string           // BaseURL is the base URL for Okta API requests.
	HTTPClient *requests.Client // HTTPClient is the client used to make HTTP requests.
	Error      *Error           // Error is the error response from the last request made by the client.
	Logger     *log.Logger      // Logger is the logger used to log messages.
}

type Embedded struct {
	Property1 map[string]interface{} `json:"property1,omitempty"` // Property1 is a map of string to interface.
	Property2 map[string]interface{} `json:"property2,omitempty"` // Property2 is a map of string to interface.
}

type Error struct {
	ErrorCauses  []ErrorCause `json:"errorCauses,omitempty"`
	ErrorCode    string       `json:"errorCode,omitempty"`
	ErrorId      string       `json:"errorId,omitempty"`
	ErrorLink    string       `json:"errorLink,omitempty"`
	ErrorSummary string       `json:"errorSummary,omitempty"`
}

type ErrorCause struct {
	ErrorSummary string `json:"errorSummary,omitempty"`
}

type Links struct {
	AccessPolicy           Link   `json:"accessPolicy,omitempty"`           // AccessPolicy is a link to the access policy.
	Activate               Link   `json:"activate,omitempty"`               // Activate is a link to activate the user.
	ChangePassword         Link   `json:"changePassword,omitempty"`         // ChangePassword is a link to change the user's password.
	ChangeRecoveryQuestion Link   `json:"changeRecoveryQuestion,omitempty"` // ChangeRecoveryQuestion is a link to change the user's recovery question.
	Deactivate             Link   `json:"deactivate,omitempty"`             // Deactivate is a link to deactivate the user.
	ExpirePassword         Link   `json:"expirePassword,omitempty"`         // ExpirePassword is a link to expire the user's password.
	ForgotPassword         Link   `json:"forgotPassword,omitempty"`         // ForgotPassword is a link to reset the user's password.
	Groups                 Link   `json:"groups,omitempty"`                 // Groups is a link to the user's groups.
	Logo                   []Link `json:"logo,omitempty"`                   // Logo is a list of links to the logo.
	Metadata               Link   `json:"metadata,omitempty"`               // Metadata is a link to the user's metadata.
	ResetFactors           Link   `json:"resetFactors,omitempty"`           // ResetFactors is a link to reset the user's factors.
	ResetPassword          Link   `json:"resetPassword,omitempty"`          // ResetPassword is a link to reset the user's password.
	Schema                 Link   `json:"schema,omitempty"`                 // Schema is a link to the user's schema.
	Self                   Link   `json:"self,omitempty"`                   // Self is a link to the user.
	Suspend                Link   `json:"suspend,omitempty"`                // Suspend is a link to suspend the user.
	Users                  Link   `json:"users,omitempty"`                  // Users is a link to the user's users.
}

type Link struct {
	Hints  Hints  `json:"hints,omitempty"`  // Hints is a list of hints for the link.
	Href   string `json:"href,omitempty"`   // Href is the URL for the link.
	Method string `json:"method,omitempty"` // Method is the HTTP method for the link.
	Type   string `json:"type,omitempty"`   // Type is the type of link.
}

type Hints struct {
	Allow []string `json:"allow,omitempty"` // Allow is a list of allowed methods.
}

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%s", url, id)
	}
	return url
}

// NewClient returns a new Okta API client.
func NewClient(verbosity int) *Client {

	// org_name := config.GetEnv("OKTA_ORG_NAME", "yourOktaDomain")
	org_name := config.GetEnv("OKTA_SANDBOX_ORG_NAME", "yourOktaDomain")
	org_name = strings.TrimPrefix(org_name, "https://")
	org_name = strings.TrimPrefix(org_name, "http://")
	org_name = strings.TrimSuffix(org_name, ".okta.com")

	// url := config.GetEnv("OKTA_BASE_URL", "okta.com")
	base := config.GetEnv("OKTA_SANDBOX_BASE_URL", "oktapreview.com")
	base = strings.Trim(base, "./")
	base = strings.TrimSuffix(base, ".com")

	// token := config.GetEnv("OKTA_API_TOKEN", "oktaApiKey")
	token := config.GetEnv("OKTA_SANDBOX_API_TOKEN", "oktaApiKey")
	BaseURL := fmt.Sprintf(BaseURL, org_name, base)

	headers := requests.Headers{
		"Authorization": "SSWS " + token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	return &Client{
		BaseURL:    BaseURL,
		HTTPClient: requests.NewClient(nil, headers),
		Logger:     log.NewLogger("{okta}", verbosity),
	}
}
