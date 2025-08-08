package lenel_s2_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/lenel_s2"
)

// TestMode defines the test mode for integration tests
type TestMode string

const (
	TestModeLive    TestMode = "live"
	TestModeFixture TestMode = "fixture"
	TestModeRecord  TestMode = "record"
)

// IntegrationTest manages fixture-based testing
type IntegrationTest struct {
	T          *testing.T
	Mode       TestMode
	FixtureDir string
	Client     *lenel_s2.Client
	mockServer *httptest.Server
	recordings map[string][]byte // For record mode
}

// NewIntegrationTest creates a new integration test
func NewIntegrationTest(t *testing.T, fixtureName string) *IntegrationTest {
	mode := TestModeFixture // Default to fixture mode
	if modeEnv := os.Getenv("REGO_TEST_MODE"); modeEnv != "" {
		mode = TestMode(modeEnv)
	}

	return &IntegrationTest{
		T:          t,
		Mode:       mode,
		FixtureDir: filepath.Join("testdata", "fixtures", fixtureName),
		recordings: make(map[string][]byte),
	}
}

// Setup initializes the test client based on mode
func (it *IntegrationTest) Setup() error {
	switch it.Mode {
	case TestModeLive:
		it.T.Log("Running in LIVE mode - connecting to real S2 system")
		url := os.Getenv("S2_URL")
		if url == "" {
			return fmt.Errorf("S2_URL environment variable not set for live testing")
		}
		it.Client = lenel_s2.NewClient(url, log.INFO)
		return nil

	case TestModeRecord:
		it.T.Log("Running in RECORD mode - recording API responses")
		url := os.Getenv("S2_URL")
		if url == "" {
			return fmt.Errorf("S2_URL environment variable not set for recording")
		}

		// Create a recording proxy server
		it.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqBody, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

			// Forward to real S2 server
			proxyReq, _ := http.NewRequest(r.Method, url+r.URL.Path, r.Body)
			proxyReq.Header = r.Header

			resp, err := http.DefaultClient.Do(proxyReq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			respBody, _ := io.ReadAll(resp.Body)

			// Save the response
			var cmd lenel_s2.NetboxCommand
			xml.Unmarshal(reqBody, &cmd)
			it.SaveXMLFixture(cmd.Command.Name, respBody)

			// Forward response
			for k, v := range resp.Header {
				w.Header()[k] = v
			}
			w.WriteHeader(resp.StatusCode)
			w.Write(respBody)
		}))

		it.Client = lenel_s2.NewClient(it.mockServer.URL, log.INFO)
		return nil

	default: // TestModeFixture
		it.T.Log("Running in FIXTURE mode - using saved responses")
		// Set up mock server with fixture data
		it.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse request to determine which fixture to load
			reqBody, _ := io.ReadAll(r.Body)
			var cmd lenel_s2.NetboxCommand
			if err := xml.Unmarshal(reqBody, &cmd); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			// Load fixture
			fixture, err := it.LoadXMLFixture(cmd.Command.Name)
			if err != nil {
				// Fall back to default mock behavior
				handleMockRequest(it.T, w, r, &cmd, string(reqBody))
				return
			}

			w.Header().Set("Content-Type", "text/xml")
			w.Write(fixture)
		}))

		it.Client = lenel_s2.NewClient(it.mockServer.URL, log.INFO)
		return nil
	}
}

// Cleanup cleans up test resources
func (it *IntegrationTest) Cleanup() {
	if it.mockServer != nil {
		it.mockServer.Close()
	}

	// Save any recorded fixtures
	if it.Mode == TestModeRecord && len(it.recordings) > 0 {
		for name, data := range it.recordings {
			it.SaveXMLFixture(name, data)
		}
	}
}

