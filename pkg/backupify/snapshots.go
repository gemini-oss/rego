/*
# Backupify - Snapshots

This package initializes all the methods for functions which interact with the Backupify's Exports:
https://{node-url}.backupify.com/{customerID}/{service}/serviceSnaps

* Currently only supports GoogleDrive exports

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/backupify/snapshots.go
package backupify

import (
	"time"
)

// SnapshotClient for chaining methods
type SnapshotClient struct {
	*Client
}

// Entry point for export-related operations
func (c *Client) Snapshots() *SnapshotClient {
	return &SnapshotClient{
		Client: c,
	}
}

func (c *SnapshotClient) GetSnapshotDates(appType AppType, user *User) (*Snapshots, error) {
	url := c.BuildURL(serviceSnapshots)
	c.Log.Println("Getting formatted snapshots from Backupify...")

	var cache Snapshots
	if c.GetCache(user.Email, &cache) {
		return &cache, nil
	}

	snapshotsPayload := SnapshotsPayload{
		AppType:   appType,
		ServiceID: user.ID,
	}

	snapshots, err := do[Snapshots](c.Client, "POST", url, nil, snapshotsPayload)
	if err != nil {
		c.Log.Fatal(err)
	}

	c.SetCache(user.Email, snapshots, 3*time.Hour)
	return &snapshots, nil
}
