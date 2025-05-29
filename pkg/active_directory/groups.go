/*
# Active Directory - Group Operations

This file contains functions for group-related operations in Active Directory.

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/active_directory/groups.go
package active_directory

import (
	"time"
)

// ListAllGroups retrieves all groups from Active Directory
func (c *Client) ListAllGroups() (*Groups, error) {
	cacheKey := "ad_all_groups"
	var cache Groups

	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	attributes := &[]Attribute{DistinguishedName, CommonName, Description}
	groups, err := do[Groups](c, "(objectClass=group)", attributes)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, groups, 30*time.Minute)
	return &groups, nil
}
