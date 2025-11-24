// pkg/lenel_s2/entities.go
package lenel_s2

import (
	"crypto/tls"
	"encoding/xml"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strconv"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ### Lenel S2 Client Structs
// ---------------------------------------------------------------------
// Credentials for S2
type Client struct {
	BaseURL string
	Session *NetboxResponse[any]
	HTTP    *requests.Client
	Log     *log.Logger
	Cache   *cache.Cache
}

// Session holds the session ID from login responses
type NetboxResponse[E any] struct {
	XMLName  xml.Name    `xml:"NETBOX"`
	ID       string      `xml:"sessionid,attr"`
	Response Response[E] `xml:"RESPONSE"`
}

type Response[E any] struct {
	APIError int    `xml:"APIERROR"`         // APIError holds API-level error codes if present (e.g. "1", "2", etc.)
	Code     string `xml:"CODE"`             // CODE is the command response code ("SUCCESS" or "FAIL")
	Details  *E     `xml:"DETAILS"`          // Can be an object of any type
	Error    string `xml:"ERRMSG,omitempty"` // Error stores a human-readable error message from command-level failures.
}

// UnmarshalXML provides a custom unmarshaller that reads the
// elements within <RESPONSE> only once.
// It will process:
//
//   - <APIERROR>: Decodes API-level errors (e.g., authentication failure).
//   - <CODE>: Decodes the command-level code.
//   - <DETAILS>: Depending on the previously decoded CODE, either:
//   - For "FAIL", decode only <ERRMSG> inside DETAILS.
//   - Otherwise, decode the full DETAILS into the generic type.
func (r *Response[E]) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break // End of input.
			}
			return err
		}

		switch tok := token.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "APIERROR":
				// Decode the API error as an int.
				var code int
				if err := d.DecodeElement(&code, &tok); err != nil {
					return err
				}
				r.APIError = code
				if msg, ok := apiErrors[code]; ok {
					r.Error = msg
				}
				// Drain the rest of <RESPONSE> so the element is fully consumed.
				if err := drainElements(d, start.Name.Local); err != nil {
					return err
				}
				// Since APIERROR should be first, return immediately.
				return nil

			case "CODE":
				if err := d.DecodeElement(&r.Code, &tok); err != nil {
					return err
				}

			case "DETAILS":
				if r.Code == "FAIL" {
					// For command-level failures, decode only the <ERRMSG>.
					var error struct {
						ErrMsg string `xml:"ERRMSG"`
					}
					if err := d.DecodeElement(&error, &tok); err != nil {
						return err
					}
					r.Error = error.ErrMsg
				} else {
					// Otherwise, decode the full details into the generic type.
					var details E
					if err := d.DecodeElement(&details, &tok); err != nil {
						return err
					}
					r.Details = &details
				}

			default:
				// Skip any unknown element.
				if err := d.Skip(); err != nil {
					return err
				}
			}

		case xml.EndElement:
			if tok.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
	return nil
}

// drainElements reads XML tokens until the closing element of
// the provided response name is encountered, ensuring the decoder
// state is clean.
func drainElements(d *xml.Decoder, responseName string) error {
	for {
		t, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// Return when we see the end of the RESPONSE element.
		if end, ok := t.(xml.EndElement); ok && end.Name.Local == responseName {
			return nil
		}
	}
	return nil
}

// NetboxEventResponse handles streaming event responses
// This is used specifically for StreamEvents which returns EVENT instead of DETAILS
type NetboxEventResponse[E any] struct {
	XMLName  xml.Name         `xml:"NETBOX"`
	Response EventResponse[E] `xml:"RESPONSE"`
}

// EventResponse handles the RESPONSE element for streaming events
type EventResponse[E any] struct {
	Command  string `xml:"command,attr"`     // Command name (e.g., "StreamEvents")
	APIError int    `xml:"APIERROR"`         // API-level error codes
	Code     string `xml:"CODE"`             // SUCCESS or FAIL
	Event    *E     `xml:"EVENT"`            // Event data (instead of DETAILS)
	Error    string `xml:"ERRMSG,omitempty"` // Error message for failures
}

// UnmarshalXML provides custom unmarshalling for EventResponse
// Similar to Response[E] but handles EVENT tag instead of DETAILS
func (r *EventResponse[E]) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Process command attribute
	for _, attr := range start.Attr {
		if attr.Name.Local == "command" {
			r.Command = attr.Value
		}
	}

	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch tok := token.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "APIERROR":
				var code int
				if err := d.DecodeElement(&code, &tok); err != nil {
					return err
				}
				r.APIError = code
				if msg, ok := apiErrors[code]; ok {
					r.Error = msg
				}
				// Drain the rest of <RESPONSE> so the element is fully consumed.
				if err := drainElements(d, start.Name.Local); err != nil {
					return err
				}
				return nil

			case "CODE":
				if err := d.DecodeElement(&r.Code, &tok); err != nil {
					return err
				}

			case "EVENT":
				if r.Code == "FAIL" {
					// For failures, just skip the EVENT content
					if err := d.Skip(); err != nil {
						return err
					}
				} else {
					// Decode the event data
					var event E
					if err := d.DecodeElement(&event, &tok); err != nil {
						return err
					}
					r.Event = &event
				}

			case "ERRMSG":
				if err := d.DecodeElement(&r.Error, &tok); err != nil {
					return err
				}

			default:
				// Skip unknown elements
				if err := d.Skip(); err != nil {
					return err
				}
			}

		case xml.EndElement:
			if tok.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
	return nil
}

type Credentials struct {
	Username string
	Password string
}

type NetboxCommand struct {
	XMLName   xml.Name `xml:"NETBOX-API"`
	SessionID string   `xml:"sessionid,attr,omitempty"`
	Command   Command  `xml:"COMMAND"`
}

type Command struct {
	Name       string      `xml:"name,attr"`
	Num        string      `xml:"num,attr"`
	DateFormat string      `xml:"dateformat,attr,omitempty"`
	Params     interface{} `xml:"PARAMS,omitempty"`
}

// END OF LENEL S2 CLIENT STRUCTS
//---------------------------------------------------------------------

// ### Lenel S2 Client Configuration Options
// ---------------------------------------------------------------------

type clientConfig struct {
	insecureSkipVerify bool
	httpClient         *http.Client
}

type Option func(*clientConfig)

// WithInsecureSkipVerify disables TLS certificate verification.
// WARNING: This should only be used in development or with trusted self-signed certificates.
func WithInsecureSkipVerify() Option {
	return func(cfg *clientConfig) {
		cfg.insecureSkipVerify = true

		// Configure HTTP client with TLS settings
		if cfg.httpClient == nil {
			cfg.httpClient = &http.Client{}
		}

		// Ensure Transport is configured with TLS settings
		if cfg.httpClient.Transport == nil {
			cfg.httpClient.Transport = &http.Transport{}
		}

		if transport, ok := cfg.httpClient.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig == nil {
				transport.TLSClientConfig = &tls.Config{}
			}
			transport.TLSClientConfig.InsecureSkipVerify = true
		}
	}
}

// END OF LENEL S2 CLIENT CONFIGURATION OPTIONS
// ---------------------------------------------------------------------

// ### Lenel S2 People Structs
// ---------------------------------------------------------------------

type People struct {
	People  *[]*Person `xml:"PEOPLE>PERSON"`
	NextKey string     `xml:"NEXTKEY,omitempty"` // This field is used for pagination.
}

// Append merges another People into the receiver by appending its Accesses.
func (p People) Append(result PaginatedResponse) PaginatedResponse {
	if p.People == nil {
		p.People = new([]*Person)
	}
	more, ok := result.(People)
	if !ok {
		return p
	}
	*p.People = append(*p.People, *more.People...)
	return p
}

// NextToken returns the token used for getting the next page of People results.
// It handles the special case where "-1" (or an empty value) indicates no further pages.
func (p People) NextToken() string {
	if p.NextKey == "-1" || p.NextKey == "" {
		return ""
	}

	return p.NextKey
}

func (p People) SetCommand(nb *Client, cmd *NetboxCommand) NetboxCommand {

	nextKey, err := strconv.Atoi(p.NextKey)
	if err != nil {
		return *cmd // If we can't parse the next key, return the original command
	}

	params := reflect.ValueOf(cmd.Command.Params)

	if params.Kind() == reflect.Ptr {
		params = params.Elem()
	}

	// If it's not settable (struct passed by value), make it settable
	if params.Kind() == reflect.Struct && !params.CanSet() {
		// Create a new pointer to a copy of the command
		paramsPtr := reflect.New(params.Type())
		paramsPtr.Elem().Set(params)
		params = paramsPtr.Elem()
	}

	if params.Kind() == reflect.Struct {
		field := params.FieldByName("StartFromKey")
		if field.IsValid() && field.CanSet() && field.Kind() == reflect.Int {
			field.SetInt(int64(nextKey))
		}
	}

	return nb.BuildRequest(
		cmd.Command.Name,
		params.Interface(),
	)
}

