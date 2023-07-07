/*
# Okta Users - Test

This package tests functions related to the Okta Users API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/#tag/User

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/okta/users_test.go
package okta_test

import (
	"testing"
)

// Test ListAllUsers
func TestListAllUsers(t *testing.T) {
	server, cleanup := setupTestServer(t, "/users?limit=200&search=status+eq+%22STAGED%22+or+status+eq+%22PROVISIONED%22+or+status+eq+%22ACTIVE%22+or+status+eq+%22RECOVERY%22+or+status+eq+%22LOCKED_OUT%22+or+status+eq+%22PASSWORD_EXPIRED%22+or+status+eq+%22SUSPENDED%22+or+status+eq+%22DEPROVISIONED%22",
		`[
			{
				"id": "1",
				"status": "STAGED"
			},
			{
				"id": "2",
				"status": "PROVISIONED"
			},
			{
				"id": "3",
				"status": "ACTIVE"
			},
			{
				"id": "4",
				"status": "RECOVERY"
			},
			{
				"id": "5",
				"status": "LOCKED_OUT"
			},
			{
				"id": "6",
				"status": "PASSWORD_EXPIRED"
			},
			{
				"id": "7",
				"status": "SUSPENDED"
			},
			{
				"id": "8",
				"status": "DEPROVISIONED"
			}
		]`)
	defer cleanup()

	client := setupTestClient(server.URL)
	users, err := client.ListAllUsers()

	if err != nil {
		t.Errorf("Expected no error, got `%v`", err)
	}

	if len(*users) != 8 {
		t.Errorf("Expected `8` users, got `%d`", len(*users))
	}
}

// Test ListActiveUsers
func TestListActiveUsers(t *testing.T) {
	server, cleanup := setupTestServer(t, "/users?limit=200&search=status+eq+%22ACTIVE%22",
		`[
			{
				"id": "1",
				"status": "ACTIVE"
			},
			{
				"id": "2",
				"status": "STAGED"
			}
		]`)
	defer cleanup()

	client := setupTestClient(server.URL)
	users, err := client.ListActiveUsers()

	if err != nil {
		t.Errorf("Expected no error, got `%v`", err)
	}

	for _, user := range *users {
		if user.Status != "ACTIVE" {
			t.Errorf("Expected user status `ACTIVE`, got `%s`", user.Status)
		}
	}
}

// Test GetUser
func TestGetUser(t *testing.T) {
	server, cleanup := setupTestServer(t, "/users/1",
		`{
			"id": "1",
			"status": "ACTIVE"
		}`)
	defer cleanup()

	client := setupTestClient(server.URL)
	user, err := client.GetUser("1")

	if err != nil {
		t.Errorf("Expected no error, got `%v`", err)
	}

	if user.ID != "1" {
		t.Errorf("Expected user ID `1`, got `%s`", user.ID)
	}
}
