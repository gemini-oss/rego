// pkg/internal/tests/common/requests/download_test.go
package requests_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	// Prepare a temporary directory for downloads
	tmpDir = filepath.Join(os.TempDir(), "/rego_tests")
)

func TestDownloadFile(t *testing.T) {
	// Artificially large content to simulate a 5GB file
	content := strings.Repeat("3D", 2.5*1024*1024*1024) // 5GB of '3D'

	// Set up mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Length", strconv.Itoa(len(content)))
		if r.Method == "HEAD" {
			w.WriteHeader(http.StatusOK)
		} else if r.Method == "GET" {
			rangeHeader := r.Header.Get("Range")
			if rangeHeader == "" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(content))
			} else {
				// Properly handle range requests
				parts := strings.SplitN(rangeHeader[6:], "-", 2)
				start, _ := strconv.Atoi(parts[0])
				var end int
				if parts[1] == "" {
					end = len(content) - 1
				} else {
					end, _ = strconv.Atoi(parts[1])
				}
				if end >= len(content) {
					end = len(content) - 1
				}
				w.Header().Set("Content-Length", strconv.Itoa(end-start+1))
				w.WriteHeader(http.StatusPartialContent)
				w.Write([]byte(content[start : end+1]))
			}
		}
	}))
	defer server.Close()

	tests := []struct {
		name             string
		client           *http.Client
		url              string
		directory        string
		filename         string
		allowDuplicates  bool
		expectedContent  string
		expectedFileName string
		expectedError    error
	}{
		{
			name:             "Successful Download",
			client:           http.DefaultClient,
			url:              server.URL,
			directory:        tmpDir,
			filename:         "rego_testfile.txt",
			allowDuplicates:  false,
			expectedContent:  content[:1024], // Check the first 1KB for validation
			expectedFileName: "rego_testfile.txt",
			expectedError:    nil,
		},
		{
			name:             "Resume Download",
			client:           http.DefaultClient,
			url:              server.URL,
			directory:        tmpDir,
			filename:         "rego_testfile.txt",
			allowDuplicates:  false,
			expectedContent:  content[:1024],
			expectedFileName: "rego_testfile.txt",
			expectedError:    nil,
		},
		{
			name:             "Handle Duplicate Downloads",
			client:           http.DefaultClient,
			url:              server.URL,
			directory:        tmpDir,
			filename:         "rego_testfile.txt",
			allowDuplicates:  true,
			expectedContent:  content[:1024],
			expectedFileName: "rego_testfile (1).txt",
			expectedError:    nil,
		},
		{
			name:             "Error Handling - Network Error",
			client:           mockHTTPClient("", 0, errors.New("network error")),
			url:              "http://gemini.com/fail",
			directory:        tmpDir,
			filename:         "rego_network_fail.txt",
			allowDuplicates:  false,
			expectedContent:  "",
			expectedFileName: "rego_network_fail.txt",
			expectedError:    errors.New("network error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "Resume Download" {
				resumePoints := []float32{0.25, 0.5, 0.75}
				for _, point := range resumePoints {
					// Simulate partial download at 25%, 50%, and 75%
					partialSize := int64(float32(len(content)) * point)
					truncateDownload(filepath.Join(tt.directory, tt.filename), partialSize)

					// Now resume download
					err := requests.NewClient(tt.client, nil, nil).DownloadFile(tt.url, tt.directory, tt.filename, tt.allowDuplicates)
					if err != nil {
						t.Errorf("Failed to resume download at %.0f%%: %v", point*100, err)
						continue
					}

					// Verify the resumed download by checking file content length
					downloadedFilePath := filepath.Join(tt.directory, tt.filename)
					data, err := os.ReadFile(downloadedFilePath)
					if err != nil {
						t.Errorf("Failed to read downloaded file at %.0f%%: %v", point*100, err)
						continue
					}
					if len(data) != len(content) {
						t.Errorf("Downloaded content mismatch at %f%%. Expected %d, got %d", point*100, len(content), len(data))
					}
				}
				return
			}

			client := requests.NewClient(tt.client, nil, nil)
			err := client.DownloadFile(tt.url, tt.directory, tt.filename, tt.allowDuplicates)

			if tt.expectedError != nil {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify file content
			downloadedFilePath := filepath.Join(tt.directory, tt.expectedFileName)
			data, err := os.ReadFile(downloadedFilePath)
			if err != nil {
				t.Errorf("Failed to read downloaded file: %v", err)
				return
			}

			if string(data[:1024]) != tt.expectedContent {
				t.Errorf("Downloaded content mismatch. Expected first 1KB to be %s, got %s", tt.expectedContent, string(data[:1024]))
			}
		})
	}
	os.RemoveAll(tmpDir)
}

func TestDownloadRealFile(t *testing.T) {
	filename := "GitHubDesktop.zip"
	url := "https://central.github.com/deployments/desktop/desktop/latest/darwin"

	client := requests.NewClient(http.DefaultClient, nil, nil)
	filePath := filepath.Join(tmpDir, filename)

	// Cleanup before test runs
	os.Remove(filePath)

	// Perform the initial download
	err := client.DownloadFile(url, tmpDir, filename, false)
	if err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}

	// Check if the file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Errorf("The file was not downloaded: %s", filePath)
	} else if err != nil {
		t.Fatalf("Failed to stat downloaded file: %v", err)
	}

	// Attempt to download again to test duplicate handling
	err = client.DownloadFile(url, tmpDir, filename, false)
	if err == nil {
		t.Logf("Duplicate download should not have occurred")
	}

	// Simulate a partial download (e.g., first half of the file)
	halfSize := fileInfo.Size() / 2
	truncateDownload(filePath, halfSize)

	// Try resuming the download
	err = client.DownloadFile(url, tmpDir, filename, false)
	if err != nil {
		t.Errorf("Failed to resume download: %v", err)
	}

	// Verify the file has been fully downloaded after resume
	completedFileInfo, _ := os.Stat(filePath)
	if completedFileInfo.Size() != fileInfo.Size() {
		t.Errorf("File size after resume does not match expected: got %v, want %v", completedFileInfo.Size(), fileInfo.Size())
	}

	// Clean up after test
	os.Remove(filePath)
}

func TestDownloadRealFileNoFilename(t *testing.T) {
	tmpDir := os.TempDir()
	url := "https://codeload.github.com/gemini-oss/rego/zip/refs/heads/main"

	client := requests.NewClient(http.DefaultClient, nil, nil)

	// Perform the download without specifying a filename
	err := client.DownloadFile(url, tmpDir, "", false)
	if err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}

	// Expected filename that should have been determined from the URL or headers
	expectedFilename := "rego-main.zip"
	filePath := filepath.Join(tmpDir, expectedFilename)

	// Check if the file with the expected filename exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("The file was not downloaded with the expected filename: %s", expectedFilename)
	} else if err != nil {
		t.Fatalf("Failed to stat downloaded file: %v", err)
	}

	// Clean up after test
	os.Remove(filePath)
}

// truncateDownload truncates a file to simulate a partial download.
func truncateDownload(filePath string, size int64) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic("failed to open file for simulation: " + err.Error())
	}
	defer file.Close()

	if err := file.Truncate(size); err != nil {
		panic("failed to truncate file for simulation: " + err.Error())
	}
}
