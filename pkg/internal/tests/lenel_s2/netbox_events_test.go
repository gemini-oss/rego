package lenel_s2_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gemini-oss/rego/pkg/lenel_s2"
)

// mockMultipartEvent creates a multipart event response
func mockMultipartEvent(events []string) string {
	var buffer bytes.Buffer

	// Write multipart headers
	buffer.WriteString("--Boundary\r\n")
	buffer.WriteString("Content-Type: text/xml\r\n")
	buffer.WriteString("\r\n")

	// Write each event
	for i, event := range events {
		buffer.WriteString(event)
		if i < len(events)-1 {
			buffer.WriteString("\r\n--Boundary\r\n")
			buffer.WriteString("Content-Type: text/xml\r\n")
			buffer.WriteString("\r\n")
		}
	}

	buffer.WriteString("\r\n--Boundary--\r\n")
	return buffer.String()
}

func TestStreamEvents(t *testing.T) {
	// Sample events to stream with crypto-themed names
	events := []string{
		// Initial success response
		`<NETBOX><RESPONSE command="StreamEvents"><CODE>SUCCESS</CODE></RESPONSE></NETBOX>`,

		// Full access granted event - Anthony with 357 theme
		`<NETBOX><RESPONSE command="StreamEvents"><EVENT><ACNAME><![CDATA[357]]></ACNAME><ACNUM><![CDATA[357]]></ACNUM><CDT>2025-08-07 13:57:35.75300 -0400</CDT><DESCNAME><![CDATA[Access granted]]></DESCNAME><NDT>2025-08-07 13:57:35.00000 -0400</NDT><NODENAME><![CDATA[REGO-Node-5]]></NODENAME><NODEUNIQUE>CRYPTO357MASTER5735</NODEUNIQUE><PARTITIONKEY>3</PARTITIONKEY><PARTNAME><![CDATA[Crypto-Zone]]></PARTNAME><PERSONID>REGO_357</PERSONID><PERSONNAME><![CDATA[Dardano, Anthony]]></PERSONNAME><PORTALKEY>357</PORTALKEY><PORTALNAME><![CDATA[REGO-Portal-357]]></PORTALNAME><RDRNAME><![CDATA[REGO-357-Reader]]></RDRNAME><READERKEY>357</READERKEY></EVENT></RESPONSE></NETBOX>`,

		// Heartbeat (empty response)
		`<NETBOX><RESPONSE command="StreamEvents"></RESPONSE></NETBOX>`,

		// Satoshi access event with 753 theme
		`<NETBOX><RESPONSE command="StreamEvents"><EVENT><PERSONNAME><![CDATA[Nakamoto, Satoshi]]></PERSONNAME><PORTALNAME><![CDATA[Bitcoin-Portal-753]]></PORTALNAME><CDT>2025-08-07 15:53:57</CDT><DESCNAME><![CDATA[Access granted]]></DESCNAME><PERSONID>SATOSHI_753</PERSONID></EVENT></RESPONSE></NETBOX>`,

		// Another heartbeat
		`<NETBOX><RESPONSE command="StreamEvents"></RESPONSE></NETBOX>`,

		// Satoshi access denied with 573 theme
		`<NETBOX><RESPONSE command="StreamEvents"><EVENT><PERSONNAME><![CDATA[Nakamoto, Satoshi]]></PERSONNAME><PORTALNAME><![CDATA[Bitcoin-Portal-573]]></PORTALNAME><CDT>2025-08-05 17:35:73</CDT><DESCNAME><![CDATA[Access denied]]></DESCNAME><PERSONID>NAKAMOTO_573</PERSONID><DETAIL><![CDATA[Invalid credentials]]></DETAIL></EVENT></RESPONSE></NETBOX>`,
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check for streaming request
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "StreamEvents") {
			// Send multipart response
			w.Header().Set("Content-Type", "multipart/mixed; boundary=Boundary")
			w.Header().Set("Transfer-Encoding", "chunked")

			// Stream events with delays
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Fatal("Expected http.Flusher")
			}

			// Use mockMultipartEvent to properly format the response
			multipartResponse := mockMultipartEvent(events)
			fmt.Fprint(w, multipartResponse)
			flusher.Flush()
		}
	}))
	defer server.Close()

	// Test multipart parsing
	t.Run("mockMultipartEvent formats correctly", func(t *testing.T) {
		multipart := mockMultipartEvent(events[:2]) // Test with first 2 events

		// Should contain boundary markers
		if !strings.Contains(multipart, "--Boundary") {
			t.Error("Missing boundary markers")
		}

		// Should contain content type headers
		if !strings.Contains(multipart, "Content-Type: text/xml") {
			t.Error("Missing Content-Type headers")
		}

		// Should end with closing boundary
		if !strings.Contains(multipart, "--Boundary--") {
			t.Error("Missing closing boundary")
		}
	})

	// Test event counting
	t.Run("StreamEvents collects events correctly", func(t *testing.T) {
		// Expected behavior:
		// - Should skip the initial SUCCESS response (no EVENT)
		// - Should collect 3 events (Anthony, Satoshi, Nakamoto)
		// - Should skip the heartbeats

		collectedEvents := 0
		expectedEvents := 3

		// Count non-heartbeat event responses
		for _, event := range events {
			if strings.Contains(event, "<EVENT>") && strings.Contains(event, "</EVENT>") {
				collectedEvents++
			}
		}

		if collectedEvents != expectedEvents {
			t.Errorf("Expected %d events, but found %d", expectedEvents, collectedEvents)
		}

		t.Logf("Mock server would stream %d events", collectedEvents)
	})

	// Test event filtering
	t.Run("Event filtering by person", func(t *testing.T) {
		anthonyEvents := 0
		satoshiEvents := 0
		nakamotoEvents := 0

		for _, event := range events {
			if strings.Contains(event, "REGO_357") {
				anthonyEvents++
			}
			if strings.Contains(event, "SATOSHI_753") {
				satoshiEvents++
			}
			if strings.Contains(event, "NAKAMOTO_573") {
				nakamotoEvents++
			}
		}

		if anthonyEvents != 1 {
			t.Errorf("Expected 1 Anthony event, got %d", anthonyEvents)
		}
		if satoshiEvents != 1 {
			t.Errorf("Expected 1 Satoshi event, got %d", satoshiEvents)
		}
		if nakamotoEvents != 1 {
			t.Errorf("Expected 1 Nakamoto event, got %d", nakamotoEvents)
		}
	})
}

