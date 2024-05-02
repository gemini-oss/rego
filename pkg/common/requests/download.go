// pkg/common/requests/download.go
package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DownloadMetadata stores state for managing downloads.
type DownloadMetadata struct {
	URL           string    // Source URL
	FilePath      string    // Full path to the downloaded file
	FileName      string    // Name of the file
	BytesReceived int64     // Bytes downloaded so far
	TotalSize     int64     // Total size of the file
	LastModified  time.Time // Last modified time from server
	Checksum      string    // For checksum validation (optional)
}

func (c *Client) DownloadFile(url, directory, filename string, allowDuplicates bool) error {

	cacheKey := "download_meta_" + filename

	var metadata *DownloadMetadata
	cached, found := c.Cache.Get(cacheKey)
	if found {
		err := json.Unmarshal(cached, metadata)
		if err != nil {
			return fmt.Errorf("error unmarshalling cached metadata: %w", err)
		}
	}

	metadata = &DownloadMetadata{
		URL:           url,
		FilePath:      directory,
		FileName:      filename,
		BytesReceived: 0,
		TotalSize:     -1,
	}

	// Use a HEAD request to fetch headers for filename extraction
	// https://developer.mozilla.org/en-US/docs/web/http/methods/head
	req, _ := c.CreateRequest("HEAD", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing HEAD request: %w", err)
	}
	defer resp.Body.Close()

	// Extract filename if not provided
	if filename == "" {
		extractFilename(url, filename, resp, metadata)
	}

	// Ensure the directory exists
	if directory == "" {
		directory = "rego_downloads"
	}
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, os.ModePerm); err != nil { // os.ModePerm
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	// Find the correct download if duplicates exist
	completeFilePath, bytesReceived, err := findLatestDownload(directory, metadata.FileName)
	if err != nil {
		return err
	}
	metadata.BytesReceived = bytesReceived
	metadata.FileName = filepath.Base(completeFilePath)

	fileInfo, err := os.Stat(completeFilePath)
	if err == nil {
		// Check if the file is already completely downloaded.
		if allowDuplicates && fileInfo.Size()-resp.ContentLength == 0 {
			completeFilePath = generateNewFilepath(directory, metadata.FileName)
			cacheKey = "download_meta_" + filepath.Base(completeFilePath) // Update cache key for new file
			metadata.FileName = filepath.Base(completeFilePath)
			metadata.BytesReceived = 0 // Reset since we will consider this a new download
		}
		if !allowDuplicates && fileInfo.Size() == resp.ContentLength {
			c.Log.Printf("Duplicates disabled. File already downloaded: %s\n", metadata.FileName)
			return nil
		}
	}

	// Start or resume the download
	req, err = c.CreateRequest("GET", url)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set the Range header if part of the file already exists
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Range
	if metadata.BytesReceived > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", bytesReceived))
	}

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	// File creation/resumption
	var out *os.File
	if resp.StatusCode == http.StatusPartialContent && metadata.BytesReceived > 0 {
		out, err = os.OpenFile(completeFilePath, os.O_APPEND|os.O_WRONLY, 0600)
		metadata.TotalSize = resp.ContentLength + metadata.BytesReceived
	} else {
		out, err = os.Create(completeFilePath)
		metadata.TotalSize = resp.ContentLength
	}
	if err != nil {
		return err
	}
	defer out.Close()

	// Set up progress tracking
	lineNum := getLineNumber()       // Get a unique line number for this download
	defer releaseLineNumber(lineNum) // Ensure the line number is released after use

	progressCh := make(chan progressData)
	pr := &progress{
		Reader:       resp.Body,
		totalBytes:   metadata.TotalSize,
		currentBytes: metadata.BytesReceived,
		lastUpdate:   0,
		startTime:    time.Now(),
		progressCh:   progressCh,
		lineNum:      lineNum,
	}
	go pr.trackProgress(metadata.FileName)

	_, err = io.Copy(out, pr)
	close(progressCh)
	if err != nil {
		return fmt.Errorf("error writing response to file: %w", err)
	}

	return nil
}

func findLatestDownload(directory, filename string) (string, int64, error) {
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	extension := filepath.Ext(filename)
	pattern := fmt.Sprintf("%s*%s", baseName, extension) // Match any file starting with baseName and ending with extension
	files, err := filepath.Glob(filepath.Join(directory, pattern))
	if err != nil {
		return "", 0, err
	}

	var latestFile string
	var maxIndex = -1                                                                     // Start with -1 to include original file without suffix as a valid option
	regex := regexp.MustCompile(fmt.Sprintf(`\((\d+)\)%s$`, regexp.QuoteMeta(extension))) // Correct regex to match numbers within parentheses just before the extension
	for _, file := range files {
		if file == filepath.Join(directory, filename) && maxIndex == -1 {
			// It's the original file without any suffix
			latestFile = file
			continue
		}
		matches := regex.FindStringSubmatch(file)
		if len(matches) > 1 {
			index, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			if index > maxIndex {
				maxIndex = index
				latestFile = file
			}
		}
	}

	if latestFile == "" {
		latestFile = filepath.Join(directory, filename) // Default to the original filename if no variants found
	}
	fileInfo, err := os.Stat(latestFile)
	if err != nil {
		return latestFile, 0, nil // File does not exist, no bytes received
	}
	return latestFile, fileInfo.Size(), nil
}

func generateNewFilepath(directory, originalFilename string) string {
	baseName := strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename))
	extension := filepath.Ext(originalFilename)
	for i := 1; ; i++ {
		newFilename := fmt.Sprintf("%s (%d)%s", baseName, i, extension)
		newPath := filepath.Join(directory, newFilename)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

func extractFilename(url, filename string, resp *http.Response, metadata *DownloadMetadata) {
	// Extract filename from headers if not provided
	if filename == "" {
		contentDisposition := resp.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil && params["filename"] != "" {
			metadata.FileName = params["filename"]
		} else {
			// Check if a 'referer' header is present, use its URL to extract the filename
			if referer := resp.Request.Header.Get("Referer"); referer != "" {
				url = resp.Request.URL.String()
			}

			// Extract filename from the URL
			re := regexp.MustCompile(`/([^/]+\.(?:[a-zA-Z0-9]+))(\?.*ext=([^&]+))?(\#|$)`)
			matches := re.FindStringSubmatch(url)
			if len(matches) > 1 {
				metadata.FileName = matches[1]
			} else {
				// Fallback default filename
				urlSegment := strings.Split(url, "/")[2]
				baseFilename := fmt.Sprintf("%s_%s", urlSegment, time.Now().Format("20060102_150405"))
				metadata.FileName = baseFilename + ".rego_download.txt"
			}
		}
	}
}
