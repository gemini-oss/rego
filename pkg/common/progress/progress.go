/*
# Common - Progress

This package provides progress tracking utilities for long-running operations.
It supports both determinate (known total) and indeterminate (unknown total) progress.

:Copyright: (c) 2026 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/common/progress/progress.go
package progress

import (
	"container/list"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Terminal ANSI escape codes
const (
	clearLine          = "\033[2K"   // Clear the entire line
	clearLineRemainder = "\033[K"    // Clear the remainder of the line
	clearToEndOfScreen = "\033[J"    // Clear from cursor to end of screen
	clearScreen        = "\033[2J"   // Clear entire screen
	hideCursor         = "\033[?25l" // Hide cursor
	showCursor         = "\033[?25h" // Show cursor
	saveCursor         = "\033[s"    // Save cursor position
	restoreCursor      = "\033[u"    // Restore cursor position
	// Alternate screen buffer - gives tracker its own screen without affecting scrollback
	enterAltScreen = "\033[?1049h" // Enter alternate screen buffer
	exitAltScreen  = "\033[?1049l" // Exit alternate screen buffer (returns to main)
	moveCursorHome = "\033[H"      // Move cursor to top-left (1,1)
	resetScrollRgn = "\033[r"      // Reset scroll region to full screen
)

// setScrollRegion sets the scroll region to lines top through bottom (1-indexed)
func setScrollRegion(top, bottom int) string {
	return fmt.Sprintf("\033[%d;%dr", top, bottom)
}

// moveCursorTo moves cursor to specified line and column (1-indexed)
func moveCursorTo(line, col int) string {
	return fmt.Sprintf("\033[%d;%dH", line, col)
}

// getTerminalSize returns the current terminal dimensions (columns, rows)
// Returns default values (120, 50) if unable to determine size
func getTerminalSize() (cols, rows int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 120, 50 // Default fallback
	}
	return int(ws.Col), int(ws.Row)
}

// Progress separator line
const progressSeparator = "════════════════════════════════════════════════════════════════════════════════"

// Spinner frames for indeterminate progress
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Terminal line management (shared across all progress trackers)
var (
	lineMu        sync.Mutex
	lineNumbers   = list.New()         // Queue to hold reusable line numbers
	usedLines     = make(map[int]bool) // Set to track used line numbers
	maxLineUsed   = -1                 // Tracks the highest line number used
	numberPrinter = message.NewPrinter(language.English)
)

// getLineNumber allocates a unique terminal line number for progress display
func getLineNumber() int {
	lineMu.Lock()
	defer lineMu.Unlock()

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

// releaseLineNumber returns a line number to the pool for reuse
func releaseLineNumber(line int) {
	lineMu.Lock()
	defer lineMu.Unlock()

	if usedLines[line] {
		delete(usedLines, line)
		lineNumbers.PushBack(line) // Add back to reusable queue
	}
}

// moveToLine moves the cursor to the specified line number
func moveToLine(lineNum int) {
	if lineNum >= 0 {
		fmt.Printf("\033[%dA", lineNum+1)
	}
}

// resetCursor resets the cursor position after updating a line
func resetCursor(lineNum int) {
	if lineNum >= 0 {
		fmt.Printf("\033[%dB", lineNum)
	}
}

// formatDuration formats a duration to a human-readable format (XXm XXs)
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	s := (d % time.Minute) / time.Second
	if h > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", h, m, s)
	}
	return fmt.Sprintf("%02dm %02ds", m, s)
}

// formatNumber formats a number with commas (e.g., 1234567 -> "1,234,567")
func formatNumber(n int64) string {
	return numberPrinter.Sprintf("%d", n)
}

// truncatePath truncates a path to fit within maxLen, keeping the end visible
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return "..."
	}
	return "..." + path[len(path)-(maxLen-3):]
}

// Update represents a progress update message
type Update struct {
	ItemsFound  int64  // Count of items processed so far
	CurrentPath string // Current item being processed (e.g., folder path)
}

// indeterminateHeaderLines is the fixed number of lines reserved for the indeterminate tracker header
const indeterminateHeaderLines = 5 // separator + stats + current path + separator + buffer

// IndeterminateTracker displays progress for operations with unknown totals
// It shows a spinner, item count, current path, and elapsed time
type IndeterminateTracker struct {
	label        string         // Operation label (e.g., "Scanning Drive")
	itemsFound   int64          // Count of items processed
	currentPath  string         // Current item being processed
	startTime    time.Time      // When tracking started
	updateCh     chan Update    // Channel for receiving updates
	lineNum      int            // Terminal line number for display
	spinnerIndex int            // Current spinner frame index
	mu           sync.Mutex     // Protects internal state
	done         bool           // Whether tracking is complete
	ticker       *time.Ticker   // For spinner animation
	stopCh       chan struct{}  // Signal to stop the tracker
	mode         TrackerMode    // Rendering mode (alt screen vs scroll region)
	terminalRows int            // Terminal height for scroll region mode
	terminalCols int            // Terminal width for dynamic sizing
	resizeCh     chan os.Signal // Channel for terminal resize signals
}

// NewIndeterminate creates a new indeterminate progress tracker with default mode (simple line)
// label: The operation label to display (e.g., "Scanning Drive")
func NewIndeterminate(label string) *IndeterminateTracker {
	return NewIndeterminateWithMode(label, ModeScrollRegion) // Default to scroll region for better log visibility
}

// NewIndeterminateWithMode creates a new indeterminate progress tracker with specified mode
// label: The operation label to display
// mode: ModeAltScreen for separate screen, ModeScrollRegion for sticky header with logs below
func NewIndeterminateWithMode(label string, mode TrackerMode) *IndeterminateTracker {
	cols, rows := getTerminalSize()
	return &IndeterminateTracker{
		label:        label,
		updateCh:     make(chan Update, 100), // Buffered to prevent blocking
		stopCh:       make(chan struct{}),
		lineNum:      -1, // Will be allocated on Start()
		mode:         mode,
		terminalRows: rows,
		terminalCols: cols,
		resizeCh:     make(chan os.Signal, 1),
	}
}

// Start begins displaying progress. Call this in a goroutine.
// Example: go tracker.Start()
func (t *IndeterminateTracker) Start() {
	t.mu.Lock()
	t.startTime = time.Now()
	t.ticker = time.NewTicker(100 * time.Millisecond) // Update spinner every 100ms
	mode := t.mode
	t.mu.Unlock()

	// Register for terminal resize signals
	signal.Notify(t.resizeCh, syscall.SIGWINCH)
	defer signal.Stop(t.resizeCh)

	// Setup terminal based on mode
	t.setupTerminal(mode)

	for {
		select {
		case <-t.stopCh:
			t.ticker.Stop()
			return
		case update, ok := <-t.updateCh:
			if !ok {
				t.ticker.Stop()
				return
			}
			t.mu.Lock()
			t.itemsFound = update.ItemsFound
			t.currentPath = update.CurrentPath
			t.mu.Unlock()
			t.render()
		case <-t.resizeCh:
			t.handleResize()
		case <-t.ticker.C:
			t.render()
		}
	}
}

// setupTerminal configures the terminal for the current mode
func (t *IndeterminateTracker) setupTerminal(mode TrackerMode) {
	switch mode {
	case ModeAltScreen:
		fmt.Print(enterAltScreen)
		fmt.Print(hideCursor)
	case ModeScrollRegion:
		fmt.Print(hideCursor)
		fmt.Print(clearScreen)
		fmt.Print(moveCursorHome)
		t.mu.Lock()
		fmt.Print(setScrollRegion(indeterminateHeaderLines+1, t.terminalRows))
		t.mu.Unlock()
		fmt.Print(moveCursorTo(indeterminateHeaderLines+1, 1))
	}
}

// handleResize updates terminal dimensions and reconfigures the display
func (t *IndeterminateTracker) handleResize() {
	cols, rows := getTerminalSize()

	t.mu.Lock()
	t.terminalCols = cols
	t.terminalRows = rows
	mode := t.mode
	t.mu.Unlock()

	if mode == ModeScrollRegion {
		lineMu.Lock()
		fmt.Print(saveCursor)
		fmt.Print(setScrollRegion(indeterminateHeaderLines+1, rows))
		fmt.Print(restoreCursor)
		lineMu.Unlock()
	}
}

// Update sends a progress update to the tracker
func (t *IndeterminateTracker) Update(itemsFound int64, currentPath string) {
	select {
	case t.updateCh <- Update{ItemsFound: itemsFound, CurrentPath: currentPath}:
	default:
		// Channel full, skip this update (non-blocking)
	}
}

// Done stops the tracker and displays a completion message
func (t *IndeterminateTracker) Done() {
	t.mu.Lock()
	if t.done {
		t.mu.Unlock()
		return
	}
	t.done = true
	itemsFound := t.itemsFound
	elapsed := time.Since(t.startTime)
	label := t.label
	mode := t.mode
	t.mu.Unlock()

	// Signal stop and close channels
	close(t.stopCh)
	close(t.updateCh)

	lineMu.Lock()

	// Cleanup based on mode
	switch mode {
	case ModeAltScreen:
		fmt.Print(showCursor)
		fmt.Print(exitAltScreen)
	case ModeScrollRegion:
		fmt.Print(resetScrollRgn)
		fmt.Print(moveCursorTo(indeterminateHeaderLines+1, 1))
		fmt.Print(clearToEndOfScreen)
		fmt.Print(showCursor)
	}

	// Print final completion message
	timestamp := time.Now().Format("2006/01/02 03:04:05 PM")
	finalMsg := fmt.Sprintf("[%s] \033[32m✓\033[0m %s | %s files | Complete [%s]",
		timestamp,
		label,
		formatNumber(itemsFound),
		formatDuration(elapsed),
	)
	fmt.Println(finalMsg)
	lineMu.Unlock()
}

// render updates the progress display using double-buffered rendering
func (t *IndeterminateTracker) render() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.done {
		return
	}

	// Advance spinner
	t.spinnerIndex = (t.spinnerIndex + 1) % len(spinnerFrames)
	spinner := spinnerFrames[t.spinnerIndex]

	elapsed := time.Since(t.startTime)
	timestamp := time.Now().Format("2006/01/02 03:04:05 PM")

	// Use dynamic terminal width
	termWidth := t.terminalCols
	if termWidth < 80 {
		termWidth = 80
	}

	// Truncate path to fit terminal
	displayPath := truncatePath(t.currentPath, termWidth-60)
	if displayPath == "" {
		displayPath = "initializing..."
	}

	// Double-buffered rendering: build entire frame in memory first
	var buf strings.Builder
	buf.Grow(1024)

	// For scroll region mode: save cursor position before moving to header
	if t.mode == ModeScrollRegion {
		buf.WriteString(saveCursor)
	}

	// Move cursor to home position
	buf.WriteString(moveCursorHome)

	// Helper to write a padded line
	writePaddedLine := func(content string) {
		buf.WriteString(content)
		buf.WriteString(clearLineRemainder)
		buf.WriteString("\n")
	}

	// Header separator
	writePaddedLine(progressSeparator)

	// Main stats line
	statsLine := fmt.Sprintf("[%s] %s %s | %s files [%s]",
		timestamp,
		spinner,
		t.label,
		formatNumber(t.itemsFound),
		formatDuration(elapsed),
	)
	writePaddedLine(statsLine)

	// Current path line
	pathLine := fmt.Sprintf("  \033[36m→\033[0m %s", displayPath)
	writePaddedLine(pathLine)

	// Footer separator
	writePaddedLine(progressSeparator)

	// For scroll region mode: restore cursor to continue logging in scroll area
	if t.mode == ModeScrollRegion {
		buf.WriteString(restoreCursor)
	}

	// Single atomic write to terminal
	lineMu.Lock()
	fmt.Print(buf.String())
	lineMu.Unlock()
}

// ItemsFound returns the current count of items found (thread-safe)
func (t *IndeterminateTracker) ItemsFound() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.itemsFound
}

// Elapsed returns the elapsed time since tracking started
func (t *IndeterminateTracker) Elapsed() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return time.Since(t.startTime)
}

// HierarchicalUpdate represents a progress update for hierarchical operations
type HierarchicalUpdate struct {
	ItemsFound       int64  // Total items found across all levels
	FoldersScanned   int64  // Total folders scanned
	CurrentPath      string // Current item being processed
	TopLevelComplete int    // Number of top-level items completed
}

// ActiveFolderUpdate represents an update to the active folders list
type ActiveFolderUpdate struct {
	Path   string // Folder path
	Action string // "add" or "remove"
}

// maxActiveFolders is the maximum number of active folders to display
const maxActiveFolders = 8

// TrackerMode defines how the progress tracker renders
type TrackerMode int

const (
	// ModeAltScreen uses alternate screen buffer (tracker gets its own screen)
	ModeAltScreen TrackerMode = iota
	// ModeScrollRegion uses scroll regions (tracker at top, logs scroll below)
	ModeScrollRegion
)

// trackerHeaderLines is the fixed number of lines reserved for the tracker header
// Includes: separator + main stats + progress bar + current path + active workers header + max active folders + separator + buffer
const trackerHeaderLines = 3 + maxActiveFolders + 4 // ~15 lines

// HierarchicalTracker displays multi-line progress for hierarchical operations
// Uses save/restore cursor with separators for stable rendering
// Line 1: Spinner with total files/folders found and elapsed time
// Line 2: Progress bar showing completion of top-level items
// Line 3: Current path being processed
// Lines 4+: Currently active folders being scanned (up to maxActiveFolders)
type HierarchicalTracker struct {
	label            string                  // Operation label
	itemsFound       int64                   // Total items found
	foldersScanned   int64                   // Total folders scanned
	currentPath      string                  // Current path being processed
	topLevelTotal    int                     // Total top-level items (set once)
	topLevelComplete int                     // Completed top-level items
	activeFolders    []string                // Currently active folders being scanned
	startTime        time.Time               // When tracking started
	updateCh         chan HierarchicalUpdate // Channel for receiving updates
	folderCh         chan ActiveFolderUpdate // Channel for active folder updates
	spinnerIndex     int                     // Current spinner frame index
	mu               sync.Mutex              // Protects internal state
	done             bool                    // Whether tracking is complete
	ticker           *time.Ticker            // For spinner animation
	stopCh           chan struct{}           // Signal to stop the tracker
	mode             TrackerMode             // Rendering mode (alt screen vs scroll region)
	terminalRows     int                     // Terminal height for scroll region mode
	terminalCols     int                     // Terminal width for dynamic sizing
	resizeCh         chan os.Signal          // Channel for terminal resize signals
}

// NewHierarchical creates a new hierarchical progress tracker with alternate screen mode (default)
// label: The operation label to display
// topLevelTotal: The number of top-level items to track (for progress bar)
func NewHierarchical(label string, topLevelTotal int) *HierarchicalTracker {
	return NewHierarchicalWithMode(label, topLevelTotal, ModeAltScreen)
}

// NewHierarchicalWithMode creates a new hierarchical progress tracker with specified mode
// label: The operation label to display
// topLevelTotal: The number of top-level items to track (for progress bar)
// mode: ModeAltScreen for separate screen, ModeScrollRegion for sticky header with logs below
func NewHierarchicalWithMode(label string, topLevelTotal int, mode TrackerMode) *HierarchicalTracker {
	cols, rows := getTerminalSize()
	return &HierarchicalTracker{
		label:         label,
		topLevelTotal: topLevelTotal,
		activeFolders: make([]string, 0, maxActiveFolders),
		updateCh:      make(chan HierarchicalUpdate, 100),
		folderCh:      make(chan ActiveFolderUpdate, 100),
		stopCh:        make(chan struct{}),
		mode:          mode,
		terminalRows:  rows,
		terminalCols:  cols,
		resizeCh:      make(chan os.Signal, 1),
	}
}

// Start begins displaying progress. Call this in a goroutine.
func (t *HierarchicalTracker) Start() {
	t.mu.Lock()
	t.startTime = time.Now()
	// 80ms ≈ 12fps - smooth spinner animation without being too fast
	t.ticker = time.NewTicker(80 * time.Millisecond)
	mode := t.mode
	t.mu.Unlock()

	// Register for terminal resize signals (SIGWINCH)
	signal.Notify(t.resizeCh, syscall.SIGWINCH)
	defer signal.Stop(t.resizeCh)

	// Setup terminal based on mode
	t.setupTerminal(mode)

	for {
		select {
		case <-t.stopCh:
			t.ticker.Stop()
			return
		case update, ok := <-t.updateCh:
			if !ok {
				t.ticker.Stop()
				return
			}
			// Just update state - ticker handles rendering at 60fps
			t.mu.Lock()
			t.itemsFound = update.ItemsFound
			t.foldersScanned = update.FoldersScanned
			t.currentPath = update.CurrentPath
			t.topLevelComplete = update.TopLevelComplete
			t.mu.Unlock()
		case folderUpdate, ok := <-t.folderCh:
			if !ok {
				continue
			}
			t.handleFolderUpdate(folderUpdate)
		case <-t.resizeCh:
			// Terminal was resized - update dimensions and re-setup
			t.handleResize()
		case <-t.ticker.C:
			t.render()
		}
	}
}

// setupTerminal configures the terminal for the current mode
func (t *HierarchicalTracker) setupTerminal(mode TrackerMode) {
	switch mode {
	case ModeAltScreen:
		// Enter alternate screen buffer - gives us a clean screen that doesn't
		// interfere with the main terminal scrollback. All log output goes to
		// the main screen, while the tracker has its own dedicated display.
		fmt.Print(enterAltScreen)
		fmt.Print(hideCursor)
	case ModeScrollRegion:
		// Scroll region mode: tracker at top, logs scroll below
		// Clear screen and set up scroll region
		fmt.Print(hideCursor)
		fmt.Print(clearScreen)
		fmt.Print(moveCursorHome)
		// Reserve top lines for tracker, set scroll region for remaining lines
		// Logs will appear in the scroll region below the tracker
		t.mu.Lock()
		fmt.Print(setScrollRegion(trackerHeaderLines+1, t.terminalRows))
		t.mu.Unlock()
		// Position cursor in the scroll region for log output
		fmt.Print(moveCursorTo(trackerHeaderLines+1, 1))
	}
}

// handleResize updates terminal dimensions and reconfigures the display
func (t *HierarchicalTracker) handleResize() {
	cols, rows := getTerminalSize()

	t.mu.Lock()
	t.terminalCols = cols
	t.terminalRows = rows
	mode := t.mode
	t.mu.Unlock()

	// Re-setup terminal with new dimensions
	if mode == ModeScrollRegion {
		// Save cursor position, update scroll region, restore cursor
		lineMu.Lock()
		fmt.Print(saveCursor)
		fmt.Print(setScrollRegion(trackerHeaderLines+1, rows))
		fmt.Print(restoreCursor)
		lineMu.Unlock()
	}
}

// handleFolderUpdate processes active folder add/remove operations
func (t *HierarchicalTracker) handleFolderUpdate(update ActiveFolderUpdate) {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch update.Action {
	case "add":
		// Add folder if not already present and under limit
		for _, f := range t.activeFolders {
			if f == update.Path {
				return // Already tracking
			}
		}
		if len(t.activeFolders) < maxActiveFolders {
			t.activeFolders = append(t.activeFolders, update.Path)
		}
	case "remove":
		// Remove folder from active list
		for i, f := range t.activeFolders {
			if f == update.Path {
				t.activeFolders = append(t.activeFolders[:i], t.activeFolders[i+1:]...)
				break
			}
		}
	}
}

// Update sends a progress update to the tracker
func (t *HierarchicalTracker) Update(itemsFound, foldersScanned int64, currentPath string, topLevelComplete int) {
	select {
	case t.updateCh <- HierarchicalUpdate{
		ItemsFound:       itemsFound,
		FoldersScanned:   foldersScanned,
		CurrentPath:      currentPath,
		TopLevelComplete: topLevelComplete,
	}:
	default:
		// Channel full, skip this update
	}
}

// AddActiveFolder marks a folder as currently being scanned
func (t *HierarchicalTracker) AddActiveFolder(path string) {
	select {
	case t.folderCh <- ActiveFolderUpdate{Path: path, Action: "add"}:
	default:
		// Channel full, skip this update
	}
}

// RemoveActiveFolder marks a folder as no longer being scanned
func (t *HierarchicalTracker) RemoveActiveFolder(path string) {
	select {
	case t.folderCh <- ActiveFolderUpdate{Path: path, Action: "remove"}:
	default:
		// Channel full, skip this update
	}
}

// SetTopLevelTotal updates the total number of top-level items
// Call this after the first API response reveals the count
func (t *HierarchicalTracker) SetTopLevelTotal(total int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.topLevelTotal = total
}

// Done stops the tracker and displays completion
func (t *HierarchicalTracker) Done() {
	t.mu.Lock()
	if t.done {
		t.mu.Unlock()
		return
	}
	t.done = true
	itemsFound := t.itemsFound
	foldersScanned := t.foldersScanned
	elapsed := time.Since(t.startTime)
	label := t.label
	mode := t.mode
	t.mu.Unlock()

	close(t.stopCh)
	close(t.updateCh)
	close(t.folderCh)

	lineMu.Lock()

	// Cleanup based on mode
	switch mode {
	case ModeAltScreen:
		// Exit alternate screen buffer - returns to main terminal with all scrollback intact
		fmt.Print(showCursor)
		fmt.Print(exitAltScreen)
	case ModeScrollRegion:
		// Reset scroll region to full screen
		fmt.Print(resetScrollRgn)
		// Move cursor below the tracker area
		fmt.Print(moveCursorTo(trackerHeaderLines+1, 1))
		// Clear from cursor to end of screen to remove any tracker remnants
		fmt.Print(clearToEndOfScreen)
		fmt.Print(showCursor)
	}

	// Print final completion message to main terminal
	timestamp := time.Now().Format("2006/01/02 03:04:05 PM")
	finalMsg := fmt.Sprintf("[%s] \033[32m✓\033[0m %s | %s files | %s folders | Complete [%s]",
		timestamp,
		label,
		formatNumber(itemsFound),
		formatNumber(foldersScanned),
		formatDuration(elapsed),
	)
	fmt.Println(finalMsg)
	lineMu.Unlock()
}

// render updates the multi-line progress display using double-buffered rendering
// This builds the entire frame in memory, then writes it in a single syscall
// to eliminate flicker even at 60fps
func (t *HierarchicalTracker) render() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.done {
		return
	}

	// Advance spinner
	t.spinnerIndex = (t.spinnerIndex + 1) % len(spinnerFrames)
	spinner := spinnerFrames[t.spinnerIndex]

	elapsed := time.Since(t.startTime)
	timestamp := time.Now().Format("2006/01/02 03:04:05 PM")

	// Calculate progress percentage
	var percent int
	if t.topLevelTotal > 0 {
		percent = (t.topLevelComplete * 100) / t.topLevelTotal
	}

	// Build progress bar (20 chars wide for better visibility)
	barWidth := 20
	filled := (percent * barWidth) / 100
	if filled > barWidth {
		filled = barWidth
	}
	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	// Use dynamic terminal width (with minimum of 80 for safety)
	termWidth := t.terminalCols
	if termWidth < 80 {
		termWidth = 80
	}

	// Double-buffered rendering: build entire frame in memory first
	var buf strings.Builder
	buf.Grow(2048) // Pre-allocate to reduce allocations

	// For scroll region mode: save cursor position before moving to header
	if t.mode == ModeScrollRegion {
		buf.WriteString(saveCursor)
	}

	// Move cursor to home position (no clear - we overwrite in place)
	buf.WriteString(moveCursorHome)

	// Helper to write a padded line (overwrites previous content without clear)
	writePaddedLine := func(content string) {
		buf.WriteString(content)
		// Pad with spaces to overwrite any previous longer content, then clear to EOL
		buf.WriteString(clearLineRemainder)
		buf.WriteString("\n")
	}

	// Header separator
	writePaddedLine(progressSeparator)

	// Line 1: Spinner with totals
	line1 := fmt.Sprintf("[%s] %s %s | %s files | %s folders [%s]",
		timestamp,
		spinner,
		t.label,
		formatNumber(t.itemsFound),
		formatNumber(t.foldersScanned),
		formatDuration(elapsed),
	)
	writePaddedLine(line1)

	// Line 2: Progress bar
	line2 := fmt.Sprintf("[%s] %3d%% |%s| %d/%d top-level",
		timestamp,
		percent,
		bar,
		t.topLevelComplete,
		t.topLevelTotal,
	)
	writePaddedLine(line2)

	// Line 3: Current path being processed
	currentPathDisplay := truncatePath(t.currentPath, termWidth-6)
	if currentPathDisplay == "" {
		currentPathDisplay = "initializing..."
	}
	line3 := fmt.Sprintf("  \033[36m→\033[0m %s", currentPathDisplay)
	writePaddedLine(line3)

	// Lines 4+: Active folders being scanned (concurrent workers)
	if len(t.activeFolders) > 0 {
		writePaddedLine(fmt.Sprintf("  \033[33mActive workers (%d):\033[0m", len(t.activeFolders)))
	}
	for i, folder := range t.activeFolders {
		prefix := "├─"
		if i == len(t.activeFolders)-1 {
			prefix = "└─"
		}
		displayFolder := truncatePath(folder, termWidth-12)
		folderLine := fmt.Sprintf("    %s %s %s",
			prefix,
			spinnerFrames[(t.spinnerIndex+i)%len(spinnerFrames)],
			displayFolder,
		)
		writePaddedLine(folderLine)
	}

	// Closing separator
	writePaddedLine(progressSeparator)

	// Calculate how many lines we actually used:
	// 1 (top separator) + 1 (stats) + 1 (progress bar) + 1 (current path) +
	// (1 if activeFolders > 0 for header) + len(activeFolders) + 1 (bottom separator)
	linesUsed := 5 // separator + 3 main lines + separator
	if len(t.activeFolders) > 0 {
		linesUsed += 1 + len(t.activeFolders) // header + folder lines
	}

	// Clear any remaining lines up to trackerHeaderLines to remove stale content
	// This is needed in both modes when the number of active folders decreases
	for i := linesUsed; i < trackerHeaderLines; i++ {
		buf.WriteString(clearLine)
		buf.WriteString("\n")
	}

	// For scroll region mode: restore cursor to continue logging in scroll area
	if t.mode == ModeScrollRegion {
		buf.WriteString(restoreCursor)
	}

	// Single atomic write to terminal - eliminates flicker
	lineMu.Lock()
	fmt.Print(buf.String())
	lineMu.Unlock()
}

// ItemsFound returns the current count of items found
func (t *HierarchicalTracker) ItemsFound() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.itemsFound
}

// FoldersScanned returns the current count of folders scanned
func (t *HierarchicalTracker) FoldersScanned() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.foldersScanned
}

// Elapsed returns the duration since tracking started
func (t *HierarchicalTracker) Elapsed() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return time.Since(t.startTime)
}

// TopLevelProgress returns the current top-level progress (complete, total)
func (t *HierarchicalTracker) TopLevelProgress() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.topLevelComplete, t.topLevelTotal
}

// ActiveFolderCount returns the number of currently active folders
func (t *HierarchicalTracker) ActiveFolderCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.activeFolders)
}

// Stats returns a snapshot of all current statistics
type TrackerStats struct {
	ItemsFound       int64
	FoldersScanned   int64
	TopLevelComplete int
	TopLevelTotal    int
	ActiveFolders    int
	Elapsed          time.Duration
	Rate             float64 // items per minute
}

// GetStats returns a snapshot of all current statistics (thread-safe)
func (t *HierarchicalTracker) GetStats() TrackerStats {
	t.mu.Lock()
	defer t.mu.Unlock()

	elapsed := time.Since(t.startTime)
	rate := 0.0
	if elapsed.Minutes() > 0 {
		rate = float64(t.itemsFound) / elapsed.Minutes()
	}

	return TrackerStats{
		ItemsFound:       t.itemsFound,
		FoldersScanned:   t.foldersScanned,
		TopLevelComplete: t.topLevelComplete,
		TopLevelTotal:    t.topLevelTotal,
		ActiveFolders:    len(t.activeFolders),
		Elapsed:          elapsed,
		Rate:             rate,
	}
}

// Summary returns a human-readable summary string of current progress
// Useful for checkpoint logging
func (t *HierarchicalTracker) Summary() string {
	stats := t.GetStats()
	return fmt.Sprintf("Files: %s | Folders: %s | Top-level: %d/%d | Active: %d | Elapsed: %s | Rate: %.1f/min",
		formatNumber(stats.ItemsFound),
		formatNumber(stats.FoldersScanned),
		stats.TopLevelComplete,
		stats.TopLevelTotal,
		stats.ActiveFolders,
		formatDuration(stats.Elapsed),
		stats.Rate,
	)
}
