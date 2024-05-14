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
	"fmt"
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

func (c *ActivityClient) GetActivities() (*Activities, error) {
	url := c.BuildURL(getActivities)
	cache_key := fmt.Sprintf("%s_%s", getActivities, c.AppType)
	c.Log.Println("Getting activities from Backupify...")

	var cache Activities
	if c.GetCache(cache_key, &cache) {
		return &cache, nil
	}

	activitiesPayload := ActivitiesPayload{
		AppType: c.AppType,
	}

	activities, err := do[ActivitiesResponse](c.Client, "POST", url, nil, activitiesPayload)
	if err != nil {
		c.Log.Fatal(err)
	}

	c.SetCache(cache_key, activities.Activities, 5*time.Minute)
	return &activities.Activities, nil
}