type Person struct {
	PersonID         string       `xml:"PERSONID"`                 // ID number of the person whose record information is to be returned
	FirstName        string       `xml:"FIRSTNAME"`                // The person's first name
	MiddleName       string       `xml:"MIDDLENAME"`               // Middle name
	LastName         string       `xml:"LASTNAME"`                 // The person's last name
	Username         string       `xml:"USERNAME"`                 // Username associated with a person’s NetBox login account
	Role             string       `xml:"ROLE"`                     // Role for a person’s record when they login to the NetBox
	AuthType         string       `xml:"AUTHTYPE"`                 // (Required) Authorization type for a person’s record. Valid values are `DB`, `LDAP`, or `SSO`
	Partition        string       `xml:"PARTITION"`                // Partition of person record
	ActivationDate   string       `xml:"ACTDATE"`                  // Activation date of the person record
	ExpirationDate   string       `xml:"EXPDATE"`                  // Expiration date for the person record
	UDF1             string       `xml:"UDF1"`                     // UDF1 through UDF20: User defined fields (20)
	UDF2             string       `xml:"UDF2"`                     // User Defined Field 2
	UDF3             string       `xml:"UDF3"`                     // User Defined Field 3
	UDF4             string       `xml:"UDF4"`                     // User Defined Field 4
	UDF5             string       `xml:"UDF5"`                     // User Defined Field 5
	UDF6             string       `xml:"UDF6"`                     // User Defined Field 6
	UDF7             string       `xml:"UDF7"`                     // User Defined Field 7
	UDF8             string       `xml:"UDF8"`                     // User Defined Field 8
	UDF9             string       `xml:"UDF9"`                     // User Defined Field 9
	UDF10            string       `xml:"UDF10"`                    // User Defined Field 10
	UDF11            string       `xml:"UDF11"`                    // User Defined Field 11
	UDF12            string       `xml:"UDF12"`                    // User Defined Field 12
	UDF13            string       `xml:"UDF13"`                    // User Defined Field 13
	UDF14            string       `xml:"UDF14"`                    // User Defined Field 14
	UDF15            string       `xml:"UDF15"`                    // User Defined Field 15
	UDF16            string       `xml:"UDF16"`                    // User Defined Field 16
	UDF17            string       `xml:"UDF17"`                    // User Defined Field 17
	UDF18            string       `xml:"UDF18"`                    // User Defined Field 18
	UDF19            string       `xml:"UDF19"`                    // User Defined Field 19
	UDF20            string       `xml:"UDF20"`                    // User Defined Field 20
	Pin              string       `xml:"PIN"`                      // PIN number that may be required for a card reader with a numeric pad
	Notes            string       `xml:"NOTES"`                    // Notes field of the person record
	Deleted          bool         `xml:"DELETED"`                  // Specifies if the person record has been deleted
	PictureURL       string       `xml:"PICTUREURL"`               // The picture data file returned as text between the PICTUREURL elements. The data file is stored in the following directory on the system: /usr/local/s2/web/upload/pics
	BadgeLayout      string       `xml:"BADGELAYOUT"`              // Name of the photo ID badging layout file.
	LastModified     string       `xml:"LASTMOD"`                  // (Deprecated) Last time the person was modified in YYYY-MM-DD format.
	LastEdit         string       `xml:"LASTEDIT"`                 // Date and time the contents of the person record were last changed, in YYYY-MM-DD HH:MM:SS format.
	LastEditPersonID string       `xml:"LASTEDITPERSONID"`         // ID of the person who last edited the person record.
	Phone            string       `xml:"CONTACTPHONE"`             // Office phone number
	Mobile           string       `xml:"MOBILEPHONE"`              // Mobile phone number
	Email            string       `xml:"CONTACTEMAIL"`             // Office email address
	SMSEmail         string       `xml:"CONTACTSMSEMAIL"`          // Office SMS email address
	Location         string       `xml:"CONTACTLOCATION"`          // Emergency contact location
	OtherName        string       `xml:"OTHERCONTACTNAME"`         // Emergency contact name
	OtherPhone       string       `xml:"OTHERCONTACTPHONE"`        // Emergency contact phone number
	Vehicles         []Vehicle    `xml:"VEHICLES>VEHICLE"`         // Element blocks containing a VEHICLE element block for each vehicle defined in the person record
	AccessLevels     []string     `xml:"ACCESSLEVELS>ACCESSLEVEL"` // Element block containing one or more access levels (maximum of 32) to be associated with the person. Only the access levels currently assigned are returned, and if there are none assigned, none are returned
	AccessCards      []AccessCard `xml:"ACCESSCARDS>ACCESSCARD"`   // Element block containing one or more credentials to be associated with the person. Only the credentials currently assigned are returned, and if there are none assigned, none are returned
}

type Vehicle struct {
	Color   string `xml:"VEHICLECOLOR"`  // The vehicle color
	Make    string `xml:"VEHICLEMAKE"`   // The vehicle make
	Model   string `xml:"VEHICLEMODEL"`  // The vehicle model
	State   string `xml:"VEHICLESTATE"`  // The vehicle state
	License string `xml:"VEHICLELICNUM"` // The vehicle license number
	Tag     string `xml:"VEHICLETAGNUM"` // The vehicle tag number
}

type AccessCard struct {
	EncodedNum     string `xml:"ENCODEDNUM"`            // ENDCODEDNUM is a representation of the actual data on the credential. Credentials are interpreted using a set of rules defined by the card format type
	HotStamp       string `xml:"HOTSTAMP,omitempty"`    // HOTSTAMP is a value optionally stamped on the card or recorded for reference. Some deployments will choose to have these fields use the same value
	Format         string `xml:"CARDFORMAT"`            // Name of the format to be used to decode the credential.
	Disabled       bool   `xml:"DISABLED,omitempty"`    // DISABLED is a flag that indicates if the credential is disabled
	Status         string `xml:"CARDSTATUS,omitempty"`  // Text string that specifies the status of the credential. If a CARDSTATUS is not included in the command, the default status ACTIVE is assigned to the CARDSTATUS parameter
	ExpirationDate string `xml:"CARDEXPDATE,omitempty"` // Expiration date for the credential.
}

// END OF LENEL S2 PEOPLE STRUCTS
//---------------------------------------------------------------------

// ### Lenel S2 UDF Structs
// ---------------------------------------------------------------------

type UDFLists struct {
	UserDefinedFields []*UDF `xml:"UDFLISTS>UDFLIST"` // Element block containing one or more user defined fields (maximum of 20) to be associated with the person. Only the user defined fields currently assigned are returned, and if there are none assigned, none are returned
}

type UDF struct {
	Key         string `xml:"UDFLISTKEY"`  // The unique key for the UDF value list
	Name        string `xml:"NAME"`        // The name of the UDF value list
	Description string `xml:"DESCRIPTION"` // A description of the UDF value list, if available
}

// END OF LENEL S2 UDF STRUCTS
//---------------------------------------------------------------------

// ### Lenel S2 (Netbox) Access Structs
// ---------------------------------------------------------------------
type AccessHistory struct {
	Accesses  []*Access `xml:"ACCESSES>ACCESS"`
	NextLogID string    `xml:"NEXTLOGID,omitempty"` // This field is used for pagination.
}

// NextToken returns the token used for getting the next page of AccessHistory results.
// It handles the special case where "-1" (or an empty value) indicates no further pages.
func (ah *AccessHistory) NextToken() string {
	if ah.NextLogID == "-1" || ah.NextLogID == "" {
		return ""
	}

	return ah.NextLogID
}

// Append merges another AccessHistory into the receiver by appending its Accesses.
func (ah *AccessHistory) Append(resp PaginatedResponse) PaginatedResponse {
	other, ok := resp.(*AccessHistory)
	if !ok {
		return ah
	}
	ah.Accesses = append(ah.Accesses, other.Accesses...)
	return ah
}

func (ah *AccessHistory) SetCommand(nb *Client, cmd *NetboxCommand) NetboxCommand {
	nextKey, err := strconv.Atoi(ah.NextLogID)
	if err != nil {
		return *cmd // If we can't parse the next key, return the original command
	}

	params := reflect.ValueOf(cmd.Command.Params)

	if params.Kind() == reflect.Ptr {
		params = params.Elem()
	}

	// If it's not settable (struct passed by value), make it settable
	if params.Kind() == reflect.Struct && !params.CanSet() {
		// Create a new pointer to a copy of the command
		paramsPtr := reflect.New(params.Type())
		paramsPtr.Elem().Set(params)
		params = paramsPtr.Elem()
	}

	if params.Kind() == reflect.Struct {
		field := params.FieldByName("AfterLogID")
		if field.IsValid() && field.CanSet() && field.Kind() == reflect.Int {
			field.SetInt(int64(nextKey))
		}
	}

	return nb.BuildRequest(
		cmd.Command.Name,
		params.Interface(),
	)
}

type Access struct {
	LogID     string `xml:"LOGID"`     // Number which identifies the data log record
	PersonID  string `xml:"PERSONID"`  // External Person ID associated with the person who owns the specified credential. This is the field in the person record labeled "ID #.
	Reader    string `xml:"READER"`    // Name of the card reader
	DTTM      string `xml:"DTTM"`      // System date and time associated with the data log
	NodeDTTM  string `xml:"NODEDTM"`   // Node date and time associated with the data log
	Type      int    `xml:"TYPE"`      // Reason type which specifies a valid or invalid access. And invalid access also returns a Reason code
	Reason    int    `xml:"REASON"`    // Reason code which specifies the reason for an invalid access
	ReaderKey string `xml:"READERKEY"` // Unique identifier for the reader
	PortalKey string `xml:"PORTALKEY"` // Unique identifier for the portal
}

// END OF LENEL S2 (NETBOX) ACCESS STRUCTS
//---------------------------------------------------------------------

// ### Lenel S2 (Netbox) Card Structs
// ---------------------------------------------------------------------

type CardFormats struct {
	Formats []string `xml:"CARDFORMATS>CARDFORMAT"`
}

// END OF LENEL S2 (NETBOX) CARD STRUCTS
//---------------------------------------------------------------------

// ### Lenel S2 (Netbox) Event Structs
// ---------------------------------------------------------------------

