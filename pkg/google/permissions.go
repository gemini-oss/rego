/*
# Google Workspace - Google Drive Permissions

This package initializes all the methods for functions which interact with the Google Drive API for Permissions:
https://developers.google.com/drive/api/reference/rest/v3/permissions/

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/permissions.go
package google

import (
	"time"
)

// PermissionsClient for chaining methods
type PermissionsClient struct {
	*Client
}

// Entry point for permissions-related operations
func (c *Client) Permissions() *PermissionsClient {
	pc := &PermissionsClient{
		Client: c,
	}

	// https://developers.google.com/drive/api/guides/limits
	pc.HTTP.RateLimiter.Available = 12000
	pc.HTTP.RateLimiter.Limit = 12000
	pc.HTTP.RateLimiter.Interval = 1 * time.Minute
	pc.HTTP.RateLimiter.Log.Verbosity = c.Log.Verbosity

	return pc
}

/*
 * Query Parameters for Permissions
 * Reference: https://developers.google.com/drive/api/reference/rest/v3/permissions/create#query-parameters
 */
type PermissionsQuery struct {
	EmailMessage              string `url:"emailMessage,omitempty"`              // A plain text custom message to include in the notification email.
	IncludePermissionsForView string `url:"includePermissionsForView,omitempty"` // Specifies which additional view's permissions to include in the response.
	PageSize                  int    `url:"pageSize,omitempty"`                  // The maximum number of permissions to return per page.
	PageToken                 string `url:"pageToken,omitempty"`                 // The token for continuing a previous list request on the next page.
	MoveToNewOwnersRoot       bool   `url:"moveToNewOwnersRoot,omitempty"`       // This parameter will only take effect if the item is not in a shared drive and the request is attempting to transfer the ownership of the item.
	SendNotificationEmail     bool   `url:"sendNotificationEmail,omitempty"`     // Whether to send a notification email when sharing to users or groups.
	SupportsAllDrives         bool   `url:"supportsAllDrives,omitempty"`         // Whether the requesting application supports both My Drives and shared drives.
	TransferOwnership         bool   `url:"transferOwnership,omitempty"`         // Whether to transfer ownership to the specified user and downgrade the current owner to a writer.
	UseDomainAdminAccess      bool   `url:"useDomainAdminAccess,omitempty"`      // Issue the request as a domain administrator.
}

func (q *PermissionsQuery) SetPageToken(token string) {
	q.PageToken = token
}

/*
 * Check if the PermissionsQuery is empty
 */
func (d *PermissionsQuery) IsEmpty() bool {
	return d.EmailMessage == "" &&
		d.IncludePermissionsForView == "" &&
		d.PageSize == 0 &&
		d.PageToken == "" &&
		!d.MoveToNewOwnersRoot &&
		!d.SendNotificationEmail &&
		!d.SupportsAllDrives &&
		!d.TransferOwnership &&
		!d.UseDomainAdminAccess
}

/*
 * # Get Google Drive File Permissions
 * drive/v3/files/{fileId}/permissions
 * @param {string} fileId - The ID of the file or shortcut.
 * https://developers.google.com/drive/api/reference/rest/v3/permissions/list
 */
func (c *PermissionsClient) GetPermissionList(driveID string) (*PermissionList, error) {
	url := c.BuildURL(DriveFiles, nil, driveID, "permissions")

	var cache PermissionList
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	permissions, err := do[*PermissionList](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, permissions, 5*time.Minute)
	return permissions, nil
}

/*
 * # Get Google Drive File Permission Details
 * drive/v3/files/{fileId}/permissions/{permissionId}
 * @param {string} fileId - The ID of the file or shortcut.
 * @param {string} permissionId - The ID of the permission.
 * https://developers.google.com/drive/api/reference/rest/v3/permissions/get
 */
func (c *PermissionsClient) GetPermissionDetails(driveID string, permissionID string) (*Permission, error) {
	url := c.BuildURL(DriveFiles, nil, driveID, "permissions", permissionID)

	var cache Permission
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	permission, err := do[Permission](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, permission, 5*time.Minute)
	return &permission, nil
}

/*
 * # Transfer Google Drive File Ownership
 * drive/v3/files/{fileId}/permissions/{permissionId}/update
 * Warning: Concurrent permissions operations on the same file are not supported; only the last update is applied.
 * @param {string} fileId - The ID of the file or shortcut.
 * @param {string} permissionId - The ID of the permission.
 * https://developers.google.com/drive/api/reference/rest/v3/permissions/create
 */
func (c *PermissionsClient) TransferOwnership(driveID string, newOwner string) (*Permission, error) {
	url := c.BuildURL(DriveFiles, nil, driveID, "permissions")

	permission := &Permission{
		EmailAddress: newOwner,
		Role:         "owner",
		Type:         "user",
	}

	q := PermissionsQuery{
		TransferOwnership: true,
	}

	permission, err := do[*Permission](c.Client, "POST", url, q, permission)
	if err != nil {
		return nil, err
	}

	return permission, nil
}
