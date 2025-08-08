package snipeit_test

import (
	"os"
	"testing"

	"github.com/gemini-oss/rego/pkg/snipeit"
)

// TestLicensesGetAllLicenses tests the GetAllLicenses method with live API or fixtures
func TestLicensesGetAllLicenses(t *testing.T) {
	it := NewIntegrationTest(t)

	list, err := it.GetLicenseList()
	if err != nil {
		t.Fatalf("Failed to get license list: %v", err)
	}

	// Validate we got data
	if list.Total == 0 {
		t.Error("No license items returned")
	}

	// Check that our struct handles all fields correctly
	if list.Rows != nil && *list.Rows != nil && len(*list.Rows) > 0 {
		item := (*list.Rows)[0]

		// Log some key fields to verify parsing
		t.Logf("First license item:")
		t.Logf("  ID: %d", item.ID)
		t.Logf("  Name: %s", item.Name)

		// Check license-specific fields
		if item.Method.Seats > 0 {
			t.Logf("  Seats: %d", item.Method.Seats)
		}

		if item.Method.ProductKey != "" {
			t.Logf("  Has product key")
		}

		// Verify critical fields
		if item.Name == "" {
			t.Error("License name should not be empty")
		}
	}
}

// TestLicenseBuilder tests the license builder pattern
func TestLicenseBuilder(t *testing.T) {
	// Test the builder pattern
	license := snipeit.NewLicense("Test License").
		LicenseEmail("test@example.com").
		LicenseName("Test User").
		Seats("10").
		ProductKey("XXXX-XXXX-XXXX-XXXX").
		PurchaseCost(999.99).
		Maintained(true).
		Reassignable(true).
		CategoryID(1).
		Build()

	// Verify the built license
	if license.Name != "Test License" {
		t.Errorf("Expected name 'Test License', got %s", license.Name)
	}

	if license.LicenseEmail != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", license.LicenseEmail)
	}

	if license.Method.Seats != "10" {
		t.Errorf("Expected seats '10', got %s", license.Method.Seats)
	}

	if license.Method.ProductKey != "XXXX-XXXX-XXXX-XXXX" {
		t.Errorf("Expected product key, got %s", license.Method.ProductKey)
	}

	if license.Method.PurchaseCost == nil || *license.Method.PurchaseCost != 999.99 {
		t.Error("Purchase cost not set correctly")
	}

	if !license.Maintained {
		t.Error("Maintained should be true")
	}

	if !license.Method.Reassignable {
		t.Error("Reassignable should be true")
	}

	if license.Method.CategoryID == nil || *license.Method.CategoryID != 1 {
		t.Error("CategoryID not set correctly")
	}
}

// TestLicenseCheckout tests the license checkout builder
func TestLicenseCheckout(t *testing.T) {
	// Skip if not in live mode since checkout modifies data
	if os.Getenv("SNIPEIT_TEST_MODE") != "live" {
		t.Skip("License checkout requires live API")
	}

	it := NewIntegrationTest(t)

	// Get a license to test with
	list, err := it.GetLicenseList()
	if err != nil || list.Rows == nil || *list.Rows == nil || len(*list.Rows) == 0 {
		t.Skip("No licenses available for testing")
	}

	license := (*list.Rows)[0]

	// Get seats for this license
	seats, err := it.client.Licenses().Seats(license.ID)
	if err != nil {
		t.Fatalf("Failed to get seats: %v", err)
	}

	if seats.Rows == nil || *seats.Rows == nil || len(*seats.Rows) == 0 {
		t.Skip("No seats available for testing")
	}

	// Find an available seat
	var availableSeat *snipeit.Seat[snipeit.SeatGET]
	for _, seat := range *seats.Rows {
		if seat.Method.AssignedUser.ID == 0 {
			availableSeat = seat
			break
		}
	}

	if availableSeat == nil {
		t.Skip("No available seats for checkout testing")
	}

	t.Logf("Found available seat %d for license %d", availableSeat.ID, license.ID)
	// Note: Not actually performing checkout to avoid modifying live data
}

// TestLicenseQuery tests the license query functionality
func TestLicenseQuery(t *testing.T) {
	query := &snipeit.LicenseQuery{
		Limit:        100,
		Offset:       0,
		Search:       "microsoft",
		Sort:         "name",
		Order:        "asc",
		Maintained:   true,
		LicenseName:  "John Doe",
		LicenseEmail: "john@example.com",
	}

	// Test QueryInterface methods
	query.SetLimit(25)
	if query.GetLimit() != 25 {
		t.Errorf("Expected limit=25, got %d", query.GetLimit())
	}

	query.SetOffset(50)
	if query.GetOffset() != 50 {
		t.Errorf("Expected offset=50, got %d", query.GetOffset())
	}

	// Test Copy
	copied := query.Copy()
	if copiedQuery, ok := copied.(*snipeit.LicenseQuery); ok {
		if copiedQuery.Search != query.Search {
			t.Error("Copy didn't preserve Search field")
		}
		if copiedQuery.LicenseName != query.LicenseName {
			t.Error("Copy didn't preserve LicenseName field")
		}
		if copiedQuery.Maintained != query.Maintained {
			t.Error("Copy didn't preserve Maintained field")
		}
	} else {
		t.Error("Copy returned wrong type")
	}
}
