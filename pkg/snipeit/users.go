/*
# SnipeIT - Users

This package initializes all the methods for functions which interact with the SnipeIT Users endpoints:
https://snipe-it.readme.io/reference/users

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/users.go
package snipeit

import (
	"time"
)

// UserClient for chaining methods
type UserClient struct {
	*Client
}

// Entry point for asset-related operations
func (c *Client) Users() *UserClient {
	uc := &UserClient{
		Client: c,
	}

	return uc
}

/*
 * Query Parameters for Users
 */
type UserQuery struct {
	Search           string     `url:"search,omitempty"`            // Search for an user by string
	Limit            int        `url:"limit,omitempty"`             // Specify the number of results you wish to return. Defaults to 50
	Offset           int        `url:"offset,omitempty"`            // Specify the number of results to skip before starting to return items. Defaults to 0
	Sort             string     `url:"sort,omitempty"`              // Sort the results by the specified column. Defaults to "created_at"
	Order            string     `url:"order,omitempty"`             // Sort the results in the specified order ("asc" or "desc"). Defaults to "desc"
	FirstName        string     `url:"first_name,omitempty"`        // First name of the user
	LastName         string     `url:"last_name,omitempty"`         // Last name of the user
	Username         string     `url:"username,omitempty"`          // Username of the user
	Email            string     `url:"email,omitempty"`             // Email address of the user
	EmployeeNumber   string     `url:"employee_numb,omitempty"`     // Employee number of the user
	State            string     `url:"state,omitempty"`             // State of the user (active, inactive, etc.)
	Zip              string     `url:"zip,omitempty"`               // Zip code of the user
	Country          string     `url:"country,omitempty"`           // Country of the user
	GroupID          uint32     `url:"group_id,omitempty"`          // ID of the group the user belongs to
	DepartmentID     uint32     `url:"department_id,omitempty"`     // ID of the department the user belongs to
	CompanyID        uint32     `url:"company_id,omitempty"`        // ID of the company the user belongs to
	LocationID       uint32     `url:"location_id,omitempty"`       // ID of the location the user belongs to
	Deleted          bool       `url:"deleted,omitempty"`           // Set to true to return deleted users
	All              bool       `url:"all,omitempty"`               // Set to true to return all users, including deleted ones
	LDAPImport       bool       `url:"ldap_import,omitempty"`       // Set to true to return only users imported from LDAP
	AssetsCount      uint32     `url:"assets_count,omitempty"`      // Set to true to return only users with assets
	LicensesCount    uint32     `url:"licenses_count,omitempty"`    // Number of checked out licenses for the user
	AccessoriesCount uint32     `url:"accessories_count,omitempty"` // Number of checked out accessories for the user
	ConsumablesCount uint32     `url:"consumables_count,omitempty"` // Number of checked out consumables for the user
	Remote           bool       `url:"remote,omitempty"`            // Whether the user is marked as a remote worker or not (should be 0 or 1)
	VIP              bool       `url:"vip,omitempty"`               // Whether the user is marked as a VIP or not (should be 0 or 1)
	StartDate        *Timestamp `url:"start_date,omitempty"`        // Start date for the user
	EndDate          *Timestamp `url:"end_date,omitempty"`          // End date for the user
}

// ### UserQuery implements QueryInterface
// ---------------------------------------------------------------------
func (q *UserQuery) Copy() QueryInterface {
	return &UserQuery{
		Search:           q.Search,
		Limit:            q.Limit,
		Offset:           q.Offset,
		Sort:             q.Sort,
		Order:            q.Order,
		FirstName:        q.FirstName,
		LastName:         q.LastName,
		Username:         q.Username,
		Email:            q.Email,
		EmployeeNumber:   q.EmployeeNumber,
		State:            q.State,
		Zip:              q.Zip,
		Country:          q.Country,
		GroupID:          q.GroupID,
		DepartmentID:     q.DepartmentID,
		CompanyID:        q.CompanyID,
		LocationID:       q.LocationID,
		Deleted:          q.Deleted,
		All:              q.All,
		LDAPImport:       q.LDAPImport,
		AssetsCount:      q.AssetsCount,
		LicensesCount:    q.LicensesCount,
		AccessoriesCount: q.AccessoriesCount,
		ConsumablesCount: q.ConsumablesCount,
		Remote:           q.Remote,
		VIP:              q.VIP,
		StartDate:        q.StartDate,
		EndDate:          q.EndDate,
	}
}

func (q *UserQuery) GetLimit() int {
	return q.Limit
}

func (q *UserQuery) SetLimit(limit int) {
	q.Limit = limit
}

func (q *UserQuery) GetOffset() int {
	return q.Offset
}

func (q *UserQuery) SetOffset(offset int) {
	q.Offset = offset
}

// END OF QUERYINTERFACE METHODS
//---------------------------------------------------------------------

/*
 * List all Users in Snipe-IT
 * /api/v1/users
 * - https://snipe-it.readme.io/reference/users
 */
func (c *UserClient) GetAllUsers() (*UserList, error) {
	url := c.BuildURL(Users)

	q := UserQuery{
		Limit:  500,
		Offset: 0,
	}

	var cache UserList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	users, err := doConcurrent[UserList](c.Client, "GET", url, &q, nil)
	if err != nil {
		c.Log.Fatalf("Error fetching license list: %v", err)
	}

	c.SetCache(url, users, 5*time.Minute)
	return users, nil
}

/*
 * # Get an user in Snipe-IT by ID
 * /api/v1/users/{id}
 * - https://snipe-it.readme.io/reference/usersid
 */
func (c *UserClient) GetUser(id uint32) (*User[UserGET], error) {
	url := c.BuildURL(Users, id)

	var cache User[UserGET]
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	user, err := do[User[UserGET]](c.Client, "GET", url, nil, nil)
	if err != nil {
		c.Log.Fatalf("Error getting user: %v", err)
	}

	c.SetCache(url, user, 5*time.Minute)
	return &user, nil
}
