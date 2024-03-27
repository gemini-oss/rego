// pkg/internal/tests/jamf/devices_test.go
package jamf_test

import (
	"testing"
)

func TestListAllComputers(t *testing.T) {
	server, cleanup := setupTestServer(t, ComputersInventory)
	defer cleanup()

	client := setupTestClient(server.URL)

	computers, err := client.Devices().ListAllComputers()
	if err != nil {
		t.Fatalf("Expected no error, got `%v`", err)
	}

	if computers.TotalCount != 3 {
		t.Errorf("Expected `3` computers, got `%d`", computers.TotalCount)
	}
}

func TestGetComputerDetails(t *testing.T) {

	server, cleanup := setupTestServer(t, ComputersInventoryDetail)
	defer cleanup()

	client := setupTestClient(server.URL)

	computerDetails, err := client.Devices().GetComputerDetails("1")
	if err != nil {
		t.Fatalf("Expected no error, got `%v`", err)
	}

	if computerDetails.ID != "1" {
		t.Errorf("Expected ID `1`, got `%s`", computerDetails.ID)
	}
	if computerDetails.UDID != "123" {
		t.Errorf("Expected UDID `123`, got `%s`", computerDetails.UDID)
	}
}

func TestListAllComputerGroups(t *testing.T) {
	server, cleanup := setupTestServer(t, ComputerGroups)
	defer cleanup()

	client := setupTestClient(server.URL)

	groups, err := client.Devices().ListAllComputerGroups()
	if err != nil {
		t.Fatalf("Expected no error, got `%v`", err)
	}

	if len(*groups) != 3 {
		t.Errorf("Expected `3` groups, got `%d`", len(*groups))
	}
}

// Example test for ListAllMobileDevices
func TestListAllMobileDevices(t *testing.T) {

	server, cleanup := setupTestServer(t, MobileDevices)
	defer cleanup()

	client := setupTestClient(server.URL)

	devices, err := client.Devices().ListAllMobileDevices()
	if err != nil {
		t.Fatalf("Expected no error, got `%v`", err)
	}

	if len(*devices.Results) != 3 {
		t.Errorf("Expected `3` devices, got `%d`", len(*devices.Results))
	}
}
