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
	"sync"
)

var (
	DriveBaseURL     = fmt.Sprintf("%s/drive/v3", BaseURL)         // https://developers.google.com/drive/api/v3/reference/
	DriveAbout       = fmt.Sprintf("%s/about", DriveBaseURL)       // https://developers.google.com/drive/api/v3/reference/about
	DriveChanges     = fmt.Sprintf("%s/changes", DriveBaseURL)     // https://developers.google.com/drive/api/v3/reference/changes
	DriveChannels    = fmt.Sprintf("%s/channels", DriveBaseURL)    // https://developers.google.com/drive/api/v3/reference/channels
	DriveComments    = fmt.Sprintf("%s/comments", DriveBaseURL)    // https://developers.google.com/drive/api/v3/reference/comments
	DriveFiles       = fmt.Sprintf("%s/files", DriveBaseURL)       // https://developers.google.com/drive/api/v3/reference/files
	DrivePermissions = fmt.Sprintf("%s/permissions", DriveBaseURL) // https://developers.google.com/drive/api/v3/reference/permissions
	DriveReplies     = fmt.Sprintf("%s/replies", DriveBaseURL)     // https://developers.google.com/drive/api/v3/reference/replies
	DriveRevisions   = fmt.Sprintf("%s/revisions", DriveBaseURL)   // https://developers.google.com/drive/api/v3/reference/revisions
)

/*
 * Query Parameters for Drive Files
 * Reference: https://developers.google.com/drive/api/reference/rest/v3/files/list#query-parameters
 */
type DriveFileQuery struct {
	AcknowledgeAbuse          bool   `url:"acknowledgeAbuse,omitempty"`          // Whether the user is acknowledging the risk of downloading known malware or other abusive files. This is only applicable when alt=media.
	Corpora                   string `url:"corpora,omitempty"`                   // Bodies of items (files/documents) to which the query applies. Supported bodies are 'user', 'domain', 'drive', and 'allDrives'. Prefer 'user' or 'drive' to 'allDrives' for efficiency.
	DriveID                   string `url:"driveId,omitempty"`                   // ID of the shared drive to search.
	Depth                     int    `url:"depth,omitempty"`                     // The depth of the traversal. **ReGo only**
	IncludeItemsFromAllDrives bool   `url:"includeItemsFromAllDrives,omitempty"` // Whether both My Drive and shared drive items should be included in results.
	OrderBy                   string `url:"orderBy,omitempty"`                   // A comma-separated list of sort keys.
	PageSize                  int    `url:"pageSize,omitempty"`                  // The maximum number of files to return per page. Partial or empty result pages are possible even before the end of the files list has been reached. Default: 100. Max: 1000. https://developers.google.com/drive/api/guides/limits
	PageToken                 string `url:"pageToken,omitempty"`                 // The token for continuing a previous list request on the next page.
	Q                         string `url:"q,omitempty"`                         // A query for filtering the file results. See the [Search for Files](https://developers.google.com/drive/api/guides/search-files) guide for supported syntax.
	Spaces                    string `url:"spaces,omitempty"`                    // A comma-separated list of spaces to query within the corpora. Supported values are 'drive' and 'appDataFolder'.
	SupportsAllDrives         bool   `url:"supportsAllDrives,omitempty"`         // Whether the requesting application supports both My Drives and shared drives.
	IncludePermissionsForView string `url:"includePermissionsForView,omitempty"` // Specifies which additional view's permissions to include in the response. Only 'published' is supported.
	IncludeLabels             string `url:"includeLabels,omitempty"`             // A comma-separated list of IDs of labels to include in the labelInfo part of the response.
	Fields                    string `url:"fields,omitempty"`                    // Examples: `files(id, name, parents)` or `id,name,parents` https://developers.google.com/drive/api/guides/fields-parameter#format
	UploadType                string `url:"uploadType,omitempty"`                // https://developers.google.com/drive/api/reference/rest/v3/files/update
	AddParents                string `url:"addParents,omitempty"`                // A comma-separated list of parent IDs to add.
	KeepRevisionForever       bool   `url:"keepRevisionForever,omitempty"`       // Whether to set the 'keepForever' field in the new head revision. This is only applicable to files with binary content in Google Drive.
	OCRLanguage               string `url:"ocrLanguage,omitempty"`               // A language hint for OCR processing during image import (ISO 639-1 code).
	RemoveParents             string `url:"removeParents,omitempty"`             // A comma-separated list of parent IDs to remove.
	UseContentAsIndexableText bool   `url:"useContentAsIndexableText,omitempty"` // Whether to use the uploaded content as indexable text.
}

/*
 * Check if the DriveQuery is empty
 */
