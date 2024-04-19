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
	"sync"

	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ActivityClient for chaining methods
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

func (c *ExportClient) ExportUsers(users *Users) error {
	for _, user := range users.Data {
		_, err := c.ExportUser(user)
		if err != nil {
			c.Log.Fatal(err)
		}
	}
	return nil
}

func (c *ExportClient) ExportUser(user *User) (*Exports, error) {
	snapshotID := user.LatestSnap.(float64)

	export, err := c.generateExport(c.exportToken, user.ID, snapshotID)
	if err != nil {
		c.Log.Fatal(err)
	}
	exports := Exports{export}

	return &exports, nil
}

func (c *ExportClient) generateExport(token string, userID int, snapshotID float64) (*Export, error) {
	url := c.BuildURL(restoreExportAction)

	exportPayload := ExportPayload{
		ActionType:         "export",
		AppType:            "GoogleDrive",
		SnapshotID:         fmt.Sprintf("%.0f", snapshotID),
		Token:              token,
		IncludePermissions: true,
		IncludeAttachments: false,
		Services:           []interface{}{userID},
	}

	c.HTTP.Headers["Accept"] = requests.All
	export, err := do[Export](c.Client, "POST", url, nil, exportPayload)
	if err != nil {
		c.Log.Fatal(err)
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

func (c *ExportClient) DownloadAvailableExports(activities *Activities) {
	c.Log.Println("Downloading all exports from Backupify...")

	var wg sync.WaitGroup
	for _, activity := range activities.Export.Items {
		wg.Add(1)
		go func(activity *Item) {
			defer wg.Done()

			if activity.Status == "completed" && activity.Export.Status == "Download" {
				export := &Export{
					ResponseData: ResponseData{
						AppType: activity.Run.AppType,
						ID:      activity.Run.ID,
					},
				}
				c.DownloadExport(activity, export)
			} else if activity.Status == "in progress" {
				c.Log.Println("Activity is in progress. Skipping...")
			}
		}(activity)
	}
	wg.Wait()
}

func (c *ExportClient) DownloadExport(activity *Item, export *Export) error {
	url := c.BuildURL(download)
	query := ExportQuery{
		Type:    "export",
		AppType: export.ResponseData.AppType,
		ID:      export.ResponseData.ID,
		EXT:     "zip",
	}

	url = fmt.Sprintf("%s?type=%s&appType=%s&id=%d&ext=%s", url, query.Type, query.AppType, query.ID, query.EXT)
	c.Log.Println("Downloading export for: ", activity.Run.Description.Services[0].ServiceEmail, "Snapshot ID", export.ResponseData.ID)
	c.Log.Debug(url)

	err := c.HTTP.DownloadFile(url,
		fmt.Sprintf(
			"./backupify/%s/%s",
			activity.Run.AppType,
			activity.Run.Description.Services[0].ServiceEmail,
		),
		fmt.Sprintf(
			"%s-%s-%d.%s",
			activity.Run.Description.Services[0].ServiceEmail,
			activity.Run.AppType,
			export.ResponseData.ID,
			query.EXT,
		),
	)
	if err != nil {
		c.Log.Fatal(err)
	}

	return nil
}
