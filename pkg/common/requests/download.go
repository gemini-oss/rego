// pkg/common/requests/download.go
package requests

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func (c *Client) DownloadFile(url, filepath, filename string) error {

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status: %d", resp.StatusCode)
	}

	// Set filepath to current directory if not provided
	if filepath == "" {
		filepath = "."
	}

	// Extract filename from headers if not provided
	if filename == "" {
		contentDisposition := resp.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil && params["filename"] != "" {
			filename = params["filename"]
		} else {
			// Check if a 'referer' header is present, use its URL to extract the filename
			if referer := resp.Request.Header.Get("Referer"); referer != "" {
				url = resp.Request.URL.String()
			}

			// Extract filename from the URL
			re := regexp.MustCompile(`/([^/]+\.(?:[a-zA-Z0-9]+))(\?.*ext=([^&]+))?(\#|$)`)
			matches := re.FindStringSubmatch(url)
			if len(matches) > 1 {
				filename = matches[1]
			} else {
				// Fallback default filename
				urlSegment := strings.Split(url, "/")[2]
				baseFilename := fmt.Sprintf("%s_%s", urlSegment, time.Now().Format("20060102_150405"))
				filename = baseFilename + ".rego_download.txt"
			}
		}
	}

	// File creation
	out, err := os.Create(fmt.Sprintf("%s/%s", filepath, filename))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	// Set up progress tracking
	progressCh := make(chan progressData)
	pr := &progress{
		Reader:       resp.Body,
		totalBytes:   resp.ContentLength,
		currentBytes: 0,
		lastUpdate:   0,
		startTime:    time.Now(),
		progressCh:   progressCh,
	}
	go pr.trackProgress(filename)

	_, err = io.Copy(out, pr)
	close(progressCh)
	if err != nil {
		return fmt.Errorf("error writing response to file: %w", err)
	}

	return nil
}
