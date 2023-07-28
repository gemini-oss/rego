/*
# Google Workspace - Admin

This package implements logic related to the `Users` resource of the Google Admin SDK API:
https://developers.google.com/admin-sdk/directory/reference/rest/v1/users

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/users.go
package google

import (
	"encoding/json"
	"fmt"
)

/*
 * Query Parameters for Drive Files
 * Reference: https://developers.google.com/admin-sdk/directory/reference/rest/v1/users/list#query-parameters
 */
type UserQuery struct {
	CustomFieldMask string         `url:"customFieldMask,omitempty"` // A comma-separated list of schema names. All fields from these schemas are fetched. This should only be set when projection=custom.
	Customer        string         `url:"customer,omitempty"`        // The unique ID for the customer's Google Workspace account. In case of a multi-domain account, to fetch all groups for a customer, use this field instead of domain. You can also use the my_customer alias to represent your account's customerId. The customerId is also returned as part of the Users resource. You must provide either the customer or the domain parameter.
	Domain          string         `url:"domain,omitempty"`          // The domain name. Use this field to get groups from only one domain. To return all domains for a customer account, use the customer query parameter instead. Either the customer or the domain parameter must be provided.
	Event           UserEvent      `url:"event,omitempty"`           // Event on which subscription is intended (if subscribing)
	MaxResults      int            `url:"maxResults,omitempty"`      // Maximum number of results to return. Default: 100. Max 500. https://developers.google.com/admin-sdk/directory/v1/limits#api-limits-and-quotas
	OrderBy         OrderBy        `url:"orderBy,omitempty"`         // Property to use for sorting results.
	PageToken       string         `url:"pageToken,omitempty"`       // Token to specify next page in the list
	Projection      UserProjection `url:"projection,omitempty"`      // What subset of fields to fetch for this user.
	Query           string         `url:"query,omitempty"`           // Query string for searching user fields. For more information on constructing user queries, see [Search for Users](https://developers.google.com/admin-sdk/directory/v1/guides/search-users).
	ShowDeleted     string         `url:"showDeleted,omitempty"`     // If set to true, retrieves the list of deleted users. (Default: false)
	SortOrder       SortOrder      `url:"sortOrder,omitempty"`       // Whether to return results in ascending or descending order, ignoring case.
	ViewType        UserViewType   `url:"viewType,omitempty"`        // Whether to fetch the administrator-only or domain-wide public view of the user. For more information, see Retrieve a user as a non-administrator.
}

/*
 * Check if the UserQuery is empty
 */
func (u *UserQuery) IsEmpty() bool {
	return u.CustomFieldMask == "" &&
		u.Customer == "" &&
		u.Domain == "" &&
		u.Event == "" &&
		u.MaxResults == 0 &&
		u.OrderBy == "" &&
		u.PageToken == "" &&
		u.Projection == "" &&
		u.Query == "" &&
		u.ShowDeleted == "" &&
		u.SortOrder == "" &&
		u.ViewType == ""
}

/*
 * Validate the query parameters for the Users resource
 */
func (u *UserQuery) ValidateQuery() error {
	if u.IsEmpty() {
		u.Customer = "my_customer"
		u.MaxResults = 100
		return nil
	}

	if u.Customer != "" && u.Domain != "" {
		return fmt.Errorf("cannot specify both customer and domain")
	}

	if u.Customer == "" && u.Domain == "" {
		u.Customer = "my_customer"
	}

	if u.MaxResults > 500 {
		return fmt.Errorf("maxResults cannot exceed %d", 500)
	}

	if u.MaxResults == 0 {
		u.MaxResults = 100
	}

	return nil
}

/*
 * List all users
 * /admin/directory/v1/users
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/users/list
 */
func (c *Client) ListAllUsers() (*Users, error) {
	users := &Users{}

	q := UserQuery{}

	err := q.ValidateQuery()
	if err != nil {
		return nil, err
	}
	q.MaxResults = 500
	q.Projection = Basic

	url := DirectoryUsers
	c.Logger.Debug("url:", url)

	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status: ", res.Status)
	c.Logger.Debug("Response Body: ", string(body))

	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	for users.NextPageToken != "" {
		usersPage := &Users{}
		q = UserQuery{
			Customer:   q.Customer,
			Domain:     q.Domain,
			MaxResults: 500,
			PageToken:  users.NextPageToken,
		}

		res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
		if err != nil {
			return nil, err
		}
		c.Logger.Println("Response Status: ", res.Status)
		c.Logger.Debug("Response Body: ", string(body))

		err = json.Unmarshal(body, &usersPage)
		if err != nil {
			return nil, err
		}
		users.Users = append(users.Users, usersPage.Users...)
		users.NextPageToken = usersPage.NextPageToken
	}

	return users, nil
}

/*
 * Search for users based on filter conditions
 * /admin/directory/v1/users
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/users/list
 */
func (c *Client) SearchUsers(q UserQuery) (*Users, error) {
	users := &Users{}

	err := q.ValidateQuery()
	if err != nil {
		return nil, err
	}

	url := DirectoryUsers
	c.Logger.Debug("url:", url)

	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status: ", res.Status)
	c.Logger.Debug("Response Body: ", string(body))

	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

/*
 * Retrieves a User's Profile
 * /admin/directory/v1/users/{userKey}
 * https://developers.google.com/admin-sdk/directory/v1/reference/users/get
 */
func (c *Client) GetUser(userKey string) (*User, error) {
	url := fmt.Sprintf(DirectoryUsers+"/%s", userKey)
	c.Logger.Debug("url:", url)

	res, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status: ", res.Status)
	c.Logger.Debug("Response Body: ", string(body))

	user := &User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
