/*
# Google Workspace - Drive

This package initializes all the methods for functions which interact with the Google Drive API:
https://developers.google.com/drive/api/v3/reference/

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/drive.go
package google

import (
	"encoding/json"
	"fmt"
)

var (
	DriveBaseURL     = fmt.Sprintf("%s/drive/v3", BaseURL)
	DriveAbout       = fmt.Sprintf("%s/about", DriveBaseURL)
	DriveChanges     = fmt.Sprintf("%s/changes", DriveBaseURL)
	DriveChannels    = fmt.Sprintf("%s/channels", DriveBaseURL)
	DriveComments    = fmt.Sprintf("%s/comments", DriveBaseURL)
	DriveFiles       = fmt.Sprintf("%s/files", DriveBaseURL)
	DrivePermissions = fmt.Sprintf("%s/permissions", DriveBaseURL)
	DriveReplies     = fmt.Sprintf("%s/replies", DriveBaseURL)
	DriveRevisions   = fmt.Sprintf("%s/revisions", DriveBaseURL)
)

/*
 * Query Parameters for Drive Files
 * Reference: https://developers.google.com/drive/api/reference/rest/v3/files/list#query-parameters
 */
type DriveFileQuery struct {
	AcknowledgeAbuse          bool   `json:"acknowledgeAbuse,omitempty"`          // Whether the user is acknowledging the risk of downloading known malware or other abusive files. This is only applicable when alt=media.
	Corpora                   string `json:"corpora,omitempty"`                   // Bodies of items (files/documents) to which the query applies. Supported bodies are 'user', 'domain', 'drive', and 'allDrives'. Prefer 'user' or 'drive' to 'allDrives' for efficiency.
	DriveId                   string `json:"driveId,omitempty"`                   // ID of the shared drive to search.
	IncludeItemsFromAllDrives bool   `json:"includeItemsFromAllDrives,omitempty"` // Whether both My Drive and shared drive items should be included in results.
	OrderBy                   string `json:"orderBy,omitempty"`                   // A comma-separated list of sort keys.
	PageSize                  int    `json:"pageSize,omitempty"`                  // The maximum number of files to return per page. Partial or empty result pages are possible even before the end of the files list has been reached.
	PageToken                 string `json:"pageToken,omitempty"`                 // The token for continuing a previous list request on the next page.
	Q                         string `json:"q,omitempty"`                         // A query for filtering the file results.
	Spaces                    string `json:"spaces,omitempty"`                    // A comma-separated list of spaces to query within the corpora. Supported values are 'drive' and 'appDataFolder'.
	SupportsAllDrives         bool   `json:"supportsAllDrives,omitempty"`         // Whether the requesting application supports both My Drives and shared drives.
	IncludePermissionsForView string `json:"includePermissionsForView,omitempty"` // Specifies which additional view's permissions to include in the response. Only 'published' is supported.
	IncludeLabels             string `json:"includeLabels,omitempty"`             // A comma-separated list of IDs of labels to include in the labelInfo part of the response.
	Fields                    string `json:"fields,omitempty"`                    // https://developers.google.com/drive/api/guides/fields-parameter#format
	UploadType                string `json:"uploadType,omitempty"`                // https://developers.google.com/drive/api/reference/rest/v3/files/update
	AddParents                string `json:"addParents,omitempty"`                // A comma-separated list of parent IDs to add.
	KeepRevisionForever       bool   `json:"keepRevisionForever,omitempty"`       // Whether to set the 'keepForever' field in the new head revision. This is only applicable to files with binary content in Google Drive.
	OCRLanguage               string `json:"ocrLanguage,omitempty"`               // A language hint for OCR processing during image import (ISO 639-1 code).
	RemoveParents             string `json:"removeParents,omitempty"`             // A comma-separated list of parent IDs to remove.
	UseContentAsIndexableText bool   `json:"useContentAsIndexableText,omitempty"` // Whether to use the uploaded content as indexable text.
}

/*
 * # Get Google Drive File
 * drive/v3/files/{fileId}
 * @param {string} fileId - The ID of the file or shortcut.
 * https://developers.google.com/drive/api/v3/reference/files/get
 */
func (c *Client) GetDocument(driveID string) (*Document, error) {
	document := &Document{}

	q := DriveFileQuery{
		Fields: "*",
	}

	url := fmt.Sprintf("%s/%s", DriveFiles, driveID)
	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &document)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return document, nil
}

/*
 * # Move Google Drive File/Folder
 * drive/v3/files/{fileId}
 * @param {string} fileId - The ID of the file or shortcut.
 * https://developers.google.com/drive/api/v3/reference/files/update
 */
func (c *Client) MoveFileToFolder(fileID string, folderID string) error {
	url := fmt.Sprintf("%s/%s", DriveFiles, fileID)

	q := DriveFileQuery{
		AddParents:    folderID,
		RemoveParents: "root", // From the perspective of a freshly created file
		Fields:        "id,parents",
	}

	res, body, err := c.HTTPClient.DoRequest("PATCH", url, q, nil)
	if err != nil {
		return err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	return nil
}
