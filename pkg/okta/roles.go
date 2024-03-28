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
	"fmt"
	"sync"
	"time"
)

/*
 * # Lists all roles with pagination support.
 * - By default, only custom roles can be listed from this endpoint
 * /api/v1/iam/roles
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/#tag/Role/operation/listRoles
 */
func (c *Client) ListAllRoles() (*RolesList, error) {
	url := c.BuildURL(OktaRoles)

	var cache RolesList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	roles, err := doPaginatedStruct[RolesList](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, roles, 5*time.Minute)
	return roles, nil
}

/*
 * # Generate a report of all Okta roles and their users
 */
func (c *Client) GenerateRoleReport() (*RoleReports, error) {
	cacheKey := "Okta_Role_Report"

	var cache RoleReports
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	roleReports := &RoleReports{}
	rolesMap := make(map[*Role]map[*User]struct{})

	users, err := c.ListActiveUsers()
	if err != nil {
		return nil, err
	}

	var rolesMutex sync.Mutex
	var rolesErrMutex sync.Mutex
	var rolesErrors []error

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, user := range *users {
		wg.Add(1)

		go func(user *User) {
			defer wg.Done()

			sem <- struct{}{}
			roles, err := c.GetUserRoles(user.ID)
			if err != nil {
				rolesErrMutex.Lock()
				rolesErrors = append(rolesErrors, err)
				rolesErrMutex.Unlock()
				<-sem
				return
			}

			userRoles := make(map[Role]map[*User]struct{})
			for _, role := range *roles {
				if userRoles[*role] == nil {
					userRoles[*role] = make(map[*User]struct{})
				}
				userRoles[*role][user] = struct{}{}
			}

			rolesMutex.Lock()
			// Add user roles to rolesMap
			for role, users := range userRoles {
				if rolesMap[&role] == nil {
					rolesMap[&role] = make(map[*User]struct{})
				}
				for user := range users {
					rolesMap[&role][user] = struct{}{}
				}
			}
			rolesMutex.Unlock()
			<-sem // release semaphore
		}(user)
	}

	wg.Wait()
	close(sem)

	if len(rolesErrors) > 0 {
		return nil, fmt.Errorf("error generating role report: %v", rolesErrors)
	}

	// Add roles to roleReports
	for role, userSet := range rolesMap {
		var users Users
		for user := range userSet {
			users = append(users, user)
		}
		*roleReports = append(*roleReports, &RoleReport{
			Role:  role,
			Users: &users,
		})
	}

	c.SetCache(cacheKey, roleReports, 60*time.Minute)
	return roleReports, nil
}

/*
 * # Retrieves a role by `roleIdOrLabel`
 * /api/v1/iam/roles/{roleIdOrLabel}
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/#tag/Role/operation/getRole
 */
func (c *Client) GetRole(roleID string) (*Role, error) {
	url := c.BuildURL(OktaRoles, roleID)

	var cache Role
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	role, err := do[Role](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, role, 5*time.Minute)
	return &role, nil
}

/*
 * Lists all roles assigned to a user identified by `userId``
 * /api/v1/users/{userId}/roles
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/#tag/RoleAssignment/operation/listAssignedRolesForUser
 */
func (c *Client) GetUserRoles(userID string) (*Roles, error) {
	url := c.BuildURL(OktaUsers, userID, "roles")

	var cache Roles
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	roles, err := doPaginated[Roles](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, roles, 15*time.Minute)
	return roles, nil
}

/*
 * # Get all Users with Role Assignments
 * /api/v1/iam/assignees/users
 * - https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/#tag/RoleAssignment/operation/listUsersWithRoleAssignments
 */
func (c *Client) ListAllUsersWithRoleAssignments() (*Users, error) {
	url := c.BuildURL(OktaIAM, "assignees", "users")

	var cache Users
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	users, err := doPaginated[Users](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, users, 30*time.Minute)
	return users, nil
}
