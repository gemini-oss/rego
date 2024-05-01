/*
# Backupify - Entities [Structs]

This package contains many structs for handling payloads / responses from Backupify's WebUI:

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/backupify/entities.go
package backupify

import (
	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ### Backupify Client Structs
// ---------------------------------------------------------------------
type Client struct {
	BaseURL     string           // BaseURL is the base URL for Backupify requests.
	HTTP        *requests.Client // HTTPClient is the client used to make HTTP requests.
	Error       string           // Error is the error message returned from the Backupify WebUI.
	Log         *log.Logger      // Log is the logger used to log messages.
	Cache       *cache.Cache     // Cache is the cache used to store responses from the Backupify WebUI.
	exportToken string           // exportToken is the token used to export data from Backupify.
}

type AppType string // AppType is the type of Backupify application.
// END OF BACKUPIFY CLIENT STRUCTS
//----------------------------------------------------------------------

// ### Backupify Activity Structs
// ---------------------------------------------------------------------
type ActivitiesResponse struct {
	Activities Activities `json:"activities,omitempty"` // Activities {Exports, Restores, Backups}
}

type Activities struct {
	Export  ActivityDetail `json:"export,omitempty"`  // Details of export activities
	Restore ActivityDetail `json:"restore,omitempty"` // Details of restore activities
	Backups ActivityDetail `json:"backups,omitempty"` // Details of backup activities
}

type ActivityDetail struct {
	HasActive bool    `json:"hasActive,omitempty"` // Indicates if there is an active status
	Items     []*Item `json:"items,omitempty"`     // List of individual items in the activity
}

type Item struct {
	BytesWritten string  `json:"bytesWritten,omitempty"` // The amount of data written, can be in various units
	DetailsPath  *string `json:"detailsPath,omitempty"`  // Path to further details, can be null
	Destination  string  `json:"destination,omitempty"`  // Destination of the data
	Export       Export  `json:"export,omitempty"`       // Export details
	Items        string  `json:"items,omitempty"`        // Summary of items processed, in the format "processed / total"
	RecoveredBy  string  `json:"recoveredBy,omitempty"`  // Who recovered the item
	Reason       string  `json:"reason,omitempty"`       // Reason for the current status (typically for failed or cancelled tasks)
	Run          Run     `json:"run,omitempty"`          // Running details
	RunMode      string  `json:"runMode,omitempty"`      // Mode of the run
	Source       string  `json:"source,omitempty"`       // Source of the data
	Stats        Stats   `json:"stats,omitempty"`        // Statistical data
	Status       string  `json:"status,omitempty"`       // Current status of the item
	Timestamp    int64   `json:"timestamp,omitempty"`    // Timestamp of the item creation or modification
	Timetaken    string  `json:"timetaken,omitempty"`    // Time taken for the operation
	Type         string  `json:"type,omitempty"`         // Type of the item
}

type Run struct {
	ActionType            string      `json:"actionType,omitempty"`            // Type of action, e.g., Export or Restore
	AppType               string      `json:"appType,omitempty"`               // Application type involved
	CompletedAt           int64       `json:"completedAt,omitempty"`           // Completion timestamp
	CreatedAt             int64       `json:"createdAt,omitempty"`             // Creation timestamp
	CustomerId            int         `json:"customerId,omitempty"`            // ID of the customer
	Description           Description `json:"description,omitempty"`           // Description of the run
	ID                    int         `json:"id,omitempty"`                    // ID of the run
	TimeTakenMilliseconds int         `json:"timeTakenMilliseconds,omitempty"` // Time taken in milliseconds
}

type Description struct {
	Filters            interface{} `json:"filters,omitempty"`            // Filters applied during the run, can be null
	IncludePermissions bool        `json:"includePermissions,omitempty"` // Indicates if permissions were included
	ItemCount          int         `json:"itemCount,omitempty"`          // Number of items considered
	Query              string      `json:"query,omitempty"`              // Query terms used
	RecoveredBy        string      `json:"recoveredBy,omitempty"`        // Who recovered the run
	Services           []*Service  `json:"services,omitempty"`           // Services involved in the run
	Snapshot           int64       `json:"snapshot,omitempty"`           // Snapshot ID
	TargetService      *Service    `json:"targetService,omitempty"`      // Target service for the run, can be null
	Type               string      `json:"type,omitempty"`               // Type of the description, e.g., full or selected
}

type Service struct {
	ServiceEmail string `json:"serviceEmail,omitempty"` // Email associated with the service
	ServiceId    int    `json:"serviceId,omitempty"`    // ID of the service
	ServiceName  string `json:"serviceName,omitempty"`  // Name of the service
}

type Stats struct {
	BytesWritten int `json:"BytesWritten,omitempty"` // Number of bytes written
	FailureCount int `json:"FailureCount,omitempty"` // Count of failed operations
	SkippedCount int `json:"SkippedCount,omitempty"` // Count of skipped operations
	SuccessCount int `json:"SuccessCount,omitempty"` // Count of successful operations
	TotalCount   int `json:"TotalCount,omitempty"`   // Total number of operations
}

type Filters struct {
	IsDeleted string `json:"isDeleted,omitempty"` // Whether deleted items were included
}

type ActivitiesPayload struct {
	AppType AppType `json:"appType"` // Type of Backupify application. e.g., "GoogleDrive", "GoogleTeamDrives", etc.
}

// END OF BACKUPIFY ACTIVITY STRUCTS
//----------------------------------------------------------------------

// ### Backupify Export Structs
// ---------------------------------------------------------------------
type Exports []*Export

type Export struct {
	ResponseData ResponseData `json:"responseData,omitempty"` // Container for the response data
	Status       string       `json:"state,omitempty"`        // Current status of the export
}

type ResponseData struct {
	Action     string `json:"action,omitempty"`     // Action taken, e.g., "Export"
	AppType    string `json:"appType,omitempty"`    // Type of application involved, e.g., "GoogleDrive"
	CustomerId int    `json:"customerId,omitempty"` // Numeric ID of the customer
	ID         int    `json:"id,omitempty"`         // Numeric ID associated with the responseData
	Status     string `json:"status,omitempty"`     // Current status, e.g., "started"
}

type ExportPayload struct {
	ActionType         string        `json:"actionType"`         // Type of action, e.g., "export"
	AppType            AppType       `json:"appType"`            // Type of application, e.g., "GoogleDrive"
	SnapshotID         string        `json:"snapshotId"`         // ID of the snapshot
	Token              string        `json:"token"`              // Placeholder for a variable
	IncludePermissions bool          `json:"includePermissions"` // Include permissions associated with the files being exported
	IncludeAttachments bool          `json:"includeAttachments"` // Include attachments
	Services           []interface{} `json:"services"`           // Identity to target. e.g. [userID]
}

// END OF BACKUPIFY EXPORT STRUCTS
//----------------------------------------------------------------------

// ### Backupify User Structs
// ---------------------------------------------------------------------
type Users struct {
	Draw            int     `json:"draw,omitempty"`            // The draw number
	Data            []*User `json:"data,omitempty"`            // List of data items
	RecordsTotal    int     `json:"recordsTotal,omitempty"`    // Total number of records
	RecordsFiltered int     `json:"recordsFiltered,omitempty"` // Number of filtered records
}

func (u *Users) Map() map[string]*User {
	userMap := make(map[string]*User)
	for _, user := range u.Data {
		userMap[user.Email] = user
	}
	return userMap
}

type User struct {
	AppType        string      `json:"appType,omitempty"`        // Type of the application
	CreatedAt      int64       `json:"createdAt,omitempty"`      // Creation timestamp
	CustomerId     int         `json:"customerId,omitempty"`     // ID of the customer
	Deleted        bool        `json:"deleted,omitempty"`        // Deletion flag
	Email          string      `json:"email,omitempty"`          // Email address
	ID             int         `json:"id,omitempty"`             // Unique identifier
	LatestSnap     interface{} `json:"latestSnap,omitempty"`     // Latest snapshot ID
	LocalSize      int64       `json:"localSize,omitempty"`      // Local size
	Name           string      `json:"name,omitempty"`           // Name of the item
	OwnSize        int64       `json:"ownSize,omitempty"`        // Owned size
	Path           string      `json:"path,omitempty"`           // File path
	PerfectBackups []Snapshot  `json:"perfectBackups,omitempty"` // List of perfect backups
	ReferencedSize int         `json:"referencedSize,omitempty"` // Referenced size
	Snapshots      []Snapshot  `json:"snapshots,omitempty"`      // List of snapshots
	SnapshotDates  *Snapshots  `json:"snapshotDates,omitempty"`  // Map of snapshot dates
	Status         string      `json:"status,omitempty"`         // Status of the item
	StorageFormat  string      `json:"storageFormat,omitempty"`  // Storage format
	UpdatedAt      int64       `json:"updatedAt,omitempty"`      // Update timestamp
	UsedBytes      string      `json:"usedBytes,omitempty"`      // Used bytes in string format
	UsedBytesFloat float64     `json:"usedBytesFloat,omitempty"` // Used bytes in float format
}

type Snapshots map[string][]Snapshot // Map of snapshot dates

type Snapshot struct {
	ID   int64  `json:"snapshotId,omitempty"`         // ID of the snapshot
	Date string `json:"formattedForButton,omitempty"` // Text formatted for display on a button.
}

type UserPayload struct {
	Draw    string   `json:"draw"`    // The draw number
	Columns []Column `json:"columns"` // List of columns
	Order   []Order  `json:"order"`   // List of order items
	Start   int      `json:"start"`   // Start index
	Length  int      `json:"length"`  // Length of the request
	Search  Search   `json:"search"`  // Search criteria
	AppType AppType  `json:"appType"` // Type of Backupify application
}

type Column struct {
	Data       string `json:"data"`       // Data field
	Name       string `json:"name"`       // Name of the column
	Searchable bool   `json:"searchable"` // Searchable flag
	Orderable  bool   `json:"orderable"`  // Orderable flag
	Search     Search `json:"search"`     // Search criteria
}

type Order struct {
	Column string `json:"column"` // Column index
	Dir    string `json:"dir"`    // Direction
}

type Search struct {
	Value string `json:"value"` // Search value
	Regex bool   `json:"regex"` // Regex flag
}

type SnapshotsPayload struct {
	AppType   AppType `json:"appType"`   // Type of Backupify application
	ServiceID int     `json:"serviceId"` // Identity to target. e.g. [userID]
}

// END OF BACKUPIFY USER STRUCTS
//----------------------------------------------------------------------

type DeletePayload struct {
	Type    string  `json:"type"`    // Type of deletion
	AppType AppType `json:"appType"` // Type of Backupify application
	ID      int     `json:"id"`      // Identity to target. e.g. [snapshotID]
}

type UserCounts struct {
	Count        int
	TotalStorage float64
}
