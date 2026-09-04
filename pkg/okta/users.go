/*
# Okta Users

This package contains all the methods to interact with the Okta Users API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/users.go
package okta

import (
	"time"

	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// UsersClient for chaining methods
type UsersClient struct {
	*Client
}

// Entry point for user-related operations
func (c *Client) Users() *UsersClient {
	// Return cached client if it exists
	if c.usersClient != nil {
		return c.usersClient
	}

	// Shallow copy a new client with Users-specific rate limiter
	// https://developer.okta.com/docs/reference/rl-best-practices/
	usersRL := ratelimit.NewRateLimiter()
	usersRL.ResetHeaders = true
	usersRL.Log.Verbosity = c.Log.Verbosity

	usersHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), usersRL)
	usersHTTP.BodyType = c.HTTP.BodyType

	usersClient := &Client{
		BaseURL: c.BaseURL,
		HTTP:    usersHTTP,
		Error:   c.Error,
		Log:     c.Log,
		Cache:   c.Cache,
	}

	c.usersClient = &UsersClient{
		Client: usersClient,
	}

	return c.usersClient
}

/*
 * Query Parameters for Users
 */
type UserQuery struct {
	Q         string // Searches the records for matching value
	After     string // The cursor to use for pagination. It is an opaque string that specifies your current location in the list and is obtained from the `Link` response header.
	Limit     string // Default: 200. Specifies the number of results returned. Defaults to 10 if `q` is provided
	Filter    string // Filters users with a supported expression for a subset of properties
	Search    string // A SCIM filter expression for most properties. Okta recommends using this parameter for search for best performance
	SortBy    string // Specifies the attribute by which to sort the results. Valid values are `id`, `created`, `activated`, `status`, and `lastUpdated`. The default is `id`
	SoftOrder string // Sorting is done in ASCII sort order (that is, by ASCII character value), but isn't case sensitive
	Expand    string `url:"expand,omitempty"` // An optional parameter to return embedded Groups in the `_embedded` property. Valid value: `groups`
}

/*
 * # Get all users, regardless of status
 * /api/v1/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers
 */
func (c *UsersClient) ListAllUsers() (*Users, error) {
	url := c.BuildURL(OktaUsers)

	var cache Users
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := &UserQuery{
		Limit:  `200`,
		Search: `status eq "STAGED" or status eq "PROVISIONED" or status eq "ACTIVE" or status eq "RECOVERY" or status eq "LOCKED_OUT" or status eq "PASSWORD_EXPIRED" or status eq "SUSPENDED" or status eq "DEPROVISIONED"`,
	}

	users, err := doPaginated[Users](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, users, 30*time.Minute)
	return users, nil
}

/*
 * # List all ACTIVE users
 * /api/v1/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listUsers
 */
func (c *UsersClient) ListActiveUsers() (*Users, error) {
	url := c.BuildURL(OktaUsers)

	var cache Users
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := &UserQuery{
		Limit:  `200`,
		Search: `status eq "ACTIVE"`,
	}

	users, err := doPaginated[Users](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, users, 30*time.Minute)
	return users, nil
}

/*
 * # Get a user by ID
 * /api/v1/users/{userId}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/getUser
 */
func (c *UsersClient) GetUser(userID string) (*User, error) {
	url := c.BuildURL(OktaUsers, userID)

	var cache User
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	q := &UserQuery{
		Expand: "groups",
	}

	user, err := do[User](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, user, 5*time.Minute)
	return &user, nil
}

/*
 * # Update a user's properties by ID
 * /api/v1/users/{userId}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/updateUser
 */
func (c *UsersClient) UpdateUser(userID string, u *User) (*User, error) {

	url := c.BuildURL(OktaUsers, userID)

	user, err := do[User](c.Client, "POST", url, nil, &u)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

/*
 * # Deactivate a User
 * /api/v1/users/{userId}/lifecycle/deactivate
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/deactivateUser
 */
func (c *UsersClient) DeactivateUser(userID string) error {
	url := c.BuildURL(OktaUsers, userID, "lifecycle", "deactivate")

	_, err := do[any](c.Client, "POST", url, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

/*
 * # Suspend a User
 * Suspends a user. This operation can only be performed on users with an ACTIVE status.
 * The user's status changes to SUSPENDED when the process is complete.
 * /api/v1/users/{userId}/lifecycle/suspend
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/suspendUser
 */
func (c *UsersClient) SuspendUser(userID string) error {
	url := c.BuildURL(OktaUsers, userID, "lifecycle", "suspend")

	_, err := do[any](c.Client, "POST", url, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

/*
 * # Unsuspend a User
 * Unsuspends a user and returns them to the ACTIVE state.
 * This operation can only be performed on users with a SUSPENDED status.
 * /api/v1/users/{userId}/lifecycle/unsuspend
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/unsuspendUser
 */
func (c *UsersClient) UnsuspendUser(userID string) error {
	url := c.BuildURL(OktaUsers, userID, "lifecycle", "unsuspend")

	_, err := do[any](c.Client, "POST", url, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

/*
 * # Get all Assigned Application Links for a User
 * /api/v1/users/{userId}/appLinks
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User/operation/listAppLinks
 */
func (c *UsersClient) GetUserAppLinks(userID string) (*AppLinks, error) {
	url := c.BuildURL(OktaUsers, userID, "appLinks")

	var cache AppLinks
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	appLinks, err := do[AppLinks](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, appLinks, 5*time.Minute)
	return &appLinks, nil
}

/*
 * # List all Groups for a User
 * /api/v1/users/{userId}/groups
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/UserResources/#tag/UserResources/operation/listUserGroups
 */
func (c *UsersClient) GetUserGroups(userID string) (*Groups, error) {
	url := c.BuildURL(OktaUsers, userID, "groups")

	var cache Groups
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	groups, err := do[Groups](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, groups, 5*time.Minute)
	return &groups, nil
}

/*
 * # List Direct (non-rule) Groups for a User
 * Filters out groups managed by active group rules
 */
func (c *UsersClient) GetDirectUserGroups(userID string) (*Groups, error) {
	allGroups, err := c.GetUserGroups(userID)
	if err != nil {
		return nil, err
	}

	ruleManaged, err := c.Groups().getRuleManagedGroupIDs()
	if err != nil {
		return nil, err
	}

	direct := make(Groups, 0)
	for _, g := range *allGroups {
		if _, ok := ruleManaged[g.ID]; !ok {
			direct = append(direct, g)
		}
	}

	return &direct, nil
}

/*
 * # List all Devices for a User
 * /api/v1/users/{userId}/devices
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/UserResources/#tag/UserResources/operation/listUserDevices
 */
func (c *UsersClient) GetUserDevices(userID string) (*UserDevices, error) {
	url := c.BuildURL(OktaUsers, userID, "devices")

	var cache UserDevices
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	devices, err := do[UserDevices](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, devices, 5*time.Minute)
	return &devices, nil
}

/*
- # Revoke User Sessions
- /api/v1/users/{userId}/sessions
- https://developer.okta.com/docs/api/openapi/okta-management/management/tag/UserSessions/#tag/UserSessions/operation/revokeUserSessions
*/
func (c *UsersClient) RevokeUserSessions(userID string) error {
	url := c.BuildURL(OktaUsers, userID, "sessions")

	q := struct {
		OAuthTokens   bool `url:"oauthTokens"`
		ForgetDevices bool `url:"forgetDevices"`
	}{
		OAuthTokens:   true,
		ForgetDevices: true,
	}

	_, err := do[any](c.Client, "DELETE", url, q, nil)
	if err != nil {
		return err
	}

	return nil
}
