/*
# Okta Roles

This package contains all the methods to interact with the Okta Roles API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/#tag/Role

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/roles.go
package okta

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Roles struct {
	Roles []Role `json:"roles,omitempty"`
}

type Role struct {
	AssignmentType string    `json:"assignmentType,omitempty"`
	Created        time.Time `json:"created,omitempty"`
	Description    string    `json:"description,omitempty"`
	ID             string    `json:"id,omitempty"`
	Label          string    `json:"label,omitempty"`
	LastUpdated    time.Time `json:"lastUpdated,omitempty"`
	Links          *Links    `json:"_links,omitempty"`
	Status         string    `json:"status,omitempty"`
	Type           string    `json:"type,omitempty"`
}

type Permission struct {
	Created     time.Time `json:"created,omitempty"`
	Label       string    `json:"label,omitempty"`
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
	Links       *Links    `json:"_links,omitempty"`
}

type RoleReport struct {
	Role  *Role
	Users []*User
}

/*
 * # Lists all roles with pagination support.
 * - By default, only custom roles can be listed from this endpoint
 * /api/v1/iam/roles
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/#tag/Role/operation/listRoles
 */
func (c *Client) ListAllRoles() (*Roles, error) {
	allRoles := &Roles{}

	url := c.BuildURL(OktaRoles)
	res, err := c.HTTPClient.PaginatedRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res[0], &allRoles)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling roles: %w", err)
	}

	return allRoles, nil
}

/*
 * # Generate a report of all Okta roles and their users
 */
func (c *Client) GenerateRoleReport() ([]*RoleReport, error) {
	roleReports := []*RoleReport{}
	rolesMap := make(map[string][]*User)

	users, _ := c.ListAllUsers()

	// Use a buffered channel as a semaphore to limit concurrent requests.
	sem := make(chan struct{}, 10)

	// WaitGroup to ensure all go routines complete their tasks.
	var wg sync.WaitGroup

	// Buffered channel to hold user roles result from each goroutine
	userRolesCh := make(chan map[string][]*User, len(*users))

	// Buffered channel to hold any errors that occur while getting user roles
	rolesErrCh := make(chan error)

	for _, user := range *users {
		wg.Add(1)

		go func(user *User) {
			// Release one semaphore resource when the goroutine completes
			defer wg.Done()

			sem <- struct{}{} // acquire one semaphore resource
			roles, err := c.GetUserRoles(user.ID)
			if err != nil {
				rolesErrCh <- err
				return
			}

			userRoles := make(map[string][]*User)
			for _, role := range roles.Roles {
				userRoles[role.Type] = append(userRoles[role.Type], user)
			}
			userRolesCh <- userRoles
			<-sem // release one semaphore resource
		}(user)
	}

	// Wait for all goroutines to finish and close channels
	go func() {
		wg.Wait()
		close(userRolesCh)
		close(rolesErrCh)
	}()

	// Collect roles from all users
	for userRoles := range userRolesCh {
		for roleType, roleUsers := range userRoles {
			rolesMap[roleType] = append(rolesMap[roleType], roleUsers...)
		}
	}

	// Check if there were any errors
	if len(rolesErrCh) > 0 {
		// Handle or return errors. For simplicity, only returning the first error here
		return nil, <-rolesErrCh
	}

	// Append system roles
	for roleType, users := range rolesMap {
		roleReports = append(roleReports, &RoleReport{
			&Role{
				ID:   roleType,
				Type: "System",
			},
			users,
		})
	}

	// Get and append custom roles
	customRoles, err := c.ListAllRoles()
	if err != nil {
		return nil, err
	}

	for _, role := range customRoles.Roles {
		roleReports = append(roleReports, &RoleReport{
			Role:  &role,
			Users: rolesMap[role.Type],
		})
	}

	return roleReports, nil
}

/*
 * # Retrieves a role by `roleIdOrLabel`
 * /api/v1/iam/roles/{roleIdOrLabel}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/#tag/Role/operation/getRole
 */
func (c *Client) GetRole(roleID string) (*Role, error) {
	role := &Role{}

	// url := fmt.Sprintf("%s/%s", OktaRoles, roleID)
	url := c.BuildURL(OktaRoles, roleID)
	res, err := c.HTTPClient.PaginatedRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res[0], &role)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling role: %w", err)
	}

	return role, nil
}

/*
 * Lists all roles assigned to a user identified by `userId``
 * /api/v1/users/{userId}/roles
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/#tag/RoleAssignment/operation/listAssignedRolesForUser
 */
func (c *Client) GetUserRoles(userID string) (*Roles, error) {

	userRoles := Roles{}

	// url := fmt.Sprintf("%s/%s/roles", OktaUsers, userID)
	url := c.BuildURL(OktaUsers, userID, "roles")
	res, err := c.HTTPClient.PaginatedRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		role := Role{}
		err := json.Unmarshal(r, &role)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user: %w", err)
		}
		userRoles.Roles = append(userRoles.Roles, role)
	}

	return &userRoles, nil
}

/*
 * # Get all Users with Role Assignments
 * /api/v1/iam/assignees/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/#tag/RoleAssignment/operation/listUsersWithRoleAssignments
 */
func (c *Client) ListAllUsersWithRoleAssignments() (*Users, error) {

	type U struct {
		Value []*User `json:"value,omitempty"`
	}

	// users := &Users{}

	// url := fmt.Sprintf("%s/assignees/users", OktaIAM)
	url := c.BuildURL(OktaIAM, "assignees", "users")
	res, err := c.HTTPClient.PaginatedRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Printf("res: %+v", res)

	return nil, nil
}
