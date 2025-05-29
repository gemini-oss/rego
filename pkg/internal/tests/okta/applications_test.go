/*
# Okta Applications - Test

This package tests functions related to the Okta Applications API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/#tag/Application

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/okta/applications_test.go
package okta_test

import (
	"testing"
)

func TestListAllApplications(t *testing.T) {
	expectedResponse := `
	[
		{
			"id": "app1",
			"status": "active",
			"profile": {
				"displayName": "Application 1"
			}
		},
		{
			"id": "app2",
			"status": "inactive",
			"profile": {
				"displayName": "Application 2"
			}
		}
	]`

	server, cleanup := setupTestServer(t, "/apps", expectedResponse)
	defer cleanup()

	client := setupTestClient(server.URL)

	apps, err := client.Applications().ListAllApplications()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(*apps) != 2 {
		t.Fatalf("Expected 2 applications, got %d", len(*apps))
	}

	if (*apps)[0].ID != "app1" {
		t.Errorf("Expected app ID `app1`, got `%s`", (*apps)[0].ID)
	}
}
