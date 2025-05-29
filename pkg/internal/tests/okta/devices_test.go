/*
# Okta Devices - Test

This package tests functions related to the Okta Devices API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/#tag/Device

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/okta/devices_test.go
package okta_test

import (
	"testing"
)

func TestListAllDevices(t *testing.T) {
	expectedResponse := `
	[
		{
			"id": "device1",
			"status": "active",
			"profile": {
				"displayName": "Device 1"
			}
		},
		{
			"id": "device2",
			"status": "inactive",
			"profile": {
				"displayName": "Device 2"
			}
		}
	]`

	server, cleanup := setupTestServer(t, "/devices", expectedResponse)
	defer cleanup()

	client := setupTestClient(server.URL)

	devices, err := client.Devices().ListAllDevices()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(*devices) != 2 {
		t.Fatalf("Expected 2 devices, got %d", len(*devices))
	}

	if (*devices)[0].ID != "device1" {
		t.Errorf("Expected device ID `device1`, got `%s`", (*devices)[0].ID)
	}
}

func TestListUsersForDevice(t *testing.T) {
	expectedResponse := `
	[
		{
			"created": "2023-06-25T17:55:56.227Z",
			"managementStatus": "managed",
			"user": {
				"id": "user1",
				"status": "ACTIVE",
				"profile": {
					"login": "test@example.com",
					"firstName": "Test",
					"lastName": "User"
				}
			}
		}
	]`

	server, cleanup := setupTestServer(t, "/devices/device1/users", expectedResponse)
	defer cleanup()

	client := setupTestClient(server.URL)

	users, err := client.Devices().ListUsersForDevice("device1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(*users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(*users))
	}

	if (*users)[0].User.ID != "user1" {
		t.Errorf("Expected user ID `user1`, got `%s`", (*users)[0].User.ID)
	}
}
