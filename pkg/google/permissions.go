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
	"encoding/json"
	"fmt"
)

/*
 * Query Parameters for Permissions
 * Reference: https://developers.google.com/drive/api/reference/rest/v3/permissions/create#query-parameters
 */
type PermissionsQuery struct {
	EmailMessage              string `json:"emailMessage,omitempty"`              // A plain text custom message to include in the notification email.
	IncludePermissionsForView string `json:"includePermissionsForView,omitempty"` // Specifies which additional view's permissions to include in the response.
	PageSize                  int    `json:"pageSize,omitempty"`                  // The maximum number of permissions to return per page.
	PageToken                 string `json:"pageToken,omitempty"`                 // The token for continuing a previous list request on the next page.
	MoveToNewOwnersRoot       bool   `json:"moveToNewOwnersRoot,omitempty"`       // This parameter will only take effect if the item is not in a shared drive and the request is attempting to transfer the ownership of the item.
	SendNotificationEmail     bool   `json:"sendNotificationEmail,omitempty"`     // Whether to send a notification email when sharing to users or groups.
	SupportsAllDrives         bool   `json:"supportsAllDrives,omitempty"`         // Whether the requesting application supports both My Drives and shared drives.
	TransferOwnership         bool   `json:"transferOwnership,omitempty"`         // Whether to transfer ownership to the specified user and downgrade the current owner to a writer.
	UseDomainAdminAccess      bool   `json:"useDomainAdminAccess,omitempty"`      // Issue the request as a domain administrator.
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
func (c *Client) GetPermissionList(driveID string) (*PermissionList, error) {
	permissions := &PermissionList{}

	q := PermissionsQuery{}

	url := fmt.Sprintf("%s/%s/permissions", DriveFiles, driveID)
	c.Logger.Debug("url:", url)
	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &permissions)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return permissions, nil
}

/*
 * # Get Google Drive File Permission Details
 * drive/v3/files/{fileId}/permissions/{permissionId}
 * @param {string} fileId - The ID of the file or shortcut.
 * @param {string} permissionId - The ID of the permission.
 * https://developers.google.com/drive/api/reference/rest/v3/permissions/get
 */
func (c *Client) GetPermissionDetails(driveID string, permissionID string) (*Permission, error) {
	permission := &Permission{}

	url := fmt.Sprintf("%s/%s/permissions/%s", DriveFiles, driveID, permissionID)
	c.Logger.Debug("url:", url)
	res, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &permission)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return permission, nil
}

/*
 * # Transfer Google Drive File Ownership
 * drive/v3/files/{fileId}/permissions/{permissionId}/update
 * Warning: Concurrent permissions operations on the same file are not supported; only the last update is applied.
 * @param {string} fileId - The ID of the file or shortcut.
 * @param {string} permissionId - The ID of the permission.
 * https://developers.google.com/drive/api/reference/rest/v3/permissions/create
 */
func (c *Client) TransferOwnership(driveID string, newOwner string) (*Permission, error) {
	permission := &Permission{
		EmailAddress: newOwner,
		Role:         "owner",
		Type:         "user",
	}

	q := PermissionsQuery{
		TransferOwnership: true,
	}

	url := fmt.Sprintf("%s/%s/permissions", DriveFiles, driveID)
	c.Logger.Debug("url:", url)
	res, body, err := c.HTTPClient.DoRequest("POST", url, q, permission)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &permission)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return permission, nil
}
