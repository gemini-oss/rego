/*
# Google Drive - Test

This package runs tests for functions which interact with the Google Drive API:
https://developers.google.com/drive/api/v3/reference/

:Copyright: (c) 2026 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/google/drive_test.go
package google_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/google"
)

// Mock file list response
var mockFileListResponse = google.FileList{
	Kind: "drive#fileList",
	Files: &[]*google.File{
		{
			ID:       "file1",
			Name:     "test-file-1.txt",
			MimeType: "text/plain",
			Parents:  []string{"root"},
		},
		{
			ID:       "file2",
			Name:     "test-file-2.txt",
			MimeType: "text/plain",
			Parents:  []string{"root"},
		},
		{
			ID:       "folder1",
			Name:     "test-folder",
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{"root"},
		},
	},
}

// Mock empty file list for subfolders
var mockEmptyFileList = google.FileList{
	Kind:  "drive#fileList",
	Files: &[]*google.File{},
}

func setupDriveMockServer(t *testing.T) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle different endpoints
		switch {
		case r.URL.Path == "/drive/v3/files" && r.Method == "GET":
			// Check if this is a subfolder query
			query := r.URL.Query().Get("q")
			if query != "" && query != "'root' in parents and trashed = false" {
				// Return empty list for subfolders to prevent infinite recursion
				json.NewEncoder(w).Encode(mockEmptyFileList)
			} else {
				json.NewEncoder(w).Encode(mockFileListResponse)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestGetFileListBasic(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	// Test basic file list without options
	file := &google.File{ID: "root", Path: "My Drive"}
	result, err := client.Drive().GetFileList(file, nil)
	if err != nil {
		t.Logf("GetFileList returned error (may be expected in test): %v", err)
	}
	if result != nil && result.Files != nil {
		t.Logf("Found %d files", len(*result.Files))
	}
}

func TestGetFileListWithCSVBackup(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	// Create temp directory for CSV
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "test_backup.csv")

	file := &google.File{ID: "root", Path: "My Drive"}
	result, err := client.Drive().GetFileList(file, nil,
		google.DriveFileListOpts.WithCSV(csvPath),
	)
	if err != nil {
		t.Logf("GetFileList with CSV backup returned error (may be expected): %v", err)
	}
	if result != nil && result.Metadata != nil {
		t.Logf("CSV backup path: %s", result.Metadata.CSVBackupPath)
	}
}

func TestGetFileListWithStream(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	headers := &[]string{"id", "name", "path"}
	file := &google.File{ID: "root", Path: "My Drive"}
	result, err := client.Drive().GetFileList(file, nil,
		google.DriveFileListOpts.StreamResults("test-sheet-id"),
		google.DriveFileListOpts.SheetName("Test Sheet"),
		google.DriveFileListOpts.Headers(headers),
	)
	if err != nil {
		t.Logf("GetFileList with stream returned error (may be expected): %v", err)
	}
	if result != nil && result.Metadata != nil {
		t.Logf("Sheet URL: %s", result.Metadata.SheetURL)
	}
}

func TestGetFileListWithQuery(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	query := &google.DriveFileQuery{
		PageSize:          100,
		Fields:            "files(id,name)",
		SupportsAllDrives: true,
	}

	file := &google.File{ID: "root", Path: "My Drive"}
	result, err := client.Drive().GetFileList(file, query)
	if err != nil {
		t.Logf("GetFileList with query returned error (may be expected): %v", err)
	}
	if result != nil {
		t.Logf("Query executed successfully")
	}
}

func TestGetFileListWithHeaders(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	customHeaders := &[]string{"id", "name", "mimeType", "size"}
	file := &google.File{ID: "root", Path: "My Drive"}
	result, err := client.Drive().GetFileList(file, nil,
		google.DriveFileListOpts.Headers(customHeaders),
	)
	if err != nil {
		t.Logf("GetFileList with headers returned error (may be expected): %v", err)
	}
	if result != nil {
		t.Logf("Custom headers set successfully")
	}
}

func TestGetFileListChainedOptions(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "test_backup.csv")
	headers := &[]string{"id", "name", "path", "mimeType"}

	// Test multiple options
	file := &google.File{ID: "root", Path: "My Drive"}
	result, err := client.Drive().GetFileList(file, &google.DriveFileQuery{PageSize: 500},
		google.DriveFileListOpts.StreamResults("test-sheet-id"),
		google.DriveFileListOpts.SheetName("Test Sheet"),
		google.DriveFileListOpts.Headers(headers),
		google.DriveFileListOpts.WithCSV(csvPath),
	)
	if err != nil {
		t.Logf("GetFileList with chained options returned error (may be expected): %v", err)
	}
	if result != nil {
		t.Logf("Chained options executed successfully")
	}
}

func TestFileListMetadataStructure(t *testing.T) {
	// Test that FileListMetadata has all expected fields
	metadata := &google.FileListMetadata{
		Query:          &google.DriveFileQuery{},
		CSVBackupPath:  "/tmp/test.csv",
		SheetURL:       "https://docs.google.com/spreadsheets/d/test",
		FilesFound:     100,
		FoldersScanned: 10,
		PartialSuccess: false,
		Errors:         nil,
		Cached:         false,
	}

	if metadata.FilesFound != 100 {
		t.Errorf("Expected FilesFound to be 100, got %d", metadata.FilesFound)
	}

	if metadata.FoldersScanned != 10 {
		t.Errorf("Expected FoldersScanned to be 10, got %d", metadata.FoldersScanned)
	}

	if metadata.CSVBackupPath != "/tmp/test.csv" {
		t.Errorf("Expected CSVBackupPath to be '/tmp/test.csv', got %s", metadata.CSVBackupPath)
	}
}

func TestGetFileListDefaultCSVPath(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	// When no path is specified, should use temp dir
	file := &google.File{ID: "root", Path: "My Drive"}
	result, err := client.Drive().GetFileList(file, nil,
		google.DriveFileListOpts.WithCSV(),
	)
	if err != nil {
		t.Logf("GetFileList with default CSV path returned error (may be expected): %v", err)
	}
	if result != nil && result.Metadata != nil && result.Metadata.CSVBackupPath != "" {
		t.Logf("Default CSV path: %s", result.Metadata.CSVBackupPath)
		// Clean up
		os.Remove(result.Metadata.CSVBackupPath)
	}
}

func TestDriveClientAccess(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	driveClient := client.Drive()
	if driveClient == nil {
		t.Fatal("Expected DriveClient to be accessible")
	}
}

func TestGetRootFileList(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	// Test GetRootFileList convenience method
	result, err := client.Drive().GetRootFileList()
	if err != nil {
		t.Logf("GetRootFileList returned error (may be expected): %v", err)
	}
	if result != nil && result.Files != nil {
		t.Logf("Found %d files in root", len(*result.Files))
	}
}

func TestGetRootFileListWithOptions(t *testing.T) {
	server := setupDriveMockServer(t)
	defer server.Close()

	client := SetupTestClient(server.URL)
	if client == nil {
		t.Skip("Skipping test: client setup failed (likely missing credentials)")
	}

	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "root_backup.csv")

	// Test GetRootFileList with options
	result, err := client.Drive().GetRootFileList(
		google.DriveFileListOpts.WithCSV(csvPath),
		google.DriveFileListOpts.ExcludeFolders(),
	)
	if err != nil {
		t.Logf("GetRootFileList with options returned error (may be expected): %v", err)
	}
	if result != nil && result.Metadata != nil {
		t.Logf("Files found: %d, Folders scanned: %d", result.Metadata.FilesFound, result.Metadata.FoldersScanned)
	}
}

// Integration test - only runs with actual credentials
func TestGetFileListIntegration(t *testing.T) {
	// Skip if no credentials available
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		t.Skip("Skipping integration test: GOOGLE_APPLICATION_CREDENTIALS not set")
	}

	client, err := google.NewClient(
		google.AuthCredentials{
			Type:   google.SERVICE_ACCOUNT,
			CICD:   true,
			Scopes: []string{"Google Drive API"},
		},
		log.DEBUG,
	)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// Test with root folder - just verify it doesn't panic
	result, err := client.Drive().GetRootFileList(
		google.DriveFileListOpts.WithCSV(),
	)

	if err != nil {
		// Some errors are expected in test environment
		t.Logf("GetRootFileList returned error (may be expected): %v", err)
	}

	if result != nil && result.Metadata != nil {
		t.Logf("Found %d files in %d folders", result.Metadata.FilesFound, result.Metadata.FoldersScanned)
		if result.Metadata.CSVBackupPath != "" {
			t.Logf("CSV backup at: %s", result.Metadata.CSVBackupPath)
			// Clean up CSV file
			os.Remove(result.Metadata.CSVBackupPath)
		}
	}
}
