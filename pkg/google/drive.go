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
	"fmt"
	"sync"
	"time"
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
	if d == nil {
		return true
	}
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
	url := c.BuildURL(DriveFiles, nil, driveID)

	q := DriveFileQuery{
		Fields:            "*",
		SupportsAllDrives: true,
	}

	file, err := do[*File](c, "GET", url, q, nil)
	if err != nil {
		return nil, err
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

	file, err := do[*File](c, "POST", url, nil, &file)
	if err != nil {
		return nil, err
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
	url := c.BuildURL(DriveFiles, nil, file.ID)

	if file.Parents == nil {
		c.Log.Println("File has no parents")
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

	_, err := do[any](c, "PATCH", url, q, nil)
	if err != nil {
		return err
	}

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
	url := c.BuildURL(DriveFiles, nil, file.ID, "copy")

	q := DriveFileQuery{
		SupportsAllDrives: true,
	}

	copy, err := do[*File](c, "POST", url, q, nil)
	if err != nil {
		return err
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
	return c.GetFileList(&file, nil)
}

/*
 * # Get File List
 * Fetches all files in a folder, recursively
 * drive/v3/files
 * @param {File} file - The file to source the list from
 * @param {DriveFileQuery} q - The query parameters to use
 * https://developers.google.com/drive/api/v3/reference/files/list
 */
func (c *Client) GetFileList(file *File, q *DriveFileQuery) (*FileList, error) {
	cacheKey := fmt.Sprintf("drive_filelist_%s", file.ID)

	var cache FileList
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	if file.Path == "" {
		if err := c.initFilePath(file); err != nil {
			return nil, err
		}
	}

	if q.IsEmpty() {
		q = &DriveFileQuery{}
		initFileListQuery(q, file.ID)
	} else if err := q.ValidateQuery(); err != nil {
		return nil, err
	}

	allFiles := &FileList{Files: &[]*File{file}}
	parentPath := determineParentPath(file)

	err := c.processFileList(q, parentPath, allFiles)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, allFiles, 5*time.Minute)

	return allFiles, nil
}

func (c *Client) initFilePath(file *File) error {
	if file.ID != "root" {
		var err error
		file.Path, err = c.GetFilePath(file.ID)
		return err
	}
	file.Path = "My Drive"
	return nil
}

func initFileListQuery(q *DriveFileQuery, fileId string) {
	*q = DriveFileQuery{
		Fields:            `files(id, name, md5Checksum, mimeType, originalFilename, owners, parents, shortcutDetails/targetId, shortcutDetails/targetMimeType)`,
		PageSize:          1000,
		IncludeLabels:     "*",
		Q:                 fmt.Sprintf(`'%s' in parents and trashed = false`, fileId),
		SupportsAllDrives: true,
	}
}

func determineParentPath(file *File) string {
	if file.ID == "root" || file.Path == "/" {
		return "My Drive"
	}
	return file.Path
}

func (c *Client) processFileList(q *DriveFileQuery, parentPath string, allFiles *FileList) error {
	sem := make(chan struct{}, 10)
	filesChannel := make(chan *FileList)
	filesErrChannel := make(chan error)

	var wg sync.WaitGroup

	for {
		filesPage, err := c.fetchFilesPage(*q)
		if err != nil {
			return err
		}

		for _, file := range *filesPage.Files {
			file.Path = parentPath + "/" + file.Name
			c.Log.Println("File Path:", file.Path)
			*allFiles.Files = append(*allFiles.Files, file)

			if file.MimeType == "application/vnd.google-apps.folder" {
				wg.Add(1)
				go c.fetchSubFiles(file, file.Path, sem, filesChannel, filesErrChannel, &wg)
			}
		}

		if filesPage.NextPageToken == "" {
			break
		}
		q.PageToken = filesPage.NextPageToken
	}

	go func() {
		wg.Wait()
		close(sem)
		close(filesChannel)
		close(filesErrChannel)
	}()

	for file := range filesChannel {
		*allFiles.Files = append(*allFiles.Files, *file.Files...)
	}

	for err := range filesErrChannel {
		return err
	}

	return nil
}

func (c *Client) fetchSubFiles(file *File, parentPath string, sem chan struct{}, filesChannel chan<- *FileList, filesErrChannel chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	sem <- struct{}{}
	defer func() { <-sem }()
	subFiles, err := c.GetFileList(&File{ID: file.ID, Path: parentPath}, &DriveFileQuery{})
	if err != nil {
		filesErrChannel <- err
		return
	}
	filesChannel <- subFiles
}

/*
# Save File List to Google Sheet
*/
func (c *Client) SaveFileListToSheet(fileList *FileList, sheetID string, headers *[]string) error {

	if sheetID == "" {
		c.Log.Println("No sheet ID provided, creating new sheet")
		c.Log.Println("Creating new spreadsheet for file list")
		sheet, err := c.CreateSpreadsheet()
		if err != nil {
			return err
		}
		sheetID = sheet.SpreadsheetID
	}

	sheetData := make([]interface{}, len(*fileList.Files))
	for i, v := range *fileList.Files {
		sheetData[i] = v
	}

	c.Log.Println("Saving File List to spreadsheet")
	if headers == nil {
		headers = &[]string{"id", "name", "path", "md5Checksum", "mimeType", "originalFilename", "owners", "parents", "shortcutDetails"}
	}
	vr := GenerateValueRange(sheetData, headers)
	err := c.UpdateSpreadsheet(sheetID, vr)
	if err != nil {
		return err
	}

	c.Log.Println("Formatting spreadsheet")
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
	url := c.BuildURL(DriveFiles, nil)

	filesPage, err := do[*FileList](c, "GET", url, q, nil)
	if err != nil {
		return nil, err
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
