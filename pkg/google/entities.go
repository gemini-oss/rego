/*
# Google Workspace - Entities [Structs]

This package contains many structs for handling responses from the Google API:

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/sheets.go
package google

import (
	"github.com/gemini-oss/rego/pkg/common/auth"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
	"golang.org/x/oauth2/jwt"
)

// ### Google Client Structs
// ---------------------------------------------------------------------
type AuthCredentials struct {
	Type        string // api_key, oauth_client, service_account
	Credentials string
	CICD        bool     // If true, will use environmental variables
	Scopes      []string // Scopes to use for OAuth
	Subject     string   // Subject to impersonate
}

type GoogleConfig struct {
	Web       auth.OAuthConfig `json:"web"`
	Installed auth.OAuthConfig `json:"installed"`
}

type Client struct {
	Auth       AuthCredentials // Credentials to use for authentication
	BaseURL    string          // Base URL to use for API calls
	OAuth      *auth.OAuthConfig
	JWT        *jwt.Config      // JWT Config
	HTTPClient *requests.Client // HTTP Client
	Error      *Error           // Error
	Logger     *log.Logger      // Logger
}

type Error struct {
	Error struct {
		Errors  []ErrorDetail `json:"errors"`
		Code    int           `json:"code"`
		Message string        `json:"message"`
	} `json:"error"`
}

type ErrorDetail struct {
	Domain       string `json:"domain"`
	Reason       string `json:"reason"`
	Message      string `json:"message"`
	LocationType string `json:"locationType"`
	Location     string `json:"location"`
}

type ServiceAccount struct {
	Type         string `json:"type"`
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	AuthURL      string `json:"auth_uri"`
	TokenURL     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`
}

// END OF GOOGLE CLIENT STRUCTS
//---------------------------------------------------------------------

// ### Google Admin SDK Structs
//---------------------------------------------------------------------

// https://developers.google.com/admin-sdk/reports/reference/rest/v1/activities/list#Activity
type Report struct {
	Kind          string     `json:"kind,omitempty"`          // The type of API resource
	Etag          string     `json:"etag,omitempty"`          // ETag of the entry
	Items         []Report   `json:"items,omitempty"`         // Activity events in the report
	OwnerDomain   string     `json:"ownerDomain,omitempty"`   // Domain that is affected by the event
	IPAddress     string     `json:"ipAddress,omitempty"`     // IP address of the user doing the action
	Events        []Event    `json:"events,omitempty"`        // Activity events in the report
	ID            ActivityID `json:"id,omitempty"`            // Unique identifier for each activity record
	Actor         Actor      `json:"actor,omitempty"`         // User doing the action
	Warnings      []Warning  `json:"warnings,omitempty"`      // Warnings, if any
	Date          string     `json:"date,omitempty"`          // The date of the report request
	Entity        Entity     `json:"entity,omitempty"`        // Information about the type of the item
	NextPageToken string     `json:"nextPageToken,omitempty"` // Token to specify next page
	UsageReports  []Report   `json:"usageReports,omitempty"`  // Various application parameter records
}

type Event struct {
	Type       string            `json:"type,omitempty"`       // Type of event
	Name       string            `json:"name,omitempty"`       // Name of the event
	Parameters []ReportParameter `json:"parameters,omitempty"` // Parameter value pairs for various applications
}

type ReportParameter struct {
	Name              string            `json:"name,omitempty"`              // The name of the parameter
	Value             string            `json:"value,omitempty"`             // String value of the parameter
	MultiValue        []string          `json:"multiValue,omitempty"`        // String values of the parameter
	IntValue          string            `json:"intValue,omitempty"`          // Integer value of the parameter
	MultiIntValue     []string          `json:"multiIntValue,omitempty"`     // Integer values of the parameter
	BoolValue         bool              `json:"boolValue,omitempty"`         // Boolean value of the parameter
	MessageValue      []ReportParameter `json:"messageValue,omitempty"`      // Nested parameter value pairs
	MultiMessageValue []ReportParameter `json:"multiMessageValue,omitempty"` // Activities list of messageValue objects
	StringValue       string            `json:"stringValue,omitempty"`       // String value of the parameter
	DatetimeValue     string            `json:"datetimeValue,omitempty"`     // The RFC 3339 formatted value of the parameter
}

type ActivityID struct {
	Time            string `json:"time,omitempty"`            // Time of occurrence of the activity
	UniqueQualifier string `json:"uniqueQualifier,omitempty"` // Unique qualifier if multiple events have the same time
	ApplicationName string `json:"applicationName,omitempty"` // Application name to which the event belongs
	CustomerID      string `json:"customerId,omitempty"`      // The unique identifier for a Google Workspace account
}

type Actor struct {
	ProfileID  string `json:"profileId,omitempty"`  // The unique Google Workspace profile ID of the actor
	Email      string `json:"email,omitempty"`      // The primary email address of the actor
	CallerType string `json:"callerType,omitempty"` // The type of actor
	Key        string `json:"key,omitempty"`        // Key present when callerType is KEY
}

type Warning struct {
	Code    string        `json:"code,omitempty"`    // Machine readable code or warning type
	Message string        `json:"message,omitempty"` // The human readable messages for a warning
	Data    []WarningData `json:"data,omitempty"`    // Key-value pairs to give detailed information on the warning
}

type WarningData struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type Entity struct {
	CustomerID string `json:"customerId,omitempty"` // The unique identifier of the customer's account
	UserEmail  string `json:"userEmail,omitempty"`  // The user's email address
	ProfileID  string `json:"profileId,omitempty"`  // The user's immutable Google Workspace profile identifier
	EntityID   string `json:"entityId,omitempty"`   // Object key
	Type       string `json:"type,omitempty"`       // The type of item
}

type Roles struct {
	Etag  string `json:"etag,omitempty"`  // ETag of the resource
	Kind  string `json:"kind,omitempty"`  // The type of the API resource
	Items []Role `json:"items,omitempty"` // A list of Roles
}

type Role struct {
	Etag             string          `json:"etag,omitempty"`             // ETag of the resource
	IsSuperAdminRole bool            `json:"isSuperAdminRole,omitempty"` // Returns true if the role is a super admin role
	IsSystemRole     bool            `json:"isSystemRole,omitempty"`     // Returns true if this is a pre-defined system role
	Kind             string          `json:"kind,omitempty"`             // The type of the API resource
	RoleDescription  string          `json:"roleDescription,omitempty"`  // A short description of the role
	RoleID           string          `json:"roleId,omitempty"`           // ID of the role
	RoleName         string          `json:"roleName,omitempty"`         // Name of the role
	RolePrivileges   []RolePrivilege `json:"rolePrivileges,omitempty"`   // The set of privileges that are granted to this role
}

type RolePrivilege struct {
	PrivilegeName string `json:"privilegeName,omitempty"` // The name of the privilege
	ServiceID     string `json:"serviceId,omitempty"`     // The obfuscated ID of the service this privilege is for
}

type RoleAssignment struct {
	RoleAssignmentId string           `json:"roleAssignmentId,omitempty"` // The unique ID of the role assignment
	RoleId           string           `json:"roleId,omitempty"`           // The ID of the role that is assigned
	Kind             string           `json:"kind,omitempty"`             // The type of the API resource
	Etag             string           `json:"etag,omitempty"`             // ETag of the resource
	AssignedTo       string           `json:"assignedTo,omitempty"`       // The unique ID of the user this role is assigned to
	AssigneeType     string           `json:"assigneeType,omitempty"`     // The type of entity this role is assigned to.
	ScopeType        string           `json:"scopeType,omitempty"`        // The type of the scope.
	OrgUnitId        string           `json:"orgUnitId,omitempty"`        // If the role is restricted to an organization unit, this contains the ID for the organization unit the exercise of this role is restricted to.
	Condition        string           `json:"condition,omitempty"`        // The condition which determines if this assignment is active. If not specified, the assignment is always active.
	NextPageToken    string           `json:"nextPageToken,omitempty"`    // Token to specify the next page in the list
	Items            []RoleAssignment `json:"items,omitempty"`            // The list of matching role assignments.
}

type RoleReport struct {
	Role  *Role
	Users []*User
}

// END OF GOOGLE ADMIN SDK STRUCTS
//---------------------------------------------------------------------

// ### Google Drive Structs
// ---------------------------------------------------------------------
type Document struct {
	Kind                         string               `json:"kind,omitempty"`                         // drive#file
	DriveID                      string               `json:"driveId,omitempty"`                      // The ID of the shared drive the file resides in. Only populated for items in shared drives.
	FileExtension                string               `json:"fileExtension,omitempty"`                // The extension of the file. This is populated even when Drive is unable to determine the extension. This field can be cleared by writing a new empty value to this field.
	CopyRequiresWriterPermission bool                 `json:"copyRequiresWriterPermission,omitempty"` // Whether the file has been created or opened in a Google editor. This field is only populated for files with content stored in Drive; it is not populated for Google Docs or shortcut files.
	MD5Checksum                  string               `json:"md5Checksum,omitempty"`                  // The MD5 checksum for the content of the file. This is populated only for files with content stored in Drive.
	ContentHints                 ContentHints         `json:"contentHints,omitempty"`                 // Additional information about the content of the file. These fields are never populated in responses.
	WritersCanShare              bool                 `json:"writersCanShare,omitempty"`              // Whether writers can share the document with other users. Not populated for items in shared drives.
	ViewedByMe                   bool                 `json:"viewedByMe,omitempty"`                   // Whether the file has been viewed by this user.
	MimeType                     string               `json:"mimeType,omitempty"`                     // The MIME type of the file. Drive will attempt to automatically detect an appropriate value from uploaded content if no value is provided. The value cannot be changed unless a new revision is uploaded. If a file is created with a Google Doc MIME type, the uploaded content will be imported if possible. The supported import formats are published in the About resource.
	ExportLinks                  map[string]string    `json:"exportLinks,omitempty"`                  // A map of the id of each of the user's apps to a link to open this file with that app. Only populated when the drive.apps.readonly scope is used.
	Parents                      []string             `json:"parents,omitempty"`                      // The IDs of the parent folders which contain the file. If not specified as part of a create request, the file will be placed directly in the user's My Drive folder. If not specified as part of a copy request, the file will inherit any discoverable parents of the source file. Update requests must use the addParents and removeParents parameters to modify the parents list.
	ThumbnailLink                string               `json:"thumbnailLink,omitempty"`                // A short-lived link to the file's thumbnail, if available. Typically lasts on the order of hours. Only populated when the requesting app can access the file's content.
	IconLink                     string               `json:"iconLink,omitempty"`                     // A static, unauthenticated link to the file's icon.
	Shared                       bool                 `json:"shared,omitempty"`                       // Whether the file has been shared. Not populated for items in shared drives.
	LastModifyingUser            User                 `json:"lastModifyingUser,omitempty"`            // The user who last modified the file.
	Owners                       []User               `json:"owners,omitempty"`                       // The owners of the file. Currently, only certain legacy files may have more than one owner. Not populated for items in shared drives.
	HeadRevisionID               string               `json:"headRevisionId,omitempty"`               // The ID of the file's head revision. This field is only populated for files with content stored in Drive; it is not populated for Google Docs or shortcut files.
	SharingUser                  User                 `json:"sharingUser,omitempty"`                  // The user who shared the file with the requesting user, if applicable.
	WebViewLink                  string               `json:"webViewLink,omitempty"`                  // A link for opening the file in a relevant Google editor or viewer in a browser.
	WebContentLink               string               `json:"webContentLink,omitempty"`               // A link for downloading the content of the file in a browser. This is only available for files with binary content in Drive.
	Size                         string               `json:"size,omitempty"`                         // The size of the file's content in bytes. This is only applicable to files with binary content in Drive.
	ViewersCanCopyContent        bool                 `json:"viewersCanCopyContent,omitempty"`        // Whether users with only reader or commenter permission can copy the file's content. This affects copy, download, and print operations.
	Permissions                  []Permission         `json:"permissions,omitempty"`                  // The full list of permissions for the file. This is only available if the requesting user can share the file. Not populated for items in shared drives.
	HasThumbnail                 bool                 `json:"hasThumbnail,omitempty"`                 // Whether the file has a thumbnail. This does not indicate whether the requesting app has access to the thumbnail. To check access, look for the presence of the thumbnailLink field.
	Spaces                       []string             `json:"spaces,omitempty"`                       // The spaces which contain the file. The currently supported values are 'drive', 'appDataFolder' and 'photos'.
	FolderColorRgb               string               `json:"folderColorRgb,omitempty"`               // The color for a folder as an RGB hex string. The supported colors are published in the folderColorPalette field of the About resource.
	ID                           string               `json:"id,omitempty"`                           // The ID of the file.
	Name                         string               `json:"name,omitempty"`                         // The name of the file. This is not necessarily unique within a folder. Note that for immutable items such as the top level folders of shared drives, My Drive root folder, and Application Data folder the name is constant.
	Description                  string               `json:"description,omitempty"`                  // A short description of the file.
	Starred                      bool                 `json:"starred,omitempty"`                      // Whether the user has starred the file.
	Trashed                      bool                 `json:"trashed,omitempty"`                      // Whether the file has been trashed, either explicitly or from a trashed parent folder. Only the owner may trash a file. The trashed item is excluded from all files.list responses returned for any user who does not own the file. However, all users with access to the file can see the trashed item metadata in an API response. All users with access can copy, download, export, and share the file.
	ExplicitlyTrashed            bool                 `json:"explicitlyTrashed,omitempty"`            // Whether the file has been explicitly trashed, as opposed to recursively trashed from a parent folder.
	CreatedTime                  string               `json:"createdTime,omitempty"`                  // The time at which the file was created (RFC 3339 date-time).
	ModifiedTime                 string               `json:"modifiedTime,omitempty"`                 // The last time the file was modified by anyone (RFC 3339 date-time). Note that setting modifiedTime will also update modifiedByMeTime for the user.
	ModifiedByMeTime             string               `json:"modifiedByMeTime,omitempty"`             // The last time the file was modified by the user (RFC 3339 date-time). If the file has been modified by anyone other than the user, this will be set to the time the file was last updated. (Read-only)
	ViewedByMeTime               string               `json:"viewedByMeTime,omitempty"`               // The last time the file was viewed by the user (RFC 3339 date-time).
	SharedWithMeTime             string               `json:"sharedWithMeTime,omitempty"`             // The time at which the file was shared with the user, if applicable (RFC 3339 date-time).
	QuotaBytesUsed               string               `json:"quotaBytesUsed,omitempty"`               // The number of storage quota bytes used by the file. This includes the head revision as well as previous revisions with keepForever enabled.
	Version                      string               `json:"version,omitempty"`                      // A monotonically increasing version number for the file. This reflects every change made to the file on the server, even those not visible to the user.
	OriginalFilename             string               `json:"originalFilename,omitempty"`             // The original filename of the uploaded content if available, or else the original value of the name field. This is only available for files with binary content in Drive.
	OwnedByMe                    bool                 `json:"ownedByMe,omitempty"`                    // Whether the user owns the file. Not populated for items in shared drives.
	FullFileExtension            string               `json:"fullFileExtension,omitempty"`            // The full file extension extracted from the name field. May contain multiple concatenated extensions, such as "tar.gz". This is only available for files with binary content in Drive.
	Properties                   map[string]string    `json:"properties,omitempty"`                   // Additional metadata about video media. This may not be available immediately upon upload.
	AppProperties                map[string]string    `json:"appProperties,omitempty"`                // Additional metadata about image media, if available.
	IsAppAuthorized              bool                 `json:"isAppAuthorized,omitempty"`              // Whether the file has been shared. Not populated for items in shared drives.
	TeamDriveID                  string               `json:"teamDriveId,omitempty"`                  // The ID of the Team Drive that owns the file. Not populated for items in shared drives.
	Capabilities                 Capabilities         `json:"capabilities,omitempty"`                 // Capabilities the current user has on this file. Each capability corresponds to a fine-grained action that a user may take.
	HasAugmentedPermissions      bool                 `json:"hasAugmentedPermissions,omitempty"`      // Whether the options to copy, print, or download this file, should be disabled for readers and commenters.
	TrashingUser                 User                 `json:"trashingUser,omitempty"`                 // The user who trashed the file. Only populated for items in shared drives.
	ThumbnailVersion             string               `json:"thumbnailVersion,omitempty"`             // A monotonically increasing version number for the thumbnail image for this file. This reflects every change made to the thumbnail on the server, including those not visible to the requesting user.
	TrashedTime                  string               `json:"trashedTime,omitempty"`                  // The time that the item was trashed (RFC 3339 date-time). Only populated for items in shared drives.
	ModifiedByMe                 bool                 `json:"modifiedByMe,omitempty"`                 // Whether the file has been modified by this user.
	PermissionIds                []string             `json:"permissionIds,omitempty"`                // A collection of arbitrary key-value pairs which are private to the requesting app. Entries with null values are cleared in update and copy requests.
	ImageMediaMetadata           ImageMediaMetadata   `json:"imageMediaMetadata,omitempty"`           // Additional metadata about image media, if available.
	VideoMediaMetadata           VideoMediaMetadata   `json:"videoMediaMetadata,omitempty"`           // Additional metadata about video media. This may not be available immediately upon upload.
	ShortcutDetails              ShortcutDetails      `json:"shortcutDetails,omitempty"`              // Shortcut file details. Only populated for shortcut files, which have the mimeType field set to application/vnd.google-apps.shortcut.
	ContentRestrictions          []ContentRestriction `json:"contentRestrictions,omitempty"`          // Restrictions for accessing the content of the file. Only populated if such a restriction exists.
	ResourceKey                  string               `json:"resourceKey,omitempty"`                  // A key needed to access the item via a shared link.
	LinkShareMetadata            LinkShareMetadata    `json:"linkShareMetadata,omitempty"`            // Metadata about the shared link.
	LabelInfo                    LabelInfo            `json:"labelInfo,omitempty"`                    // Additional information about the content of the file. These fields are never populated in responses.
	SHA1Checksum                 string               `json:"sha1Checksum,omitempty"`                 // The SHA1 checksum for the content of the file. It is computed by Drive and guaranteed to be up-to-date at all times. A change in the content of the file will cause a change in its SHA256 checksum.
	SHA256Checksum               string               `json:"sha256Checksum,omitempty"`               // The SHA256 checksum for the content of the file. It is computed by Drive and guaranteed to be up-to-date at all times. A change in the content of the file will cause a change in its SHA256 checksum.
}

type Permission struct {
	// Permission fields here
}

type ContentHints struct {
	IndexableText string    `json:"indexableText,omitempty"` // Text to be indexed for the file to improve fullText queries. This is limited to 128KB in length and may contain HTML elements.
	Thumbnail     Thumbnail `json:"thumbnail,omitempty"`     // A thumbnail for the file. This will only be used if Drive cannot generate a standard thumbnail.
}

type Thumbnail struct {
	Image    string `json:"image,omitempty"`    // The thumbnail data encoded with URL-safe Base64 (RFC 4648 section 5).
	MimeType string `json:"mimeType,omitempty"` // The MIME type of the thumbnail.
}

type Capabilities struct {
	CanAddChildren                        bool `json:"canAddChildren,omitempty"`                        // Whether the current user can add children to this folder. This is always false when the item is not a folder.
	CanAddFolderFromAnotherDrive          bool `json:"canAddFolderFromAnotherDrive,omitempty"`          // Whether the current user can add a folder from another drive (different shared drive or My Drive) to this folder. This is false when the item is not a folder. Only populated for items in shared drives.
	CanAddMyDriveParent                   bool `json:"canAddMyDriveParent,omitempty"`                   // Whether the current user can add a parent for the item without removing an existing parent in the same request. Not populated for shared drive files.
	CanChangeCopyRequiresWriterPermission bool `json:"canChangeCopyRequiresWriterPermission,omitempty"` // Whether the current user can change the copyRequiresWriterPermission restriction of this file.
	CanChangeSecurityUpdateEnabled        bool `json:"canChangeSecurityUpdateEnabled,omitempty"`        // Whether the current user can modify the content restrictions of this file.
	CanChangeViewersCanCopyContent        bool `json:"canChangeViewersCanCopyContent,omitempty"`        // Whether the current user can modify the viewersCanCopyContent restriction of this file.
	CanComment                            bool `json:"canComment,omitempty"`                            // Whether the current user can comment on this file.
	CanCopy                               bool `json:"canCopy,omitempty"`                               // Whether the current user can copy this file. For a Team Drive item, whether the current user can copy non-folder descendants of this item, or this item itself if it is not a folder.
	CanDelete                             bool `json:"canDelete,omitempty"`                             // Whether the current user can delete this file.
	CanDeleteChildren                     bool `json:"canDeleteChildren,omitempty"`                     // Whether the current user can delete children of this folder. This is false when the item is not a folder. Only populated for items in shared drives.
	CanDownload                           bool `json:"canDownload,omitempty"`                           // Whether the current user can download this file.
	CanEdit                               bool `json:"canEdit,omitempty"`                               // Whether the current user can edit this file.
	CanListChildren                       bool `json:"canListChildren,omitempty"`                       // Whether the current user can list the children of this folder. This is always false when the item is not a folder.
	CanModifyContent                      bool `json:"canModifyContent,omitempty"`                      // Whether the current user can modify the content of this file.
	CanModifyContentRestriction           bool `json:"canModifyContentRestriction,omitempty"`           // Whether the current user can modify restrictions on content of this file.
	CanModifyLabels                       bool `json:"canModifyLabels,omitempty"`                       // Whether the current user can modify the file's metadata.
	CanMoveChildrenOutOfDrive             bool `json:"canMoveChildrenOutOfDrive,omitempty"`             // Whether the current user can move children of this folder outside of the shared drive. This is false when the item is not a folder. Only populated for items in shared drives.
	CanMoveChildrenOutOfTeamDrive         bool `json:"canMoveChildrenOutOfTeamDrive,omitempty"`         // Deprecated - use canMoveChildrenOutOfDrive instead.
	CanMoveChildrenWithinDrive            bool `json:"canMoveChildrenWithinDrive,omitempty"`            // Whether the current user can move children of this folder within this drive. This is false when the item is not a folder. Note that a request to move the child may still fail depending on the current user's access to the child and to the destination folder.
	CanMoveChildrenWithinTeamDrive        bool `json:"canMoveChildrenWithinTeamDrive,omitempty"`        // Deprecated - use canMoveChildrenWithinDrive instead.
	CanMoveItemIntoTeamDrive              bool `json:"canMoveItemIntoTeamDrive,omitempty"`              // Deprecated - use canMoveItemWithinDrive or canMoveItemOutOfDrive instead.
	CanMoveItemOutOfDrive                 bool `json:"canMoveItemOutOfDrive,omitempty"`                 // Whether the current user can move this item outside of this drive by changing its parent. Note that a request to change the parent of the item may still fail depending on the new parent that is being added.
	CanMoveItemOutOfTeamDrive             bool `json:"canMoveItemOutOfTeamDrive,omitempty"`             // Deprecated - use canMoveItemOutOfDrive instead.
	CanMoveItemWithinDrive                bool `json:"canMoveItemWithinDrive,omitempty"`                // Whether the current user can move this item within this drive. Note that a request to change the parent of the item may still fail depending on the new parent that is being added and the parent that is being removed.
	CanMoveItemWithinTeamDrive            bool `json:"canMoveItemWithinTeamDrive,omitempty"`            // Deprecated - use canMoveItemWithinDrive instead.
	CanMoveTeamDriveItem                  bool `json:"canMoveTeamDriveItem,omitempty"`                  // Deprecated - use canMoveItemWithinDrive or canMoveItemOutOfDrive instead.
	CanReadDrive                          bool `json:"canReadDrive,omitempty"`                          // Whether the current user can read the shared drive to which this file belongs. Only populated for items in shared drives.
	CanReadLabels                         bool `json:"canReadLabels,omitempty"`                         // Whether the current user can read the revisions resource of this file. For a Team Drive item, whether revisions of non-folder descendants of this item, or this item itself if it is not a folder, can be read.
	CanReadRevisions                      bool `json:"canReadRevisions,omitempty"`                      // Whether the current user can read the revisions resource of this file. For a Team Drive item, whether revisions of non-folder descendants of this item, or this item itself if it is not a folder, can be read.
	CanReadTeamDrive                      bool `json:"canReadTeamDrive,omitempty"`                      // Deprecated - use canReadDrive instead.
	CanRemoveChildren                     bool `json:"canRemoveChildren,omitempty"`                     // Whether the current user can remove children from this folder. This is always false when the item is not a folder. For a Team Drive item, whether the current user can remove descendants of this item, or this item itself if it is not a folder, from a shared drive.
	CanRemoveMyDriveParent                bool `json:"canRemoveMyDriveParent,omitempty"`                // Whether the current user can remove a parent from the item without adding another parent in the same request. Not populated for shared drive files.
	CanRename                             bool `json:"canRename,omitempty"`                             // Whether the current user can rename this file.
	CanShare                              bool `json:"canShare,omitempty"`                              // Whether the current user can modify the sharing settings for this file.
	CanTrash                              bool `json:"canTrash,omitempty"`                              // Whether the current user can move this file to trash.
	CanTrashChildren                      bool `json:"canTrashChildren,omitempty"`                      // Whether the current user can trash children of this folder. This is false when the item is not a folder. Only populated for items in shared drives.
	CanUntrash                            bool `json:"canUntrash,omitempty"`                            // Whether the current user can restore this file from trash.
}

type ImageMediaMetadata struct {
	FlashUsed        bool     `json:"flashUsed,omitempty"`        // Whether a flash was used to create the photo.
	MeteringMode     string   `json:"meteringMode,omitempty"`     // The metering mode used to create the photo.
	Sensor           string   `json:"sensor,omitempty"`           // The type of sensor used to create the photo.
	ExposureMode     string   `json:"exposureMode,omitempty"`     // The exposure mode used to create the photo.
	ColorSpace       string   `json:"colorSpace,omitempty"`       // The color space of the photo.
	WhiteBalance     string   `json:"whiteBalance,omitempty"`     // The white balance mode used to create the photo.
	Width            int      `json:"width,omitempty"`            // The width of the image in pixels.
	Height           int      `json:"height,omitempty"`           // The height of the image in pixels.
	Location         Location `json:"location,omitempty"`         // Geographic location information stored in the image.
	Rotation         int      `json:"rotation,omitempty"`         // The rotation in clockwise degrees from the image's original orientation.
	Time             string   `json:"time,omitempty"`             // The date and time the photo was taken (EXIF DateTime).
	CameraMake       string   `json:"cameraMake,omitempty"`       // The make of the camera used to create the photo.
	CameraModel      string   `json:"cameraModel,omitempty"`      // The model of the camera used to create the photo.
	ExposureTime     float64  `json:"exposureTime,omitempty"`     // The length of the exposure, in seconds.
	Aperture         float64  `json:"aperture,omitempty"`         // The aperture used to create the photo (f-number).
	FocalLength      float64  `json:"focalLength,omitempty"`      // The focal length used to create the photo, in millimeters.
	ISOSpeed         int      `json:"isoSpeed,omitempty"`         // The ISO speed used to create the photo.
	ExposureBias     float64  `json:"exposureBias,omitempty"`     // The exposure bias of the photo (APEX value).
	MaxApertureValue float64  `json:"maxApertureValue,omitempty"` // The smallest f-number of the lens at the focal length used to create the photo (APEX value).
	SubjectDistance  int      `json:"subjectDistance,omitempty"`  // The distance to the subject of the photo, in meters.
	Lens             string   `json:"lens,omitempty"`             // The lens used to create the photo.
}

type Location struct {
	Latitude  float64 `json:"latitude,omitempty"`  // The latitude stored in the image.
	Longitude float64 `json:"longitude,omitempty"` // The longitude stored in the image.
	Altitude  float64 `json:"altitude,omitempty"`  // The altitude stored in the image.
}

type VideoMediaMetadata struct {
	Width          int    `json:"width,omitempty"`          // The width of the video in pixels.
	Height         int    `json:"height,omitempty"`         // The height of the video in pixels.
	DurationMillis string `json:"durationMillis,omitempty"` // The duration of the video in milliseconds.
}

type ShortcutDetails struct {
	TargetID          string `json:"targetId,omitempty"`          // The ID of the file that this shortcut points to.
	TargetMimeType    string `json:"targetMimeType,omitempty"`    // The MIME type of the file that this shortcut points to. The value of this field is a snapshot of the target's MIME type, captured when the shortcut is created.
	TargetResourceKey string `json:"targetResourceKey,omitempty"` // The resource key of the target file. This is a unique identifier of the target file and is guaranteed to be immutable across file renames.
}

type ContentRestriction struct {
	// Content restriction fields here
}

type LinkShareMetadata struct {
	SecurityUpdateEligible bool `json:"securityUpdateEligible,omitempty"` // Indicates whether this revision is protected with the security update shield. Only populated and used in Meet.
	SecurityUpdateEnabled  bool `json:"securityUpdateEnabled,omitempty"`  // Indicates whether users with only link sharing permissions can copy the file's content. This field is only populated for drive internal copies (not copies created by copying the file in the Drive UI). This is only applicable to files with binary content in Google Drive.
}

type LabelInfo struct {
	Labels []Label `json:"labels,omitempty"` // The list of labels belonging to this account.
}

type Label struct {
	// Label fields here
}

// END OF GOOGLE DRIVE STRUCTS
//---------------------------------------------------------------------

// ### Spreadsheet Structs
// ---------------------------------------------------------------------
// Spreadsheet represents a spreadsheet.
type Spreadsheet struct {
	DataSources         []DataSource                `json:"dataSources,omitempty"`         // List of data sources in the spreadsheet
	DataSourceSchedules []DataSourceRefreshSchedule `json:"dataSourceSchedules,omitempty"` // List of data source refresh schedules in the spreadsheet
	DeveloperMetadata   []DeveloperMetadata         `json:"developerMetadata,omitempty"`   // Developer metadata associated with the spreadsheet
	NamedRanges         []NamedRange                `json:"namedRanges,omitempty"`         // List of named ranges in the spreadsheet
	Properties          *SpreadsheetProperties      `json:"properties,omitempty"`          // Properties of the spreadsheet
	Sheets              []Sheet                     `json:"sheets,omitempty"`              // List of sheets in the spreadsheet
	SpreadsheetID       string                      `json:"spreadsheetId,omitempty"`       // ID of the spreadsheet
	SpreadsheetURL      string                      `json:"spreadsheetUrl,omitempty"`      // URL of the spreadsheet
}

// ValueRange represents a value range in a spreadsheet.
type ValueRange struct {
	Range          string     `json:"range"`          // The range the values cover, in A1 notation
	MajorDimension string     `json:"majorDimension"` // The major dimension of the values
	Values         [][]string `json:"values"`         // The data that was read or to be written
}

// DataSource represents a data source in a spreadsheet.
type DataSource struct {
	CalculatedColumns []DataSourceColumn `json:"calculatedColumns,omitempty"` // Calculated columns in the data source
	DataSourceID      string             `json:"dataSourceId,omitempty"`      // ID of the data source
	SheetID           int                `json:"sheetId,omitempty"`           // ID of the sheet the data source is in
	Spec              *DataSourceSpec    `json:"spec,omitempty"`              // Specification of the data source
}

// DataSourceSpec represents the specifications of a data source.
type DataSourceSpec struct {
	BigQuery   *BigQueryDataSourceSpec `json:"bigQuery,omitempty"`   // Specifications for a BigQuery data source
	Parameters []*DataSourceParameter  `json:"parameters,omitempty"` // Parameters of the data source
}

// BigQueryDataSourceSpec represents the specifications of a BigQuery data source.
type BigQueryDataSourceSpec struct {
	ProjectID string             `json:"projectId,omitempty"` // Project ID of the BigQuery data source
	QuerySpec *BigQueryQuerySpec `json:"querySpec,omitempty"` // Specifications for a BigQuery query
	TableSpec *BigQueryTableSpec `json:"tableSpec,omitempty"` // Specifications for a BigQuery table
}

// BigQueryQuerySpec represents the specifications of a BigQuery query.
type BigQueryQuerySpec struct {
	RawQuery string `json:"rawQuery,omitempty"` // Raw query string for the BigQuery
}

// BigQueryTableSpec represents the specifications of a BigQuery table.
type BigQueryTableSpec struct {
	TableProjectID string `json:"tableProjectId,omitempty"` // Project ID of the table
	TableID        string `json:"tableId,omitempty"`        // ID of the table
	DatasetID      string `json:"datasetId,omitempty"`      // ID of the dataset
}

// DataSourceParameter represents a parameter of a data source.
type DataSourceParameter struct {
	Name         string     `json:"name,omitempty"`         // Name of the parameter
	NamedRangeID string     `json:"namedRangeId,omitempty"` // ID of the named range
	Range        *GridRange `json:"range,omitempty"`        // Grid range of the parameter
}

// DataSourceColumn represents a column in the data source.
type DataSourceColumn struct {
	Formula   string                     `json:"formula,omitempty"`   // Formula of the data source column
	Reference *DataSourceColumnReference `json:"reference,omitempty"` // Reference to the data source column
}

// DataSourceColumnReference represents a reference to a column in the data source.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/other#DataSourceColumnReference
type DataSourceColumnReference struct {
	Name string `json:"name,omitempty"` // Name of the data source column reference
}

// DataSourceRefreshSchedule represents a refresh schedule of a data source.
type DataSourceRefreshSchedule struct {
	DailySchedule   *DataSourceRefreshDailySchedule   `json:"dailySchedule,omitempty"`   // Daily refresh schedule
	Enabled         bool                              `json:"enabled,omitempty"`         // Whether the refresh schedule is enabled
	MonthlySchedule *DataSourceRefreshMonthlySchedule `json:"monthlySchedule,omitempty"` // Monthly refresh schedule
	NextRun         *Interval                         `json:"nextRun,omitempty"`         // The next scheduled run
	RefreshScope    string                            `json:"refreshScope,omitempty"`    // Scope of the refresh
	WeeklySchedule  *DataSourceRefreshWeeklySchedule  `json:"weeklySchedule,omitempty"`  // Weekly refresh schedule
}

// DataSourceRefreshDailySchedule represents a daily schedule for data source refresh.
type DataSourceRefreshDailySchedule struct {
	StartTime *TimeOfDay `json:"startTime,omitempty"` // Start time of the daily schedule
}

// DataSourceRefreshWeeklySchedule represents a weekly schedule for data source refresh.
type DataSourceRefreshWeeklySchedule struct {
	DaysOfWeek []DayOfWeek `json:"daysOfWeek,omitempty"` // Days of the week for the weekly schedule
	StartTime  *TimeOfDay  `json:"startTime,omitempty"`  // Start time of the weekly schedule
}

// DataSourceRefreshMonthlySchedule represents a monthly schedule for data source refresh.
type DataSourceRefreshMonthlySchedule struct {
	DaysOfMonth []int      `json:"daysOfMonth,omitempty"` // Days of the month for the monthly schedule
	StartTime   *TimeOfDay `json:"startTime,omitempty"`   // Start time of the monthly schedule
}

// DayOfWeek represents a day of the week.
type DayOfWeek int

// TimeOfDay represents a time of a day.
type TimeOfDay struct {
	Hours   int `json:"hours,omitempty"`   // Hours of the time of day
	Minutes int `json:"minutes,omitempty"` // Minutes of the time of day
	Nanos   int `json:"nanos,omitempty"`   // Nanoseconds of the time of day
	Seconds int `json:"seconds,omitempty"` // Seconds of the time of day
}

// Interval represents a time interval.
type Interval struct {
	EndTime   string `json:"endTime,omitempty"`   // End time of the interval
	StartTime string `json:"startTime,omitempty"` // Start time of the interval
}

// DeveloperMetadataLocationType represents the type of location on which developer metadata may be associated.
type DeveloperMetadataLocationType int

// DeveloperMetadata represents metadata associated with a developer.
type DeveloperMetadata struct {
	Location      *DeveloperMetadataLocation `json:"location,omitempty"`      // Location of the metadata
	MetadataID    int                        `json:"metadataId,omitempty"`    // ID of the metadata
	MetadataKey   string                     `json:"metadataKey,omitempty"`   // Key of the metadata
	MetadataValue string                     `json:"metadataValue,omitempty"` // Value of the metadata
	Visibility    string                     `json:"visibility,omitempty"`    // Visibility of the metadata
}

type DeveloperMetadataLocation interface{}

// NamedRange represents a named range in a spreadsheet.
type NamedRange struct {
	Name         string     `json:"name,omitempty"`         // Name of the range
	NamedRangeID string     `json:"namedRangeId,omitempty"` // ID of the named range
	Range        *GridRange `json:"range,omitempty"`        // Range the named range refers to
}

// SpreadsheetProperties represents properties of a spreadsheet.
type SpreadsheetProperties struct {
	AutoRecalc                   string                        `json:"autoRecalc,omitempty"`                   // Recalculation interval setting
	DefaultFormat                *CellFormat                   `json:"defaultFormat,omitempty"`                // Default cell format for the spreadsheet
	IterativeCalculationSettings *IterativeCalculationSettings `json:"iterativeCalculationSettings,omitempty"` // Iterative calculation settings for the spreadsheet
	Locale                       string                        `json:"locale,omitempty"`                       // Locale of the spreadsheet
	SpreadsheetTheme             *SpreadsheetTheme             `json:"spreadsheetTheme,omitempty"`             // Theme applied to the spreadsheet
	TimeZone                     string                        `json:"timeZone,omitempty"`                     // Timezone of the spreadsheet
	Title                        string                        `json:"title,omitempty"`                        // Title of the spreadsheet
}

// Sheet represents a sheet within a spreadsheet.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/sheets
type Sheet struct {
	BandedRanges       []BandedRange           `json:"bandedRanges,omitempty"`       // List of banded (alternate colors) ranges in a sheet
	BasicFilter        *BasicFilter            `json:"basicFilter,omitempty"`        // Basic filter applied to the sheet data
	Charts             []EmbeddedChart         `json:"charts,omitempty"`             // List of charts in a sheet
	ColumnGroups       []DimensionGroup        `json:"columnGroups,omitempty"`       // List of column groups in a sheet
	ConditionalFormats []ConditionalFormatRule `json:"conditionalFormats,omitempty"` // List of conditional formatting rules in a sheet
	Data               []GridData              `json:"data,omitempty"`               // Data in the grid
	DeveloperMetadata  []DeveloperMetadata     `json:"developerMetadata,omitempty"`  // Developer metadata in a sheet
	FilterViews        []FilterView            `json:"filterViews,omitempty"`        // List of filter views in a sheet
	Merges             []GridRange             `json:"merges,omitempty"`             // List of merges in a sheet
	Properties         *SheetProperties        `json:"properties,omitempty"`         // Properties of a sheet
	ProtectedRanges    []ProtectedRange        `json:"protectedRanges,omitempty"`    // List of protected ranges in a sheet
	RowGroups          []DimensionGroup        `json:"rowGroups,omitempty"`          // List of row groups in a sheet
	Slicers            []Slicer                `json:"slicers,omitempty"`            // List of slicers in a sheet
}

// SheetProperties represents properties of a sheet.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/sheets#sheetproperties
type SheetProperties struct {
	DataSourceSheetProperties *DataSourceSheetProperties `json:"dataSourceSheetProperties,omitempty"` // Properties of a data source sheet
	GridProperties            *GridProperties            `json:"gridProperties,omitempty"`            // Properties of a grid
	Index                     int                        `json:"index,omitempty"`                     // Index of the sheet
	Hidden                    bool                       `json:"hidden,omitempty"`                    // Indicates whether the sheet is hidden
	RightToLeft               bool                       `json:"rightToLeft,omitempty"`               // Indicates whether the sheet is right-to-left
	SheetID                   int                        `json:"sheetId,omitempty"`                   // ID of the sheet
	SheetType                 SheetType                  `json:"sheetType,omitempty"`                 // Type of the sheet
	TabColor                  *Color                     `json:"tabColor,omitempty"`                  // Color of the tab
	TabColorStyle             *ColorStyle                `json:"tabColorStyle,omitempty"`             // Style of the tab color
	Title                     string                     `json:"title,omitempty"`                     // Title of the sheet
}

// DataSourceSheetProperties represents the properties of a data source sheet.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/sheets#datasourcesheetproperties
type DataSourceSheetProperties struct {
	DataSourceID        string               `json:"dataSourceId,omitempty"`        // ID of the data source
	Columns             []DataSourceColumn   `json:"columns,omitempty"`             // List of data source columns
	DataExecutionStatus *DataExecutionStatus `json:"dataExecutionStatus,omitempty"` // Status of the data execution
}

type DataExecutionStatus interface{}

// GridProperties represents the properties of a grid.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/sheets#gridproperties
type GridProperties struct {
	RowCount                int  `json:"rowCount,omitempty"`                // Number of rows in the grid
	ColumnCount             int  `json:"columnCount,omitempty"`             // Number of columns in the grid
	FrozenRowCount          int  `json:"frozenRowCount,omitempty"`          // Number of frozen rows in the grid
	FrozenColumnCount       int  `json:"frozenColumnCount,omitempty"`       // Number of frozen columns in the grid
	HideGridlines           bool `json:"hideGridlines,omitempty"`           // Whether to hide gridlines
	RowGroupControlAfter    bool `json:"rowGroupControlAfter,omitempty"`    // Whether to control the row group after the grid
	ColumnGroupControlAfter bool `json:"columnGroupControlAfter,omitempty"` // Whether to control the column group after the grid
}

// RowData represents data in a row.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/sheets#rowdata
type RowData struct {
	Values []CellData `json:"values,omitempty"` // List of cell data
}

// CellData represents data in a cell.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#celldata
type CellData struct {
	UserEnteredValue  *ExtendedValue      `json:"userEnteredValue,omitempty"`  // User-entered value
	EffectiveValue    *ExtendedValue      `json:"effectiveValue,omitempty"`    // Effective value
	FormattedValue    string              `json:"formattedValue,omitempty"`    // Formatted value
	UserEnteredFormat *CellFormat         `json:"userEnteredFormat,omitempty"` // User-entered format
	EffectiveFormat   *CellFormat         `json:"effectiveFormat,omitempty"`   // Effective format
	Hyperlink         string              `json:"hyperlink,omitempty"`         // Hyperlink in the cell
	Note              string              `json:"note,omitempty"`              // Note in the cell
	TextFormatRuns    []TextFormatRun     `json:"textFormatRuns,omitempty"`    // List of text format runs
	DataValidation    *DataValidationRule `json:"dataValidation,omitempty"`    // Data validation rule
	PivotTable        *PivotTable         `json:"pivotTable,omitempty"`        // Pivot table
	DataSourceTable   *DataSourceTable    `json:"dataSourceTable,omitempty"`   // Data source table
	DataSourceFormula *DataSourceFormula  `json:"dataSourceFormula,omitempty"` // Data source formula
}

type TextFormatRun interface{}
type DataSourceTable interface{}
type DataSourceFormula interface{}

// ExtendedValue represents a user-entered value.
// https://developers.google.com/sheets/api/reference/rest/v4/sheets#ExtendedValue
type ExtendedValue struct {
	NumberValue  float64     `json:"numberValue,omitempty"`  // Number value entered by the user
	StringValue  string      `json:"stringValue,omitempty"`  // String value entered by the user
	BoolValue    bool        `json:"boolValue,omitempty"`    // Boolean value entered by the user
	FormulaValue string      `json:"formulaValue,omitempty"` // Formula value entered by the user
	ErrorValue   *ErrorValue `json:"errorValue,omitempty"`   // Error value entered by the user
}

type ErrorValue interface{}

// DataValidationRule represents a rule for data validation.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/other#ExtendedValue
type DataValidationRule struct {
	Condition    bool   `json:"condition,omitempty"`    // Condition for the data validation
	InputMessage string `json:"inputMessage,omitempty"` // Input message for the data validation
	Strict       bool   `json:"strict,omitempty"`       // Whether the data validation is strict
	ShowCustomUi bool   `json:"showCustomUi,omitempty"` // Whether to show a custom UI for the data validation
}

// PivotTable represents a pivot table.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/pivot-tables
type PivotTable struct {
	Rows                []PivotGroup                `json:"rows,omitempty"`                // Rows of the pivot table
	Columns             []PivotGroup                `json:"columns,omitempty"`             // Columns of the pivot table
	Criteria            map[int]PivotFilterCriteria `json:"criteria,omitempty"`            // Criteria for the pivot table
	FilterSpecs         []PivotFilterSpec           `json:"filterSpecs,omitempty"`         // Filter specifications for the pivot table
	Values              []PivotValue                `json:"values,omitempty"`              // Values of the pivot table
	ValueLayout         PivotValueLayout            `json:"valueLayout,omitempty"`         // Layout of the pivot table values
	DataExecutionStatus *DataExecutionStatus        `json:"dataExecutionStatus,omitempty"` // Status of the data execution
	Source              *GridRange                  `json:"source,omitempty"`              // Source data range for the pivot table
	DataSourceID        string                      `json:"dataSourceId,omitempty"`        // ID of the data source
}

// PivotGroup represents a group in a pivot table.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/pivot-tables#pivotgroup
type PivotGroup struct {
	ShowTotals                bool                       `json:"showTotals,omitempty"`                // Whether to show totals
	ValueMetadata             []PivotGroupValueMetadata  `json:"valueMetadata,omitempty"`             // Metadata for pivot group values
	SortOrder                 SortOrder                  `json:"sortOrder,omitempty"`                 // Sort order of the pivot group
	ValueBucket               *PivotGroupSortValueBucket `json:"valueBucket,omitempty"`               // Bucket for sorting pivot group values
	RepeatHeadings            bool                       `json:"repeatHeadings,omitempty"`            // Whether to repeat headings
	Label                     string                     `json:"label,omitempty"`                     // Label for the pivot group
	GroupRule                 *PivotGroupRule            `json:"groupRule,omitempty"`                 // Rule for grouping
	GroupLimit                *PivotGroupLimit           `json:"groupLimit,omitempty"`                // Limit for the pivot group
	SourceColumnOffset        int                        `json:"sourceColumnOffset,omitempty"`        // Offset for source column
	DataSourceColumnReference *DataSourceColumnReference `json:"dataSourceColumnReference,omitempty"` // Reference for data source column
}

type PivotFilterCriteria interface{}
type PivotFilterSpec interface{}
type PivotValue interface{}
type PivotValueLayout interface{}
type PivotGroupValueMetadata interface{}
type SortOrder interface{}
type PivotGroupSortValueBucket interface{}
type PivotGroupRule interface{}
type PivotGroupLimit interface{}

type SheetType interface{}

// Color represents the color object
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/other#color
type Color struct {
	Alpha float64 `json:"alpha,omitempty"` // Alpha represents the alpha channel value of the color, which should be between 0 and 1 (inclusive)
	Blue  float64 `json:"blue,omitempty"`  // Blue represents the blue component of the color, which should be between 0 and 1 (inclusive)
	Green float64 `json:"green,omitempty"` // Green represents the green component of the color, which should be between 0 and 1 (inclusive)
	Red   float64 `json:"red,omitempty"`   // Red represents the red component of the color, which should be between 0 and 1 (inclusive)
}

// ColorStyle represents the color style object
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/other#colorstyle
type ColorStyle struct {
	RGBColor   *Color `json:"rgbColor,omitempty"`   // RGBColor represents the RGB color of the style
	ThemeColor string `json:"themeColor,omitempty"` // ThemeColor represents the theme color type of the style
}

type RecalculationInterval interface{}

// CellFormat represents the formatting of a cell.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#CellFormat
type CellFormat struct {
	BackgroundColor      *Color        `json:"backgroundColor,omitempty"`      // Background color of the cell
	BackgroundColorStyle *ColorStyle   `json:"backgroundColorStyle,omitempty"` // Style of the background color of the cell
	Borders              *Borders      `json:"borders,omitempty"`              // Borders of the cell
	HorizontalAlignment  string        `json:"horizontalAlignment,omitempty"`  // Horizontal alignment of the cell
	HyperlinkDisplayType string        `json:"hyperlinkDisplayType,omitempty"` // Display type of hyperlinks in the cell
	NumberFormat         *NumberFormat `json:"numberFormat,omitempty"`         // Number format of the cell
	Padding              *Padding      `json:"padding,omitempty"`              // Padding of the cell
	TextDirection        string        `json:"textDirection,omitempty"`        // Text direction in the cell
	TextFormat           *TextFormat   `json:"textFormat,omitempty"`           // Text format of the cell
	TextRotation         *TextRotation `json:"textRotation,omitempty"`         // Text rotation in the cell
	VerticalAlignment    string        `json:"verticalAlignment,omitempty"`    // Vertical alignment of the cell
	WrapStrategy         string        `json:"wrapStrategy,omitempty"`         // Wrap strategy of the cell
}

type Borders interface{}
type NumberFormat interface{}
type Padding interface{}

// TextFormat represents the text format
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/other#textformat
type TextFormat struct {
	ForegroundColor      *Color      `json:"foregroundColor,omitempty"`      // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#color
	ForegroundColorStyle *ColorStyle `json:"foregroundColorStyle,omitempty"` // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#colorstyle
	FontFamily           string      `json:"fontFamily,omitempty"`           // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#fontfamily
	FontSize             int         `json:"fontSize,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#fontsize
	Bold                 bool        `json:"bold,omitempty"`                 // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#bold
	Italic               bool        `json:"italic,omitempty"`               // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#italic
	Strikethrough        bool        `json:"strikethrough,omitempty"`        // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#strikethrough
	Underline            bool        `json:"underline,omitempty"`            // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#underline
	Link                 *SheetLink  `json:"link,omitempty"`                 // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#link
}

type TextRotation interface{}

// Link represents the link object
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#link
type SheetLink struct {
	// URI represents the uniform resource identifier
	Uri string `json:"uri,omitempty"`
}

// IterativeCalculationSettings represents the settings for iterative calculations.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets#IterativeCalculationSettings
type IterativeCalculationSettings struct {
	ConvergenceThreshold float64 `json:"convergenceThreshold,omitempty"` // Convergence threshold for iterative calculations
	MaxIterations        int     `json:"maxIterations,omitempty"`        // Maximum number of iterations for calculations
}

// ThemeColorPair represents a pair of theme color.
type ThemeColorPair struct {
	// ...
}

// SpreadsheetTheme represents a theme of a spreadsheet.
type SpreadsheetTheme struct {
	PrimaryFontFamily string           `json:"primaryFontFamily,omitempty"` // Primary font family of the theme
	ThemeColors       []ThemeColorPair `json:"themeColors,omitempty"`       // Theme color pairs
}

// GridData represents the data in a grid (or sheet).
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/sheets#griddata
type GridData struct {
	ColumnMetadata []DimensionProperties `json:"columnMetadata,omitempty"` // Metadata about the columns in the grid
	RowData        []RowData             `json:"rowData,omitempty"`        // The actual data in the rows of the grid
	RowMetadata    []DimensionProperties `json:"rowMetadata,omitempty"`    // Metadata about the rows in the grid
	StartColumn    int                   `json:"startColumn,omitempty"`    // Starting column index of the grid
	StartRow       int                   `json:"startRow,omitempty"`       // Starting row index of the grid
}

// DimensionProperties represents properties of dimensions within a sheet.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/sheets#dimensionproperties
type DimensionProperties struct {
	HiddenByFilter            bool                       `json:"hiddenByFilter,omitempty"`            // Indicates whether this dimension is hidden by a filter.
	HiddenByUser              bool                       `json:"hiddenByUser,omitempty"`              // Indicates whether this dimension is hidden by the user.
	PixelSize                 int                        `json:"pixelSize,omitempty"`                 // The size of the dimension.
	DeveloperMetadata         []*DeveloperMetadata       `json:"developerMetadata,omitempty"`         // Metadata about this dimension.
	DataSourceColumnReference *DataSourceColumnReference `json:"dataSourceColumnReference,omitempty"` // The reference to the data source column.
}

// GridRange represents a range in a grid (or sheet).
type GridRange struct {
	EndColumnIndex   int `json:"endColumnIndex,omitempty"`   // Ending column index of the range
	EndRowIndex      int `json:"endRowIndex,omitempty"`      // Ending row index of the range
	SheetID          int `json:"sheetId,omitempty"`          // ID of the sheet where the range is found
	StartColumnIndex int `json:"startColumnIndex,omitempty"` // Starting column index of the range
	StartRowIndex    int `json:"startRowIndex,omitempty"`    // Starting row index of the range
}

// BooleanRule represents a boolean rule for conditional formatting.
type BooleanRule struct {
	// ...
}

// GradientRule represents a gradient rule for conditional formatting.
type GradientRule struct {
	// ...
}

// ConditionalFormatRule represents a conditional formatting rule.
type ConditionalFormatRule struct {
	BooleanRule  *BooleanRule  `json:"booleanRule,omitempty"`  // Boolean rule for the conditional format
	GradientRule *GradientRule `json:"gradientRule,omitempty"` // Gradient rule for the conditional format
	Ranges       []GridRange   `json:"ranges,omitempty"`       // Ranges that the conditional format rule is applied to
}

// SortSpec represents a sort specification.
type SortSpec struct {
	// ...
}

// FilterCriteria represents filter criteria.
type FilterCriteria struct {
	// ...
}

// FilterSpec represents a filter specification.
type FilterSpec struct {
	// ...
}

// FilterView represents a filter view.
type FilterView struct {
	Criteria     map[string]FilterCriteria `json:"criteria,omitempty"`     // Criteria of the filter view
	FilterSpecs  []FilterSpec              `json:"filterSpecs,omitempty"`  // Specifications of the filter view
	FilterViewID int                       `json:"filterViewId,omitempty"` // ID of the filter view
	NamedRangeID string                    `json:"namedRangeId,omitempty"` // ID of the named range of the filter view
	Range        *GridRange                `json:"range,omitempty"`        // Range of the filter view
	SortSpecs    []SortSpec                `json:"sortSpecs,omitempty"`    // Sort specifications of the filter view
	Title        string                    `json:"title,omitempty"`        // Title of the filter view
}

// ProtectedRange represents a range that is protected.
type ProtectedRange struct {
	Description           string      `json:"description,omitempty"`           // Description of the protected range
	Editors               *Editors    `json:"editors,omitempty"`               // Editors of the protected range
	NamedRangeID          string      `json:"namedRangeId,omitempty"`          // ID of the named range
	ProtectedRangeID      int         `json:"protectedRangeId,omitempty"`      // ID of the protected range
	Range                 *GridRange  `json:"range,omitempty"`                 // Range that is protected
	RequestingUserCanEdit bool        `json:"requestingUserCanEdit,omitempty"` // Whether the requesting user can edit the protected range
	UnprotectedRanges     []GridRange `json:"unprotectedRanges,omitempty"`     // Unprotected ranges within the protected range
	WarningOnly           bool        `json:"warningOnly,omitempty"`           // Whether the protected range is warning only
}

type Editors interface{}

// BasicFilter represents a basic filter.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/sheets#basicfilter
type BasicFilter struct {
	Criteria    map[string]FilterCriteria `json:"criteria,omitempty"`    // Criteria of the basic filter
	FilterSpecs []FilterSpec              `json:"filterSpecs,omitempty"` // Specifications of the basic filter
	Range       *GridRange                `json:"range,omitempty"`       // Range of the basic filter
	SortSpecs   []SortSpec                `json:"sortSpecs,omitempty"`   // Sort specifications of the basic filter
}

// EmbeddedChart represents an embedded chart.
type EmbeddedChart struct {
	Border   *EmbeddedObjectBorder   `json:"border,omitempty"`   // Border of the embedded chart
	ChartID  int                     `json:"chartId,omitempty"`  // ID of the embedded chart
	Position *EmbeddedObjectPosition `json:"position,omitempty"` // Position of the embedded chart
	Spec     *ChartSpec              `json:"spec,omitempty"`     // Specification of the embedded chart
}

type EmbeddedObjectBorder interface{}
type EmbeddedObjectPosition interface{}
type ChartSpec interface{}

// BandedRange represents a banded (alternating colors) range.
type BandedRange struct {
	BandedRangeID    int                `json:"bandedRangeId,omitempty"`    // ID of the banded range
	ColumnProperties *BandingProperties `json:"columnProperties,omitempty"` // Properties of the columns in the banded range
	Range            *GridRange         `json:"range,omitempty"`            // Range of the banded range
	RowProperties    *BandingProperties `json:"rowProperties,omitempty"`    // Properties of the rows in the banded range
}

type BandingProperties interface{}

// DimensionGroup represents a group of dimensions.
type DimensionGroup struct {
	Collapsed bool            `json:"collapsed,omitempty"` // Whether the dimension group is collapsed
	Depth     int             `json:"depth,omitempty"`     // Depth of the dimension group
	Range     *DimensionRange `json:"range,omitempty"`     // Range of the dimension group
}

// DimensionRange represents the dimension range object
// https://developers.google.com/sheets/api/reference/rest/v4/DimensionRange
type DimensionRange struct {
	Dimension  string `json:"dimension,omitempty"`  // // Dimension represents the dimension type, which could be ROWS or COLUMNS
	EndIndex   int    `json:"endIndex,omitempty"`   // EndIndex represents the end index of the dimension
	SheetID    int    `json:"sheetId,omitempty"`    // SheetID represents the ID of the sheet
	StartIndex int    `json:"startIndex,omitempty"` // StartIndex represents the start index of the dimension
}

// Slicer represents a slicer.
type Slicer struct {
	Position *EmbeddedObjectPosition `json:"position,omitempty"` // Position of the slicer
	SlicerID int                     `json:"slicerId,omitempty"` // ID of the slicer
	Spec     *SlicerSpec             `json:"spec,omitempty"`     // Specification of the slicer
}

// SlicerSpec represents the specification for a slicer.
type SlicerSpec struct {
	DataRange            *GridRange      `json:"dataRange,omitempty"`            // Range for the data
	FilterCriteria       *FilterCriteria `json:"filterCriteria,omitempty"`       // Criteria for filtering
	ColumnIndex          int             `json:"columnIndex,omitempty"`          // Index of the column
	ApplyToPivotTables   bool            `json:"applyToPivotTables,omitempty"`   // Whether to apply to pivot tables
	Title                string          `json:"title,omitempty"`                // Title for the slicer
	TextFormat           *TextFormat     `json:"textFormat,omitempty"`           // Format for the text
	BackgroundColor      *Color          `json:"backgroundColor,omitempty"`      // Color for the background
	BackgroundColorStyle *ColorStyle     `json:"backgroundColorStyle,omitempty"` // Style for the background color
	HorizontalAlignment  HorizontalAlign `json:"horizontalAlignment,omitempty"`  // Horizontal alignment
}

type HorizontalAlign interface{}

// SheetBatchRequest represents a batch of updates to apply to a spreadsheet.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#Request
type SheetBatchRequest struct {
	IncludeSpreadsheetInResponse bool            `json:"includeSpreadsheetInResponse,omitempty"` // Determines if the update response should include the spreadsheet resource
	Requests                     []*SheetRequest `json:"requests,omitempty"`                     // List of requests to be processed by the API
	ResponseIncludeGridData      bool            `json:"responseIncludeGridData,omitempty"`      // Determines if the response should include grid data
	ResponseRanges               []string        `json:"responseRanges,omitempty"`               // The ranges that are returned in the response
}

// SheetRequest represents a single kind of update to apply to a spreadsheet.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#Request
type SheetRequest struct {
	AddChart                     interface{}                       `json:"addChart,omitempty"`                     // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#addchartrequest
	AddConditionalFormatRule     interface{}                       `json:"addConditionalFormatRule,omitempty"`     // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#addconditionalformatrulerequest
	AddDataSource                interface{}                       `json:"addDataSource,omitempty"`                // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#adddatasourcerequest
	AddDimensionGroup            interface{}                       `json:"addDimensionGroup,omitempty"`            // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#adddimensiongrouprequest
	AddFilterView                interface{}                       `json:"addFilterView,omitempty"`                // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#addfilterviewrequest
	AddNamedRange                interface{}                       `json:"addNamedRange,omitempty"`                // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#addnamedrangerequest
	AddProtectedRange            interface{}                       `json:"addProtectedRange,omitempty"`            // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#addprotectedrangerequest
	AddSheet                     interface{}                       `json:"addSheet,omitempty"`                     // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#addsheetrequest
	AddSlicer                    interface{}                       `json:"addSlicer,omitempty"`                    // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#addslicerrequest
	AppendCells                  interface{}                       `json:"appendCells,omitempty"`                  // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#appendcellsrequest
	AppendDimension              interface{}                       `json:"appendDimension,omitempty"`              // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#appenddimensionrequest
	AutoFill                     interface{}                       `json:"autoFill,omitempty"`                     // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#autofillrequest
	AutoResizeDimensions         interface{}                       `json:"autoResizeDimensions,omitempty"`         // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#autoresizedimensionsrequest
	ClearBasicFilter             interface{}                       `json:"clearBasicFilter,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#clearbasicfilterrequest
	CopyPaste                    interface{}                       `json:"copyPaste,omitempty"`                    // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#copypasterequest
	CreateDeveloperMetadata      interface{}                       `json:"createDeveloperMetadata,omitempty"`      // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#createdevelopermetadatarequest
	CutPaste                     interface{}                       `json:"cutPaste,omitempty"`                     // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#cutpasterequest
	DeleteBanding                interface{}                       `json:"deleteBanding,omitempty"`                // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deletebandingrequest
	DeleteConditionalFormatRule  interface{}                       `json:"deleteConditionalFormatRule,omitempty"`  // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deleteconditionalformatrulerequest
	DeleteDataSource             interface{}                       `json:"deleteDataSource,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deletedatasourcerequest
	DeleteDeveloperMetadata      interface{}                       `json:"deleteDeveloperMetadata,omitempty"`      // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deletedevelopermetadatarequest
	DeleteDimension              interface{}                       `json:"deleteDimension,omitempty"`              // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deletedimensionrequest
	DeleteDimensionGroup         interface{}                       `json:"deleteDimensionGroup,omitempty"`         // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deletedimensiongrouprequest
	DeleteDuplicates             interface{}                       `json:"deleteDuplicates,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deleteduplicatesrequest
	DeleteEmbeddedObject         interface{}                       `json:"deleteEmbeddedObject,omitempty"`         // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deleteembeddedobjectrequest
	DeleteFilterView             interface{}                       `json:"deleteFilterView,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deletefilterviewrequest
	DeleteNamedRange             interface{}                       `json:"deleteNamedRange,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deletenamedrangerequest
	DeleteProtectedRange         interface{}                       `json:"deleteProtectedRange,omitempty"`         // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deleteprotectedrangerequest
	DeleteRange                  interface{}                       `json:"deleteRange,omitempty"`                  // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deleterangerequest
	DeleteSheet                  interface{}                       `json:"deleteSheet,omitempty"`                  // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#deletesheetrequest
	DuplicateFilterView          interface{}                       `json:"duplicateFilterView,omitempty"`          // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#duplicatefilterviewrequest
	DuplicateSheet               interface{}                       `json:"duplicateSheet,omitempty"`               // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#duplicatesheetrequest
	FindReplace                  interface{}                       `json:"findReplace,omitempty"`                  // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#findreplacerequest
	InsertDimension              interface{}                       `json:"insertDimension,omitempty"`              // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#insertdimensionrequest
	InsertRange                  interface{}                       `json:"insertRange,omitempty"`                  // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#insertrangerequest
	MergeCells                   interface{}                       `json:"mergeCells,omitempty"`                   // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#mergecellsrequest
	MoveDimension                interface{}                       `json:"moveDimension,omitempty"`                // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#movedimensionrequest
	PasteData                    interface{}                       `json:"pasteData,omitempty"`                    // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#pastedatarequest
	RandomizeRange               interface{}                       `json:"randomizeRange,omitempty"`               // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#randomizerangerequest
	RefreshDataSource            interface{}                       `json:"refreshDataSource,omitempty"`            // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#refreshdatasourcerequest
	RepeatCell                   *RepeatCellRequest                `json:"repeatCell,omitempty"`                   // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#repeatcellrequest
	SetBasicFilter               *SetBasicFilterRequest            `json:"setBasicFilter,omitempty"`               // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#setbasicfilterrequest
	SetDataValidation            interface{}                       `json:"setDataValidation,omitempty"`            // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#setdatavalidationrequest
	SortRange                    interface{}                       `json:"sortRange,omitempty"`                    // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#sortrangerequest
	TextToColumns                interface{}                       `json:"textToColumns,omitempty"`                // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#texttocolumnsrequest
	TrimWhitespace               interface{}                       `json:"trimWhitespace,omitempty"`               // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#trimwhitespacerequest
	UpdateBanding                interface{}                       `json:"updateBanding,omitempty"`                // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatebandingrequest
	UpdateBorders                interface{}                       `json:"updateBorders,omitempty"`                // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatebordersrequest
	UpdateCells                  interface{}                       `json:"updateCells,omitempty"`                  // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatecellsrequest
	UpdateChartSpec              interface{}                       `json:"updateChartSpec,omitempty"`              // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatechartspecrequest
	UpdateConditionalFormatRule  interface{}                       `json:"updateConditionalFormatRule,omitempty"`  // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updateconditionalformatrulerequest
	UpdateDataSource             interface{}                       `json:"updateDataSource,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatedatasourcerequest
	UpdateDeveloperMetadata      interface{}                       `json:"updateDeveloperMetadata,omitempty"`      // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatedevelopermetadatarequest
	UpdateDimensionGroup         interface{}                       `json:"updateDimensionGroup,omitempty"`         // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatedimensiongrouprequest
	UpdateDimensionProperties    *UpdateDimensionPropertiesRequest `json:"updateDimensionProperties,omitempty"`    // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatedimensionpropertiesrequest
	UpdateEmbeddedObjectBorder   interface{}                       `json:"updateEmbeddedObjectBorder,omitempty"`   // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updateembeddedobjectborderrequest
	UpdateEmbeddedObjectPosition interface{}                       `json:"updateEmbeddedObjectPosition,omitempty"` // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updateembeddedobjectpositionrequest
	UpdateFilterView             interface{}                       `json:"updateFilterView,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatefilterviewrequest
	UpdateNamedRange             interface{}                       `json:"updateNamedRange,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatenamedrangerequest
	UpdateProtectedRange         interface{}                       `json:"updateProtectedRange,omitempty"`         // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updateprotectedrangerequest
	UpdateSheetProperties        interface{}                       `json:"updateSheetProperties,omitempty"`        // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatesheetpropertiesrequest
	UpdateSlicerSpec             interface{}                       `json:"updateSlicerSpec,omitempty"`             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updateslicerspecrequest
}

// AutoResizeDimensionsRequest represents a request to auto resize dimensions.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#autoresizedimensionsrequest
type AutoResizeDimensionsRequest struct {
	Dimensions                interface{} `json:"dimensions,omitempty"`                // The dimensions to resize on the sheet
	DataSourceSheetDimensions interface{} `json:"dataSourceSheetDimensions,omitempty"` // The dimensions to resize on the data source sheet
}

// DataSourceSheetDimensionRange represents the data source sheet dimension range object
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets#DataDourceSheetDimensionRange
type DataSourceSheetDimensionRange struct {
	ColumnReferences []DataSourceColumnReference `json:"columnReferences,omitempty"` // ColumnReferences represents the list of data source column references
	SheetID          int                         `json:"sheetId,omitempty"`          // SheetID represents the ID of the sheet
}

// SetBasicFilterRequest represents the request to set a basic filter
// Source: https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#setbasicfilterrequest
type SetBasicFilterRequest struct {
	Filter *BasicFilter `json:"filter,omitempty"` // Filter is the basic filter to be set
}

// RepeatCellRequest sets the values of cells in a range to a set of values in a certain pattern.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#repeatcellrequest
type RepeatCellRequest struct {
	Range  *GridRange `json:"range,omitempty"`  // The range within the sheet that the repeated cell will cover
	Cell   *CellData  `json:"cell,omitempty"`   // The data to be repeated in the range
	Fields string     `json:"fields,omitempty"` // The fields of the cell to be updated
}

// UpdateDimensionPropertiesRequest represents the request to update dimension properties
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/request#updatedimensionpropertiesrequest
type UpdateDimensionPropertiesRequest struct {
	Properties           *DimensionProperties           `json:"properties,omitempty"`           // Dimension properties to update
	Fields               string                         `json:"fields,omitempty"`               // Fields to update
	Range                *DimensionRange                `json:"range,omitempty"`                // Range of the dimension to update
	DataSourceSheetRange *DataSourceSheetDimensionRange `json:"dataSourceSheetRange,omitempty"` // Range of the dataSource sheet dimension to update
}

// END OF SPREADSHEET STRUCTS
//---------------------------------------------------------------------------------------

// ### User Structs
// ---------------------------------------------------------------------------------------
type User struct {
	AgreedToTerms              bool           `json:"agreedToTerms,omitempty"`              // User's agreement to terms status
	Aliases                    []string       `json:"aliases,omitempty"`                    // User aliases
	Archived                   bool           `json:"archived,omitempty"`                   // User's archival status
	ChangePasswordAtNextLogin  bool           `json:"changePasswordAtNextLogin,omitempty"`  // User's change password at next login status
	CreationTime               string         `json:"creationTime,omitempty"`               // User's creation time
	CustomerID                 string         `json:"customerId,omitempty"`                 // User's customer ID
	DeletionTime               string         `json:"deletionTime,omitempty"`               // User's deletion time
	Emails                     []Email        `json:"emails,omitempty"`                     // User's emails
	Etag                       string         `json:"etag,omitempty"`                       // ETag of the user
	ExternalIds                []ExternalID   `json:"externalIds,omitempty"`                // User's external IDs
	Gender                     Gender         `json:"gender,omitempty"`                     // User's gender
	HashFunction               string         `json:"hashFunction,omitempty"`               // User's hash function
	ID                         string         `json:"id,omitempty"`                         // User's ID
	IncludeInGlobalAddressList bool           `json:"includeInGlobalAddressList,omitempty"` // User's inclusion status in global address list
	IsAdmin                    bool           `json:"isAdmin,omitempty"`                    // User's admin status
	IsDelegatedAdmin           bool           `json:"isDelegatedAdmin,omitempty"`           // User's delegated admin status
	IsEnforcedIn2Sv            bool           `json:"isEnforcedIn2Sv,omitempty"`            // User's 2SV enforcement status
	IsEnrolledIn2Sv            bool           `json:"isEnrolledIn2Sv,omitempty"`            // User's 2SV enrolment status
	IsMailboxSetup             bool           `json:"isMailboxSetup,omitempty"`             // User's mailbox setup status
	Ims                        []IM           `json:"ims,omitempty"`                        // User's instant messaging addresses
	IpWhitelisted              bool           `json:"ipWhitelisted,omitempty"`              // User's IP whitelist status
	Kind                       string         `json:"kind,omitempty"`                       // The type of the API resource
	Languages                  []Language     `json:"languages,omitempty"`                  // User's languages
	LastLoginTime              string         `json:"lastLoginTime,omitempty"`              // User's last login time
	Locations                  []UserLocation `json:"locations,omitempty"`                  // User's locations
	Name                       Name           `json:"name,omitempty"`                       // User's name
	NonEditableAliases         []string       `json:"nonEditableAliases,omitempty"`         // User's non-editable aliases
	Notes                      Note           `json:"notes,omitempty"`                      // User's notes
	OrgUnitPath                string         `json:"orgUnitPath,omitempty"`                // User's organizational unit path
	Organizations              []Organization `json:"organizations,omitempty"`              // User's organizations
	Password                   Password       `json:"password,omitempty"`                   // User's password
	Phones                     []Phone        `json:"phones,omitempty"`                     // A list of the user's phone numbers
	PosixAccounts              []POSIXAccount `json:"posixAccounts,omitempty"`              // The list of POSIX account information for the user
	PrimaryEmail               string         `json:"primaryEmail,omitempty"`               // User's primary email
	RecoveryEmail              string         `json:"recoveryEmail,omitempty"`              // User's recovery email
	RecoveryPhone              string         `json:"recoveryPhone,omitempty"`              // User's recovery phone number
	Relations                  []Relation     `json:"relations,omitempty"`                  // User's relations
	Roles                      interface{}    `json:"roles,omitempty"`                      // User's roles
	SshPublicKeys              []SSHPublicKey `json:"sshPublicKeys,omitempty"`              // A list of SSH public keys
	Suspended                  bool           `json:"suspended,omitempty"`                  // User's suspension status
	SuspensionReason           string         `json:"suspensionReason,omitempty"`           // User's suspension reason
	ThumbnailPhotoEtag         string         `json:"thumbnailPhotoEtag,omitempty"`         // User's thumbnail photo ETag
	ThumbnailPhotoUrl          string         `json:"thumbnailPhotoUrl,omitempty"`          // User's thumbnail photo URL
	Websites                   []Website      `json:"websites,omitempty"`                   // The list of the user's websites
}

type Email struct {
	Address    string `json:"address,omitempty"`    // The user's email address
	CustomType string `json:"customType,omitempty"` // The custom value if the email address type is custom
	Primary    bool   `json:"primary,omitempty"`    // Indicator if this is the user's primary email
	Type       string `json:"type,omitempty"`       // The type of the email account
}

type ExternalID struct {
	CustomType string `json:"customType,omitempty"` // The custom value if the external ID type is custom
	Type       string `json:"type,omitempty"`       // The type of external ID
	Value      string `json:"value,omitempty"`      // The value of the external ID
}

type Gender struct {
	AddressMeAs  string `json:"addressMeAs,omitempty"`  // The proper way to refer to the profile owner by humans
	CustomGender string `json:"customGender,omitempty"` // Name of a custom gender
	Type         string `json:"type,omitempty"`         // The type of gender
}

type IM struct {
	CustomProtocol string `json:"customProtocol,omitempty"` // The custom protocol's string if the protocol value is custom_protocol
	CustomType     string `json:"customType,omitempty"`     // The custom value if the IM type is custom
	IM             string `json:"im,omitempty"`             // The user's IM network ID
	Primary        bool   `json:"primary,omitempty"`        // Indicator if this is the user's primary IM
	Protocol       string `json:"protocol,omitempty"`       // The IM protocol identifies the IM network
	Type           string `json:"type,omitempty"`           // The type of IM account
}

type Language struct {
	CustomLanguage string `json:"customLanguage,omitempty"` // User provided language name if there is no corresponding ISO 639 language code
	LanguageCode   string `json:"languageCode,omitempty"`   // ISO 639 string representation of a language
	Preference     string `json:"preference,omitempty"`     // Controls whether the specified languageCode is the user's preferred language
}

type UserLocation struct {
	Area         string `json:"area,omitempty"`         // Textual location
	BuildingId   string `json:"buildingId,omitempty"`   // Building identifier
	CustomType   string `json:"customType,omitempty"`   // The custom value if the location type is custom
	DeskCode     string `json:"deskCode,omitempty"`     // Most specific textual code of individual desk location
	FloorName    string `json:"floorName,omitempty"`    // Floor name/number
	FloorSection string `json:"floorSection,omitempty"` // Floor section
	Type         string `json:"type,omitempty"`         // The location type
}

type Name struct {
	FullName    string `json:"fullName,omitempty"`    // The user's full name
	FamilyName  string `json:"familyName,omitempty"`  // The user's last name
	GivenName   string `json:"givenName,omitempty"`   // The user's first name
	DisplayName string `json:"displayName,omitempty"` // The user's display name
}

type Note struct {
	ContentType string `json:"contentType,omitempty"` // Content type of note
	Value       string `json:"value,omitempty"`       // Contents of notes
}

type Organization struct {
	CostCenter         string `json:"costCenter,omitempty"`         // The cost center of the user's organization
	CustomType         string `json:"customType,omitempty"`         // The custom value if the organization type is custom
	Department         string `json:"department,omitempty"`         // Specifies the department within the organization
	Description        string `json:"description,omitempty"`        // The description of the organization
	Domain             string `json:"domain,omitempty"`             // The domain the organization belongs to
	FullTimeEquivalent int    `json:"fullTimeEquivalent,omitempty"` // The full-time equivalent millipercent within the organization
	Location           string `json:"location,omitempty"`           // The physical location of the organization
	Name               string `json:"name,omitempty"`               // The name of the organization
	Primary            bool   `json:"primary,omitempty"`            // Indicator if this is the user's primary organization
	Symbol             string `json:"symbol,omitempty"`             // Text string symbol of the organization
	Title              string `json:"title,omitempty"`              // The user's title within the organization
	Type               string `json:"type,omitempty"`               // The type of organization
}

type Password struct {
	Value string `json:"value,omitempty"` // The password for the user account
}

type Phone struct {
	CustomType string `json:"customType,omitempty"` // The custom value if the phone type is custom
	Primary    bool   `json:"primary,omitempty"`    // If true, this is the user's primary phone number
	Type       string `json:"type,omitempty"`       // The type of phone number
	Value      string `json:"value,omitempty"`      // A human-readable phone number
}

type POSIXAccount struct {
	AccountId           string `json:"accountId,omitempty"`           // A POSIX account field identifier
	Gecos               string `json:"gecos,omitempty"`               // The GECOS (user information) for this account
	Gid                 uint64 `json:"gid,omitempty"`                 // The default group ID
	HomeDirectory       string `json:"homeDirectory,omitempty"`       // The path to the home directory for this account
	OperatingSystemType string `json:"operatingSystemType,omitempty"` // The operating system type for this account
	Primary             bool   `json:"primary,omitempty"`             // If this is user's primary account within the SystemId
	Shell               string `json:"shell,omitempty"`               // The path to the login shell for this account
	SystemId            string `json:"systemId,omitempty"`            // System identifier for which account Username or Uid apply to
	Uid                 uint64 `json:"uid,omitempty"`                 // The POSIX compliant user ID
	Username            string `json:"username,omitempty"`            // The username of the account
}

type Relation struct {
	CustomType string `json:"customType,omitempty"` // The custom value if the relationship type is custom
	Type       string `json:"type,omitempty"`       // The type of relationship
	Value      string `json:"value,omitempty"`      // The email address of the person the user is related to
}

type SSHPublicKey struct {
	ExpirationTimeUsec int64  `json:"expirationTimeUsec,omitempty"` // An expiration time in microseconds since epoch
	Fingerprint        string `json:"fingerprint,omitempty"`        // A SHA-256 fingerprint of the SSH public key
	Key                string `json:"key,omitempty"`                // An SSH public key
}

type Website struct {
	CustomType string `json:"customType,omitempty"` // The custom value if the website type is custom
	Primary    bool   `json:"primary,omitempty"`    // If true, this is the user's primary website
	Type       string `json:"type,omitempty"`       // The type or purpose of the website
	Value      string `json:"value,omitempty"`      // The URL of the website
}

// END OF USER STRUCTS
//-----------------------------------------------------------------------------