func (d *DriveFileQuery) IsEmpty() bool {
	return !d.AcknowledgeAbuse &&
		d.Corpora == "" &&
		d.DriveID == "" &&
		!d.IncludeItemsFromAllDrives &&
		d.OrderBy == "" &&
		d.PageSize == 0 &&
		d.PageToken == "" &&
		d.Q == "" &&
		d.Spaces == "" &&
		!d.SupportsAllDrives &&
		d.IncludePermissionsForView == "" &&
		d.IncludeLabels == "" &&
		d.Fields == "" &&
		d.UploadType == "" &&
		d.AddParents == "" &&
		!d.KeepRevisionForever &&
		d.OCRLanguage == "" &&
		d.RemoveParents == "" &&
		!d.UseContentAsIndexableText
}

/*
 * Validate the query parameters for the Files resource
 */
func (d *DriveFileQuery) ValidateQuery() error {
	if d.IsEmpty() {
		d.Fields = "*"
		d.Depth = 1
		return nil
	}

	if d.Corpora == "" {
		d.Corpora = "user"
	}

	if d.Fields == "" {
		d.Fields = "*"
	}

	if d.PageSize == 0 {
		d.PageSize = 100
	}

	return nil
}

/*
# Get Google Drive File
- drive/v3/files/{fileId}
- @param {string} fileId - The ID of the file or shortcut.
- https://developers.google.com/drive/api/v3/reference/files/get
*/
func (c *Client) GetFile(driveID string) (*File, error) {
	file := &File{}

	q := DriveFileQuery{
		Fields:            "*",
		SupportsAllDrives: true,
	}

	url := fmt.Sprintf("%s/%s", DriveFiles, driveID)
	c.Logger.Debug("url:", url)
	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &file)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return file, nil
}

/*
_Create Google Drive File/Folder_
  - If no file is provided, a folder will be created
  - drive/v3/files
  - https://developers.google.com/drive/api/v3/reference/files/update
*/
func (c *Client) CreateFile(file *File) (*File, error) {
	if file == nil {
		file = &File{
			MimeType: "application/vnd.google-apps.folder",
			Name:     "New Folder",
		}
	}

	url := DriveFiles
	c.Logger.Debug("url:", url)
	res, body, err := c.HTTPClient.DoRequest("POST", url, nil, &file)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &file)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return file, nil
}

/*
 * # Move Google Drive File/Folder
 * drive/v3/files/{fileId}
 * @param {File} file - The file to move
 * @param {File} folder - The folder to move the file to
 * https://developers.google.com/drive/api/v3/reference/files/update
 */
