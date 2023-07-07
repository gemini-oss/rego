/*
# Google Discovery API - Tests

This package tests functions which generate and organize the Google API
https://developers.google.com/workspace

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/google/api_test.go
package google_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gemini-oss/rego/pkg/google"
)

func TestFetchGoogleAPIScopes(t *testing.T) {
	_, _, err := google.FetchDirectoryEndpoints()
	if err != nil {
		t.Errorf("Failed to fetch Google API scopes: %v", err)
	}
}

func TestReadDiscoveryDirectory(t *testing.T) {
	_, _, err := google.ReadDiscoveryDirectory()
	if err != nil {
		t.Errorf("Failed to load Google API Scopes: %s", err.Error())
	}

	// Load the saved scopes file to confirm it was saved correctly
	file, err := os.Open("google_endpoints.json")
	if err != nil {
		t.Errorf("Failed to open google_endpoints.json: %s", err.Error())
		return
	}
	defer file.Close()

	endpoints := &google.Endpoints{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&endpoints)
	if err != nil {
		t.Errorf("Failed to decode google_endpoints.json: %s", err.Error())
		return
	}

	if len(*endpoints) == 0 {
		t.Errorf("Failed to save Google API scopes: got %v, want non-empty", endpoints)
		return
	}
}

func TestSaveEndpoints(t *testing.T) {
	testData := &google.Endpoints{google.Endpoint{
		Title:       "test",
		Description: "data",
	}}

	err := google.SaveEndpoints(testData)
	if err != nil {
		t.Errorf("Failed to save endpoints: %v", err)
	}

	// Load the saved scopes file to confirm it was saved correctly
	e := &google.Endpoints{}

	// Get the absolute path of the currently running file
	executablePath, err := os.Executable()
	if err != nil {
		t.Errorf("%v", err)
	}
	executableDir := filepath.Dir(executablePath)

	var packageRoot string
	switch testing.Short() {
	case true:
		packageRoot, _ = filepath.Abs(filepath.Join(executableDir, "json"))
	case false:
		packageRoot, _ = filepath.Abs(filepath.Join(executableDir, "..", "..", "..", "google", "json"))
	}

	file, err := os.ReadFile(filepath.Join(packageRoot, "google_endpoints.json"))
	if err != nil {
		t.Errorf("Failed to open google_endpoints.json: %v", err)
		return
	}
	_ = json.Unmarshal([]byte(file), &e)
	if err != nil {
		t.Errorf("Failed to decode JSON from google_endpoints.json: %v", err)
	}

	if len(*e) == 0 {
		t.Errorf("Saved data does not match expected data: got %v, want %v", e, testData)
	}
}

func TestOrganizeScopes(t *testing.T) {
	err := google.OrganizeScopes()
	if err != nil {
		t.Errorf("Failed to organize scopes: %v", err)
	}
}

func TestDedupeScopes(t *testing.T) {
	scopes := []string{
		"https://www.googleapis.com/auth/admin.directory.customer",
		"https://www.googleapis.com/auth/admin.directory.customer",
		"https://www.googleapis.com/auth/admin.directory.device.chromeos",
		"https://www.googleapis.com/auth/admin.directory.device.chromeos",
		"https://www.googleapis.com/auth/admin.directory.device.mobile",
		"https://www.googleapis.com/auth/admin.directory.device.mobile",
		"https://www.googleapis.com/auth/admin.directory.domain",
		"https://www.googleapis.com/auth/admin.directory.domain",
	}
	scopes = google.DedupeScopes(scopes)
	if len(scopes) != 4 {
		t.Errorf("Failed to dedupe scopes: got %d, want 4", len(scopes))
	}
}

func TestLoadScopes(t *testing.T) {
	_, err := google.LoadScopes("Admin SDK API")
	if err != nil {
		t.Errorf("Failed to load scopes: %v", err)
	}
}
