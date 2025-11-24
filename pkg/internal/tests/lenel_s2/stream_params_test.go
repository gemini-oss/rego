package lenel_s2_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/gemini-oss/rego/pkg/lenel_s2"
)

func TestStreamEventsBuilder(t *testing.T) {
	tests := []struct {
		name        string
		buildFunc   func() *lenel_s2.StreamEventsParams
		wantTags    []string
		wantFilters map[string][]string
	}{
		{
			name: "AccessGranted with default fields",
			buildFunc: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.AccessGranted).
					Build()
			},
			wantTags: []string{"PERSONID", "PERSONNAME", "PORTALKEY", "PORTALNAME",
				"READERKEY", "READER2KEY", "RDRNAME", "ACNAME", "ACNUM",
				"NDT", "NODEADDRESS", "NODENAME", "NODEUNIQUE",
				"PARTNAME", "PARTITIONKEY", "CDT"},
			wantFilters: map[string][]string{},
		},
		{
			name: "InvalidAccess with person filter for Anthony",
			buildFunc: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.InvalidAccess).
					FilterByPersonName("Anthony Dardano", "A. Dardano", "REGO Master").
					Build()
			},
			wantTags: []string{"DETAIL", "PERSONID", "PERSONNAME", "PORTALKEY",
				"PORTALNAME", "READERKEY", "READER2KEY", "RDRNAME",
				"ACNAME", "ACNUM", "NDT", "NODEADDRESS", "NODENAME",
				"NODEUNIQUE", "PARTNAME", "PARTITIONKEY", "CDT"},
			wantFilters: map[string][]string{
				"PERSONNAME": {"Anthony Dardano", "A. Dardano", "REGO Master"},
			},
		},
		{
			name: "Multiple event types with portal filters",
			buildFunc: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.PortalHeldOpen).
					WithEventType(lenel_s2.EventTypes.PortalForcedOpen).
					WithEventType(lenel_s2.EventTypes.PortalRestored).
					FilterByPortalName("REGO-Portal-5", "NetBox-Portal-7", "S2-Portal-3").
					Build()
			},
			wantTags: []string{"PORTALKEY", "PORTALNAME", "NDT", "NODEADDRESS",
				"NODENAME", "NODEUNIQUE", "PARTNAME", "PARTITIONKEY", "CDT"},
			wantFilters: map[string][]string{
				"PORTALNAME": {"REGO-Portal-5", "NetBox-Portal-7", "S2-Portal-3"},
			},
		},
		{
			name: "NetworkNodeEvents with ShobuPrime and Apollo filters",
			buildFunc: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.NetworkNodeTimeout).
					WithEventType(lenel_s2.EventTypes.NetworkNodeRestored).
					WithFilter(lenel_s2.EventTags.NodeName, "Satoshi-Node-5", "Nakamoto-Master-7", "REGO-Node-3").
					Build()
			},
			wantTags: []string{"NDT", "NODEADDRESS", "NODENAME", "NODEUNIQUE",
				"PARTNAME", "PARTITIONKEY", "CDT"},
			wantFilters: map[string][]string{
				"NODENAME": {"Satoshi-Node-5", "Nakamoto-Master-7", "REGO-Node-3"},
			},
		},
		{
			name: "Custom fields with complex filters",
			buildFunc: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithField(lenel_s2.EventTags.PersonName).
					WithField(lenel_s2.EventTags.DescName).
					WithField(lenel_s2.EventTags.CDT).
					WithField(lenel_s2.EventTags.PartitionName).
					WithFilter(lenel_s2.EventTags.PersonName, "Anthony Dardano", "Satoshi Nakamoto", "Bitcoin").
					WithFilter(lenel_s2.EventTags.DescName, "Access granted", "Access denied", "Invalid badge").
					WithFilter(lenel_s2.EventTags.PartitionName, "Office", "REGO-5", "Crypto-7").
					Build()
			},
			wantTags: []string{"PERSONNAME", "DESCNAME", "CDT", "PARTNAME"},
			wantFilters: map[string][]string{
				"PERSONNAME": {"Anthony Dardano", "Satoshi Nakamoto", "Bitcoin"},
				"DESCNAME":   {"Access granted", "Access denied", "Invalid badge"},
				"PARTNAME":   {"Office", "REGO-5", "Crypto-7"},
			},
		},
		{
			name: "ElevatorEvents with mixed person names",
			buildFunc: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.ElevatorAccessGranted).
					WithEventType(lenel_s2.EventTypes.ElevatorAccessDenied).
					FilterByPersonName("Anthony Dardano", "Tony D", "Satoshi-757", "Bitcoin-Master").
					Build()
			},
			wantTags: []string{"PERSONID", "PERSONNAME", "RDRNAME", "UCBITLENGTH",
				"NDT", "NODEADDRESS", "NODENAME", "PARTNAME", "PARTITIONKEY", "CDT", "DETAIL"},
			wantFilters: map[string][]string{
				"PERSONNAME": {"Anthony Dardano", "Tony D", "Satoshi-757", "Bitcoin-Master"},
			},
		},
		{
			name: "AlarmEvents with panel filters",
			buildFunc: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.AlarmPanelArmed).
					WithEventType(lenel_s2.EventTypes.AlarmPanelDisarmed).
					WithFilter(lenel_s2.EventTags.AlarmPanelName, "REGO-Alarm-5", "Satoshi-Panel", "Crypto-Security-7").
					Build()
			},
			wantTags: []string{"ALARMPANELNAME", "PARTNAME", "PARTITIONKEY", "CDT"},
			wantFilters: map[string][]string{
				"ALARMPANELNAME": {"REGO-Alarm-5", "Satoshi-Panel", "Crypto-Security-7"},
			},
		},
		{
			name: "IntrusionPanelEvents with zone filters",
			buildFunc: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.IntrusionPanelAlarm).
					WithEventType(lenel_s2.EventTypes.IntrusionPanelZoneTrouble).
					WithFilter(lenel_s2.EventTags.IPanelZone, "Zone-357", "Satoshi-Z5", "Bitcoin-Z7").
					WithFilter(lenel_s2.EventTags.IPanelName, "REGO-Intrusion", "Prime-Panel").
					Build()
			},
			wantTags: []string{"IPANELAREA", "IPANELNAME", "IPANELZONE", "PARTNAME", "PARTITIONKEY", "CDT"},
			wantFilters: map[string][]string{
				"IPANELZONE": {"Zone-357", "Satoshi-Z5", "Bitcoin-Z7"},
				"IPANELNAME": {"REGO-Intrusion", "Prime-Panel"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.buildFunc()

			// Marshal to XML to verify structure
			xmlData, err := xml.Marshal(params)
			if err != nil {
				t.Fatalf("Failed to marshal params: %v", err)
			}

			xmlStr := string(xmlData)
			t.Logf("Generated XML:\n%s", xmlStr)

			// Verify all expected tags are present
			for _, tag := range tt.wantTags {
				if !strings.Contains(xmlStr, "<"+tag) {
					t.Errorf("Expected tag <%s> not found in XML", tag)
				}
			}

			// Verify filters
			for tag, expectedFilters := range tt.wantFilters {
				for _, filter := range expectedFilters {
					if !strings.Contains(xmlStr, filter) {
						t.Errorf("Expected filter value '%s' for tag %s not found in XML", filter, tag)
					}
				}
			}
		})
	}
}