// SaveXMLFixture saves XML response data to a fixture file
func (it *IntegrationTest) SaveXMLFixture(name string, xmlData []byte) error {
	if it.Mode == TestModeRecord {
		it.recordings[name] = xmlData
	}

	fixturePath := filepath.Join(it.FixtureDir, name+".xml")
	if err := os.MkdirAll(it.FixtureDir, 0755); err != nil {
		return err
	}

	// Pretty print XML for readability
	var formatted bytes.Buffer
	formatted.WriteString(xml.Header)

	// Parse and re-encode for pretty printing
	var temp interface{}
	if err := xml.Unmarshal(xmlData, &temp); err == nil {
		encoder := xml.NewEncoder(&formatted)
		encoder.Indent("", "  ")
		if err := encoder.Encode(temp); err == nil {
			return os.WriteFile(fixturePath, formatted.Bytes(), 0644)
		}
	}

	// If pretty print fails, just save raw
	return os.WriteFile(fixturePath, xmlData, 0644)
}

// LoadXMLFixture loads XML response data from a fixture file
func (it *IntegrationTest) LoadXMLFixture(name string) ([]byte, error) {
	fixturePath := filepath.Join(it.FixtureDir, name+".xml")
	return os.ReadFile(fixturePath)
}

// handleMockRequest routes mock requests to appropriate handlers
func handleMockRequest(t *testing.T, w http.ResponseWriter, r *http.Request, cmd *lenel_s2.NetboxCommand, bodyStr string) {
	switch cmd.Command.Name {
	case "Login":
		handleLogin(w, r)
	case "Logout":
		handleLogout(w, r)
	case "GetPerson":
		handleGetPerson(w, r, cmd, bodyStr)
	case "SearchPersonData":
		handleSearchPersonData(w, r, cmd)
	case "ModifyPerson":
		handleModifyPerson(w, r, cmd)
	case "GetAccessHistory":
		handleGetAccessHistory(w, r, cmd)
	case "GetUDFLists":
		handleGetUDFLists(w, r)
	case "GetCardFormats":
		handleGetCardFormats(w, r)
	default:
		t.Logf("Unhandled command: %s", cmd.Command.Name)
		sendErrorResponse(w, "FAIL", fmt.Sprintf("Unknown command: %s", cmd.Command.Name))
	}
}

// TestHelpers provides common test setup functionality
type TestHelpers struct {
	server *httptest.Server
	client *lenel_s2.Client
}

// SetupTestServer creates a mock S2 server for testing
func SetupTestServer(t *testing.T) *TestHelpers {
	th := &TestHelpers{}

	// Create mock server
	th.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the full body first
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()

		// Parse the request
		var cmd lenel_s2.NetboxCommand
		if err := xml.Unmarshal(bodyBytes, &cmd); err != nil {
			t.Logf("Failed to decode request: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Route based on command
		switch cmd.Command.Name {
		case "Login":
			handleLogin(w, r)
		case "Logout":
			handleLogout(w, r)
		case "GetPerson":
			t.Logf("GetPerson request: %+v", cmd)
			t.Logf("GetPerson body: %s", string(bodyBytes))
			handleGetPerson(w, r, &cmd, string(bodyBytes))
		case "SearchPersonData":
			handleSearchPersonData(w, r, &cmd)
		case "ModifyPerson":
			handleModifyPerson(w, r, &cmd)
		case "GetAccessHistory":
			handleGetAccessHistory(w, r, &cmd)
		case "GetUDFLists":
			handleGetUDFLists(w, r)
		case "GetCardFormats":
			handleGetCardFormats(w, r)
		default:
			t.Logf("Unhandled command: %s", cmd.Command.Name)
			sendErrorResponse(w, "FAIL", fmt.Sprintf("Unknown command: %s", cmd.Command.Name))
		}
	}))

	// Set test environment variables
	os.Setenv("S2_URL", th.server.URL)
	os.Setenv("S2_USERNAME", "anthony.nakamoto")
	os.Setenv("S2_PASSWORD", "CryptoMaster357")
	os.Setenv("REGO_ENCRYPTION_KEY", "REGOlithROCKS357NetBoxS2Gateway!") // 32 bytes

	return th
}

