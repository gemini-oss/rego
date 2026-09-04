/*
# Google Workspace - Drive

This package initializes all the methods for functions which interact with the Google Drive API:
https://developers.google.com/drive/api/v3/reference/

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/drive.go
package google

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gemini-oss/rego/pkg/common/progress"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
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

// DriveClient for chaining methods
type DriveClient struct {
	*Client
}

// Entry point for drive-related operations
func (c *Client) Drive() *DriveClient {
	// Return cached client if it exists
	if c.driveClient != nil {
		return c.driveClient
	}

	// Shallow copy a new client with Drive-specific rate limiter
	// https://developers.google.com/drive/api/guides/limits
	driveRL := ratelimit.NewRateLimiter(12000, 1*time.Minute)
	driveRL.Log.Verbosity = c.Log.Verbosity

	driveHTTP := requests.NewClient(c.HTTP.GetHTTPClient(), c.HTTP.GetHeaders(), driveRL)
	driveHTTP.BodyType = c.HTTP.BodyType

	driveClient := &Client{
		Auth:     c.Auth,
		BaseURL:  c.BaseURL,
		OAuth:    c.OAuth,
		JWT:      c.JWT,
		HTTP:     driveHTTP,
		Error:    c.Error,
		Log:      c.Log,
		Cache:    c.Cache,
		Customer: c.Customer,
	}

	c.driveClient = &DriveClient{
		Client: driveClient,
	}

	return c.driveClient
}

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

func (q *DriveFileQuery) SetPageToken(token string) {
	q.PageToken = token
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
	} else if !strings.Contains(d.Fields, "nextPageToken") {
		// When using partial field selection, nextPageToken MUST be included
		// for pagination to work. The API only returns fields you request.
		d.Fields = "nextPageToken, " + d.Fields
	}

	if d.PageSize == 0 {
		d.PageSize = 1000 // Max allowed by Google Drive API
	}

	return nil
}

type driveFileListOptions struct {
	// StreamResults enables per-folder, in-flight streaming of results to a Google Sheet.
	StreamResults func(sheetID string) FileListOption

	// SheetName sets a custom sheet name for results.
	SheetName func(name string) FileListOption

	// WithCSV enables CSV backup during Drive fetching.
	// path: Optional path for the CSV file (default: os.TempDir()/rego_drive_backup_{timestamp}.csv)
	WithCSV func(path ...string) FileListOption

	// Headers sets custom headers for both CSV and Sheet output.
	Headers func(headers *[]string) FileListOption

	// WithDelegation sets a Domain-Wide Delegation client for Google Sheets operations.
	// Useful when current client doesn't have permission to write to target sheet.
	WithDelegation func(client *Client) FileListOption

	// IgnoreMemory enables "memory-efficient mode" where results are NOT accumulated in memory.
	IgnoreMemory func() FileListOption

	// ExcludeFolders excludes folders from outputs while still traversing them if in the query.
	ExcludeFolders func() FileListOption

	// WithCache controls cache behavior for the FileList operation.
	WithCache func(enabled bool) FileListOption

	// WithProgressBar controls whether the progress tracker is displayed.
	WithProgressBar func(enabled bool) FileListOption
}

// DriveFileListOpts provides namespaced options for GetFileList.
// Usage: client.Drive().GetFileList(file, query, FileListOpts.WithStream("sheetID"))
var DriveFileListOpts = driveFileListOptions{
	StreamResults: func(sheetID string) FileListOption {
		return func(cfg *fileListConfig) {
			cfg.streamEnabled = true
			cfg.sheetID = sheetID
		}
	},
	SheetName: func(name string) FileListOption {
		return func(cfg *fileListConfig) {
			cfg.sheetName = name
		}
	},
	WithCSV: func(path ...string) FileListOption {
		return func(cfg *fileListConfig) {
			cfg.csvEnabled = true
			if len(path) > 0 && path[0] != "" {
				cfg.csvPath = path[0]
			}
		}
	},
	Headers: func(headers *[]string) FileListOption {
		return func(cfg *fileListConfig) {
			if headers != nil && len(*headers) > 0 {
				cfg.headers = *headers
			}
		}
	},
	WithDelegation: func(client *Client) FileListOption {
		return func(cfg *fileListConfig) {
			cfg.delegationClient = client
		}
	},
	IgnoreMemory: func() FileListOption {
		return func(cfg *fileListConfig) {
			cfg.ignoreMemory = true
		}
	},
	ExcludeFolders: func() FileListOption {
		return func(cfg *fileListConfig) {
			cfg.excludeFolders = true
		}
	},
	WithCache: func(enabled bool) FileListOption {
		return func(cfg *fileListConfig) {
			cfg.useCache = enabled
		}
	},
	WithProgressBar: func(enabled bool) FileListOption {
		return func(cfg *fileListConfig) {
			cfg.showProgress = enabled
		}
	},
}

/*
# Get Google Drive File
  - drive/v3/files/{fileId}
  - @param {string} fileId - The ID of the file or shortcut.
  - @param {*sync.Map} fileCache - Optional in-memory cache for files (fileID -> *File).
    When provided, avoids rego cache overhead (gob serialization).
  - https://developers.google.com/drive/api/v3/reference/files/get
*/
func (c *DriveClient) GetFile(driveID string, fileCache ...*sync.Map) (*File, error) {
	// Check in-memory cache first (fastest - no serialization)
	if len(fileCache) > 0 && fileCache[0] != nil {
		if cached, ok := fileCache[0].Load(driveID); ok {
			return cached.(*File), nil
		}
	}

	cacheKey := fmt.Sprintf("drive_file_%s", driveID)

	// Check rego cache (slower - gob serialization)
	var cachedFile File
	if c.GetCache(cacheKey, &cachedFile) {
		// Also store in in-memory cache for future fast lookups
		if len(fileCache) > 0 && fileCache[0] != nil {
			fileCache[0].Store(driveID, &cachedFile)
		}
		return &cachedFile, nil
	}

	url := c.BuildURL(DriveFiles, nil, driveID)

	q := DriveFileQuery{
		Fields:            "*",
		SupportsAllDrives: true,
	}

	file, err := do[File](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	// Store in both caches
	c.SetCache(cacheKey, file, 30*time.Minute)
	if len(fileCache) > 0 && fileCache[0] != nil {
		fileCache[0].Store(driveID, &file)
	}
	return &file, nil
}

/*
_Create Google Drive File/Folder_
  - If no file is provided, a folder will be created
  - drive/v3/files
  - https://developers.google.com/drive/api/v3/reference/files/update
*/
func (c *DriveClient) CreateFile(file *File) (*File, error) {
	if file == nil {
		file = &File{
			MimeType: "application/vnd.google-apps.folder",
			Name:     "New Folder",
		}
	}

	url := DriveFiles

	file, err := do[*File](c.Client, "POST", url, nil, &file)
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
func (c *DriveClient) MoveFileToFolder(file *File, folder *File) error {
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

	_, err := do[any](c.Client, "PATCH", url, q, nil)
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
func (c *DriveClient) CopyFileToFolder(file *File, folder *File) error {
	url := c.BuildURL(DriveFiles, nil, file.ID, "copy")

	q := DriveFileQuery{
		SupportsAllDrives: true,
	}

	copy, err := do[File](c.Client, "POST", url, q, nil)
	if err != nil {
		return err
	}

	c.MoveFileToFolder(&copy, folder)

	return nil
}

/*
 * # Get File List ("My Drive")
 * drive/v3/files
 * https://developers.google.com/drive/api/v3/reference/files/list
 */
func (c *DriveClient) GetRootFileList(opts ...FileListOption) (*FileList, error) {
	file := File{
		ID:   "root",
		Path: "/",
	}
	return c.GetFileList(&file, nil, opts...)
}

/*
 * # Get File List
 * Fetches all files in a folder, recursively
 * drive/v3/files
 * @param {File} file - The file/folder to source the list from
 * @param {DriveFileQuery} q - The query parameters to use (can be nil for defaults)
 * @param {FileListOption} opts - Optional configuration for streaming, CSV backup, cache control, etc.
 * https://developers.google.com/drive/api/v3/reference/files/list
 */
func (c *DriveClient) GetFileList(file *File, q *DriveFileQuery, opts ...FileListOption) (*FileList, error) {
	// For global queries (no parent filter), file parameter is optional
	isGlobalQuery := q != nil && !hasParentFilter(q.Q)

	if file == nil {
		if isGlobalQuery {
			// For global queries, create a placeholder file
			file = &File{ID: "global", Name: "Global Search", Path: ""}
		} else {
			return nil, fmt.Errorf("file parameter cannot be nil for hierarchical queries")
		}
	}

	// Only resolve file path for hierarchical queries (global queries don't need it)
	if !isGlobalQuery && file.Path == "" {
		if err := c.initFilePath(file); err != nil {
			return nil, err
		}
	}

	// Build config from options first (to check if cache is enabled)
	cfg := &fileListConfig{
		useCache:     false,
		showProgress: true,
		headers: []string{
			"id", "name", "path", "md5Checksum", "mimeType",
			"originalFilename", "owners", "parents", "shortcutDetails",
		},
		folderStats: make(map[string]*DriveFolderStats),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Initialize query if empty
	if q.IsEmpty() {
		q = &DriveFileQuery{}
		initFileListQuery(q, file.ID)
	} else if err := q.ValidateQuery(); err != nil {
		return nil, err
	}
	cfg.query = q

	// Cache key computation - only needed if caching is enabled
	// Uses query hash for global queries to differentiate between different searches
	cacheKey := func() string {
		if isGlobalQuery && q.Q != "" {
			hash := md5.Sum([]byte(q.Q))
			return fmt.Sprintf("drive_filelist_global_%s", hex.EncodeToString(hash[:8]))
		}
		return fmt.Sprintf("drive_filelist_%s", file.ID)
	}

	startTime := time.Now()
	// Check cache if enabled
	if cfg.useCache {
		var cache FileList
		if c.GetCache(cacheKey(), &cache) {
			cache.Metadata = &FileListMetadata{
				Query:    q,
				Duration: time.Since(startTime),
				Cached:   true,
			}
			if cache.Files != nil {
				cache.Metadata.FilesFound = len(*cache.Files)
			}
			return &cache, nil
		}
	}

	// Set sheet name from folder if not specified
	if cfg.sheetName == "" {
		if file.Name != "" {
			cfg.sheetName = file.Name
		} else {
			cfg.sheetName = file.ID
		}
	}

	// Start progress tracker - use indeterminate for global queries, hierarchical for folder traversal
	if cfg.showProgress {
		if hasParentFilter(q.Q) {
			cfg.progressTracker = progress.NewHierarchicalWithMode("Scanning Drive", 0, progress.ModeScrollRegion)
			go cfg.progressTracker.Start()
		} else {
			cfg.indeterminateTracker = progress.NewIndeterminate("Scanning Drive (Global)")
			go cfg.indeterminateTracker.Start()
		}
	}

	// Initialize CSV backup
	if cfg.csvEnabled {
		if err := c.initCSVBackup(cfg); err != nil {
			c.Log.Warning("Failed to initialize CSV backup:", err)
			cfg.csvEnabled = false
		}
	}

	// Initialize sheet stream
	if cfg.streamEnabled {
		if err := c.initSheetStream(cfg); err != nil {
			c.Log.Warning("Failed to initialize sheet stream:", err)
			cfg.streamEnabled = false
		}
	}

	// Start heartbeat logger for long-running jobs
	heartbeatStop := make(chan struct{})
	if cfg.streamEnabled || cfg.csvEnabled {
		go c.runHeartbeatLogger(cfg, startTime, heartbeatStop)
	}

	// Temporarily enable cache for path resolution operations to avoid redundant API calls
	// This is done at the entry point to avoid race conditions when processing files concurrently
	originalCacheState := c.Cache.Enabled
	c.Cache.Enabled = true
	defer func() { c.Cache.Enabled = originalCacheState }()

	// Process files - use simple pagination for global queries, recursive for hierarchical
	result := &FileList{Files: &[]*File{}}
	var err error
	if hasParentFilter(q.Q) {
		c.Log.Println("Using hierarchical processing (parent filter detected)")
		err = c.processFileList(file, file.Path, result, cfg, 0)
	} else {
		c.Log.Println("Using global query processing (no parent filter)")
		err = c.processGlobalQuery(result, cfg)
	}

	// Stop heartbeat logger
	close(heartbeatStop)

	// Stop progress tracker
	if cfg.progressTracker != nil {
		cfg.progressTracker.Done()
	} else if cfg.indeterminateTracker != nil {
		cfg.indeterminateTracker.Done()
	}

	// Close CSV backup
	if cfg.csvEnabled && cfg.csvFile != nil {
		c.closeCSVBackup(cfg)
	}

	// Populate metadata
	result.Metadata = &FileListMetadata{
		Query:          q,
		FilesFound:     cfg.filesFound,
		FoldersScanned: cfg.foldersScanned,
		Duration:       time.Since(startTime),
		PartialSuccess: len(cfg.errors) > 0 && cfg.filesFound > 0,
		Errors:         cfg.errors,
		Cached:         false,
	}

	if cfg.csvEnabled && cfg.csvPath != "" {
		result.Metadata.CSVBackupPath = cfg.csvPath
	}
	if cfg.streamEnabled && cfg.sheetID != "" {
		result.Metadata.SheetURL = fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", cfg.sheetID)
	}

	// Format the sheet after streaming is complete
	if cfg.streamEnabled && cfg.sheetID != "" && cfg.filesFound > 0 {
		c.formatSheetAfterStreaming(cfg)
	}

	// Log completion summary
	c.logCompletionSummary(cfg, result.Metadata)

	// Store in cache if enabled
	if cfg.useCache {
		c.SetCache(cacheKey(), result, 5*time.Minute)
	}

	return result, err
}

func (c *DriveClient) initFilePath(file *File) error {
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

// updateParentInQuery replaces the parent ID in a query string with a new parent ID
// e.g., "'oldId' in parents and ..." becomes "'newId' in parents and ..."
func updateParentInQuery(q string, newParentID string) string {
	re := regexp.MustCompile(`'[^']+' in parents`)
	return re.ReplaceAllString(q, fmt.Sprintf("'%s' in parents", newParentID))
}

// hasParentFilter checks if the query string contains a parent filter clause
// Returns true if query contains "'...' in parents" pattern
func hasParentFilter(q string) bool {
	re := regexp.MustCompile(`'[^']+' in parents`)
	return re.MatchString(q)
}

// truncateFolderName truncates a folder path for display, keeping the end visible
func truncateFolderName(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return "..."
	}
	return "..." + path[len(path)-(maxLen-3):]
}

// ---------------------------------------------------------------------
// GetFileList Helper Methods
// ---------------------------------------------------------------------

// initCSVBackup creates the CSV file for backup
func (c *DriveClient) initCSVBackup(cfg *fileListConfig) error {
	if cfg.csvPath == "" {
		cfg.csvPath = filepath.Join(os.TempDir(),
			fmt.Sprintf("rego_drive_backup_%s.csv", time.Now().Format("20060102_150405")))
	}

	file, err := os.Create(cfg.csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV backup: %w", err)
	}
	cfg.csvFile = file
	cfg.csvWriter = csv.NewWriter(file)

	c.Log.Println("CSV backup initialized:", cfg.csvPath)
	return nil
}

// closeCSVBackup properly closes the CSV writer and file
func (c *DriveClient) closeCSVBackup(cfg *fileListConfig) error {
	if cfg.csvWriter != nil {
		cfg.csvWriter.Flush()
	}
	if cfg.csvFile != nil {
		return cfg.csvFile.Close()
	}
	return nil
}

// initSheetStream initializes the sheets client and creates the sheet if needed
func (c *DriveClient) initSheetStream(cfg *fileListConfig) error {
	// Use DWD client if provided
	if cfg.delegationClient != nil {
		cfg.sheetsClient = cfg.delegationClient.Sheets()
	} else {
		cfg.sheetsClient = c.Client.Sheets()
	}

	// Check if sheet exists, create if not
	spreadsheet, err := cfg.sheetsClient.GetSpreadsheet(cfg.sheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	sheetExists := false
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == cfg.sheetName {
			sheetExists = true
			break
		}
	}

	if !sheetExists {
		c.Log.Println("Creating new sheet:", cfg.sheetName)
		_, err := cfg.sheetsClient.AddSheet(cfg.sheetID, &SheetProperties{
			Title: cfg.sheetName,
		})
		if err != nil {
			return fmt.Errorf("failed to create sheet %s: %w", cfg.sheetName, err)
		}
	}

	c.Log.Println("Sheet stream initialized for:", cfg.sheetName)
	return nil
}

// appendToOutputs writes files to both CSV and Sheet outputs using GenerateValueRange
func (c *DriveClient) appendToOutputs(cfg *fileListConfig, files []*File) error {
	if len(files) == 0 {
		return nil
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	// Get or create a sheets client for flattening
	sheetsClient := cfg.sheetsClient
	if sheetsClient == nil {
		if cfg.delegationClient != nil {
			sheetsClient = cfg.delegationClient.Sheets()
		} else {
			sheetsClient = c.Client.Sheets()
		}
	}

	// Convert []*File to []any for GenerateValueRange
	data := make([]any, len(files))
	for i, file := range files {
		data[i] = file
	}

	headers := &cfg.headers
	vr := sheetsClient.GenerateValueRange(data, cfg.sheetName, headers)

	// Update headers to the flattened version on first call
	if !cfg.headersWritten && !cfg.csvHeadersWritten && len(vr.Values) > 0 {
		cfg.headers = vr.Values[0]
	}

	// Write to CSV
	if cfg.csvEnabled && cfg.csvWriter != nil {
		if !cfg.csvHeadersWritten {
			if len(vr.Values) > 0 {
				if err := cfg.csvWriter.Write(vr.Values[0]); err != nil {
					return fmt.Errorf("failed to write CSV headers: %w", err)
				}
			}
			cfg.csvHeadersWritten = true
		}
		for i := 1; i < len(vr.Values); i++ {
			if err := cfg.csvWriter.Write(vr.Values[i]); err != nil {
				return err
			}
		}
		cfg.csvWriter.Flush()
		if err := cfg.csvWriter.Error(); err != nil {
			return err
		}
	}

	// Write to Sheet
	if cfg.streamEnabled && cfg.sheetsClient != nil {
		if !cfg.headersWritten {
			if err := cfg.sheetsClient.UpdateSpreadsheet(cfg.sheetID, vr); err != nil {
				return err
			}
			cfg.headersWritten = true
		} else {
			if len(vr.Values) > 1 {
				vr.Values = vr.Values[1:]
			} else {
				return nil
			}
			if err := cfg.sheetsClient.AppendSpreadsheet(cfg.sheetID, vr); err != nil {
				return err
			}
		}
	}

	return nil
}

// runHeartbeatLogger logs periodic progress updates for long-running operations
func (c *DriveClient) runHeartbeatLogger(cfg *fileListConfig, startTime time.Time, stop chan struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	heartbeatCount := 0

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			heartbeatCount++
			cfg.mu.Lock()
			files := cfg.filesFound
			folders := cfg.foldersScanned
			errs := len(cfg.errors)
			topComplete := atomic.LoadInt32(&cfg.topLevelComplete)
			topTotal := cfg.topLevelTotal
			cfg.mu.Unlock()

			elapsed := time.Since(startTime)
			rate := float64(files) / elapsed.Minutes()

			c.Log.Println("════════════════════════════════════════════════════════════════")
			c.Log.Printf("[HEARTBEAT #%d] Status: RUNNING", heartbeatCount)
			c.Log.Printf("  Elapsed:       %s", elapsed.Round(time.Second))
			c.Log.Printf("  Files found:   %d (%.1f files/min)", files, rate)
			c.Log.Printf("  Folders:       %d", folders)
			c.Log.Printf("  Top-level:     %d/%d complete", topComplete, topTotal)
			c.Log.Printf("  Errors:        %d", errs)
			if cfg.progressTracker != nil {
				c.Log.Printf("  Summary:       %s", cfg.progressTracker.Summary())
			}

			cfg.folderStatsMu.Lock()
			for path, stats := range cfg.folderStats {
				if !stats.Completed {
					folderElapsed := time.Since(stats.StartTime)
					c.Log.Printf("  [ACTIVE] %s: %d files, %d subfolders, running %s",
						truncateFolderName(path, 60), stats.FilesFound, stats.FolderCount, folderElapsed.Round(time.Second))
				}
			}
			cfg.folderStatsMu.Unlock()

			if cfg.csvEnabled && cfg.csvWriter != nil {
				cfg.mu.Lock()
				cfg.csvWriter.Flush()
				if err := cfg.csvWriter.Error(); err != nil {
					c.Log.Warning("CSV flush error:", err)
				} else {
					c.Log.Printf("  CSV flushed to disk: %s", cfg.csvPath)
				}
				cfg.mu.Unlock()
			}

			cfg.checkpointMu.Lock()
			cfg.lastCheckpoint = time.Now()
			cfg.checkpointMu.Unlock()

			c.Log.Println("════════════════════════════════════════════════════════════════")
		}
	}
}

// formatSheetAfterStreaming formats the sheet header and auto-sizes columns
func (c *DriveClient) formatSheetAfterStreaming(cfg *fileListConfig) {
	spreadsheet, err := cfg.sheetsClient.GetSpreadsheet(cfg.sheetID)
	if err != nil {
		c.Log.Warning("Failed to get spreadsheet for formatting:", err)
		return
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == cfg.sheetName {
			rows := cfg.filesFound + 1
			columns := len(cfg.headers)
			if err := cfg.sheetsClient.FormatHeaderAndAutoSize(cfg.sheetID, &sheet, rows, columns); err != nil {
				c.Log.Warning("Failed to format sheet:", err)
			} else {
				c.Log.Println("Sheet formatted successfully")
			}
			break
		}
	}
}

// logCompletionSummary logs the final summary of the file listing operation
func (c *DriveClient) logCompletionSummary(cfg *fileListConfig, meta *FileListMetadata) {
	c.Log.Println("════════════════════════════════════════════════════════════════")
	c.Log.Println("[COMPLETE] Drive file list operation finished")
	c.Log.Println("════════════════════════════════════════════════════════════════")
	c.Log.Println("  Duration:", meta.Duration.Round(time.Second))
	c.Log.Println("  Files found:", meta.FilesFound)
	c.Log.Println("  Folders scanned:", meta.FoldersScanned)
	if meta.CSVBackupPath != "" {
		c.Log.Println("  CSV backup:", meta.CSVBackupPath)
	}
	if meta.SheetURL != "" {
		c.Log.Println("  Sheet URL:", meta.SheetURL)
	}

	cfg.folderStatsMu.Lock()
	if len(cfg.folderStats) > 0 {
		c.Log.Println("────────────────────────────────────────────────────────────────")
		c.Log.Printf("  Top-level folder breakdown (%d folders):", len(cfg.folderStats))
		for path, stats := range cfg.folderStats {
			status := "✓"
			if !stats.Completed {
				status = "✗ INCOMPLETE"
			}
			duration := stats.EndTime.Sub(stats.StartTime)
			if !stats.Completed {
				duration = time.Since(stats.StartTime)
			}
			c.Log.Printf("    %s %s: %d files, %d subfolders, %s",
				status, truncateFolderName(path, 50), stats.FilesFound, stats.FolderCount, duration.Round(time.Second))
		}
	}
	cfg.folderStatsMu.Unlock()

	if len(meta.Errors) > 0 {
		c.Log.Println("────────────────────────────────────────────────────────────────")
		c.Log.Warning("  Errors encountered:", len(meta.Errors))
		for i, e := range meta.Errors {
			if i >= 5 {
				c.Log.Warning("    ... and", len(meta.Errors)-5, "more errors")
				break
			}
			c.Log.Warning("   ", e)
		}
	}
	c.Log.Println("════════════════════════════════════════════════════════════════")
}

// processFileList processes files recursively with optional streaming support
// depth: 0 = root folder (serial processing for top-level), 1+ = subfolders (parallel processing)
func (c *DriveClient) processFileList(file *File, parentPath string, result *FileList, cfg *fileListConfig, depth int) error {
	// Reset query pagination for this folder
	q := *cfg.query
	q.Q = updateParentInQuery(cfg.query.Q, file.ID)
	q.PageToken = ""

	var folderFiles []*File
	var topLevelFolders []*File

	// Fetch all files in this folder (paginated)
	for {
		filesPage, err := c.fetchFilesPage(q)
		if err != nil {
			cfg.errors = append(cfg.errors, fmt.Errorf("error fetching folder %s: %w", file.ID, err))
			return err
		}

		// Warn if the search was incomplete (API couldn't search all documents)
		if filesPage.IncompleteSearch {
			c.Log.Warning("API returned incomplete search results - some files may be missing")
			c.Log.Warning("Consider narrowing your query or using a different corpora setting")
		}

		for _, file := range *filesPage.Files {
			file.Path = parentPath + "/" + file.Name
			c.Log.Debug("File Path:", file.Path)
			folderFiles = append(folderFiles, file)

			// Only accumulate results if not ignoring memory
			if !cfg.ignoreMemory {
				*result.Files = append(*result.Files, file)
			}

			if file.MimeType == "application/vnd.google-apps.folder" && depth == 0 {
				topLevelFolders = append(topLevelFolders, file)
			}
		}

		if filesPage.NextPageToken == "" {
			break
		}
		q.PageToken = filesPage.NextPageToken
	}

	// At depth 0, set the top-level total for progress bar
	if depth == 0 && len(topLevelFolders) > 0 {
		cfg.topLevelTotal = len(topLevelFolders)
		if cfg.progressTracker != nil {
			cfg.progressTracker.SetTopLevelTotal(len(topLevelFolders))
		}
		c.Log.Println("Found", len(topLevelFolders), "top-level folders to process serially")
	}

	// Separate folders and files
	var foldersToProcess []*File
	nonFolderCount := 0
	for _, f := range folderFiles {
		if f.MimeType == "application/vnd.google-apps.folder" {
			foldersToProcess = append(foldersToProcess, f)
		} else {
			nonFolderCount++
		}
	}

	// Update counts and stream current folder's files
	if len(folderFiles) > 0 {
		cfg.mu.Lock()
		cfg.filesFound += nonFolderCount
		cfg.foldersScanned++
		filesFound := cfg.filesFound
		foldersScanned := cfg.foldersScanned
		cfg.mu.Unlock()

		// Update progress
		if cfg.progressTracker != nil {
			cfg.progressTracker.Update(
				int64(filesFound),
				int64(foldersScanned),
				parentPath,
				int(atomic.LoadInt32(&cfg.topLevelComplete)),
			)
		}

		// Filter out folders if excludeFolders is enabled
		filesToStream := folderFiles
		if cfg.excludeFolders {
			filesToStream = make([]*File, 0, nonFolderCount)
			for _, f := range folderFiles {
				if f.MimeType != "application/vnd.google-apps.folder" {
					filesToStream = append(filesToStream, f)
				}
			}
		}

		// Append to outputs (CSV and/or Sheet)
		if (cfg.csvEnabled || cfg.streamEnabled) && len(filesToStream) > 0 {
			if err := c.appendToOutputs(cfg, filesToStream); err != nil {
				c.Log.Warning("Output append error:", err)
				cfg.errors = append(cfg.errors, fmt.Errorf("output append error: %w", err))
			}
		}
	}

	// Process subfolders based on depth
	if depth == 0 {
		// SERIAL processing for top-level folders (better progress visibility)
		for i, folder := range foldersToProcess {
			folderStartTime := time.Now()

			cfg.mu.Lock()
			startingFiles := cfg.filesFound
			startingFolders := cfg.foldersScanned
			startingErrors := len(cfg.errors)
			cfg.mu.Unlock()

			cfg.folderStatsMu.Lock()
			cfg.folderStats[folder.Path] = &DriveFolderStats{
				StartTime: folderStartTime,
			}
			cfg.folderStatsMu.Unlock()

			c.Log.Println("────────────────────────────────────────────────────────────────")
			c.Log.Printf("[START %d/%d] Top-level folder: %s", i+1, len(foldersToProcess), folder.Name)
			c.Log.Printf("  Path: %s", folder.Path)
			c.Log.Printf("  Running totals: %d files, %d folders", startingFiles, startingFolders)
			c.Log.Println("────────────────────────────────────────────────────────────────")

			if cfg.progressTracker != nil {
				cfg.progressTracker.AddActiveFolder(folder.Path)
			}

			err := c.processFileList(folder, folder.Path, result, cfg, 1)
			if err != nil {
				c.Log.Error("Error processing top-level folder", folder.Name, ":", err)
				cfg.errors = append(cfg.errors, err)
			}

			if cfg.progressTracker != nil {
				cfg.progressTracker.RemoveActiveFolder(folder.Path)
			}

			folderDuration := time.Since(folderStartTime)
			cfg.mu.Lock()
			endingFiles := cfg.filesFound
			endingFolders := cfg.foldersScanned
			endingErrors := len(cfg.errors)
			cfg.mu.Unlock()

			folderFilesFound := endingFiles - startingFiles
			folderFoldersScanned := endingFolders - startingFolders
			folderErrorCount := endingErrors - startingErrors

			cfg.folderStatsMu.Lock()
			if stats, ok := cfg.folderStats[folder.Path]; ok {
				stats.EndTime = time.Now()
				stats.FilesFound = folderFilesFound
				stats.FolderCount = folderFoldersScanned
				stats.ErrorCount = folderErrorCount
				stats.Completed = true
			}
			cfg.folderStatsMu.Unlock()

			atomic.AddInt32(&cfg.topLevelComplete, 1)
			if cfg.progressTracker != nil {
				cfg.progressTracker.Update(
					int64(endingFiles),
					int64(endingFolders),
					folder.Path+" ✓",
					int(atomic.LoadInt32(&cfg.topLevelComplete)),
				)
			}

			rate := float64(folderFilesFound) / folderDuration.Minutes()
			c.Log.Println("════════════════════════════════════════════════════════════════")
			c.Log.Printf("[COMPLETE %d/%d] Top-level folder: %s", i+1, len(foldersToProcess), folder.Name)
			c.Log.Printf("  Duration:        %s", folderDuration.Round(time.Second))
			c.Log.Printf("  Files found:     %d (%.1f files/min)", folderFilesFound, rate)
			c.Log.Printf("  Subfolders:      %d", folderFoldersScanned)
			if folderErrorCount > 0 {
				c.Log.Printf("  Errors:          %d", folderErrorCount)
			}
			c.Log.Printf("  Running totals:  %d files, %d folders", endingFiles, endingFolders)
			c.Log.Println("════════════════════════════════════════════════════════════════")

			foldersToProcess[i] = nil
		}
	} else {
		// PARALLEL processing for nested subfolders (depth 1+)
		if len(foldersToProcess) > 0 {
			sem := make(chan struct{}, runtime.GOMAXPROCS(0))
			filesChannel := make(chan *FileList, len(foldersToProcess))
			filesErrChannel := make(chan error, len(foldersToProcess))
			var wg sync.WaitGroup

			for _, folder := range foldersToProcess {
				wg.Add(1)
				go c.fetchSubFiles(folder, folder.Path, cfg, sem, filesChannel, filesErrChannel, &wg)
			}

			go func() {
				wg.Wait()
				close(sem)
				close(filesChannel)
				close(filesErrChannel)
			}()

			if !cfg.ignoreMemory {
				for subFiles := range filesChannel {
					*result.Files = append(*result.Files, *subFiles.Files...)
					subFiles.Files = nil
				}
			}

			for err := range filesErrChannel {
				cfg.errors = append(cfg.errors, err)
				c.Log.Error("Subfolder error:", err)
			}
		}
	}

	return nil
}

// fetchSubFiles fetches files from a subfolder with streaming support
func (c *DriveClient) fetchSubFiles(file *File, parentPath string, cfg *fileListConfig, sem chan struct{}, filesChannel chan<- *FileList, filesErrChannel chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	sem <- struct{}{}
	defer func() { <-sem }()

	if cfg.progressTracker != nil {
		cfg.progressTracker.AddActiveFolder(parentPath)
		defer cfg.progressTracker.RemoveActiveFolder(parentPath)
	}

	subResult := &FileList{Files: &[]*File{}}

	err := c.processFileList(file, parentPath, subResult, cfg, 1)
	if err != nil {
		filesErrChannel <- err
		return
	}

	filesChannel <- subResult
}

/*
# Save File List to Google Sheet
*/
func (c *DriveClient) SaveFileListToSheet(fileList *FileList, sheetID string, headers *[]string) error {

	c.Log.Println("Saving File List to spreadsheet")
	if headers == nil {
		headers = &[]string{"id", "name", "path", "md5Checksum", "mimeType", "originalFilename", "owners", "parents", "shortcutDetails"}
	}

	err := c.Sheets().SaveToSheet(fileList.Files, sheetID, (*fileList.Files)[0].ID, headers)
	if err != nil {
		return err
	}
	return nil
}

/*
 * # Fetch Files Page
 * Fetches a page of files
 * drive/v3/files
 * https://developers.google.com/drive/api/v3/reference/files/list
 */
func (c *DriveClient) fetchFilesPage(q DriveFileQuery) (*FileList, error) {
	url := c.BuildURL(DriveFiles, nil)

	filesPage, err := do[FileList](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	return &filesPage, nil
}

// processGlobalQuery handles global queries (no parent filter) using a hybrid approach.
// Instead of resolving paths per-file (O(N) API calls), it groups files by parent and
// resolves each unique parent path once (O(unique parents) API calls).
//
// Performance strategy:
// - Phase 1: Collect all files from all pages (no blocking on path resolution)
// - Phase 2: Group files by their immediate parent ID
// - Phase 3: Resolve unique parent paths concurrently (one per unique parent)
// - Phase 4: Apply parent paths to children in-memory (no API calls)
//
// This reduces path resolution from O(files * depth) to O(unique_parents * depth),
// typically a 10-100x improvement for queries returning many files in shared folders.
func (c *DriveClient) processGlobalQuery(result *FileList, cfg *fileListConfig) error {
	c.Log.Println("Starting global query processing (hybrid approach)...")
	c.Log.Debug("Query:", cfg.query.Q)

	// ═══════════════════════════════════════════════════════════════════════
	// Phase 1: Collect all files from all pages
	// ═══════════════════════════════════════════════════════════════════════
	var allFiles []*File
	q := *cfg.query
	q.PageToken = ""
	pageNum := 0

	for {
		pageNum++
		c.Log.Debug("Fetching page", pageNum, "...")
		filesPage, err := c.fetchFilesPage(q)
		if err != nil {
			cfg.errors = append(cfg.errors, fmt.Errorf("error fetching files page %d: %w", pageNum, err))
			return err
		}

		if filesPage.IncompleteSearch {
			c.Log.Warning("API returned incomplete search results - some files may be missing")
		}

		if filesPage.Files != nil && len(*filesPage.Files) > 0 {
			allFiles = append(allFiles, *filesPage.Files...)
			c.Log.Println("Page", pageNum, "fetched", len(*filesPage.Files), "files. Total:", len(allFiles))

			// Update progress during collection phase
			if cfg.indeterminateTracker != nil {
				cfg.indeterminateTracker.Update(int64(len(allFiles)), fmt.Sprintf("Collecting files (page %d)...", pageNum))
			}
		}

		if filesPage.NextPageToken == "" {
			break
		}
		q.PageToken = filesPage.NextPageToken
	}

	if len(allFiles) == 0 {
		c.Log.Println("No files found matching query")
		return nil
	}

	c.Log.Println("Collection complete:", len(allFiles), "files from", pageNum, "pages")

	// ═══════════════════════════════════════════════════════════════════════
	// Phase 2: Group files by their immediate parent ID
	// ═══════════════════════════════════════════════════════════════════════
	parentGroups := make(map[string][]*File) // parentID -> children
	var rootFiles []*File                    // Files with no parent (root or shared)

	for _, f := range allFiles {
		if len(f.Parents) > 0 {
			parentGroups[f.Parents[0]] = append(parentGroups[f.Parents[0]], f)
		} else {
			// Files with no parent - at root or shared with me
			if f.Shared && !f.OwnedByMe {
				f.Path = "Shared with me/" + f.Name
			} else {
				f.Path = f.Name
			}
			rootFiles = append(rootFiles, f)
		}
	}

	c.Log.Println("Grouped into", len(parentGroups), "unique parents +", len(rootFiles), "root files")

	// ═══════════════════════════════════════════════════════════════════════
	// Phase 3: Resolve unique parent paths concurrently
	// ═══════════════════════════════════════════════════════════════════════
	// Initialize in-memory caches for fast path resolution (avoids rego cache gob overhead)
	cfg.pathCache = &sync.Map{}
	cfg.fileCache = &sync.Map{}
	var pathErrors sync.Map // parentID -> error
	var wg sync.WaitGroup
	var resolvedCount int32

	if cfg.indeterminateTracker != nil {
		cfg.indeterminateTracker.Update(int64(len(allFiles)), fmt.Sprintf("Resolving %d parent paths...", len(parentGroups)))
	}

	for parentID := range parentGroups {
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()

			// Pass both pathCache and fileCache for maximum performance
			// GetFilePath stores the result in pathCache, so we don't need the return value
			_, err := c.GetFilePath(pid, cfg.pathCache, cfg.fileCache)
			if err != nil {
				c.Log.Warning("Failed to resolve parent path for", pid, ":", err)
				cfg.pathCache.Store(pid, "[unknown]")
				pathErrors.Store(pid, err)
			}

			current := atomic.AddInt32(&resolvedCount, 1)
			if cfg.indeterminateTracker != nil && current%5 == 0 {
				cfg.indeterminateTracker.Update(int64(len(allFiles)), fmt.Sprintf("Resolved %d/%d parent paths", current, len(parentGroups)))
			}
		}(parentID)
	}
	wg.Wait()

	c.Log.Println("Resolved", resolvedCount, "parent paths")

	// ═══════════════════════════════════════════════════════════════════════
	// Phase 4: Apply parent paths to all children (in-memory, no API calls)
	// ═══════════════════════════════════════════════════════════════════════
	if cfg.indeterminateTracker != nil {
		cfg.indeterminateTracker.Update(int64(len(allFiles)), "Applying paths to files...")
	}

	for parentID, children := range parentGroups {
		parentPath, _ := cfg.pathCache.Load(parentID)
		pathStr := parentPath.(string)

		for _, f := range children {
			f.Path = pathStr + "/" + f.Name
		}
	}

	// ═══════════════════════════════════════════════════════════════════════
	// Phase 5: Accumulate results and stream to outputs
	// ═══════════════════════════════════════════════════════════════════════
	cfg.mu.Lock()
	cfg.filesFound = len(allFiles)
	cfg.mu.Unlock()

	// Accumulate results if not ignoring memory
	if !cfg.ignoreMemory {
		*result.Files = append(*result.Files, allFiles...)
	}

	// Final progress update
	if cfg.indeterminateTracker != nil {
		cfg.indeterminateTracker.Update(int64(len(allFiles)), "Complete")
	}

	// Stream to outputs (CSV and/or Sheet)
	filesToStream := allFiles
	if cfg.excludeFolders {
		filtered := make([]*File, 0, len(allFiles))
		for _, f := range allFiles {
			if f.MimeType != "application/vnd.google-apps.folder" {
				filtered = append(filtered, f)
			}
		}
		filesToStream = filtered
	}

	if (cfg.csvEnabled || cfg.streamEnabled) && len(filesToStream) > 0 {
		if err := c.appendToOutputs(cfg, filesToStream); err != nil {
			c.Log.Warning("Output append error:", err)
			cfg.errors = append(cfg.errors, fmt.Errorf("output append error: %w", err))
		}
	}

	// Collect path resolution errors
	pathErrors.Range(func(key, value interface{}) bool {
		cfg.errors = append(cfg.errors, fmt.Errorf("path resolution failed for parent %s: %w", key.(string), value.(error)))
		return true
	})

	return nil
}

/*
 * # Get File Path
 * Constructs the path of a file by walking up the parent chain.
 * drive/v3/files/{fileId}
 * @param {string} id - The ID of the file or shortcut to get the path of.
 * @param {*sync.Map} caches - Optional in-memory caches for performance:
 *                             caches[0]: pathCache (fileID -> path string)
 *                             caches[1]: fileCache (fileID -> *File)
 * https://developers.google.com/drive/api/v3/reference/files/get
 */
func (c *DriveClient) GetFilePath(id string, caches ...*sync.Map) (string, error) {
	var pathCache, fileCache *sync.Map
	if len(caches) > 0 {
		pathCache = caches[0]
	}
	if len(caches) > 1 {
		fileCache = caches[1]
	}

	// Check path cache first
	if pathCache != nil {
		if cached, ok := pathCache.Load(id); ok {
			return cached.(string), nil
		}
	}

	file, err := c.GetFile(id, fileCache)
	if err != nil {
		return "", err
	}

	// If file has no parents, it's in root (My Drive) or "Shared With Me" via the WebUI
	if len(file.Parents) == 0 {
		var path string
		if file.Shared && !file.OwnedByMe {
			path = "Shared with me/" + file.Name
		} else {
			path = file.Name
		}
		if pathCache != nil {
			pathCache.Store(id, path)
		}
		return path, nil
	}

	// Recursively resolve parent path
	parentPath, err := c.GetFilePath(file.Parents[0], caches...)
	if err != nil {
		return "", err
	}

	fullPath := parentPath + "/" + file.Name
	if pathCache != nil {
		pathCache.Store(id, fullPath)
	}
	return fullPath, nil
}
