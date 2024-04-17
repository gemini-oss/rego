/*
# Backupify - Activities

This package initializes all the methods for functions which interact with Backupify's Activities:
{Exports, Restores, Backups}

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/backupify/activities.go
package backupify

import (
	"time"
)

// ActivityClient for chaining methods
type ActivityClient struct {
	*Client
}

// Entry point for activity-related operations
func (c *Client) Activities() *ActivityClient {
	return &ActivityClient{
		Client: c,
	}
}

func (c *ActivityClient) GetActivities(appType AppType) (*Activities, error) {
	url := c.BuildURL(getActivities)
	c.Log.Println("Getting activities from Backupify...")

	var cache Activities
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	activitiesPayload := ActivitiesPayload{
		AppType: appType,
	}

	activities, err := do[ActivitiesResponse](c.Client, "POST", url, nil, activitiesPayload)
	if err != nil {
		c.Log.Fatal(err)
	}

	c.SetCache(url, activities.Activities, 5*time.Minute)
	return &activities.Activities, nil
}
