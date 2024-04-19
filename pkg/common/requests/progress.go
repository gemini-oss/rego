// pkg/common/requests/progress.go
package requests

import (
	"container/list"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

const (
	progressFormat = "[%s] %3d%% |%-10s| %s (%s/%s, %s) [%s]"
	progressLog    = "|%3d%%| %s (%s, %s) [%s]"
)

var (
	mu                 sync.Mutex
	lineNumbers            = list.New()         // Queue to hold reusable line numbers
	usedLines              = make(map[int]bool) // Set to track used line numbers
	maxLineUsed        int = -1                 // Tracks the highest line number used
	clearLine              = "\033[2K"          // Clear the entire line
	clearLineRemainder     = "\033[K"           // Clear the remainder of the line
)

func getLineNumber() int {
	mu.Lock()
	defer mu.Unlock()

	// Check for reusable line numbers first
	if lineNumbers.Front() != nil {
		line := lineNumbers.Remove(lineNumbers.Front()).(int)
		usedLines[line] = true
		return line
	}

	// Allocate a new line number if no reusables
	maxLineUsed++
	usedLines[maxLineUsed] = true
	return maxLineUsed
}

func releaseLineNumber(line int) {
	mu.Lock()
	defer mu.Unlock()

	if usedLines[line] {
		delete(usedLines, line)
		lineNumbers.PushBack(line) // Add back to reusable queue
	}
}

type progress struct {
	io.Reader    // Embedded reader
	totalBytes   int64
	currentBytes int64
	lastUpdate   int64
	startTime    time.Time
	progressCh   chan progressData
	lineNum      int // Terninal line number for progress bar
	Enabled      bool
}

type progressData struct {
	percentComplete int   // Progress percentage
	bytesRead       int64 // Bytes read since the last update
}

func (pr *progress) Read(b []byte) (int, error) {
	n, err := pr.Reader.Read(b)
	pr.currentBytes += int64(n)

	// Update progress every 1MB or on completion/error
	if pr.currentBytes-pr.lastUpdate > 1024*1000 || err != nil {
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
		mu.Lock()

		// Ensure the cursor is moved to the correct line
		moveToLine(pr.lineNum)

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
		completed := int(10 * data.percentComplete / 100)
		bar := strings.Repeat("█", completed) + strings.Repeat(" ", 10-completed)

		// Format elapsed time.
		elapsedTime := formatDuration(time.Duration(elapsed) * time.Second)

		// Print or Log the progress bar.
		if data.percentComplete == 100 {
			fmt.Print(clearLine)
			progressLog := fmt.Sprintf(progressLog, data.percentComplete, filename, totalBytes, speedStr, elapsedTime)
			l.Println(progressLog)
			resetCursor(pr.lineNum)
			mu.Unlock()
			time.Sleep(1 * time.Second)
			continue
		} else {
			progressLog := fmt.Sprintf(progressFormat, time.Now().Format("2006/01/02 03:04:05 PM"), data.percentComplete, bar, filename, currentBytes, totalBytes, speedStr, elapsedTime)
			fmt.Println(progressLog + clearLineRemainder)
		}

		// Reset cursor position
		resetCursor(pr.lineNum)
		mu.Unlock()
	}
}

// Move the cursor to the specified line number
func moveToLine(lineNum int) {
	fmt.Printf("\033[%dA", lineNum+1)
}

// Reset the cursor to the beginning of the line
func resetCursor(lineNum int) {
	fmt.Printf("\033[%dB", lineNum)
}