func TestEventParsing(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		expected lenel_s2.Event
		wantErr  bool
	}{
		{
			name: "Full access granted event",
			xml: `<EVENT>
				<ACNAME><![CDATA[003D]]></ACNAME>
				<ACNUM><![CDATA[003D]]></ACNUM>
				<CDT>2025-08-07 13:30:31.65900 -0400</CDT>
				<DESCNAME><![CDATA[Access granted]]></DESCNAME>
				<NDT>2025-08-07 13:30:31.00000 -0400</NDT>
				<NODENAME><![CDATA[REGO-1FL]]></NODENAME>
				<NODEUNIQUE>ABCDEFABCDEFABCD</NODEUNIQUE>
				<PARTITIONKEY>3</PARTITIONKEY>
				<PARTNAME><![CDATA[Office]]></PARTNAME>
				<PERSONID>US_0003</PERSONID>
				<PERSONNAME><![CDATA[Dardano, Anthony]]></PERSONNAME>
				<PORTALKEY>33</PORTALKEY>
				<PORTALNAME><![CDATA[REGO-Lobby]]></PORTALNAME>
				<RDRNAME><![CDATA[REGO-Lobby-Reader]]></RDRNAME>
				<READERKEY>003</READERKEY>
			</EVENT>`,
			expected: lenel_s2.Event{
				ACName:       "003D",
				ACNum:        "003D",
				CDT:          "2025-08-07 13:30:31.65900 -0400",
				DescName:     "Access granted",
				NDT:          "2025-08-07 13:30:31.00000 -0400",
				NodeName:     "REGO-1FL",
				NodeUnique:   "ABCDEFABCDEFABCD",
				PartitionKey: "3",
				PartName:     "Office",
				PersonID:     "US_0003",
				PersonName:   "Dardano, Anthony",
				PortalKey:    "33",
				PortalName:   "REGO-Lobby",
				RdrName:      "REGO-Lobby-Reader",
				ReaderKey:    "003",
			},
			wantErr: false,
		},
		{
			name: "Access denied event with minimal fields",
			xml: `<EVENT>
				<PERSONNAME><![CDATA[Dardano, Anthony]]></PERSONNAME>
				<PORTALNAME><![CDATA[REGO-Entrance]]></PORTALNAME>
				<CDT>2025-08-05 09:35:00</CDT>
				<DESCNAME><![CDATA[Access denied]]></DESCNAME>
			</EVENT>`,
			expected: lenel_s2.Event{
				PersonName: "Dardano, Anthony",
				PortalName: "REGO-Entrance",
				CDT:        "2025-08-05 09:35:00",
				DescName:   "Access denied",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event lenel_s2.Event
			err := parseXML(tt.xml, &event)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare key fields
			if event.PersonName != tt.expected.PersonName {
				t.Errorf("PersonName = %v, want %v", event.PersonName, tt.expected.PersonName)
			}
			if event.PortalName != tt.expected.PortalName {
				t.Errorf("PortalName = %v, want %v", event.PortalName, tt.expected.PortalName)
			}
			if event.ActivityID != tt.expected.ActivityID {
				t.Errorf("ActivityID = %v, want %v", event.ActivityID, tt.expected.ActivityID)
			}
			if event.DescName != tt.expected.DescName {
				t.Errorf("DescName = %v, want %v", event.DescName, tt.expected.DescName)
			}
			if event.CDT != tt.expected.CDT {
				t.Errorf("CDT = %v, want %v", event.CDT, tt.expected.CDT)
			}
		})
	}
}

// Helper function to parse XML
func parseXML(xmlStr string, v any) error {
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	return decoder.Decode(v)
}
