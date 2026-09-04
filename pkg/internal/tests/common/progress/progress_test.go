/*
# Progress - Test

This package runs tests for the progress tracking utilities.

:Copyright: (c) 2026 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/common/progress/progress_test.go
package progress_test

import (
	"sync"
	"testing"
	"time"

	"github.com/gemini-oss/rego/pkg/common/progress"
)

func TestNewIndeterminate(t *testing.T) {
	tracker := progress.NewIndeterminate("Test Operation")

	if tracker == nil {
		t.Fatal("Expected tracker to be created, got nil")
	}
}

func TestIndeterminateTrackerStartAndDone(t *testing.T) {
	tracker := progress.NewIndeterminate("Test Operation")

	// Start tracker in goroutine
	go tracker.Start()

	// Give it a moment to initialize
	time.Sleep(100 * time.Millisecond)

	// Send some updates with larger numbers
	tracker.Update(10000, "/test/path/to/folder/1")
	time.Sleep(200 * time.Millisecond)

	tracker.Update(25000, "/test/path/to/folder/2/subfolder")
	time.Sleep(200 * time.Millisecond)

	tracker.Update(50000, "/test/path/to/folder/3/deep/nested/path")
	time.Sleep(200 * time.Millisecond)

	// Verify items found
	if tracker.ItemsFound() != 50000 {
		t.Errorf("Expected ItemsFound to be 50000, got %d", tracker.ItemsFound())
	}

	// Stop tracker
	tracker.Done()

	// Calling Done again should be safe (idempotent)
	tracker.Done()
}

func TestIndeterminateTrackerConcurrentUpdates(t *testing.T) {
	tracker := progress.NewIndeterminate("Concurrent Test")

	go tracker.Start()
	time.Sleep(50 * time.Millisecond)

	var wg sync.WaitGroup
	numGoroutines := 10
	updatesPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < updatesPerGoroutine; j++ {
				tracker.Update(int64((id*updatesPerGoroutine)+j), "/concurrent/path")
			}
		}(i)
	}

	wg.Wait()
	tracker.Done()

	// Should complete without panic or race conditions
}

func TestIndeterminateTrackerElapsed(t *testing.T) {
	tracker := progress.NewIndeterminate("Elapsed Test")

	go tracker.Start()
	time.Sleep(100 * time.Millisecond)

	elapsed := tracker.Elapsed()
	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected elapsed time >= 100ms, got %v", elapsed)
	}

	tracker.Done()
}

func TestIndeterminateTrackerBufferedChannel(t *testing.T) {
	tracker := progress.NewIndeterminate("Buffer Test")

	go tracker.Start()
	time.Sleep(50 * time.Millisecond)

	// Rapidly send many updates - should not block due to buffered channel
	for i := 0; i < 200; i++ {
		tracker.Update(int64(i), "/rapid/update")
	}

	// Should complete quickly without blocking
	tracker.Done()
}

func TestMultipleTrackers(t *testing.T) {
	tracker1 := progress.NewIndeterminate("Tracker 1")
	tracker2 := progress.NewIndeterminate("Tracker 2")

	go tracker1.Start()
	go tracker2.Start()

	time.Sleep(50 * time.Millisecond)

	tracker1.Update(100, "/path/1")
	tracker2.Update(200, "/path/2")

	time.Sleep(50 * time.Millisecond)

	if tracker1.ItemsFound() != 100 {
		t.Errorf("Tracker1 expected 100 items, got %d", tracker1.ItemsFound())
	}

	if tracker2.ItemsFound() != 200 {
		t.Errorf("Tracker2 expected 200 items, got %d", tracker2.ItemsFound())
	}

	tracker1.Done()
	tracker2.Done()
}

// HierarchicalTracker Tests

func TestNewHierarchical(t *testing.T) {
	tracker := progress.NewHierarchical("Test Operation", 5)

	if tracker == nil {
		t.Fatal("Expected hierarchical tracker to be created, got nil")
	}
}

func TestHierarchicalTrackerStartAndDone(t *testing.T) {
	tracker := progress.NewHierarchical("Hierarchical Test", 10)

	// Start tracker in goroutine
	go tracker.Start()

	// Give it a moment to initialize
	time.Sleep(100 * time.Millisecond)

	// Send updates with larger numbers and more realistic paths
	tracker.Update(15000, 50, "/My Drive/Documents/Projects/2024", 1)
	time.Sleep(300 * time.Millisecond)

	tracker.Update(45000, 150, "/My Drive/Documents/Projects/Archive/Old Files", 3)
	time.Sleep(300 * time.Millisecond)

	tracker.Update(125000, 500, "/Shared Drives/Engineering/Source Code/Backend", 6)
	time.Sleep(300 * time.Millisecond)

	tracker.Update(250000, 1200, "/Shared Drives/Marketing/Campaigns/Q4-2024", 8)
	time.Sleep(300 * time.Millisecond)

	tracker.Update(500000, 2500, "/Shared Drives/Company/All Departments/Resources", 10)
	time.Sleep(200 * time.Millisecond)

	// Verify items found
	if tracker.ItemsFound() != 500000 {
		t.Errorf("Expected ItemsFound to be 500000, got %d", tracker.ItemsFound())
	}

	// Verify folders scanned
	if tracker.FoldersScanned() != 2500 {
		t.Errorf("Expected FoldersScanned to be 2500, got %d", tracker.FoldersScanned())
	}

	// Stop tracker
	tracker.Done()

	// Calling Done again should be safe (idempotent)
	tracker.Done()
}

func TestHierarchicalTrackerSetTopLevelTotal(t *testing.T) {
	tracker := progress.NewHierarchical("TopLevel Test", 0)

	go tracker.Start()
	time.Sleep(50 * time.Millisecond)

	// Set top level total after start (simulating discovery)
	tracker.SetTopLevelTotal(10)

	// Send updates
	tracker.Update(100, 10, "/test/path", 5)
	time.Sleep(50 * time.Millisecond)

	tracker.Done()
}

func TestHierarchicalTrackerActiveFolders(t *testing.T) {
	tracker := progress.NewHierarchical("ActiveFolders Test", 8)

	go tracker.Start()
	time.Sleep(100 * time.Millisecond)

	// Simulate scanning multiple folders concurrently
	tracker.AddActiveFolder("/My Drive/Documents/Work")
	tracker.Update(1000, 10, "/My Drive/Documents/Work/file1.pdf", 0)
	time.Sleep(200 * time.Millisecond)

	tracker.AddActiveFolder("/My Drive/Documents/Personal")
	tracker.Update(2500, 25, "/My Drive/Documents/Personal/photos", 0)
	time.Sleep(200 * time.Millisecond)

	tracker.AddActiveFolder("/Shared Drives/Engineering/Code")
	tracker.Update(5000, 50, "/Shared Drives/Engineering/Code/main.go", 1)
	time.Sleep(200 * time.Millisecond)

	tracker.AddActiveFolder("/Shared Drives/Marketing/Assets")
	tracker.Update(8000, 80, "/Shared Drives/Marketing/Assets/logo.png", 1)
	time.Sleep(200 * time.Millisecond)

	// First folder completes
	tracker.RemoveActiveFolder("/My Drive/Documents/Work")
	tracker.Update(12000, 120, "/My Drive/Documents/Personal/videos", 2)
	time.Sleep(200 * time.Millisecond)

	// Add more folders as workers become available
	tracker.AddActiveFolder("/Shared Drives/HR/Policies")
	tracker.AddActiveFolder("/Shared Drives/Finance/Reports")
	tracker.Update(18000, 180, "/Shared Drives/HR/Policies/handbook.pdf", 3)
	time.Sleep(300 * time.Millisecond)

	// Several complete at once
	tracker.RemoveActiveFolder("/My Drive/Documents/Personal")
	tracker.RemoveActiveFolder("/Shared Drives/Engineering/Code")
	tracker.Update(25000, 250, "/Shared Drives/Marketing/Assets/banner.jpg", 5)
	time.Sleep(200 * time.Millisecond)

	// Final folders
	tracker.RemoveActiveFolder("/Shared Drives/Marketing/Assets")
	tracker.RemoveActiveFolder("/Shared Drives/HR/Policies")
	tracker.RemoveActiveFolder("/Shared Drives/Finance/Reports")
	tracker.Update(35000, 350, "Complete", 8)
	time.Sleep(200 * time.Millisecond)

	tracker.Done()

	// Should complete without panic
}

func TestHierarchicalTrackerConcurrentUpdates(t *testing.T) {
	tracker := progress.NewHierarchical("Concurrent Hierarchical Test", 10)

	go tracker.Start()
	time.Sleep(50 * time.Millisecond)

	var wg sync.WaitGroup
	numGoroutines := 10
	updatesPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < updatesPerGoroutine; j++ {
				tracker.Update(
					int64((id*updatesPerGoroutine)+j),
					int64(id),
					"/concurrent/path",
					id%10,
				)
				// Also test active folder updates
				if j%10 == 0 {
					tracker.AddActiveFolder("/concurrent/folder" + string(rune('A'+id)))
				}
				if j%15 == 0 {
					tracker.RemoveActiveFolder("/concurrent/folder" + string(rune('A'+id)))
				}
			}
		}(i)
	}

	wg.Wait()
	tracker.Done()

	// Should complete without panic or race conditions
}

func TestHierarchicalTrackerProgressBar(t *testing.T) {
	tracker := progress.NewHierarchical("Progress Bar Test", 10)

	go tracker.Start()
	time.Sleep(100 * time.Millisecond)

	// Simulate processing 10 top-level shared drives
	driveNames := []string{
		"/Shared Drives/Engineering",
		"/Shared Drives/Marketing",
		"/Shared Drives/Sales",
		"/Shared Drives/HR",
		"/Shared Drives/Finance",
		"/Shared Drives/Legal",
		"/Shared Drives/Operations",
		"/Shared Drives/Executive",
		"/Shared Drives/IT",
		"/Shared Drives/Support",
	}

	for i := range 10 {
		// Add the drive as active
		tracker.AddActiveFolder(driveNames[i])
		time.Sleep(150 * time.Millisecond)

		// Simulate scanning with increasing totals
		filesFound := int64((i + 1) * 25000)
		foldersScanned := int64((i + 1) * 500)
		tracker.Update(filesFound, foldersScanned, driveNames[i]+"/current/file.txt", i)
		time.Sleep(200 * time.Millisecond)

		// Remove as complete
		tracker.RemoveActiveFolder(driveNames[i])
		tracker.Update(filesFound+5000, foldersScanned+100, driveNames[i]+" complete", i+1)
		time.Sleep(100 * time.Millisecond)
	}

	tracker.Done()
}
