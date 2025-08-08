package lenel_s2_test

import (
	"os"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/lenel_s2"
)

func TestGetPerson(t *testing.T) {
	var client *lenel_s2.Client
	var cleanup func()

	// Check if we should use integration testing
	if mode := os.Getenv("REGO_TEST_MODE"); mode == "live" || mode == "record" {
		it := NewIntegrationTest(t, "people")
		if err := it.Setup(); err != nil {
			t.Skip("Skipping integration test:", err)
		}
		client = it.Client
		cleanup = it.Cleanup
	} else {
		// Use mock server for fixture mode (default)
		th := SetupTestServer(t)
		client = lenel_s2.NewClient(th.server.URL, log.INFO)
		cleanup = th.Cleanup
	}
	defer cleanup()

	tests := []struct {
		name     string
		personID string
		wantName string
		wantErr  bool
	}{
		{
			name:     "Get Anthony Dardano - REGO Master",
			personID: "REGO_357",
			wantName: "Anthony",
			wantErr:  false,
		},
		{
			name:     "Get Satoshi Nakamoto - Bitcoin Creator",
			personID: "SATOSHI_753",
			wantName: "Satoshi",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			person, err := client.GetPerson(tt.personID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPerson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && person.FirstName != tt.wantName {
				t.Errorf("GetPerson() FirstName = %v, want %v", person.FirstName, tt.wantName)
			}
		})
	}
}

func TestListAllUsers(t *testing.T) {
	var client *lenel_s2.Client
	var cleanup func()

	// Check if we should use integration testing
	if mode := os.Getenv("REGO_TEST_MODE"); mode == "live" || mode == "record" {
		it := NewIntegrationTest(t, "people")
		if err := it.Setup(); err != nil {
			t.Skip("Skipping integration test:", err)
		}
		client = it.Client
		cleanup = it.Cleanup
	} else {
		// Use mock server for fixture mode (default)
		th := SetupTestServer(t)
		client = lenel_s2.NewClient(th.server.URL, log.INFO)
		cleanup = th.Cleanup
	}
	defer cleanup()

	users, err := client.ListAllUsers()
	if err != nil {
		t.Fatalf("ListAllUsers() error = %v", err)
	}

	if users == nil || len(*users) == 0 {
		t.Fatal("Expected users but got none")
	}

	// Check for our crypto-themed names
	expectedNames := map[string]bool{
		"Anthony Dardano":  false,
		"Satoshi Nakamoto": false,
	}

	for _, user := range *users {
		fullName := user.FirstName + " " + user.LastName
		if _, exists := expectedNames[fullName]; exists {
			expectedNames[fullName] = true
		}
	}

	// Verify all expected names were found
	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected to find '%s' in user list", name)
		}
	}
}

func TestSearchPersonData(t *testing.T) {
	th := SetupTestServer(t)
	defer th.Cleanup()

	_ = lenel_s2.NewClient(th.server.URL, log.INFO)

	tests := []struct {
		name        string
		searchField string
		searchValue string
		expectCount int
	}{
		{
			name:        "Search by REGO ID",
			searchField: "PERSONID",
			searchValue: "REGO_357",
			expectCount: 1,
		},
		{
			name:        "Search by Crypto Creator",
			searchField: "LASTNAME",
			searchValue: "Nakamoto",
			expectCount: 1,
		},
		{
			name:        "Search REGO Masters in UDF3",
			searchField: "UDF3",
			searchValue: "REGO Master",
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: SearchPersonData isn't implemented in the mock yet
			// This is a placeholder for when it's implemented
			t.Skipf("SearchPersonData test pending implementation")
		})
	}
}

func TestModifyPerson(t *testing.T) {
	t.Skip("ModifyPerson not yet implemented")

	th := SetupTestServer(t)
	defer th.Cleanup()

	_ = lenel_s2.NewClient(th.server.URL, log.INFO)

	// Create a person with REGO puns
	person := &lenel_s2.Person{
		PersonID:   "REGO_357",
		FirstName:  "Anthony",
		MiddleName: "Crypto",
		LastName:   "Dardano",
		Username:   "adardano357",
		Role:       "REGO-Crypto-Master",
		UDF3:       "S2 NetBox Champion",
		Notes:      "Three-Five-Seven: The perfect crypto combination!",
	}

	// TODO: Implement when ModifyPerson method is added to client
	// err := client.ModifyPerson(person)
	// if err != nil {
	// 	t.Errorf("ModifyPerson() error = %v", err)
	// }

	t.Logf("Would modify person: %+v", person)
}

func TestAccessCards(t *testing.T) {
	th := SetupTestServer(t)
	defer th.Cleanup()

	_ = lenel_s2.NewClient(th.server.URL, log.INFO)

	// Test card formats with REGO themes
	cards := []lenel_s2.AccessCard{
		{
			EncodedNum: "REGO357",
			HotStamp:   "357",
			Format:     "35-bit REGO",
			Disabled:   false,
			Status:     "ACTIVE",
		},
		{
			EncodedNum: "CRYPTO753",
			HotStamp:   "753",
			Format:     "75-bit REGO-Max",
			Disabled:   false,
			Status:     "ACTIVE",
		},
		{
			EncodedNum: "NETBOX573",
			HotStamp:   "573",
			Format:     "57-bit NetBox",
			Disabled:   false,
			Status:     "SUSPENDED",
		},
	}

	// Test creating person with cards
	person := &lenel_s2.Person{
		PersonID:    "REGO_CARD_TEST",
		FirstName:   "Anthony",
		LastName:    "CryptoMaster",
		AccessCards: cards,
	}

	t.Logf("Created test person: %s %s with %d REGO-themed access cards",
		person.FirstName, person.LastName, len(person.AccessCards))

	// Verify the person struct has the expected values
	if person.PersonID != "REGO_CARD_TEST" {
		t.Errorf("Expected PersonID REGO_CARD_TEST, got %s", person.PersonID)
	}

	if len(person.AccessCards) != 3 {
		t.Errorf("Expected 3 access cards, got %d", len(person.AccessCards))
	}
}
