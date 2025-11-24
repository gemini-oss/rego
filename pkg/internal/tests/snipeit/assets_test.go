package snipeit_test

import (
	"os"
	"testing"

	"github.com/gemini-oss/rego/pkg/snipeit"
)

// TestAssetsGetAllAssets tests the GetAllAssets method with live API or fixtures
func TestAssetsGetAllAssets(t *testing.T) {
	it := NewIntegrationTest(t)

	list, err := it.GetHardwareList()
	if err != nil {
		t.Fatalf("Failed to get hardware list: %v", err)
	}

	// Validate we got data
	if list.Total == 0 {
		t.Error("No hardware items returned")
	}

	// Check that our struct handles all fields correctly
	if list.Rows != nil && *list.Rows != nil && len(*list.Rows) > 0 {
		item := (*list.Rows)[0]

		// Log some key fields to verify parsing
		t.Logf("First hardware item:")
		t.Logf("  ID: %d", item.ID)
		t.Logf("  Serial: %s", item.Serial)
		t.Logf("  Asset Tag: %s", item.AssetTag)

		// Check fields that have caused issues
		if item.WarrantyExpires != nil {
			t.Logf("  Warranty Expires: %s", item.WarrantyExpires.Format("2006-01-02"))
		}

		if item.CustomFields != nil && *item.CustomFields != nil {
			t.Logf("  Custom Fields: %d fields", len(*item.CustomFields))
		}

		// Verify critical fields are populated
		if item.Serial == "" {
			t.Error("Serial number should not be empty")
		}
	}
}

// TestAssetsGetBySerial tests getting a specific asset by serial number
func TestAssetsGetBySerial(t *testing.T) {
	// Skip if not in live mode since we need a real serial
	if os.Getenv("SNIPEIT_TEST_MODE") != "live" {
		t.Skip("GetBySerial requires live API")
	}

	it := NewIntegrationTest(t)

	// First get list to find a valid serial
	list, err := it.GetHardwareList()
	if err != nil || list.Rows == nil || *list.Rows == nil || len(*list.Rows) == 0 {
		t.Skip("No assets available for testing")
	}

	testSerial := (*list.Rows)[0].Serial
	if testSerial == "" {
		t.Skip("First asset has no serial number")
	}

	// Test GetAssetBySerial
	asset, err := it.client.Assets().GetAssetBySerial(testSerial)
	if err != nil {
		t.Fatalf("Failed to get asset by serial %s: %v", testSerial, err)
	}

	if asset == nil {
		t.Fatal("GetBySerial returned nil")
	}

	if asset.Serial != testSerial {
		t.Errorf("Expected serial %s, got %s", testSerial, asset.Serial)
	}
}

// TestAssetsQuery tests the query functionality
func TestAssetsQuery(t *testing.T) {
	// Test the query interface implementation
	query := &snipeit.AssetQuery{
		Limit:  100,
		Offset: 0,
		Search: "test",
		Sort:   "id",
		Order:  "asc",
	}

	// Test QueryInterface methods
	query.SetLimit(50)
	if query.GetLimit() != 50 {
		t.Errorf("Expected limit=50, got %d", query.GetLimit())
	}

	query.SetOffset(10)
	if query.GetOffset() != 10 {
		t.Errorf("Expected offset=10, got %d", query.GetOffset())
	}

	// Test Copy
	copied := query.Copy()
	if copiedQuery, ok := copied.(*snipeit.AssetQuery); ok {
		if copiedQuery.Search != query.Search {
			t.Error("Copy didn't preserve Search field")
		}
		if copiedQuery.Limit != query.Limit {
			t.Error("Copy didn't preserve Limit field")
		}
	} else {
		t.Error("Copy returned wrong type")
	}
}