// Cleanup closes the test server and cleans up environment
func (th *TestHelpers) Cleanup() {
	if th.server != nil {
		th.server.Close()
	}
	os.Unsetenv("S2_URL")
	os.Unsetenv("S2_USERNAME")
	os.Unsetenv("S2_PASSWORD")
	os.Unsetenv("REGO_ENCRYPTION_KEY")
}

// Mock response handlers

func handleLogin(w http.ResponseWriter, _ *http.Request) {
	response := `<NETBOX sessionid="REGO-SESSION-357">
		<RESPONSE>
			<CODE>SUCCESS</CODE>
		</RESPONSE>
	</NETBOX>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

func handleLogout(w http.ResponseWriter, _ *http.Request) {
	response := `<NETBOX>
		<RESPONSE>
			<CODE>SUCCESS</CODE>
		</RESPONSE>
	</NETBOX>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

func handleGetPerson(w http.ResponseWriter, _ *http.Request, _ *lenel_s2.NetboxCommand, bodyStr string) {
	// Default to Anthony
	personID := "REGO_357"
	firstName := "Anthony"
	lastName := "Dardano"
	username := "adardano"
	role := "RegoMaster"
	udf3 := "REGO Master"

	// Check raw request body for person ID
	fmt.Printf("GetPerson body: %s\n", bodyStr)

	if strings.Contains(bodyStr, "SATOSHI_753") {
		personID = "SATOSHI_753"
		firstName = "Satoshi"
		lastName = "Nakamoto"
		username = "snakamoto"
		role = "BitcoinCreator"
		udf3 = "Bitcoin Creator"
	}

	response := fmt.Sprintf(`<NETBOX>
		<RESPONSE>
			<CODE>SUCCESS</CODE>
			<DETAILS>
				<PERSONID>%s</PERSONID>
				<FIRSTNAME>%s</FIRSTNAME>
				<LASTNAME>%s</LASTNAME>
				<USERNAME>%s</USERNAME>
				<ROLE>%s</ROLE>
				<UDF3>%s</UDF3>
				<UDF5>S2 Champion</UDF5>
				<UDF7>Triple Seven</UDF7>
				<DELETED>false</DELETED>
			</DETAILS>
		</RESPONSE>
	</NETBOX>`, personID, firstName, lastName, username, role, udf3)
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

func handleSearchPersonData(w http.ResponseWriter, _ *http.Request, _ *lenel_s2.NetboxCommand) {
	response := `<NETBOX>
		<RESPONSE>
			<CODE>SUCCESS</CODE>
			<DETAILS>
				<PEOPLE>
					<PERSON>
						<PERSONID>REGO_357</PERSONID>
						<FIRSTNAME>Anthony</FIRSTNAME>
						<LASTNAME>Dardano</LASTNAME>
						<UDF3>REGO Master</UDF3>
					</PERSON>
					<PERSON>
						<PERSONID>SATOSHI_753</PERSONID>
						<FIRSTNAME>Satoshi</FIRSTNAME>
						<LASTNAME>Nakamoto</LASTNAME>
						<UDF3>Bitcoin Creator</UDF3>
					</PERSON>
				</PEOPLE>
				<NEXTKEY>-1</NEXTKEY>
			</DETAILS>
		</RESPONSE>
	</NETBOX>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

func handleModifyPerson(w http.ResponseWriter, _ *http.Request, _ *lenel_s2.NetboxCommand) {
	response := `<NETBOX>
		<RESPONSE>
			<CODE>SUCCESS</CODE>
		</RESPONSE>
	</NETBOX>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

func handleGetAccessHistory(w http.ResponseWriter, _ *http.Request, _ *lenel_s2.NetboxCommand) {
	response := `<NETBOX>
		<RESPONSE>
			<CODE>SUCCESS</CODE>
			<DETAILS>
				<ACCESSES>
					<ACCESS>
						<LOGID>3573</LOGID>
						<PERSONID>REGO_357</PERSONID>
						<READER>REGO-Gate-5</READER>
						<DTTM>2025-03-05 07:35:57</DTTM>
						<TYPE>3</TYPE>
						<REASON>0</REASON>
						<PORTALKEY>357</PORTALKEY>
						<PORTALNAME>NetBox-Portal-Seven</PORTALNAME>
					</ACCESS>
					<ACCESS>
						<LOGID>5735</LOGID>
						<PERSONID>S2_753</PERSONID>
						<READER>S2-Reader-Crypto</READER>
						<DTTM>2025-05-07 15:35:57</DTTM>
						<TYPE>5</TYPE>
						<REASON>7</REASON>
						<READERKEY>753</READERKEY>
					</ACCESS>
				</ACCESSES>
				<NEXTLOGID>-1</NEXTLOGID>
			</DETAILS>
		</RESPONSE>
	</NETBOX>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

func handleGetUDFLists(w http.ResponseWriter, _ *http.Request) {
	response := `<NETBOX>
		<RESPONSE>
			<CODE>SUCCESS</CODE>
			<DETAILS>
				<UDFLISTS>
					<UDFLIST>
						<UDFLISTKEY>3</UDFLISTKEY>
						<NAME>REGO Department</NAME>
						<DESCRIPTION>Crypto Security Department</DESCRIPTION>
					</UDFLIST>
					<UDFLIST>
						<UDFLISTKEY>5</UDFLISTKEY>
						<NAME>Satoshi Building</NAME>
						<DESCRIPTION>Bitcoin Mining Tower</DESCRIPTION>
					</UDFLIST>
					<UDFLIST>
						<UDFLISTKEY>7</UDFLISTKEY>
						<NAME>NAKAMOTO Network</NAME>
						<DESCRIPTION>Bitcoin Smart Contract</DESCRIPTION>
					</UDFLIST>
				</UDFLISTS>
			</DETAILS>
		</RESPONSE>
	</NETBOX>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

func handleGetCardFormats(w http.ResponseWriter, _ *http.Request) {
	response := `<NETBOX>
		<RESPONSE>
			<CODE>SUCCESS</CODE>
			<DETAILS>
				<CARDFORMATS>
					<CARDFORMAT>35-bit REGO</CARDFORMAT>
					<CARDFORMAT>57-bit NetBox</CARDFORMAT>
					<CARDFORMAT>75-bit REGO-Max</CARDFORMAT>
					<CARDFORMAT>357-bit Crypto-Ultimate</CARDFORMAT>
				</CARDFORMATS>
			</DETAILS>
		</RESPONSE>
	</NETBOX>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

func sendErrorResponse(w http.ResponseWriter, code string, message string) {
	response := fmt.Sprintf(`<NETBOX>
		<RESPONSE>
			<CODE>%s</CODE>
			<DETAILS>
				<ERRMSG>%s</ERRMSG>
			</DETAILS>
		</RESPONSE>
	</NETBOX>`, code, message)
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, response)
}

// TestClientInitialization tests creating a new S2 client
func TestClientInitialization(t *testing.T) {
	th := SetupTestServer(t)
	defer th.Cleanup()

	client := lenel_s2.NewClient(th.server.URL, log.INFO)
	if client == nil {
		t.Fatal("Failed to create REGO client")
	}

	if client.Session == nil {
		t.Error("Client session is nil - NetBox not connected")
	}

	if client.Session.ID != "REGO-SESSION-357" {
		t.Errorf("Expected session ID 'REGO-SESSION-357', got '%s'", client.Session.ID)
	}
}

func TestNetboxEventResponse(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		wantErr bool
		check   func(t *testing.T, resp lenel_s2.NetboxEventResponse[lenel_s2.Event])
	}{
		{
			name: "StreamEvents response with event data",
			xml: `<NETBOX>
				<RESPONSE command="StreamEvents">
					<EVENT>
						<PERSONNAME><![CDATA[Anthony Dardano]]></PERSONNAME>
						<PORTALNAME><![CDATA[REGO-Portal-5]]></PORTALNAME>
						<CDT>2025-08-07 13:57:35</CDT>
						<DESCNAME><![CDATA[Access granted]]></DESCNAME>
					</EVENT>
				</RESPONSE>
			</NETBOX>`,
			wantErr: false,
			check: func(t *testing.T, resp lenel_s2.NetboxEventResponse[lenel_s2.Event]) {
				if resp.Response.Command != "StreamEvents" {
					t.Errorf("Expected command 'StreamEvents', got '%s'", resp.Response.Command)
				}
				if resp.Response.Event == nil {
					t.Fatal("Expected event data but got nil")
				}
				if resp.Response.Event.PersonName != "Anthony Dardano" {
					t.Errorf("Expected PersonName 'Anthony Dardano', got '%s'", resp.Response.Event.PersonName)
				}
				if resp.Response.Event.PortalName != "REGO-Portal-5" {
					t.Errorf("Expected PortalName 'REGO-Portal-5', got '%s'", resp.Response.Event.PortalName)
				}
			},
		},
		{
			name: "StreamEvents heartbeat (empty response)",
			xml: `<NETBOX>
				<RESPONSE command="StreamEvents">
				</RESPONSE>
			</NETBOX>`,
			wantErr: false,
			check: func(t *testing.T, resp lenel_s2.NetboxEventResponse[lenel_s2.Event]) {
				if resp.Response.Command != "StreamEvents" {
					t.Errorf("Expected command 'StreamEvents', got '%s'", resp.Response.Command)
				}
				if resp.Response.Event != nil {
					t.Error("Expected nil event for heartbeat")
				}
				if resp.Response.Code != "" {
					t.Errorf("Expected empty CODE for heartbeat, got '%s'", resp.Response.Code)
				}
			},
		},
		{
			name: "StreamEvents with SUCCESS code",
			xml: `<NETBOX>
				<RESPONSE command="StreamEvents">
					<CODE>SUCCESS</CODE>
				</RESPONSE>
			</NETBOX>`,
			wantErr: false,
			check: func(t *testing.T, resp lenel_s2.NetboxEventResponse[lenel_s2.Event]) {
				if resp.Response.Code != "SUCCESS" {
					t.Errorf("Expected CODE 'SUCCESS', got '%s'", resp.Response.Code)
				}
				if resp.Response.Event != nil {
					t.Error("Expected nil event for SUCCESS response")
				}
			},
		},
		{
			name: "StreamEvents with API error",
			xml: `<NETBOX>
				<RESPONSE command="StreamEvents">
					<APIERROR>7</APIERROR>
				</RESPONSE>
			</NETBOX>`,
			wantErr: false,
			check: func(t *testing.T, resp lenel_s2.NetboxEventResponse[lenel_s2.Event]) {
				if resp.Response.APIError != 7 {
					t.Errorf("Expected APIError 7, got %d", resp.Response.APIError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp lenel_s2.NetboxEventResponse[lenel_s2.Event]
			err := xml.Unmarshal([]byte(tt.xml), &resp)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}

func TestEventResponseUnmarshalXML(t *testing.T) {
	// Test the custom UnmarshalXML for EventResponse
	xmlData := `<RESPONSE command="StreamEvents">
		<EVENT>
			<PERSONID>REGO_757</PERSONID>
			<PERSONNAME><![CDATA[Anthony Dardano]]></PERSONNAME>
			<ACNAME><![CDATA[757]]></ACNAME>
			<ACNUM><![CDATA[757]]></ACNUM>
		</EVENT>
	</RESPONSE>`

	var resp lenel_s2.EventResponse[lenel_s2.Event]
	decoder := xml.NewDecoder(strings.NewReader(xmlData))

	// Find the start element
	for {
		tok, err := decoder.Token()
		if err != nil {
			t.Fatal("Failed to find start element:", err)
		}
		if start, ok := tok.(xml.StartElement); ok && start.Name.Local == "RESPONSE" {
			err = resp.UnmarshalXML(decoder, start)
			if err != nil {
				t.Fatal("UnmarshalXML failed:", err)
			}
			break
		}
	}

	if resp.Command != "StreamEvents" {
		t.Errorf("Expected command 'StreamEvents', got '%s'", resp.Command)
	}

	if resp.Event == nil {
		t.Fatal("Expected event data but got nil")
	}

	if resp.Event.PersonID != "REGO_757" {
		t.Errorf("Expected PersonID 'REGO_757', got '%s'", resp.Event.PersonID)
	}

	if resp.Event.ACNum != "757" {
		t.Errorf("Expected ACNum '757', got '%s'", resp.Event.ACNum)
	}
}

func TestMultipleEventTypes(t *testing.T) {
	// Test different event types that might come through StreamEvents
	eventTypes := []struct {
		name      string
		eventXML  string
		checkFunc func(t *testing.T, event lenel_s2.Event)
	}{
		{
			name: "Access Granted Event",
			eventXML: `<EVENT>
				<PERSONNAME><![CDATA[Anthony Dardano]]></PERSONNAME>
				<PORTALNAME><![CDATA[REGO-Portal-7]]></PORTALNAME>
				<DESCNAME><![CDATA[Access granted]]></DESCNAME>
				<CDT>2025-08-07 15:57:35</CDT>
			</EVENT>`,
			checkFunc: func(t *testing.T, event lenel_s2.Event) {
				if event.DescName != "Access granted" {
					t.Errorf("Expected DescName 'Access granted', got '%s'", event.DescName)
				}
			},
		},
		{
			name: "System Event",
			eventXML: `<EVENT>
				<ACTIVITYID>5737</ACTIVITYID>
				<DESCNAME><![CDATA[System startup]]></DESCNAME>
				<PARTNAME><![CDATA[REGO]]></PARTNAME>
				<CDT>2025-08-07 05:05:05</CDT>
			</EVENT>`,
			checkFunc: func(t *testing.T, event lenel_s2.Event) {
				if event.ActivityID != "5737" {
					t.Errorf("Expected ActivityID '5737', got '%s'", event.ActivityID)
				}
				if event.PartName != "REGO" {
					t.Errorf("Expected PartName 'REGO', got '%s'", event.PartName)
				}
			},
		},
		{
			name: "Network Event",
			eventXML: `<EVENT>
				<NODENAME><![CDATA[REGO-Node-5]]></NODENAME>
				<NODEADDRESS>10.5.7.35</NODEADDRESS>
				<NODEUNIQUE>ABCD5735EFGH</NODEUNIQUE>
				<DESCNAME><![CDATA[Network node connected]]></DESCNAME>
			</EVENT>`,
			checkFunc: func(t *testing.T, event lenel_s2.Event) {
				if event.NodeName != "REGO-Node-5" {
					t.Errorf("Expected NodeName 'REGO-Node-5', got '%s'", event.NodeName)
				}
				if event.NodeAddress != "10.5.7.35" {
					t.Errorf("Expected NodeAddress '10.5.7.35', got '%s'", event.NodeAddress)
				}
			},
		},
	}

	for _, tt := range eventTypes {
		t.Run(tt.name, func(t *testing.T) {
			var event lenel_s2.Event
			err := xml.Unmarshal([]byte(tt.eventXML), &event)
			if err != nil {
				t.Fatalf("Failed to unmarshal event: %v", err)
			}

			tt.checkFunc(t, event)
		})
	}
}
