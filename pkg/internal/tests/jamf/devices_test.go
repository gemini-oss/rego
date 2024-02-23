// pkg/internal/tests/jamf/devices_test.go
package jamf_test

import (
	"testing"
)

// Example test for ListAllMobileDevices
func TestListAllMobileDevices(t *testing.T) {
	responseMap := map[string]string{
		"/api/v2/mobile-devices": `{"totalCount":3,"results":[{"id":"1","name":"iPad"}, {"id":"2","name":"iPad"}, {"id":"3","name":"iPad"}]}`,
	}

	server, cleanup := setupTestServer(t, responseMap)
	defer cleanup()

	client := setupTestClient(server.URL)

	devices, err := client.ListAllMobileDevices()
	if err != nil {
		t.Fatalf("Expected no error, got `%v`", err)
	}

	if len(*devices.Results) != 3 {
		t.Errorf("Expected `3` devices, got `%d`", len(*devices.Results))
	}
}