// Event represents a single event in the StreamEvents response
// All fields are pointers to handle optional presence in different event types
type Event struct {
	XMLName             xml.Name `xml:"EVENT"`
	ActivityID          string   `xml:"ACTIVITYID,omitempty"`          // Internal identifier for the activity
	AlarmID             string   `xml:"ALARMID,omitempty"`             // Internal identifier for the alarm
	AlarmPanelName      string   `xml:"ALARMPANELNAME,omitempty"`      // Configured name of the alarm panel
	AlarmStateName      string   `xml:"ALARMSTATENAME,omitempty"`      // Alarm state name
	AlarmTimerName      string   `xml:"ALARMTIMERNAME,omitempty"`      // Alarm timer name
	AlarmTransitionName string   `xml:"ALARMTRANSITIONNAME,omitempty"` // Alarm transition name
	ACName              string   `xml:"ACNAME,omitempty"`              // Access Card Hot Stamp
	ACNum               string   `xml:"ACNUM,omitempty"`               // Access Card Encoded Number
	BladeSlot           string   `xml:"BLADESLOT,omitempty"`           // Slot number of the node blade
	CDT                 string   `xml:"CDT,omitempty"`                 // Controller date/time
	DescName            string   `xml:"DESCNAME,omitempty"`            // General description of the activity
	Detail              string   `xml:"DETAIL,omitempty"`              // Additional text detail
	EventID             string   `xml:"EVENTID,omitempty"`             // Internal identifier for the event
	EvtName             string   `xml:"EVTNAME,omitempty"`             // Configured name of the event
	EvtPrio             string   `xml:"EVTPRIO,omitempty"`             // Configured priority number
	IPanelArea          string   `xml:"IPANELAREA,omitempty"`          // Intrusion panel area
	IPanelName          string   `xml:"IPANELNAME,omitempty"`          // Intrusion panel name
	IPanelOutput        string   `xml:"IPANELOUTPUT,omitempty"`        // Intrusion panel output
	IPanelUser          string   `xml:"IPANELUSER,omitempty"`          // Intrusion panel user
	IPanelZone          string   `xml:"IPANELZONE,omitempty"`          // Intrusion panel zone
	LevelKey            string   `xml:"LEVELKEY,omitempty"`            // Internal identifier for threat level
	LocationKey         string   `xml:"LOCATIONKEY,omitempty"`         // Internal identifier for location
	LocationName        string   `xml:"LOCATIONNAME,omitempty"`        // Configured location name
	LoginAddress        string   `xml:"LOGINADDRESS,omitempty"`        // Host from which user logged in
	NDT                 string   `xml:"NDT,omitempty"`                 // Node date/time
	NodeAddress         string   `xml:"NODEADDRESS,omitempty"`         // IP address of the node
	NodeName            string   `xml:"NODENAME,omitempty"`            // Configured name for the node
	NodeUnique          string   `xml:"NODEUNIQUE,omitempty"`          // Unique identifier for the node
	PartitionKey        string   `xml:"PARTITIONKEY,omitempty"`        // Internal identifier for partition
	PartName            string   `xml:"PARTNAME,omitempty"`            // Partition name
	PersonID            string   `xml:"PERSONID,omitempty"`            // Configured person identifier
	PersonName          string   `xml:"PERSONNAME,omitempty"`          // Configured person name
	PortalKey           string   `xml:"PORTALKEY,omitempty"`           // Internal identifier for portal
	PortalName          string   `xml:"PORTALNAME,omitempty"`          // Configured name of the portal
	RdrName             string   `xml:"RDRNAME,omitempty"`             // Configured name of card reader
	ReaderKey           string   `xml:"READERKEY,omitempty"`           // Internal identifier for reader
	Reader2Key          string   `xml:"READER2KEY,omitempty"`          // Internal identifier for second reader
	ThreatName          string   `xml:"THREATNAME,omitempty"`          // Name of the threat level
	UCBitLength         string   `xml:"UCBITLENGTH,omitempty"`         // Number of bits in card format
}

// EventCategory represents the category of an event
// Categories are based on the NetBox documentation and can be used for filtering or processing events
type EventCategory string

const (
	CategoryAccess         EventCategory = "Access"
	CategoryPortal         EventCategory = "Portal"
	CategoryNetwork        EventCategory = "Network"
	CategoryAuthentication EventCategory = "Authentication"
	CategoryEvent          EventCategory = "Event"
	CategoryAlarm          EventCategory = "Alarm"
	CategoryElevator       EventCategory = "Elevator"
	CategoryThreat         EventCategory = "Threat"
	CategoryIntrusion      EventCategory = "Intrusion"
	CategorySystem         EventCategory = "System"
	CategoryBackup         EventCategory = "Backup"
)

// EventType represents a specific event with its metadata
type EventType struct {
	Name           string
	Category       EventCategory
	RequiredFields []tagName
}

