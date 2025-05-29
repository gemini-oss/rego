// pkg/lenel_s2/entities.go
package lenel_s2

import (
	"encoding/xml"
	"io"

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

// ### Lenel S2 People Structs
// ---------------------------------------------------------------------

type People struct {
	People []*Person `xml:"PEOPLE>PERSON"`
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

func (ah *AccessHistory) SetCommand(nb *Client, cmd NetboxCommand) NetboxCommand {
	newParams := []any{
		cmd.Command.Params,
		struct {
			StartLogID string `xml:"STARTLOGID"`
		}{
			StartLogID: ah.NextLogID,
		},
	}

	return nb.BuildRequest(
		cmd.Command.Name,
		newParams,
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

type Events struct {
	ActivityID string `xml:"ACTIVITYID"` // Activity ID of the event
	DESCNAME   string `xml:"DESCNAME"`   // Description of the event
	CDT        string `xml:"CDT"`        // CDT of the event
	PARTNAME   string `xml:"PARTNAME"`   // Partition name of the event
}

type TagNames struct {
	ACName       string   `xml:"ACNAME,omitempty"`
	ActivityID   string   `xml:"ACTIVITYID,omitempty"`
	CDT          string   `xml:"CDT,omitempty"`
	DescName     string   `xml:"DESCNAME,omitempty"`
	LoginAddress string   `xml:"LOGINADDRESS,omitempty"`
	PersonName   string   `xml:"PERSONNAME,omitempty"`
	Detail       string   `xml:"DETAIL,omitempty"`
	PartName     PartName `xml:"PARTNAME,omitempty"`
}

type PartName struct {
	Filters []string `xml:"FILTERS>FILTER"`
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