func TestStreamEventsParamsMarshal(t *testing.T) {
	// Test specific XML marshaling scenarios
	tests := []struct {
		name      string
		params    *lenel_s2.StreamEventsParams
		wantXML   string
		checkFunc func(t *testing.T, xmlStr string)
	}{
		{
			name: "Empty tag (no filter)",
			params: lenel_s2.NewStreamEventsBuilder().
				WithField(lenel_s2.EventTags.PersonName).
				WithField(lenel_s2.EventTags.CDT).
				Build(),
			checkFunc: func(t *testing.T, xmlStr string) {
				// Should have empty PERSONNAME and CDT tags
				if !strings.Contains(xmlStr, "<PERSONNAME/>") && !strings.Contains(xmlStr, "<PERSONNAME></PERSONNAME>") {
					t.Error("Expected empty PERSONNAME tag")
				}
				if !strings.Contains(xmlStr, "<CDT/>") && !strings.Contains(xmlStr, "<CDT></CDT>") {
					t.Error("Expected empty CDT tag")
				}
			},
		},
		{
			name: "Tag with single filter",
			params: lenel_s2.NewStreamEventsBuilder().
				WithField(lenel_s2.EventTags.PersonName).
				WithFilter(lenel_s2.EventTags.PersonName, "Anthony Dardano").
				Build(),
			checkFunc: func(t *testing.T, xmlStr string) {
				if !strings.Contains(xmlStr, "<PERSONNAME>") {
					t.Error("Expected PERSONNAME tag with content")
				}
				if !strings.Contains(xmlStr, "<FILTER>Anthony Dardano</FILTER>") {
					t.Error("Expected FILTER with 'Anthony Dardano'")
				}
			},
		},
		{
			name: "Tag with multiple filters",
			params: lenel_s2.NewStreamEventsBuilder().
				WithField(lenel_s2.EventTags.PortalName).
				WithFilter(lenel_s2.EventTags.PortalName, "REGO-Portal-5", "Satoshi-Gate", "Bitcoin-Entry-7").
				Build(),
			checkFunc: func(t *testing.T, xmlStr string) {
				filters := []string{"REGO-Portal-5", "Satoshi-Gate", "Bitcoin-Entry-7"}
				for _, filter := range filters {
					if !strings.Contains(xmlStr, "<FILTER>"+filter+"</FILTER>") {
						t.Errorf("Expected FILTER with '%s'", filter)
					}
				}
			},
		},
		{
			name: "Mixed event types with filters",
			params: lenel_s2.NewStreamEventsBuilder().
				WithEventType(lenel_s2.EventTypes.LoggedIn).
				WithEventType(lenel_s2.EventTypes.LoggedOut).
				FilterByPersonName("Anthony Dardano", "Satoshi Nakamoto").
				WithFilter(lenel_s2.EventTags.LoginAddress, "10.5.7.35", "192.168.3.57").
				Build(),
			checkFunc: func(t *testing.T, xmlStr string) {
				// Check for person filters
				if !strings.Contains(xmlStr, "Anthony Dardano") {
					t.Error("Expected 'Anthony Dardano' in filters")
				}
				if !strings.Contains(xmlStr, "Satoshi Nakamoto") {
					t.Error("Expected 'Satoshi Nakamoto' in filters")
				}
				// Check for login address filters
				if !strings.Contains(xmlStr, "10.5.7.35") {
					t.Error("Expected '10.5.7.35' in filters")
				}
				if !strings.Contains(xmlStr, "192.168.3.57") {
					t.Error("Expected '192.168.3.57' in filters")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlData, err := xml.Marshal(tt.params)
			if err != nil {
				t.Fatalf("Failed to marshal params: %v", err)
			}

			xmlStr := string(xmlData)
			t.Logf("Generated XML:\n%s", xmlStr)

			if tt.checkFunc != nil {
				tt.checkFunc(t, xmlStr)
			}
		})
	}
}

func TestEventTypeFieldUniqueness(t *testing.T) {
	// Test that multiple event types properly merge their required fields
	builder := lenel_s2.NewStreamEventsBuilder()

	// Add multiple event types that share some fields
	builder.WithEventType(lenel_s2.EventTypes.AccessGranted)
	builder.WithEventType(lenel_s2.EventTypes.InvalidAccess)
	builder.WithEventType(lenel_s2.EventTypes.NetworkNodeTimeout)

	params := builder.Build()
	xmlData, err := xml.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}

	xmlStr := string(xmlData)

	// Count occurrences of PERSONNAME - should only appear once despite being in multiple event types
	personNameCount := strings.Count(xmlStr, "<PERSONNAME")
	if personNameCount != 1 {
		t.Errorf("Expected PERSONNAME to appear once, but found %d occurrences", personNameCount)
	}

	// Verify DETAIL is included (only in InvalidAccess)
	if !strings.Contains(xmlStr, "<DETAIL") {
		t.Error("Expected DETAIL tag from InvalidAccess event type")
	}
}

