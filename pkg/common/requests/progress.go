// pkg/common/requests/progress.go
package requests

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	progressFormat = "\r[%s] %3d%% |%-25s| %s (%s/%s, %s) [%s]"
)

type progress struct {
	io.Reader
	totalBytes   int64
	currentBytes int64
	lastUpdate   int64
	startTime    time.Time
	progressCh   chan progressData
}

type progressData struct {
	percentComplete int   // Progress percentage
	bytesRead       int64 // Bytes read since the last update
}

func (pr *progress) Read(b []byte) (int, error) {
	n, err := pr.Reader.Read(b)
	pr.currentBytes += int64(n)

	// Update progress every 10KB or on completion/error
	if pr.currentBytes-pr.lastUpdate > 1024*10 || err != nil {
		update := progressData{
			percentComplete: int(100 * pr.currentBytes / pr.totalBytes),
			bytesRead:       pr.currentBytes,
		}
		pr.progressCh <- update
		pr.lastUpdate = pr.currentBytes
	}

	return n, err
}

// Converts bytes to a human-readable format.
func byteHuman(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Formats duration to a more human-readable format.
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	return fmt.Sprintf("%02dm %02ds", m, s)
}

// trackProgress tracks the progress of a download.
func (pr *progress) trackProgress(filename string) {
	for data := range pr.progressCh {
		// Calculate the total elapsed time.
		elapsed := time.Since(pr.startTime).Seconds()

		// Calculate the average download speed based on the total bytes read so far and the elapsed time.
		speed := float64(pr.currentBytes) / elapsed

		// Use byteHuman to convert current and total bytes to a human-readable format.
		currentBytes := byteHuman(pr.currentBytes)
		totalBytes := byteHuman(pr.totalBytes)

		// Convert download speed to a human-readable format.
		speedStr := byteHuman(int64(speed)) + "/s"

		// Calculate the number of '█' symbols for the progress bar.
		completed := int(25 * data.percentComplete / 100)
		bar := strings.Repeat("█", completed) + strings.Repeat(" ", 25-completed)

		// Format elapsed time.
		elapsedTime := formatDuration(time.Duration(elapsed) * time.Second)

		// Print the formatted progress bar.
		progressLog := fmt.Sprintf(progressFormat, time.Now().Format("2006/01/02 03:04:05 PM"), data.percentComplete, bar, filename, currentBytes, totalBytes, speedStr, elapsedTime)
		fmt.Print(progressLog)
	}
	fmt.Println()
}