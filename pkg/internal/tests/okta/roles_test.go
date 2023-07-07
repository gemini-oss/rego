/*
# Okta Roles - Test

This package tests functions related to the Okta Roles API:
https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/#tag/Role

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/okta/roles_test.go
package okta_test

import (
	"testing"
)

func TestListAllRoles(t *testing.T) {
	expectedResponse := `
	{
		"roles": [
			{
				"id": "role1",
				"assignmentType": "GROUP",
				"label": "Role 1"
			},
			{
				"id": "role2",
				"assignmentType": "USER",
				"label": "Role 2"
			}
		]
	}`

	server, cleanup := setupTestServer(t, "/iam/roles", expectedResponse)
	defer cleanup()

	client := setupTestClient(server.URL)

	roles, err := client.ListAllRoles()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(roles.Roles) != 2 {
		t.Fatalf("Expected 2 roles, got %d", len(roles.Roles))
	}

	if roles.Roles[0].ID != "role1" {
		t.Errorf("Expected role ID `role1`, got `%s`", roles.Roles[0].ID)
	}
}

func TestGetRole(t *testing.T) {
	expectedResponse := `
	{
		"id": "role1",
		"assignmentType": "GROUP",
		"label": "Role 1"
	}`

	server, cleanup := setupTestServer(t, "/iam/roles/role1", expectedResponse)
	defer cleanup()

	client := setupTestClient(server.URL)

	role, err := client.GetRole("role1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if role.ID != "role1" {
		t.Errorf("Expected role ID `role1`, got `%s`", role.ID)
	}
}

func TestGetUserRoles(t *testing.T) {
	expectedResponse := `
	[
		{
			"id": "role1",
			"assignmentType": "GROUP",
			"label": "Role 1"
		},
		{
			"id": "role2",
			"assignmentType": "USER",
			"label": "Role 2"
		}
	]`

	server, cleanup := setupTestServer(t, "/users/user1/roles", expectedResponse)
	defer cleanup()

	client := setupTestClient(server.URL)

	userRoles, err := client.GetUserRoles("user1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(userRoles.Roles) != 2 {
		t.Fatalf("Expected 2 roles, got %d", len(userRoles.Roles))
	}

	if userRoles.Roles[0].ID != "role1" {
		t.Errorf("Expected role ID `role1`, got `%s`", userRoles.Roles[0].ID)
	}
}