// EventTypes provides access to all event types in the order they appear in the NetBox documentation
var EventTypes = struct {
	AccessGranted                       EventType // DESCNAME: "Access Granted"
	AccessDenied                        EventType // DESCNAME: "Access Denied"
	InvalidAccess                       EventType // DESCNAME: "Invalid Access"
	PortalHeldOpen                      EventType // DESCNAME: "Portal Held Open"
	PortalForcedOpen                    EventType // DESCNAME: "Portal Forced Open"
	PortalRestored                      EventType // DESCNAME: "Portal Restored"
	NetworkControllerStartup            EventType // DESCNAME: "Network Controller Startup"
	NetworkNodeStartup                  EventType // DESCNAME: "Network Node Startup"
	NetworkControllerShutdown           EventType // DESCNAME: "Network Controller Shutdown"
	MomentaryUnlock                     EventType // DESCNAME: "Momentary Unlock"
	Unlock                              EventType // DESCNAME: "Unlock"
	Relock                              EventType // DESCNAME: "Relock"
	NetworkNodeTimeout                  EventType // DESCNAME: "Network Node Timeout"
	NetworkNodeRestored                 EventType // DESCNAME: "Network Node Restored"
	NetworkNodeDisconnect               EventType // DESCNAME: "Network Node Disconnect"
	NetworkNodeBadConfiguration         EventType // DESCNAME: "Network Node Bad Configuration"
	NetworkNodeConnected                EventType // DESCNAME: "Network Node Connected"
	NetworkNodeIdentification           EventType // DESCNAME: "Network Node Identification"
	NetworkNodeDataDisconnect           EventType // DESCNAME: "Network Node Data Disconnect"
	LogArchiveSuccess                   EventType // DESCNAME: "Log Archive Success"
	LogArchiveFailure                   EventType // DESCNAME: "Log Archive Failure"
	LoggedIn                            EventType // DESCNAME: "Logged In"
	LoggedOut                           EventType // DESCNAME: "Logged Out"
	LoginFailed                         EventType // DESCNAME: "Login Failed"
	RequestMomentaryUnlock              EventType // DESCNAME: "Request Momentary Unlock"
	SessionExpired                      EventType // DESCNAME: "Session Expired"
	EventTriggered                      EventType // DESCNAME: "Event Triggered"
	EventNormal                         EventType // DESCNAME: "Event Normal"
	EventActivated                      EventType // DESCNAME: "Event Activated"
	EventTrouble                        EventType // DESCNAME: "Event Trouble"
	NetworkNodeTamperAlarm              EventType // DESCNAME: "Network Node Tamper Alarm"
	NetworkNodeDHCPFailed               EventType // DESCNAME: "Network Node DHCP Failed"
	ElevatorAccessGranted               EventType // DESCNAME: "Elevator Access Granted"
	ElevatorAccessDenied                EventType // DESCNAME: "Elevator Access Denied"
	ThreatLevelSet                      EventType // DESCNAME: "Threat Level Set"
	ThreatLevelSetAPI                   EventType // DESCNAME: "Threat Level Set API"
	ThreatLevelSetALM                   EventType // DESCNAME: "Threat Level Set ALM"
	LicenseReadFailure                  EventType // DESCNAME: "License Read Failure"
	FTPBackupComplete                   EventType // DESCNAME: "FTP Backup Complete"
	FTPBackupFailed                     EventType // DESCNAME: "FTP Backup Failed"
	AlarmActionsCleared                 EventType // DESCNAME: "Alarm Actions Cleared"
	AlarmAcknowledged                   EventType // DESCNAME: "Alarm Acknowledged"
	AlarmPanelArmRequest                EventType // DESCNAME: "Alarm Panel Arm Request"
	AlarmPanelDisarmRequest             EventType // DESCNAME: "Alarm Panel Disarm Request"
	AlarmPanelArmed                     EventType // DESCNAME: "Alarm Panel Armed"
	AlarmPanelDisarmed                  EventType // DESCNAME: "Alarm Panel Disarmed"
	AlarmPanelArmFailure                EventType // DESCNAME: "Alarm Panel Arm Failure"
	AlarmPanelDisarmFailure             EventType // DESCNAME: "Alarm Panel Disarm Failure"
	AlarmPanelArmInterrupted            EventType // DESCNAME: "Alarm Panel Arm Interrupted"
	NetworkNodeBladeNotResponding       EventType // DESCNAME: "Network Node Blade Not Responding"
	NetworkNodeBladeResponding          EventType // DESCNAME: "Network Node Blade Responding"
	NetworkNodeCoprocessorNotResponding EventType // DESCNAME: "Network Node Coprocessor Not Responding"
	NetworkNodeCoprocessorResponding    EventType // DESCNAME: "Network Node Coprocessor Responding"
	NASBackupComplete                   EventType // DESCNAME: "NAS Backup Complete"
	NASBackupFailed                     EventType // DESCNAME: "NAS Backup Failed"
	EventAcknowledged                   EventType // DESCNAME: "Event Acknowledged"
	EventActionsCleared                 EventType // DESCNAME: "Event Actions Cleared"
	AccessNotCompleted                  EventType // DESCNAME: "Access Not Completed"
	DutyLogEntry                        EventType // DESCNAME: "Duty Log Entry"
	BatteryVoltageLow                   EventType // DESCNAME: "Battery Voltage Low"
	BatteryFailed                       EventType // DESCNAME: "Battery Failed"
	BatteryReplaced                     EventType // DESCNAME: "Battery Replaced"
	AccessDeniedRadioBusy               EventType // DESCNAME: "Access Denied Because Radio Busy"
	NetworkNodeDiscovered               EventType // DESCNAME: "Network Node Discovered"
	NetworkNodeConfigReloaded           EventType // DESCNAME: "Network Node Configuration and Card Info Reloaded"
	IntrusionPanelConnected             EventType // DESCNAME: "Intrusion Panel Connected"
	IntrusionPanelNotConnected          EventType // DESCNAME: "Intrusion Panel Not Connected"
	IntrusionPanelRequestArmArea        EventType // DESCNAME: "Intrusion Panel Request Arm Area"
	IntrusionPanelRequestDisarmArea     EventType // DESCNAME: "Intrusion Panel Request Disarm Area"
	IntrusionPanelRequestBypassZone     EventType // DESCNAME: "Intrusion Panel Request Bypass Zone"
	IntrusionPanelRequestResetBypass    EventType // DESCNAME: "Intrusion Panel Request Reset Bypass"
	IntrusionPanelAreaArmed             EventType // DESCNAME: "Intrusion Panel Area Armed"
	IntrusionPanelAreaDisarmed          EventType // DESCNAME: "Intrusion Panel Area Disarmed"
	IntrusionPanelAlarm                 EventType // DESCNAME: "Intrusion Panel Alarm"
	IntrusionPanelRestored              EventType // DESCNAME: "Intrusion Panel Restored"
	IntrusionPanelConnectionRestored    EventType // DESCNAME: "Intrusion Panel Connection Restored"
	IntrusionPanelZoneBypassed          EventType // DESCNAME: "Intrusion Panel Zone Bypassed"
	IntrusionPanelZoneReset             EventType // DESCNAME: "Intrusion Panel Zone Reset"
	IntrusionPanelAreaLateToAlarm       EventType // DESCNAME: "Intrusion Panel Area Late to Alarm"
	IntrusionPanelZoneTrouble           EventType // DESCNAME: "Intrusion Panel Zone Trouble"
	IntrusionPanelZoneFault             EventType // DESCNAME: "Intrusion Panel Zone Fault"
	IntrusionPanelZoneRestored          EventType // DESCNAME: "Intrusion Panel Zone Restored"
	IntrusionPanelRequestToggleOutput   EventType // DESCNAME: "Intrusion Panel Request Toggle Output"
	IntrusionPanelOutputToggled         EventType // DESCNAME: "Intrusion Panel Output Toggled"
	IntrusionPanelCommPathTrouble       EventType // DESCNAME: "Intrusion Panel Communication Path Trouble"
	IntrusionPanelCommPathRestored      EventType // DESCNAME: "Intrusion Panel Communication Path Restored"
	SystemBackupStarted                 EventType // DESCNAME: "System Backup Started"
	SystemBackupInProgress              EventType // DESCNAME: "System Backup In Progress"
	SystemBackupSuccessful              EventType // DESCNAME: "System Backup Successful"
	SystemBackupFailed                  EventType // DESCNAME: "System Backup Failed"
	VideoEvent                          EventType // DESCNAME: "Video Event"
	CauseInactive                       EventType // DESCNAME: "Cause Inactive"
	KeypadTimedUnlockExpired            EventType // DESCNAME: "Keypad Timed Unlock Expired"
	TemporaryCredentialIssued           EventType // DESCNAME: "Temporary Credential Issued"
	TemporaryCredentialReturned         EventType // DESCNAME: "Temporary Credential Returned"
	RequestPersistentUnlock             EventType // DESCNAME: "Request Persistent Unlock"
	RequestPersistentLock               EventType // DESCNAME: "Request Persistent Lock"
	RequestDisablePortal                EventType // DESCNAME: "Request Disable Portal"
	RequestEnablePortal                 EventType // DESCNAME: "Request Enable Portal"
	PortalDisabled                      EventType // DESCNAME: "Portal Disabled"
	PortalEnabled                       EventType // DESCNAME: "Portal Enabled"
	KeypadCommandExecuted               EventType // DESCNAME: "Keypad Command Executed"
	ReaderTamperAlarm                   EventType // DESCNAME: "Reader Tamper Alarm"
	ReaderTamperNormal                  EventType // DESCNAME: "Reader Tamper Normal"
	ReaderBatteryAlarm                  EventType // DESCNAME: "Reader Battery Alarm"
	ReaderBatteryNormal                 EventType // DESCNAME: "Reader Battery Normal"
	BladeTamperAlarm                    EventType // DESCNAME: "Blade Tamper Alarm"
	BladeTamperNormal                   EventType // DESCNAME: "Blade Tamper Normal"
	ManualKeyOverride                   EventType // DESCNAME: "Manual Key Override"
	Evacuation                          EventType // DESCNAME: "Evacuation"
	MusteringForEvacuation              EventType // DESCNAME: "Mustering for Evacuation"
	SystemHealth                        EventType // DESCNAME: "System Health"
	ReaderCommunicationAlarm            EventType // DESCNAME: "Reader Communication Alarm"
	ReaderCommunicationNormal           EventType // DESCNAME: "Reader Communication Normal"
	FTPBackupFailedConfigured           EventType // DESCNAME: "FTP Backup Failed: FTP is Configured and Enabled"
	NASBackupFailedConfigured           EventType // DESCNAME: "NAS Backup Failed: FTP is Configured and Enabled"
	BackupCopiedToFTP                   EventType // DESCNAME: "Backup Successfully Copied to FTP Server"
	BackupCopiedToNAS                   EventType // DESCNAME: "Backup Successfully Copied to NAS Server"
	ElevatorFreeAccess                  EventType // DESCNAME: "Elevator Free Access"
	ElevatorAccessNotCompleted          EventType // DESCNAME: "Elevator Access Not Completed"
	EmergencyCallActivated              EventType // DESCNAME: "Emergency Call Activated for Elevator"
	EmergencyCallRestored               EventType // DESCNAME: "Emergency Call Restored for Elevator"
	PrivacyEnabled                      EventType // DESCNAME: "Privacy Enabled"
	InteriorPushButton                  EventType // DESCNAME: "Interior Push Button Pressed"
	DoorBolted                          EventType // DESCNAME: "Door Bolted"
	SystemLicenseExpires60Days          EventType // DESCNAME: "System License Expires in 60 Days"
	SystemLicenseExpires30Days          EventType // DESCNAME: "System License Expires in 30 Days"
	SystemLicenseExpired                EventType // DESCNAME: "System License Expired"
	SystemLicenseExpiredAck             EventType // DESCNAME: "System License Expired Acknowledged"
	NetworkNodeLicenseNotDetected       EventType // DESCNAME: "Network Node System License Not Detected"
	NetworkNodeLicenseNotDetected21Days EventType // DESCNAME: "Network Node System License Not Detected for 21 Days"
	NetworkNodeLicenseNotDetected30Days EventType // DESCNAME: "Network Node System License Not Detected for 30 Days"
	NetworkNodeLicenseReestablished     EventType // DESCNAME: "Network Node System License Reestablished"
	NetworkNodeLicenseWarningAck        EventType // DESCNAME: "Network Node System License Warning Acknowledged"
	NetworkNodeLicenseErrorAck          EventType // DESCNAME: "Network Node System License Error Acknowledged"
	NetworkControllerTakeover           EventType // DESCNAME: "Network Controller Takeover"
	NetworkControllerPrimary            EventType // DESCNAME: "Network Controller Configured As Primary"
	NetworkControllerStandby            EventType // DESCNAME: "Network Controller Configured As Standby"
	NetworkNodeSecureConfigError        EventType // DESCNAME: "Network Node Secure Configuration Error"
	NetworkNodeSecureCommFailed         EventType // DESCNAME: "Network Node Secure Communication Failed"
}{
	// Access Events
	AccessGranted: EventType{
		Name:     "Access Granted",
		Category: CategoryAccess,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.ReaderName,
			EventTags.ReaderKey,
			EventTags.Reader2Key,
			EventTags.ACName,
			EventTags.ACNum,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AccessDenied: EventType{
		Name:     "Access Denied",
		Category: CategoryAccess,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.ReaderKey,
			EventTags.Reader2Key,
			EventTags.ReaderName,
			EventTags.ACName,
			EventTags.ACNum,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	InvalidAccess: EventType{
		Name:     "Invalid Access",
		Category: CategoryAccess,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.ReaderKey,
			EventTags.Reader2Key,
			EventTags.ReaderName,
			EventTags.ACName,
			EventTags.ACNum,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Portal Events
	PortalHeldOpen: EventType{
		Name:     "Portal Held Open",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	PortalForcedOpen: EventType{
		Name:     "Portal Forced Open",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	PortalRestored: EventType{
		Name:     "Portal Restored",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Network Controller Events
	NetworkControllerStartup: EventType{
		Name:     "Network Controller Startup",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeStartup: EventType{
		Name:     "Network Node Startup",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkControllerShutdown: EventType{
		Name:     "Network Controller Shutdown",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Lock/Unlock Events
	MomentaryUnlock: EventType{
		Name:     "Momentary Unlock",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	Unlock: EventType{
		Name:     "Unlock",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	Relock: EventType{
		Name:     "Relock",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PortalName,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Network Node Events
	NetworkNodeTimeout: EventType{
		Name:     "Network Node Timeout",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeRestored: EventType{
		Name:     "Network Node Restored",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeDisconnect: EventType{
		Name:     "Network Node Disconnect",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeBadConfiguration: EventType{
		Name:     "Network Node Bad Configuration",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeConnected: EventType{
		Name:     "Network Node Connected",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeIdentification: EventType{
		Name:     "Network Node Identification",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeDataDisconnect: EventType{
		Name:     "Network Node Data Disconnect",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Archive Events
	LogArchiveSuccess: EventType{
		Name:     "Log Archive Success",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	LogArchiveFailure: EventType{
		Name:     "Log Archive Failure",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Authentication Events
	LoggedIn: EventType{
		Name:     "Logged In",
		Category: CategoryAuthentication,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.LoginAddress,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	LoggedOut: EventType{
		Name:     "Logged Out",
		Category: CategoryAuthentication,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.LoginAddress,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	LoginFailed: EventType{
		Name:     "Login Failed",
		Category: CategoryAuthentication,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.LoginAddress,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	RequestMomentaryUnlock: EventType{
		Name:     "Request Momentary Unlock",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PortalName,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	SessionExpired: EventType{
		Name:     "Session Expired",
		Category: CategoryAuthentication,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.LoginAddress,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Event System Events
	EventTriggered: EventType{
		Name:     "Event Triggered",
		Category: CategoryEvent,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	EventNormal: EventType{
		Name:     "Event Normal",
		Category: CategoryEvent,
		RequiredFields: []tagName{
			EventTags.AlarmID,
			EventTags.EventID,
			EventTags.EventName,
			EventTags.EventPriorityNumber,
			EventTags.NDT,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	EventActivated: EventType{
		Name:     "Event Activated",
		Category: CategoryEvent,
		RequiredFields: []tagName{
			EventTags.AlarmID,
			EventTags.EventID,
			EventTags.EventName,
			EventTags.EventPriorityNumber,
			EventTags.NDT,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	EventTrouble: EventType{
		Name:     "Event Trouble",
		Category: CategoryEvent,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Tamper/Security Events
	NetworkNodeTamperAlarm: EventType{
		Name:     "Network Node Tamper Alarm",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeDHCPFailed: EventType{
		Name:     "Network Node DHCP Failed",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Elevator Events
	ElevatorAccessGranted: EventType{
		Name:     "Elevator Access Granted",
		Category: CategoryElevator,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.ReaderName,
			EventTags.UCBitLength,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	ElevatorAccessDenied: EventType{
		Name:     "Elevator Access Denied",
		Category: CategoryElevator,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.ReaderName,
			EventTags.UCBitLength,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Threat Level Events
	ThreatLevelSet: EventType{
		Name:     "Threat Level Set",
		Category: CategoryThreat,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.LevelKey,
			EventTags.LocationKey,
			EventTags.LocationName,
			EventTags.ThreatName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	ThreatLevelSetAPI: EventType{
		Name:     "Threat Level Set (API)",
		Category: CategoryThreat,
		RequiredFields: []tagName{
			EventTags.LevelKey,
			EventTags.LocationKey,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	ThreatLevelSetALM: EventType{
		Name:     "Threat Level Set (ALM)",
		Category: CategoryThreat,
		RequiredFields: []tagName{
			EventTags.LevelKey,
			EventTags.LocationKey,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// License Events
	LicenseReadFailure: EventType{
		Name:     "License Read Failure",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Backup Events
	FTPBackupComplete: EventType{
		Name:     "FTP Backup Complete",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	FTPBackupFailed: EventType{
		Name:     "FTP Backup Failed",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Alarm Events
	AlarmActionsCleared: EventType{
		Name:     "Alarm Actions Cleared",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmID,
			EventTags.EventID,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AlarmAcknowledged: EventType{
		Name:     "Alarm Acknowledged",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmID,
			EventTags.EventID,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AlarmPanelArmRequest: EventType{
		Name:     "Alarm Panel Arm Request",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AlarmPanelDisarmRequest: EventType{
		Name:     "Alarm Panel Disarm Request",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AlarmPanelArmed: EventType{
		Name:     "Alarm Panel Armed",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AlarmPanelDisarmed: EventType{
		Name:     "Alarm Panel Disarmed",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AlarmPanelArmFailure: EventType{
		Name:     "Alarm Panel Arm Failure",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AlarmPanelDisarmFailure: EventType{
		Name:     "Alarm Panel Disarm Failure",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	AlarmPanelArmInterrupted: EventType{
		Name:     "Alarm Panel Arm Interrupted",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.AlarmPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Blade/Coprocessor Events
	NetworkNodeBladeNotResponding: EventType{
		Name:     "Network Node Blade Not Responding",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.BladeSlot,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeBladeResponding: EventType{
		Name:     "Network Node Blade Responding",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.BladeSlot,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeCoprocessorNotResponding: EventType{
		Name:     "Network Node Coprocessor Not Responding",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeCoprocessorResponding: EventType{
		Name:     "Network Node Coprocessor Responding",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// NAS Backup Events
	NASBackupComplete: EventType{
		Name:     "NAS Backup Complete",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NASBackupFailed: EventType{
		Name:     "NAS Backup Failed",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Event Management
	EventAcknowledged: EventType{
		Name:     "Event Acknowledged",
		Category: CategoryEvent,
		RequiredFields: []tagName{
			EventTags.EventID,
			EventTags.EventName,
			EventTags.EventPriorityNumber,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	EventActionsCleared: EventType{
		Name:     "Event Actions Cleared",
		Category: CategoryEvent,
		RequiredFields: []tagName{
			EventTags.EventID,
			EventTags.EventName,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Additional Access Events
	AccessNotCompleted: EventType{
		Name:     "Access Not Completed",
		Category: CategoryAccess,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PortalName,
			EventTags.ReaderName,
			EventTags.ReaderKey,
			EventTags.Reader2Key,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	DutyLogEntry: EventType{
		Name:     "Duty Log Entry",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Battery Events
	BatteryVoltageLow: EventType{
		Name:     "Battery Voltage Low",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	BatteryFailed: EventType{
		Name:     "Battery Failed",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	BatteryReplaced: EventType{
		Name:     "Battery Replaced",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	AccessDeniedRadioBusy: EventType{
		Name:     "Access Denied Because Radio Busy",
		Category: CategoryAccess,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Network Node Discovery
	NetworkNodeDiscovered: EventType{
		Name:     "Network Node Discovered",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeConfigReloaded: EventType{
		Name:     "Network Node Configuration and Card Info Reloaded",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Intrusion Panel Events
	IntrusionPanelConnected: EventType{
		Name:     "Intrusion Panel Connected",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelNotConnected: EventType{
		Name:     "Intrusion Panel Not Connected",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelRequestArmArea: EventType{
		Name:     "Intrusion Panel Request Arm Area",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelRequestDisarmArea: EventType{
		Name:     "Intrusion Panel Request Disarm Area",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelUser,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelRequestBypassZone: EventType{
		Name:     "Intrusion Panel Request Bypass Zone",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelUser,
			EventTags.IPanelZone,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelRequestResetBypass: EventType{
		Name:     "Intrusion Panel Request Reset Bypass",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelUser,
			EventTags.IPanelZone,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelAreaArmed: EventType{
		Name:     "Intrusion Panel Area Armed",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelAreaDisarmed: EventType{
		Name:     "Intrusion Panel Area Disarmed",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelAlarm: EventType{
		Name:     "Intrusion Panel Alarm",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelZone,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelRestored: EventType{
		Name:     "Intrusion Panel Restored",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelConnectionRestored: EventType{
		Name:     "Intrusion Panel Connection Restored",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelZoneBypassed: EventType{
		Name:     "Intrusion Panel Zone Bypassed",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelZone,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelZoneReset: EventType{
		Name:     "Intrusion Panel Zone Reset",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelZone,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelAreaLateToAlarm: EventType{
		Name:     "Intrusion Panel Area Late to Alarm",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelZoneTrouble: EventType{
		Name:     "Intrusion Panel Zone Trouble",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelZone,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelZoneFault: EventType{
		Name:     "Intrusion Panel Zone Fault",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelZone,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelZoneRestored: EventType{
		Name:     "Intrusion Panel Zone Restored",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelZone,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelRequestToggleOutput: EventType{
		Name:     "Intrusion Panel Request Toggle Output",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.IPanelArea,
			EventTags.IPanelName,
			EventTags.IPanelOutput,
			EventTags.IPanelUser,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelOutputToggled: EventType{
		Name:     "Intrusion Panel Output Toggled",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.IPanelName,
			EventTags.IPanelOutput,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelCommPathTrouble: EventType{
		Name:     "Intrusion Panel Communication Path Trouble",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	IntrusionPanelCommPathRestored: EventType{
		Name:     "Intrusion Panel Communication Path Restored",
		Category: CategoryIntrusion,
		RequiredFields: []tagName{
			EventTags.IPanelName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// System Backup Events
	SystemBackupStarted: EventType{
		Name:     "System Backup Started",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	SystemBackupInProgress: EventType{
		Name:     "System Backup In Progress",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	SystemBackupSuccessful: EventType{
		Name:     "System Backup Successful",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	SystemBackupFailed: EventType{
		Name:     "System Backup Failed",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Video Event
	VideoEvent: EventType{
		Name:     "Video Event",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Misc Events
	CauseInactive: EventType{
		Name:     "Cause Inactive",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	KeypadTimedUnlockExpired: EventType{
		Name:     "Keypad Timed Unlock Expired",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	TemporaryCredentialIssued: EventType{
		Name:     "Temporary Credential Issued",
		Category: CategoryAccess,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	TemporaryCredentialReturned: EventType{
		Name:     "Temporary Credential Returned",
		Category: CategoryAccess,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Portal Request Events
	RequestPersistentUnlock: EventType{
		Name:     "Request Persistent Unlock",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PortalName,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	RequestPersistentLock: EventType{
		Name:     "Request Persistent Lock",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PortalKey,
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.PortalName,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	RequestDisablePortal: EventType{
		Name:     "Request Disable Portal",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	RequestEnablePortal: EventType{
		Name:     "Request Enable Portal",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	PortalDisabled: EventType{
		Name:     "Portal Disabled",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	PortalEnabled: EventType{
		Name:     "Portal Enabled",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Keypad Events
	KeypadCommandExecuted: EventType{
		Name:     "Keypad Command Executed",
		Category: CategoryAccess,
		RequiredFields: []tagName{
			EventTags.EventID,
			EventTags.EventName,
			EventTags.ReaderName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Reader Tamper Events
	ReaderTamperAlarm: EventType{
		Name:     "Reader Tamper Alarm",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	ReaderTamperNormal: EventType{
		Name:     "Reader Tamper Normal",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	ReaderBatteryAlarm: EventType{
		Name:     "Reader Battery Alarm",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	ReaderBatteryNormal: EventType{
		Name:     "Reader Battery Normal",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Blade Tamper Events
	BladeTamperAlarm: EventType{
		Name:     "Blade Tamper Alarm",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	BladeTamperNormal: EventType{
		Name:     "Blade Tamper Normal",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Manual Override
	ManualKeyOverride: EventType{
		Name:     "Manual Key Override",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Evacuation Events
	Evacuation: EventType{
		Name:     "Evacuation",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	MusteringForEvacuation: EventType{
		Name:     "Mustering For Evacuation",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.LoginAddress,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// System Health
	SystemHealth: EventType{
		Name:     "System Health",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Reader Communication Events
	ReaderCommunicationAlarm: EventType{
		Name:     "Reader Communication Alarm",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	ReaderCommunicationNormal: EventType{
		Name:     "Reader Communication Normal",
		Category: CategoryAlarm,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// FTP/NAS Configuration Events
	FTPBackupFailedConfigured: EventType{
		Name:     "FTP Backup Failed: FTP is Configured and Enabled",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NASBackupFailedConfigured: EventType{
		Name:     "NAS Backup Failed: FTP is Configured and Enabled",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	BackupCopiedToFTP: EventType{
		Name:     "Backup Successfully Copied to FTP Server",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	BackupCopiedToNAS: EventType{
		Name:     "Backup Successfully Copied to NAS Server",
		Category: CategoryBackup,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Elevator Additional Events
	ElevatorFreeAccess: EventType{
		Name:     "Elevator Free Access",
		Category: CategoryElevator,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	ElevatorAccessNotCompleted: EventType{
		Name:     "Elevator Access Not Completed",
		Category: CategoryElevator,
		RequiredFields: []tagName{
			EventTags.PersonID,
			EventTags.PersonName,
			EventTags.ReaderName,
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	EmergencyCallActivated: EventType{
		Name:     "Emergency Call Activated for Elevator",
		Category: CategoryElevator,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	EmergencyCallRestored: EventType{
		Name:     "Emergency Call Restored for Elevator",
		Category: CategoryElevator,
		RequiredFields: []tagName{
			EventTags.NDT,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Privacy/Door Events
	PrivacyEnabled: EventType{
		Name:     "Privacy Enabled",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	InteriorPushButton: EventType{
		Name:     "Interior Push Button",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	DoorBolted: EventType{
		Name:     "Door Bolted",
		Category: CategoryPortal,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// License Events
	SystemLicenseExpires60Days: EventType{
		Name:     "System License Expires in 60 Days",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	SystemLicenseExpires30Days: EventType{
		Name:     "System License Expires in 30 Days",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	SystemLicenseExpired: EventType{
		Name:     "System License Expired",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	SystemLicenseExpiredAck: EventType{
		Name:     "System License Expired Acknowledged",
		Category: CategorySystem,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeLicenseNotDetected: EventType{
		Name:     "Network Node System License Not Detected",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeLicenseNotDetected21Days: EventType{
		Name:     "Network Node System License Not Detected for 21 Days",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeLicenseNotDetected30Days: EventType{
		Name:     "Network Node System License Not Detected for 30 Days",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeLicenseReestablished: EventType{
		Name:     "Network Node System License Reestablished",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeLicenseWarningAck: EventType{
		Name:     "Network Node System License Warning Acknowledged",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeLicenseErrorAck: EventType{
		Name:     "Network Node System License Error Acknowledged",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Network Controller Events
	NetworkControllerTakeover: EventType{
		Name:     "Network Controller Takeover",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkControllerPrimary: EventType{
		Name:     "Network Controller Configured As Primary",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkControllerStandby: EventType{
		Name:     "Network Controller Configured As Standby",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},

	// Security Events
	NetworkNodeSecureConfigError: EventType{
		Name:     "Network Node Secure Configuration Error",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
	NetworkNodeSecureCommFailed: EventType{
		Name:     "Network Node Secure Communication Failed",
		Category: CategoryNetwork,
		RequiredFields: []tagName{
			EventTags.Detail,
			EventTags.NodeAddress,
			EventTags.NodeName,
			EventTags.NodeUnique,
			EventTags.PartitionName,
			EventTags.PartitionKey,
			EventTags.CDT,
		},
	},
}

// TagName represents a strongly-typed event field with its XML tag
type tagName struct {
	XMLName        string
	SupportsFilter bool
	Description    string
}

// TagNames provides type-safe access to all event fields
type TagNames struct {
	XMLName             xml.Name `xml:"TAGNAMES,omitempty"`
	ACName              tagName  `xml:"ACNAME,omitempty"`              // Access Card Hot Stamp (supports filters)
	ACNum               tagName  `xml:"ACNUM,omitempty"`               // Access Card Encoded Number (supports filters)
	ActivityID          tagName  `xml:"ACTIVITYID,omitempty"`          // Internal identifier for activity (no filters)
	AlarmID             tagName  `xml:"ALARMID,omitempty"`             // Internal identifier for alarm (no filters)
	AlarmPanelName      tagName  `xml:"ALARMPANELNAME,omitempty"`      // Configured name of the alarm panel (supports filters)
	AlarmStateName      tagName  `xml:"ALARMSTATENAME,omitempty"`      // Internal name for alarm state (Active/Escalated) (supports filters)
	AlarmTimerName      tagName  `xml:"ALARMTIMERNAME,omitempty"`      // Internal timer name for alarm state changes (supports filters)
	AlarmTransitionName tagName  `xml:"ALARMTRANSITIONNAME,omitempty"` // Internal name for alarm transition (supports filters)
	BladeSlot           tagName  `xml:"BLADESLOT,omitempty"`           // Slot number of node blade (supports filters)
	CDT                 tagName  `xml:"CDT,omitempty"`                 // Controller date/time (no filters)
	DescName            tagName  `xml:"DESCNAME,omitempty"`            // General description of the activity (supports filters)
	Detail              tagName  `xml:"DETAIL,omitempty"`              // Additional text detail about activity (supports filters)
	EventID             tagName  `xml:"EVENTID,omitempty"`             // Internal identifier for event (no filters)
	EventName           tagName  `xml:"EVTNAME,omitempty"`             // Configured event name (supports filters)
	EventPriorityNumber tagName  `xml:"EVTPRIO,omitempty"`             // Configured priority number (no filters)
	IPanelArea          tagName  `xml:"IPANELAREA,omitempty"`          // Intrusion panel area (supports filters)
	IPanelName          tagName  `xml:"IPANELNAME,omitempty"`          // Intrusion panel name (supports filters)
	IPanelOutput        tagName  `xml:"IPANELOUTPUT,omitempty"`        // Intrusion panel output (supports filters)
	IPanelUser          tagName  `xml:"IPANELUSER,omitempty"`          // Intrusion panel user (supports filters)
	IPanelZone          tagName  `xml:"IPANELZONE,omitempty"`          // Intrusion panel zone (supports filters)
	LevelKey            tagName  `xml:"LEVELKEY,omitempty"`            // Internal identifier for threat level (no filters)
	LocationKey         tagName  `xml:"LOCATIONKEY,omitempty"`         // Internal identifier for location (no filters)
	LocationName        tagName  `xml:"LOCATIONNAME,omitempty"`        // Configured location (supports filters)
	LoginAddress        tagName  `xml:"LOGINADDRESS,omitempty"`        // Host from which user logged in (supports filters)
	NodeAddress         tagName  `xml:"NODEADDRESS,omitempty"`         // IP address of node (supports filters)
	NodeName            tagName  `xml:"NODENAME,omitempty"`            // Configured name for node (supports filters)
	NodeUnique          tagName  `xml:"NODEUNIQUE,omitempty"`          // Unique identifier for node (no filters)
	NDT                 tagName  `xml:"NDT,omitempty"`                 // Node date/time (no filters)
	PartitionKey        tagName  `xml:"PARTITIONKEY,omitempty"`        // Internal identifier for partition (no filters)
	PartitionName       tagName  `xml:"PARTNAME,omitempty"`            // Partition name (supports filters)
	PersonName          tagName  `xml:"PERSONNAME,omitempty"`          // Configured person name (supports filters)
	PersonID            tagName  `xml:"PERSONID,omitempty"`            // Configured person identifier (no filters)
	PortalKey           tagName  `xml:"PORTALKEY,omitempty"`           // Internal identifier for portal (no filters)
	PortalName          tagName  `xml:"PORTALNAME,omitempty"`          // Configured portal name (supports filters)
	ReaderName          tagName  `xml:"RDRNAME,omitempty"`             // Configured card reader name (supports filters)
	ReaderKey           tagName  `xml:"READERKEY,omitempty"`           // Internal identifier for reader (no filters)
	Reader2Key          tagName  `xml:"READER2KEY,omitempty"`          // Internal identifier for second reader (no filters)
	ThreatName          tagName  `xml:"THREATNAME,omitempty"`          // Name of threat level (supports filters)
	UCBitLength         tagName  `xml:"UCBITLENGTH,omitempty"`         // Number of bits in card format (no filters)
}

// Event Tags is the singleton instance for type-safe field access
var EventTags = TagNames{
	// Access Card Fields
	ACName: tagName{
		XMLName:        "ACNAME",
		SupportsFilter: true,
		Description:    "Access Card Hot Stamp",
	},
	ACNum: tagName{
		XMLName:        "ACNUM",
		SupportsFilter: true,
		Description:    "Access Card Encoded Number",
	},

	// Activity Fields
	ActivityID: tagName{
		XMLName:        "ACTIVITYID",
		SupportsFilter: false,
		Description:    "Internal identifier for the activity",
	},

	// Alarm Fields
	AlarmID: tagName{
		XMLName:        "ALARMID",
		SupportsFilter: false,
		Description:    "Internal identifier for the alarm",
	},
	AlarmPanelName: tagName{
		XMLName:        "ALARMPANELNAME",
		SupportsFilter: true,
		Description:    "Configured name of the alarm panel",
	},
	AlarmStateName: tagName{
		XMLName:        "ALARMSTATENAME",
		SupportsFilter: true,
		Description:    "Internal name for alarm state (Active/Escalated)",
	},
	AlarmTimerName: tagName{
		XMLName:        "ALARMTIMERNAME",
		SupportsFilter: true,
		Description:    "Internal timer name for alarm state changes",
	},
	AlarmTransitionName: tagName{
		XMLName:        "ALARMTRANSITIONNAME",
		SupportsFilter: true,
		Description:    "Internal name for alarm transition",
	},

	// Node Fields
	BladeSlot: tagName{
		XMLName:        "BLADESLOT",
		SupportsFilter: true,
		Description:    "Slot number of node blade",
	},
	NodeAddress: tagName{
		XMLName:        "NODEADDRESS",
		SupportsFilter: true,
		Description:    "IP address of the node",
	},
	NodeName: tagName{
		XMLName:        "NODENAME",
		SupportsFilter: true,
		Description:    "Configured name for the node",
	},
	NodeUnique: tagName{
		XMLName:        "NODEUNIQUE",
		SupportsFilter: false,
		Description:    "Unique identifier for the node",
	},
	NDT: tagName{
		XMLName:        "NDT",
		SupportsFilter: false,
		Description:    "Node date/time",
	},

	// Time Fields
	CDT: tagName{
		XMLName:        "CDT",
		SupportsFilter: false,
		Description:    "Controller date/time",
	},

	// Description Fields
	DescName: tagName{
		XMLName:        "DESCNAME",
		SupportsFilter: true,
		Description:    "General description of the activity",
	},
	Detail: tagName{
		XMLName:        "DETAIL",
		SupportsFilter: true,
		Description:    "Additional text detail about activity",
	},

	// Event Fields
	EventID: tagName{
		XMLName:        "EVENTID",
		SupportsFilter: false,
		Description:    "Internal identifier for the event",
	},
	EventName: tagName{
		XMLName:        "EVTNAME",
		SupportsFilter: true,
		Description:    "Configured name of the event",
	},
	EventPriorityNumber: tagName{
		XMLName:        "EVTPRIO",
		SupportsFilter: false,
		Description:    "Configured priority number",
	},

	// Intrusion Panel Fields
	IPanelArea: tagName{
		XMLName:        "IPANELAREA",
		SupportsFilter: true,
		Description:    "Intrusion panel area",
	},
	IPanelName: tagName{
		XMLName:        "IPANELNAME",
		SupportsFilter: true,
		Description:    "Intrusion panel name",
	},
	IPanelOutput: tagName{
		XMLName:        "IPANELOUTPUT",
		SupportsFilter: true,
		Description:    "Intrusion panel output",
	},
	IPanelUser: tagName{
		XMLName:        "IPANELUSER",
		SupportsFilter: true,
		Description:    "Intrusion panel user",
	},
	IPanelZone: tagName{
		XMLName:        "IPANELZONE",
		SupportsFilter: true,
		Description:    "Intrusion panel zone",
	},

	// Location Fields
	LocationKey: tagName{
		XMLName:        "LOCATIONKEY",
		SupportsFilter: false,
		Description:    "Internal identifier for location",
	},
	LocationName: tagName{
		XMLName:        "LOCATIONNAME",
		SupportsFilter: true,
		Description:    "Configured location name",
	},

	// Login Fields
	LoginAddress: tagName{
		XMLName:        "LOGINADDRESS",
		SupportsFilter: true,
		Description:    "Host from which user logged in",
	},

	// Partition Fields
	PartitionKey: tagName{
		XMLName:        "PARTITIONKEY",
		SupportsFilter: false,
		Description:    "Internal identifier for partition",
	},
	PartitionName: tagName{
		XMLName:        "PARTNAME",
		SupportsFilter: true,
		Description:    "Partition name",
	},

	// Person Fields
	PersonID: tagName{
		XMLName:        "PERSONID",
		SupportsFilter: false,
		Description:    "Configured person identifier",
	},
	PersonName: tagName{
		XMLName:        "PERSONNAME",
		SupportsFilter: true,
		Description:    "Configured person name",
	},

	// Portal Fields
	PortalKey: tagName{
		XMLName:        "PORTALKEY",
		SupportsFilter: false,
		Description:    "Internal identifier for portal",
	},
	PortalName: tagName{
		XMLName:        "PORTALNAME",
		SupportsFilter: true,
		Description:    "Configured name of the portal",
	},

	// Reader Fields
	ReaderName: tagName{
		XMLName:        "RDRNAME",
		SupportsFilter: true,
		Description:    "Configured name of card reader",
	},
	ReaderKey: tagName{
		XMLName:        "READERKEY",
		SupportsFilter: false,
		Description:    "Internal identifier for reader",
	},
	Reader2Key: tagName{
		XMLName:        "READER2KEY",
		SupportsFilter: false,
		Description:    "Internal identifier for second reader",
	},

	// Threat Level Fields
	LevelKey: tagName{
		XMLName:        "LEVELKEY",
		SupportsFilter: false,
		Description:    "Internal identifier for threat level",
	},
	ThreatName: tagName{
		XMLName:        "THREATNAME",
		SupportsFilter: true,
		Description:    "Name of the threat level",
	},

	// Misc Fields
	UCBitLength: tagName{
		XMLName:        "UCBITLENGTH",
		SupportsFilter: false,
		Description:    "Number of bits in card format",
	},
}

type TagFilter struct {
	Filters []string `xml:"FILTERS>FILTER"`
}

// StreamEventsBuilder provides a fluent interface for building event stream requests
type StreamEventsBuilder struct {
	fields  map[string]tagName
	filters map[string][]string
}

// NewStreamEventsBuilder creates a new builder
func NewStreamEventsBuilder() *StreamEventsBuilder {
	return &StreamEventsBuilder{
		fields:  make(map[string]tagName),
		filters: make(map[string][]string),
	}
}

// WithField adds a field to monitor
func (b *StreamEventsBuilder) WithField(field tagName) *StreamEventsBuilder {
	b.fields[field.XMLName] = field
	return b
}

// WithFields adds multiple fields to monitor
func (b *StreamEventsBuilder) WithFields(fields ...tagName) *StreamEventsBuilder {
	for _, field := range fields {
		b.fields[field.XMLName] = field
	}
	return b
}

// WithFilter adds a filter for a field (automatically adds the field if not present)
func (b *StreamEventsBuilder) WithFilter(field tagName, values ...string) *StreamEventsBuilder {
	if !field.SupportsFilter {
		// Silently ignore filters for fields that don't support them
		// Could also log a warning here if desired
		return b
	}

	// Ensure the field is added
	b.fields[field.XMLName] = field

	// Add the filter values
	if b.filters[field.XMLName] == nil {
		b.filters[field.XMLName] = []string{}
	}
	b.filters[field.XMLName] = append(b.filters[field.XMLName], values...)

	return b
}

// WithEventType adds all required fields for a specific event type
func (b *StreamEventsBuilder) WithEventType(eventType EventType) *StreamEventsBuilder {
	for _, field := range eventType.RequiredFields {
		b.fields[field.XMLName] = field
	}
	// Also add the DESCNAME field to capture the event type
	b.fields[EventTags.DescName.XMLName] = EventTags.DescName
	return b
}

// WithEventTypes adds fields for multiple event types
func (b *StreamEventsBuilder) WithEventTypes(eventTypes ...EventType) *StreamEventsBuilder {
	for _, et := range eventTypes {
		b.WithEventType(et)
	}
	return b
}

// WithCategory adds all events from a category
func (b *StreamEventsBuilder) WithCategory(category EventCategory) *StreamEventsBuilder {
	// Always include DESCNAME to identify events
	b.fields[EventTags.DescName.XMLName] = EventTags.DescName

	// Collect all unique fields from events in this category
	v := reflect.ValueOf(EventTypes)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if eventType, ok := field.Interface().(EventType); ok {
			if eventType.Category == category {
				for _, reqField := range eventType.RequiredFields {
					b.fields[reqField.XMLName] = reqField
				}
			}
		}
	}
	return b
}

// FilterByPersonName adds a person name filter
func (b *StreamEventsBuilder) FilterByPersonName(names ...string) *StreamEventsBuilder {
	return b.WithFilter(EventTags.PersonName, names...)
}

// FilterByPortalName adds a portal name filter
func (b *StreamEventsBuilder) FilterByPortalName(portals ...string) *StreamEventsBuilder {
	return b.WithFilter(EventTags.PortalName, portals...)
}

// FilterByNodeName adds a node name filter
func (b *StreamEventsBuilder) FilterByNodeName(nodes ...string) *StreamEventsBuilder {
	return b.WithFilter(EventTags.NodeName, nodes...)
}

// FilterByEventName adds an event name filter
func (b *StreamEventsBuilder) FilterByEventName(events ...string) *StreamEventsBuilder {
	return b.WithFilter(EventTags.EventName, events...)
}

// FilterByDescName adds a description filter (for filtering by event type names)
func (b *StreamEventsBuilder) FilterByDescName(descriptions ...string) *StreamEventsBuilder {
	return b.WithFilter(EventTags.DescName, descriptions...)
}

// Build creates the final parameters structure for the API call
func (b *StreamEventsBuilder) Build() *StreamEventsParams {
	// Create a map to hold our field definitions
	fields := make(map[string]any)

	// Process each field
	for xmlName, field := range b.fields {
		// Check if this field has filters
		if filterValues, hasFilters := b.filters[xmlName]; hasFilters && len(filterValues) > 0 {
			// Field with filters
			fields[xmlName] = &struct {
				XMLName xml.Name `xml:""`
				*TagFilter
			}{
				XMLName: xml.Name{Local: field.XMLName},
				TagFilter: &TagFilter{
					Filters: filterValues,
				},
			}
		} else {
			// Field without filters (empty tag)
			fields[xmlName] = &struct {
				XMLName xml.Name `xml:""`
			}{
				XMLName: xml.Name{Local: field.XMLName},
			}
		}
	}

	// If no fields specified, include some defaults for basic functionality
	if len(fields) == 0 {
		fields["DESCNAME"] = &struct {
			XMLName xml.Name `xml:""`
		}{
			XMLName: xml.Name{Local: "DESCNAME"},
		}

		fields["CDT"] = &struct {
			XMLName xml.Name `xml:""`
		}{
			XMLName: xml.Name{Local: "CDT"},
		}

		fields["PERSONNAME"] = &struct {
			XMLName xml.Name `xml:""`
		}{
			XMLName: xml.Name{Local: "PERSONNAME"},
		}
		fields["PORTALNAME"] = &struct {
			XMLName xml.Name `xml:""`
		}{
			XMLName: xml.Name{Local: "PORTALNAME"},
		}
	}

	// Create the wrapper structure
	return &StreamEventsParams{
		XMLName: xml.Name{Local: "PARAMS"},
		TagNames: &DynamicTagNames{
			Fields: fields,
		},
	}
}

// StreamEventsParams wraps the tag names for the request
type StreamEventsParams struct {
	XMLName  xml.Name         `xml:"PARAMS,omitempty"`
	TagNames *DynamicTagNames `xml:"TAGNAMES"`
}

// DynamicTagNames holds the dynamic field definitions
type DynamicTagNames struct {
	XMLName xml.Name               `xml:"TAGNAMES"`
	Fields  map[string]interface{} `xml:"-"`
}

// MarshalXML implements custom XML marshaling for DynamicTagNames
func (d *DynamicTagNames) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Start the TAGNAMES element
	if err := e.EncodeToken(xml.StartElement{Name: xml.Name{Local: "TAGNAMES"}}); err != nil {
		return err
	}

	// Sort field names for consistent output
	var fieldNames []string
	for name := range d.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	// Encode each field
	for _, name := range fieldNames {
		if err := e.Encode(d.Fields[name]); err != nil {
			return err
		}
	}

	// End the TAGNAMES element
	if err := e.EncodeToken(xml.EndElement{Name: xml.Name{Local: "TAGNAMES"}}); err != nil {
		return err
	}

	return nil
}

// END OF LENEL S2 (NETBOX) EVENT STRUCTS
//---------------------------------------------------------------------

// apiPatination maps some XML elements to see if they are used for pagination
// This is used to determine if the next page of results should be requested
// when the API returns a response with a next page of results.
var apiPagination = map[string]bool{
	"NEXTKEY": true,
}

// ### Enums
// --------------------------------------------------------------------
// netboxAPI is the top-level struct that nests all command categories.
type netboxAPI struct {
	Actions       netboxActions
	Configuration netboxConfiguration
	Events        netboxEvents
	History       netboxHistory
	People        netboxPeople
	Portals       netboxPortals
	ThreatLevels  netboxThreatLevels
	Utility       netboxUtility
}

// Actions represents commands in the "Actions" category.
type netboxActions struct {
	ActivateOutput        string
	DeactivateOutput      string
	DogOnNextExitPortal   string
	LockPortal            string
	MomentaryUnlockPortal string
	SetThreatLevel        string
	UnlockPortal          string
}

// Configuration represents commands in the "Configuration" category.
type netboxConfiguration struct {
	AddAccessLevel         string
	AddAccessLevelGroup    string
	AddHoliday             string
	AddPartition           string
	AddPortalGroup         string
	AddReaderGroup         string
	AddTimeSpec            string
	AddTimeSpecGroup       string
	AddThreatLevel         string
	AddThreatLevelGroup    string
	DeleteAccessLevel      string
	DeleteAccessLevelGroup string
	DeleteHoliday          string
	DeletePortalGroup      string
	DeleteReaderGroup      string
	DeleteTimeSpec         string
	GetAccessLevel         string
	GetAccessLevels        string
	GetAccessLevelGroup    string
	GetAccessLevelGroups   string
	GetAccessLevelNames    string
	GetCardFormats         string
	GetElevators           string
	GetFloors              string
	GetHoliday             string
	GetHolidays            string
	GetOutputs             string
	GetPartitions          string
	GetPortalGroup         string
	GetPortalGroups        string
	GetReaderGroup         string
	GetReaderGroups        string
	GetReaders             string
	GetTimeSpecGroup       string
	GetTimeSpecGroups      string
	GetTimeSpecs           string
	GetUDFLists            string
	GetUDFListItems        string
	ModifyAccessLevelGroup string
	ModifyHoliday          string
	ModifyPortalGroup      string
	ModifyReaderGroup      string
	ModifyThreatLevel      string
	ModifyThreatLevelGroup string
	ModifyTimeSpec         string
	ModifyTimeSpecGroup    string
	ModifyUDFListItems     string
	RemoveThreatLevel      string
	RemoveThreatLevelGroup string
	SetThreatLevel         string
}

// Events represents commands in the "Events" category.
type netboxEvents struct {
	ListEvents   string
	StreamEvents string
	TriggerEvent string
}

// History represents commands in the "History" category.
type netboxHistory struct {
	GetEventHistory      string
	GetCardAccessDetails string
}

// People represents commands in the "People" category.
type netboxPeople struct {
	AddAccessLevelGroup string
	AddCredential       string // The AddCredential command adds a credential to a person record in the system database.
	AddPerson           string // The AddPerson command allows you to add a new person record
	GetAccessLevelNames string
	GetPerson           string
	GetPicture          string
	ModifyAccessLevel   string
	ModifyCredential    string
	ModifyPerson        string
	RemoveCredential    string
	RemovePerson        string
	SearchPersonData    string
}

// Portals represents commands in the "Portals" category.
type netboxPortals struct {
	AddPortalGroup    string
	AddReaderGroup    string
	DeletePortalGroup string
	DeleteReaderGroup string
	GetPortalGroup    string
	GetPortalGroups   string
	GetReader         string
	GetReaders        string
	GetReaderGroup    string
	GetReaderGroups   string
	ModifyPortalGroup string
	ModifyReaderGroup string
}

// ThreatLevels represents commands in the "Threat Levels" category.
type netboxThreatLevels struct {
	AddThreatLevel         string
	AddThreatLevelGroup    string
	ModifyThreatLevel      string
	ModifyThreatLevelGroup string
	SetThreatLevel         string
}

// Utility represents commands in the "Utility" category.
type netboxUtility struct {
	GetAPIVersion   string
	GetPartitions   string
	Login           string
	Logout          string
	PingApp         string
	SwitchPartition string
}

// Commands is an instance of NBAPICommands populated with string constants for each command.
var NetboxCommands = netboxAPI{
	Actions: netboxActions{
		ActivateOutput:        "Activate Output",
		DeactivateOutput:      "Deactivate Output",
		DogOnNextExitPortal:   "DogOnNextExitPortal",
		LockPortal:            "LockPortal",
		MomentaryUnlockPortal: "MomentaryUnlockPortal",
		SetThreatLevel:        "SetThreatLevel",
		UnlockPortal:          "UnlockPortal",
	},
	Configuration: netboxConfiguration{
		AddAccessLevel:         "AddAccessLevel",
		AddAccessLevelGroup:    "AddAccessLevelGroup",
		AddHoliday:             "AddHoliday",
		AddPartition:           "AddPartition",
		AddPortalGroup:         "AddPortalGroup",
		AddReaderGroup:         "AddReaderGroup",
		AddTimeSpec:            "AddTimeSpec",
		AddTimeSpecGroup:       "AddTimeSpecGroup",
		AddThreatLevel:         "AddThreatLevel",
		AddThreatLevelGroup:    "AddThreatLevelGroup",
		DeleteAccessLevel:      "DeleteAccessLevel",
		DeleteAccessLevelGroup: "DeleteAccessLevelGroup",
		DeleteHoliday:          "DeleteHoliday",
		DeletePortalGroup:      "DeletePortalGroup",
		DeleteReaderGroup:      "DeleteReaderGroup",
		DeleteTimeSpec:         "DeleteTimeSpec",
		GetAccessLevel:         "GetAccessLevel",
		GetAccessLevels:        "GetAccessLevels",
		GetAccessLevelGroup:    "GetAccessLevelGroup",
		GetAccessLevelGroups:   "GetAccessLevelGroups",
		GetAccessLevelNames:    "GetAccessLevelNames",
		GetCardFormats:         "GetCardFormats",
		GetElevators:           "GetElevators",
		GetFloors:              "GetFloors",
		GetHoliday:             "GetHoliday",
		GetHolidays:            "GetHolidays",
		GetOutputs:             "GetOutputs",
		GetPartitions:          "GetPartitions",
		GetPortalGroup:         "GetPortalGroup",
		GetPortalGroups:        "GetPortalGroups",
		GetReaderGroup:         "GetReaderGroup",
		GetReaderGroups:        "GetReaderGroups",
		GetReaders:             "GetReaders",
		GetTimeSpecGroup:       "GetTimeSpecGroup",
		GetTimeSpecGroups:      "GetTimeSpecGroups",
		GetTimeSpecs:           "GetTimeSpecs",
		GetUDFLists:            "GetUDFLists",
		GetUDFListItems:        "GetUDFListItems",
		ModifyAccessLevelGroup: "ModifyAccessLevelGroup",
		ModifyHoliday:          "ModifyHoliday",
		ModifyPortalGroup:      "ModifyPortalGroup",
		ModifyReaderGroup:      "ModifyReaderGroup",
		ModifyThreatLevel:      "ModifyThreatLevel",
		ModifyThreatLevelGroup: "ModifyThreatLevelGroup",
		ModifyTimeSpec:         "ModifyTimeSpec",
		ModifyTimeSpecGroup:    "ModifyTimeSpecGroup",
		ModifyUDFListItems:     "ModifyUDFListItems",
		RemoveThreatLevel:      "RemoveThreatLevel",
		RemoveThreatLevelGroup: "RemoveThreatLevelGroup",
		SetThreatLevel:         "SetThreatLevel",
	},
	Events: netboxEvents{
		ListEvents:   "ListEvents",
		StreamEvents: "StreamEvents",
		TriggerEvent: "TriggerEvent",
	},
	History: netboxHistory{
		GetEventHistory:      "GetEventHistory",
		GetCardAccessDetails: "GetCardAccessDetails",
	},
	People: netboxPeople{
		AddAccessLevelGroup: "AddAccessLevelGroup",
		AddCredential:       "AddCredential",
		AddPerson:           "AddPerson",
		GetAccessLevelNames: "GetAccessLevelNames",
		GetPerson:           "GetPerson",
		GetPicture:          "GetPicture",
		ModifyAccessLevel:   "ModifyAccessLevel",
		ModifyCredential:    "ModifyCredential",
		ModifyPerson:        "ModifyPerson",
		RemoveCredential:    "RemoveCredential",
		RemovePerson:        "RemovePerson",
		SearchPersonData:    "SearchPersonData",
	},
	Portals: netboxPortals{
		AddPortalGroup:    "AddPortalGroup",
		AddReaderGroup:    "AddReaderGroup",
		DeletePortalGroup: "DeletePortalGroup",
		DeleteReaderGroup: "DeleteReaderGroup",
		GetPortalGroup:    "GetPortalGroup",
		GetPortalGroups:   "GetPortalGroups",
		GetReader:         "GetReader",
		GetReaders:        "GetReaders",
		GetReaderGroup:    "GetReaderGroup",
		GetReaderGroups:   "GetReaderGroups",
		ModifyPortalGroup: "ModifyPortalGroup",
		ModifyReaderGroup: "ModifyReaderGroup",
	},
	ThreatLevels: netboxThreatLevels{
		AddThreatLevel:         "AddThreatLevel",
		AddThreatLevelGroup:    "AddThreatLevelGroup",
		ModifyThreatLevel:      "ModifyThreatLevel",
		ModifyThreatLevelGroup: "ModifyThreatLevelGroup",
		SetThreatLevel:         "SetThreatLevel",
	},
	Utility: netboxUtility{
		GetAPIVersion:   "GetAPIVersion",
		GetPartitions:   "GetPartitions",
		Login:           "Login",
		Logout:          "Logout",
		PingApp:         "PingApp",
		SwitchPartition: "SwitchPartition",
	},
}

// apiErrors maps API error codes to their description.
var apiErrors = map[int]string{
	1: "The API failed to initialize.",
	2: "The API is not enabled on the system.",
	3: "The call contains an invalid API command.",
	4: "The API was unable to parse the command request.",
	5: "There was an authentication failure. Refer to options for configuring authentication to work with the API.",
	6: "The XML code contains an unknown command. Check the syntax of the command request.",
}

// END OF LENEL S2 (NETBOX) ENUMS
//---------------------------------------------------------------------
