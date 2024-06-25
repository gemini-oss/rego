/*
# Backupify - Exports

This package initializes all the methods for functions which interact with the Backupify's Exports:
https://{node-url}.backupify.com/{customerID}/{service}/export

* Currently only supports GoogleDrive exports

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/backupify/exports.go
package backupify

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ExportClient for chaining methods
type ExportClient struct {
	*Client
}

// Entry point for export-related operations
func (c *Client) Exports() *ExportClient {
	return &ExportClient{
		Client: c,
	}
}

type ExportQuery struct {
	Type    string `json:"type"`    // Type of query. e.g. 'export'
	AppType string `json:"appType"` // Type of application. e.g. 'GoogleDrive'
	ID      int    `json:"id"`      // ID of the export
	EXT     string `json:"ext"`     // Extension of the file. e.g. 'zip'
}

func (c *ExportClient) ExportUsers(users *Users) ([]Exports, error) {
	var totalExports []Exports
	for _, user := range users.Data {
		exports, err := c.ExportUser(user)
		if err != nil {
			c.Log.Warning("Error exporting user: ", user.ID, " ", user.Email)
			c.Log.Warning(err)
			continue
		}
		totalExports = append(totalExports, *exports)
		time.Sleep(1 * time.Second)
	}
	return totalExports, nil
}

func (c *ExportClient) ExportUser(user *User) (*Exports, error) {
	c.Log.Println("Perfect Backups: ", len(user.PerfectBackups))
	c.Log.Println("Snapshots: ", len(user.Snapshots))

	var exports Exports
	for _, snapshot := range user.Snapshots {
		export, err := c.generateExport(user.ID, snapshot.ID)
		if err != nil {
			c.Log.Warning("Error exporting snapshot: ", user.ID, " ", snapshot.ID)
			c.Log.Warning(err)
			continue
		}
		exports = append(exports, export)
		time.Sleep(1 * time.Second)
	}

	return &exports, nil
}

func (c *ExportClient) generateExport(userID int, snapshotID int64) (*Export, error) {
	url := c.BuildURL(restoreExportAction)

	exportPayload := ExportPayload{
		ActionType:         "export",
		AppType:            c.AppType,
		SnapshotID:         fmt.Sprintf("%d", snapshotID),
		Token:              c.exportToken,
		IncludePermissions: true,
		IncludeAttachments: false,
		Services:           []interface{}{userID},
	}

	export, err := do[Export](c.Client, "POST", url, nil, exportPayload)
	if err != nil {
		return nil, err
	}

	c.Log.Println("Export started: ", export.ResponseData.ID)
	return &export, nil
}

func (c *Client) CheckExportFilters(activities *Activities) {
	for _, activity := range activities.Export.Items {
		switch filters := activity.Run.Description.Filters.(type) {
		case map[string]interface{}:
			if isDeleted, ok := filters["isDeleted"].(string); ok {
				fmt.Printf("Activity %s has filter isDeleted with value: %s\n", activity.Status, isDeleted)
			}
		case []interface{}:
		default:
		}
	}
}

func (c *ExportClient) DownloadAvailableExports(activities *Activities) [][]string {
	c.Log.Println("Downloading all exports from Backupify...")

	// Initialize exportReports with headers
	var exportReports = make([][]string, 1, len(activities.Export.Items)+1)
	exportReports[0] = []string{"Service Email", "Snapshot ID", "URL", "Download Path", "File Name", "Export ID", "Downloaded At"}

	// Buffered channel to limit concurrent downloads
	sem := make(chan struct{}, 5)

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, activity := range activities.Export.Items {
		wg.Add(1)
		go func(activity *Item) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if activity.Status == "completed" && activity.Export.Status == "Download" {
				export := &Export{
					ResponseData: ResponseData{
						AppType: activity.Run.AppType,
						ID:      activity.Run.ID,
					},
					SnapshotDate: activity.Export.SnapshotDate,
				}
				report, err := c.DownloadExport(activity, export)
				if err != nil {
					c.Log.Fatal(err)
				}
				mu.Lock()
				exportReports = append(exportReports, report)
				mu.Unlock()
			} else if activity.Status == "in progress" {
				c.Log.Println("Activity is in progress. Skipping...")
			}
		}(activity)
	}
	wg.Wait()

	return exportReports
}

func (c *ExportClient) DownloadExport(activity *Item, export *Export) ([]string, error) {
	url := c.BuildURL(download)
	query := ExportQuery{
		Type:    "export",
		AppType: export.ResponseData.AppType,
		ID:      export.ResponseData.ID,
		EXT:     "zip",
	}

	url = fmt.Sprintf("%s?type=%s&appType=%s&id=%d&ext=%s", url, query.Type, query.AppType, query.ID, query.EXT)
	c.Log.Println("Downloading Export for: ", activity.Run.Description.Services[0].ServiceEmail, "Snapshot ID: ", activity.Run.Description.Snapshot, "Export ID: ", export.ResponseData.ID)
	c.Log.Debug(url)

	pwd, err := os.Getwd()
	if err != nil {
		c.Log.Fatal(err)
	}

	downloadPath := filepath.Join(pwd, fmt.Sprintf(
		"backupify/%s/%s",
		activity.Run.AppType,
		activity.Run.Description.Services[0].ServiceEmail,
	))

	switch export.SnapshotDate {
	case "":
		c.Log.Warning("Snapshot date is empty. Using export ID in place of snapshot date.")
		export.SnapshotDate = fmt.Sprintf("exp_%d", export.ResponseData.ID)
	default:
		// Convert `MM/DD/YYYY` to `YYYY_MM_DD`
		export.SnapshotDate = strings.ReplaceAll(strings.ReplaceAll(export.SnapshotDate[6:]+"_"+export.SnapshotDate[:2]+"_"+export.SnapshotDate[3:5], "/", ""), " ", "")
	}

	fileName := fmt.Sprintf(
		"%s-%s-snap_%d-%s.%s",
		strings.Split(activity.Run.Description.Services[0].ServiceEmail, "@")[0],
		activity.Run.AppType,
		activity.Run.Description.Snapshot,
		export.SnapshotDate,
		query.EXT,
	)

	err = c.HTTP.DownloadFile(url, downloadPath, fileName, false)
	if err != nil {
		c.Log.Fatal(err)
	}

	downloadReport := []string{
		activity.Run.Description.Services[0].ServiceEmail,    // Service Email
		fmt.Sprintf("%d", activity.Run.Description.Snapshot), // Snapshot ID
		url,          // URL
		downloadPath, // Download Path
		fileName,     // File Name
		fmt.Sprintf("%d", export.ResponseData.ID), // Export ID
		time.Now().Format("2006-01-02 15:04:05"),  // Downloaded At
	}
	return downloadReport, nil
}

func (c *ExportClient) DeleteExport(activity *Item, export *Export) error {
	url := c.BuildURL(delete)

	deleteQuery := DeletePayload{
		Type:    "export",
		AppType: AppType(export.ResponseData.AppType),
		ID:      export.ResponseData.ID,
	}

	_, err := do[Export](c.Client, "POST", url, deleteQuery, nil)
	if err != nil {
		c.Log.Fatal(err)
	}

	return nil
}