func (c *Client) MoveFileToFolder(file *File, folder *File) error {
	url := fmt.Sprintf("%s/%s", DriveFiles, file.ID)
	c.Logger.Debug("url:", url)

	if file.Parents == nil {
		c.Logger.Println("File has no parents")
		var err error
		file, err = c.GetFile(file.ID)
		if err != nil {
			return err
		}
	}

	q := DriveFileQuery{
		AddParents:        folder.ID,
		RemoveParents:     file.Parents[0],
		Fields:            "id,name,parents",
		SupportsAllDrives: true,
	}

	res, body, err := c.HTTPClient.DoRequest("PATCH", url, q, nil)
	if err != nil {
		return err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	return nil
}

/*
 * # Move Google Drive File/Folder
 * drive/v3/files/{fileId}
 * @param {File} file - The file to move
 * @param {File} folder - The folder to move the file to
 * https://developers.google.com/drive/api/v3/reference/files/update
 */
func (c *Client) CopyFileToFolder(file *File, folder *File) error {
	url := fmt.Sprintf("%s/%s/copy", DriveFiles, file.ID)
	c.Logger.Debug("url:", url)

	q := DriveFileQuery{
		SupportsAllDrives: true,
	}

	res, body, err := c.HTTPClient.DoRequest("POST", url, q, nil)
	if err != nil {
		return err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	copy := &File{}
	err = json.Unmarshal(body, &copy)
	if err != nil {
		return nil
	}

	c.MoveFileToFolder(copy, folder)

	return nil
}

/*
 * # Get File List ("My Drive")
 * drive/v3/files
 * https://developers.google.com/drive/api/v3/reference/files/list
 */
func (c *Client) GetRootFileList() (*FileList, error) {
	file := File{
		ID:   "root",
		Path: "/",
	}
	q := DriveFileQuery{}
	return c.GetFileList(&file, q)
}

/*
 * # Get File List
 * Fetches all files in a folder, recursively
 * drive/v3/files
 * @param {File} file - The file to source the list from
 * @param {DriveFileQuery} q - The query parameters to use
 * https://developers.google.com/drive/api/v3/reference/files/list
 */
func (c *Client) GetFileList(file *File, q DriveFileQuery) (*FileList, error) {
	allFiles := &FileList{}
	allFiles.Files = append(allFiles.Files, *file)

	var parentPath string
	if file.Path == "" && file.ID != "root" {
		var err error // Prevent Variable shadowing
		file.Path, err = c.GetFilePath(file.ID)
		if err != nil {
			return nil, err
		}
	} else if file.ID == "root" || file.Path == "/" {
		file.Path = "My Drive"
		parentPath = file.Path
	}
	parentPath = "Search Results"

	if q.IsEmpty() {
		q.Fields = `files(id, name, md5Checksum, mimeType, originalFilename, owners, parents, shortcutDetails/targetId, shortcutDetails/targetMimeType)`
		q.PageSize = 1000
		q.IncludeLabels = "*"
		q.Q = fmt.Sprintf(`'%s' in parents and trashed = false`, file.ID)
		q.SupportsAllDrives = true
		parentPath = "Search Results"
	} else {
		err := q.ValidateQuery()
		if err != nil {
			return nil, err
		}
		if q.Q == "" {
			q.Q = fmt.Sprintf(`'%s' in parents and trashed = false`, file.ID)
		}
	}

	sem := make(chan struct{}, 10)
	filesChannel := make(chan *FileList)
	filesErrChannel := make(chan error)
	var wg sync.WaitGroup

	for {
		filesPage, err := c.fetchFilesPage(q)
		if err != nil {
			return nil, err
		}

		for _, file := range filesPage.Files {
			// Generate file's path and append it to the list
			file.Path = parentPath + "/" + file.Name
			c.Logger.Println("File Path:", file.Path)
			allFiles.Files = append(allFiles.Files, file)
			if file.MimeType == "application/vnd.google-apps.folder" {
				wg.Add(1)
				go func(fileId string, filePath string) {
					defer wg.Done()
					sem <- struct{}{}
					subFiles, err := c.GetFileList(&File{ID: fileId, Path: filePath + "/"}, DriveFileQuery{})
					<-sem
					if err != nil {
						filesErrChannel <- err
					} else {
						filesChannel <- subFiles
					}
				}(file.ID, file.Path)
			}
		}

		if filesPage.NextPageToken == "" {
			break
		}

		q.PageToken = filesPage.NextPageToken
	}

	go func() {
		wg.Wait()
		close(filesChannel)
		close(filesErrChannel)
	}()

	for file := range filesChannel {
		allFiles.Files = append(allFiles.Files, file.Files...)
	}

	for err := range filesErrChannel {
		return nil, err
	}

	return allFiles, nil
}

/*
# Save File List to Google Sheet
*/
func (c *Client) SaveFileListToSheet(fileList *FileList, sheetID string, headers *[]string) error {

	if sheetID == "" {
		c.Logger.Println("No sheet ID provided, creating new sheet")
		c.Logger.Println("Creating new spreadsheet for file list")
		sheet, err := c.CreateSpreadsheet()
		if err != nil {
			return err
		}
		sheetID = sheet.SpreadsheetID
	}

	sheetData := make([]interface{}, len(fileList.Files))
	for i, v := range fileList.Files {
		sheetData[i] = v
	}

	c.Logger.Println("Saving File List to spreadsheet")
	if headers == nil {
		headers = &[]string{"id", "name", "path", "md5Checksum", "mimeType", "originalFilename", "owners", "parents", "shortcutDetails"}
	}
	vr := GenerateValueRange(sheetData, headers)
	err := c.UpdateSpreadsheet(sheetID, vr)
	if err != nil {
		return err
	}

	c.Logger.Println("Formatting spreadsheet")
	rows := len(vr.Values)
	columns := len(vr.Values[0])
	c.FormatHeaderAndAutoSize(sheetID, rows, columns)

	return nil
}

/*
 * # Fetch Files Page
 * Fetches a page of files
 * drive/v3/files
 * https://developers.google.com/drive/api/v3/reference/files/list
 */
func (c *Client) fetchFilesPage(q DriveFileQuery) (*FileList, error) {
	url := DriveFiles
	c.Logger.Debug("url:", url)

	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		c.Logger.Println(err)
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	filesPage := &FileList{}
	err = json.Unmarshal(body, &filesPage)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling user: %w", err)
	}

	return filesPage, nil
}

/*
 * # Get File Path
 * Constructs the path of a file
 * drive/v3/files/{fileId}
 * @param {string} id - The ID of the file or shortcut to get the path of.
 * https://developers.google.com/drive/api/v3/reference/files/get
 */
func (c *Client) GetFilePath(id string) (string, error) {
	file, err := c.GetFile(id)
	if err != nil {
		return "", err
	}

	// If file has no parents, it's in root (My Drive) or "Shared With Me" via the WebUI
	if len(file.Parents) == 0 {
		if file.Shared && !file.OwnedByMe {
			return "Shared with me/" + file.Name, nil
		}
		return file.Name, nil
	}

	// We will take only the first parent, because a folder can have multiple parents in Google Drive
	parent, err := c.GetFile(file.Parents[0])
	if err != nil {
		return "", err
	}

	parentPath, err := c.GetFilePath(parent.ID)
	if err != nil {
		return "", err
	}

	return parentPath + "/" + file.Name, nil
}
