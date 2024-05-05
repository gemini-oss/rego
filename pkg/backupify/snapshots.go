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
	"fmt"
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
	cacheKey := fmt.Sprintf("%d_%s_%s", user.ID, user.Name, string(appType))
	c.Log.Println("Get dates for", appType, "Backupify snapshots:", user.Email, "(", user.ID, ")")

	if (*user).Snapshots == nil || len((*user).Snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots found for user %d", user.ID)
	}

	var cache Snapshots
	if c.GetCache(cacheKey, &cache) {
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

	c.SetCache(cacheKey, snapshots, 24*time.Hour)
	return &snapshots, nil
}