func TestComplexFilterScenarios(t *testing.T) {
	// Test complex real-world filtering scenarios
	scenarios := []struct {
		name     string
		scenario string
		builder  func() *lenel_s2.StreamEventsParams
	}{
		{
			name:     "Monitor VIP access for multiple users",
			scenario: "Track Anthony, Satoshi, and Bitcoin accessing VIP areas",
			builder: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.AccessGranted).
					WithEventType(lenel_s2.EventTypes.InvalidAccess).
					FilterByPersonName("Anthony Dardano", "Satoshi Nakamoto", "Bitcoin").
					FilterByPortalName("VIP-Portal-5", "Executive-Suite-7", "REGO-Penthouse-3").
					Build()
			},
		},
		{
			name:     "Security breach monitoring",
			scenario: "Monitor all forced/held open events on critical portals",
			builder: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.PortalForcedOpen).
					WithEventType(lenel_s2.EventTypes.PortalHeldOpen).
					FilterByPortalName("Vault-Door-357", "Server-Room-5", "Satoshi-Secure", "Crypto-DataCenter-7").
					Build()
			},
		},
		{
			name:     "After hours access monitoring",
			scenario: "Track all access between 7PM and 5AM",
			builder: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.AccessGranted).
					WithEventType(lenel_s2.EventTypes.InvalidAccess).
					WithEventType(lenel_s2.EventTypes.AccessNotCompleted).
					FilterByDescName("After Hours Access", "Night Shift Entry", "Emergency Access").
					Build()
			},
		},
		{
			name:     "Multi-building threat monitoring",
			scenario: "Monitor threat level changes across REGO, Satoshi, and Bitcoin buildings",
			builder: func() *lenel_s2.StreamEventsParams {
				return lenel_s2.NewStreamEventsBuilder().
					WithEventType(lenel_s2.EventTypes.ThreatLevelSet).
					WithEventType(lenel_s2.EventTypes.ThreatLevelSetAPI).
					WithEventType(lenel_s2.EventTypes.ThreatLevelSetALM).
					WithFilter(lenel_s2.EventTags.LocationName, "REGO-Building-5", "Satoshi-Tower", "Bitcoin-Complex-7").
					Build()
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			t.Logf("Scenario: %s", s.scenario)

			params := s.builder()
			xmlData, err := xml.Marshal(params)
			if err != nil {
				t.Fatalf("Failed to marshal params: %v", err)
			}

			xmlStr := string(xmlData)

			// Verify it produces valid XML
			var testParams lenel_s2.StreamEventsParams
			if err := xml.Unmarshal(xmlData, &testParams); err != nil {
				t.Errorf("Generated XML is not valid: %v", err)
			}

			// Log the size for performance considerations
			t.Logf("XML size: %d bytes", len(xmlStr))
		})
	}
}
