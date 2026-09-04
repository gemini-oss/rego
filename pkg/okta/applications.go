/*
# Okta Applications

This package contains all the methods to interact with the Okta Applications API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/#tag/Application

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/applications.go
package okta

import (
	"fmt"
	"time"

	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ApplicationsClient for chaining methods
type ApplicationsClient struct {
	*Client
}

// Entry point for application-related operations
func (c *Client) Applications() *ApplicationsClient {
	// Return cached client if it exists
	if c.applicationsClient != nil {
		return c.applicationsClient
	}

	// Shallow copy a new client with Applications-specific rate limiter
	// https://developer.okta.com/docs/reference/rl-best-practices/
	applicationsRL := ratelimit.NewRateLimiter()
	applicationsRL.ResetHeaders = true
	applicationsRL.Log.Verbosity = c.Log.Verbosity

	applicationsHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), applicationsRL)
	applicationsHTTP.BodyType = c.HTTP.BodyType

	applicationsClient := &Client{
		BaseURL: c.BaseURL,
		HTTP:    applicationsHTTP,
		Error:   c.Error,
		Log:     c.Log,
		Cache:   c.Cache,
	}

	c.applicationsClient = &ApplicationsClient{
		Client: applicationsClient,
	}

	return c.applicationsClient
}

/*
 * Query parameters for Applications
 */
type AppQuery struct {
	Q                 string // Searches the records for matching value
	After             string // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the "Link" response header.
	Limit             string // Default: -1. Specifies the number of results for a page
	Filter            string // Filters apps by "status", "user.id", "group.id" or "credentials.signing.kid" expression
	Expand            string // Traverses users link relationship and optionally embeds Application User resource
	IncludeNonDeleted bool   // Default: false.
}

/*
 * # List All Applications
 * Lists all applications with pagination. A subset of apps can be returned that match a supported filter expression or query.
 * /api/v1/apps
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/#tag/Application/operation/listApplications
 */
func (c *ApplicationsClient) ListAllApplications() (*Applications, error) {
	url := c.BuildURL(OktaApps)

	var cache Applications
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := AppQuery{
		IncludeNonDeleted: false,
	}

	applications, err := doPaginated[Applications](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, applications, 5*time.Minute)
	return applications, nil
}

/*
 * # Get Application
 * Retrieves an application from your Okta organization by id.
 * /api/v1/apps/{appId}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/#tag/Application/operation/listApplications
 */
func (c *ApplicationsClient) GetApplication(appID string) (*Application, error) {
	url := c.BuildURL(OktaApps, appID)

	var cache Application
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := AppQuery{
		IncludeNonDeleted: false,
	}

	application, err := do[*Application](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, application, 5*time.Minute)
	return application, nil
}

/*
 * # List all Application Users
 * Retrieves all users assigned to an application
 * /api/v1/apps/{appid}/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/ApplicationUsers/
 */
func (c *ApplicationsClient) ListAllApplicationUsers(appID string) (*Users, error) {
	url := c.BuildURL(OktaApps, appID, "users")

	var cache Users
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := AppQuery{
		Limit:  "500",
		Expand: "user",
	}

	appUsers, err := doPaginated[Users](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, appUsers, 5*time.Minute)
	return appUsers, nil
}

/*
 * # Get Application User
 * Retrieves a single user assigned to an application
 * /api/v1/apps/{appid}/users/{userid}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/ApplicationUsers/#tag/ApplicationUsers/operation/getApplicationUser
 */
func (c *ApplicationsClient) GetApplicationUser(appID string, userID string) (*User, error) {
	url := c.BuildURL(OktaApps, appID, "users", userID)

	var cache User
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := AppQuery{
		Expand: "user",
	}

	user, err := do[User](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, user, 5*time.Minute)
	return &user, nil
}

/*
 * Get all applications assigned to a user
 * /api/v1/apps
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/#tag/Application/operation/listApplications
 */
func (c *ApplicationsClient) GetUserApplications(userID string) (*Applications, error) {
	url := c.BuildURL(OktaApps)

	var cache Applications
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := AppQuery{
		Filter: fmt.Sprintf("user.id eq \"%s\"", userID),
		Expand: fmt.Sprintf("user/%s", userID),
	}

	apps, err := doPaginated[Applications](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, apps, 5*time.Minute)
	return apps, nil
}

/*
 * # Convert Application Assignment
 * Retrieves a user assigned to an application and converts the scope to the opposite of its current value
 * /api/v1/apps/{appid}/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/ApplicationUsers/#tag/ApplicationUsers/operation/assignUserToApplication
 */
func (c *ApplicationsClient) ConvertApplicationAssignment(appID string, userID string) (*User, error) {
	url := c.BuildURL(OktaApps, appID, "users")

	// Get the user assigned to the application to determine the scope
	user, err := c.GetApplicationUser(appID, userID)
	if err != nil {
		return nil, err
	}

	// Switch the scope
	scopeSwitch := map[string]string{
		"GROUP": "USER",
		"USER":  "GROUP",
	}

	// Update the user's scope
	payload := &map[string]string{
		"id":    user.ID,
		"scope": scopeSwitch[user.Scope],
	}

	user, err = do[*User](c.Client, "POST", url, nil, payload)
	if err != nil {
		return nil, err
	}

	// Cache the changed user
	c.SetCache(c.BuildURL(OktaApps, appID, "users", userID), user, 5*time.Minute)
	return user, nil
}

/*
 * # Remove Application Assignment
 * Retrieves a user assigned to an application and removes the assignment
 * /api/v1/apps/{appid}/users/{userid}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/ApplicationUsers/#tag/ApplicationUsers/operation/unassignUserFromApplication
 */
func (c *ApplicationsClient) RemoveApplicationAssignment(appID string, userID string) error {
	url := c.BuildURL(OktaApps, appID, "users", userID)

	_, err := do[any](c.Client, "DELETE", url, nil, nil)
	if err != nil {
		// If error message is not: "unexpected end of JSON input"
		// if fmt.Errorf("unexpected end of JSON input").Error() != err.Error() {
		// 	return err
		// }
		return nil
	}

	return nil
}
