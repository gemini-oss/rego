/*
# Active Directory - User Operations

This file contains functions for user-related operations in Active Directory.

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/active_directory/users.go
package active_directory

import (
	"fmt"
	"time"
)

// ListAllAdmins retrieves all admins from Active Directory
func (c *Client) ListAllAdmins() (*Users, error) {
	cacheKey := "rego_ad_all_admins"

	var cache Users
	if c.GetCache(cacheKey, cache) {
		return &cache, nil
	}

	attributes := DefaultUserAttributes
	users, err := do[Users](c, FILTER_USER_ADMIN, attributes)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return &users, nil
}

// ListAllUsers retrieves all users from Active Directory
func (c *Client) ListAllUsers() (*Users, error) {
	cacheKey := "rego_ad_all_users"

	var cache Users
	if c.GetCache(cacheKey, cache) {
		return &cache, nil
	}

	attributes := DefaultUserAttributes
	users, err := do[Users](c, "(&(objectCategory=person)(objectClass=user))", attributes)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return &users, nil
}

// ActiveUsers retrieves all active users from Active Directory
func (c *Client) ActiveUsers() (*Users, error) {
	cacheKey := "rego_ad_active_users"

	var cache Users
	if c.GetCache(cacheKey, cache) {
		return &cache, nil
	}

	attributes := DefaultUserAttributes
	users, err := do[Users](c, FILTER_USER_ACTIVE, attributes)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return &users, nil
}

// DisabledUsers retrieves all disabled users from Active Directory
func (c *Client) DisabledUsers() (*Users, error) {
	cacheKey := "rego_ad_disabled_users"

	var cache Users
	if c.GetCache(cacheKey, cache) {
		return &cache, nil
	}

	attributes := DefaultUserAttributes
	users, err := do[Users](c, FILTER_USER_DISABLED, attributes)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return &users, nil
}

// LockedUsers retrieves all locked users from Active Directory
func (c *Client) LockedUsers() (*Users, error) {
	cacheKey := "rego_ad_locked_users"

	var cache Users
	if c.GetCache(cacheKey, cache) {
		return &cache, nil
	}

	attributes := DefaultUserAttributes
	users, err := do[Users](c, FILTER_USER_LOCKED, attributes)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return &users, nil
}

// PasswordNeverExpiresUsers retrieves all users with passwords that never expire from Active Directory
func (c *Client) PasswordNeverExpiresUsers() (*Users, error) {
	cacheKey := "rego_ad_password_never_expires_users"

	var cache Users
	if c.GetCache(cacheKey, cache) {
		return &cache, nil
	}

	attributes := DefaultUserAttributes
	users, err := do[Users](c, FILTER_USER_PASSWORD_NEVER_EXPIRES, attributes)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return &users, nil
}

func (c *Client) MemberOf(group string) (*Users, error) {
	cacheKey := "rego_memberof_" + group

	var cache Users
	if c.GetCache(cacheKey, cache) {
		return &cache, nil
	}

	attributes := DefaultUserAttributes
	users, err := do[Users](c, fmt.Sprintf(FILTER_USER_NESTED_GROUP, group, "OU=Groups", c.BaseDN), attributes)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, users, 30*time.Minute)
	return &users, nil
}
