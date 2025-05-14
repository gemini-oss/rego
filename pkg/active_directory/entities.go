/*
# Active Directory - Entities

This package contains various structs for handling Active Directory objects like Users, Groups, etc.
- https://learn.microsoft.com/en-us/windows/win32/adschema/classes-all

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/active_directory/entities.go
package active_directory

import (
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/log"

	"github.com/go-ldap/ldap/v3"
)

// ### Active Directory Client Entities
// ---------------------------------------------------------------------
// Client holds Active Directory client data
type Client struct {
	Server   string
	BaseDN   string
	username string
	password string
	LDAP     *ldap.Conn
	Log      *log.Logger
	Cache    *cache.Cache
}

// Slice is an interface that ensures T is a slice type.
type Slice[T any] interface {
	~[]T
}

// END OF ACTIVE DIRECTORY CLIENT ENTITIES
//---------------------------------------------------------------------

// ### Active Directory Entities
// ---------------------------------------------------------------------
type Users []*User

// User represents an AD user with detailed fields (AKA: Contact)
// https://learn.microsoft.com/en-us/windows/win32/adschema/c-user
type User struct {
	AccountExpires             time.Time   `ldap:"accountExpires"`             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-accountexpires
	AdminCount                 int         `ldap:"adminCount"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-admincount
	AltSecurityIdentities      []string    `ldap:"altSecurityIdentities"`      // https://learn.microsoft.com/en-us/windows/win32/adschema/a-altsecurityidentities
	BadPasswordTime            int64       `ldap:"badPasswordTime"`            // https://learn.microsoft.com/en-us/windows/win32/adschema/a-badpasswordtime
	BadPwdCount                int         `ldap:"badPwdCount"`                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-badpwdcount
	City                       string      `ldap:"l"`                          // https://learn.microsoft.com/en-us/windows/win32/adschema/a-l
	CodePage                   int         `ldap:"codePage"`                   // https://learn.microsoft.com/en-us/windows/win32/adschema/a-codepage
	CommonName                 string      `ldap:"cn"`                         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-cn
	Country                    string      `ldap:"c"`                          // https://learn.microsoft.com/en-us/windows/win32/adschema/a-c
	CountryCode                int         `ldap:"countryCode"`                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-countrycode
	Department                 string      `ldap:"department"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-department
	DistinguishedName          string      `ldap:"dn"`                         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-distinguishedName
	DSCorePropagationData      []time.Time `ldap:"dSCorePropagationData"`      // https://learn.microsoft.com/en-us/windows/win32/adschema/a-dscorepropagationdata
	DisplayName                string      `ldap:"displayName"`                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-displayname
	Division                   string      `ldap:"division"`                   // https://learn.microsoft.com/en-us/windows/win32/adschema/a-division
	EmployeeID                 string      `ldap:"employeeID"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-employeeid
	EmployeeNumber             string      `ldap:"employeeNumber"`             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-employeenumber
	GivenName                  string      `ldap:"givenName"`                  // https://learn.microsoft.com/en-us/windows/win32/adschema/a-givenname
	InstanceType               int         `ldap:"instanceType"`               // https://learn.microsoft.com/en-us/windows/win32/adschema/a-instancetype
	LastLogoff                 time.Time   `ldap:"lastLogoff"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-lastlogoff
	LastLogon                  time.Time   `ldap:"lastLogon"`                  // https://learn.microsoft.com/en-us/windows/win32/adschema/a-lastlogon
	LastLogonTimestamp         time.Time   `ldap:"lastLogonTimestamp"`         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-lastlogontimestamp
	Manager                    string      `ldap:"manager"`                    // https://learn.microsoft.com/en-us/windows/win32/adschema/a-manager
	Mail                       string      `ldap:"mail"`                       // https://learn.microsoft.com/en-us/windows/win32/adschema/a-mail
	MemberOf                   []string    `ldap:"memberOf"`                   // https://learn.microsoft.com/en-us/windows/win32/adschema/a-memberof
	Mobile                     string      `ldap:"mobile"`                     // https://learn.microsoft.com/en-us/windows/win32/adschema/a-mobile
	Name                       string      `ldap:"name"`                       // https://learn.microsoft.com/en-us/windows/win32/adschema/a-name
	ObjectCategory             string      `ldap:"objectCategory"`             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-objectcategory
	ObjectClass                []string    `ldap:"objectClass"`                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-objectclass
	ObjectGUID                 string      `ldap:"objectGUID"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-objectguid
	ObjectSID                  string      `ldap:"objectSid"`                  // https://learn.microsoft.com/en-us/windows/win32/adschema/a-objectsid
	PhysicalDeliveryOfficeName string      `ldap:"physicalDeliveryOfficeName"` // https://learn.microsoft.com/en-us/windows/win32/adschema/a-physicaldeliveryofficename
	PostalCode                 string      `ldap:"postalCode"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-postalcode
	PwdLastSet                 time.Time   `ldap:"pwdLastSet"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-pwdlastset
	ReplPropertyMetaData       string      `ldap:"replPropertyMetaData"`       // https://learn.microsoft.com/en-us/windows/win32/adschema/a-replpropertymetadata
	SAMAccountName             string      `ldap:"sAMAccountName"`             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-samaccountname
	SAMAccountType             int         `ldap:"sAMAccountType"`             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-samaccounttype
	SN                         string      `ldap:"sn"`                         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-sn
	StreetAddress              string      `ldap:"streetAddress"`              // https://learn.microsoft.com/en-us/windows/win32/adschema/a-streetaddress
	TelephoneNumber            string      `ldap:"telephoneNumber"`            // https://learn.microsoft.com/en-us/windows/win32/adschema/a-telephonenumber
	Title                      string      `ldap:"title"`                      // https://learn.microsoft.com/en-us/windows/win32/adschema/a-title
	UserAccountControl         int         `ldap:"userAccountControl"`         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-useraccountcontrol
	UserPrincipalName          string      `ldap:"userPrincipalName"`          // https://learn.microsoft.com/en-us/windows/win32/adschema/a-userprincipalname
	USNChanged                 string      `ldap:"uSNChanged"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-usnchanged
	USNCreated                 string      `ldap:"uSNCreated"`                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-usncreated
	WhenChanged                time.Time   `ldap:"whenChanged"`                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-whenchanged
	WhenCreated                time.Time   `ldap:"whenCreated"`                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-whencreated
}

type Groups []*Group

type Group struct {
	DN          string    `ldap:"dn,omitempty"`
	CommonName  string    `ldap:"commonName,omitempty"`
	Description string    `ldap:"description,omitempty"`
	Members     []string  `ldap:"members,omitempty"`
	ManagedBy   string    `ldap:"managedBy,omitempty"`
	WhenCreated time.Time `ldap:"whenCreated,omitempty"`
	WhenChanged time.Time `ldap:"whenChanged,omitempty"`
}

type Computers []*Computer

// Computer represents an AD computer account
type Computer struct {
	DN                string    `ldap:"dn,omitempty"`
	CommonName        string    `ldap:"cn,omitempty"`
	SAMAccountName    string    `ldap:"sAMAccountName,omitempty"`
	DistinguishedName string    `ldap:"distinguishedName,omitempty"`
	OperatingSystem   string    `ldap:"operatingSystem,omitempty"`
	WhenCreated       time.Time `ldap:"whenCreated,omitempty"`
	WhenChanged       time.Time `ldap:"whenChanged,omitempty"`
}

type OUs []*OrganizationalUnit

// OrganizationalUnit represents an AD Organizational Unit
type OrganizationalUnit struct {
	DN                string    `ldap:"dn,omitempty"`
	Name              string    `ldap:"name,omitempty"`
	DistinguishedName string    `ldap:"distinguishedName,omitempty"`
	Description       string    `ldap:"description,omitempty"`
	WhenCreated       time.Time `ldap:"whenCreated,omitempty"`
	WhenChanged       time.Time `ldap:"whenChanged,omitempty"`
}

// END OF ACTIVE DIRECTORY OBJECT CLASS ENTITIES
//---------------------------------------------------------------------

// ### Enums
// --------------------------------------------------------------------
// Inteded for LDAP Query parameters
// https://learn.microsoft.com/en-us/windows/win32/adschema/attributes-all

// Attribute holds possible LDAP attribute constants
type Attribute string

// Shared attributes
const (
	CommonName         Attribute = "cn"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-cn
	Description        Attribute = "description"        // https://learn.microsoft.com/en-us/windows/win32/adschema/a-description
	DN                 Attribute = "dn"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-distinguishedName
	DisplayName        Attribute = "displayName"        // https://learn.microsoft.com/en-us/windows/win32/adschema/a-displayname
	DistinguishedName  Attribute = "distinguishedName"  // https://learn.microsoft.com/en-us/windows/win32/adschema/a-distinguishedName
	LastLogon          Attribute = "lastLogon"          // https://learn.microsoft.com/en-us/windows/win32/adschema/a-lastlogon
	LastLogonTimestamp Attribute = "lastLogonTimestamp" // https://learn.microsoft.com/en-us/windows/win32/adschema/a-lastlogontimestamp
	ObjectCategory     Attribute = "objectCategory"     // https://learn.microsoft.com/en-us/windows/win32/adschema/a-objectcategory
	ObjectClass        Attribute = "objectClass"        // https://learn.microsoft.com/en-us/windows/win32/adschema/a-objectclass
	ObjectGUID         Attribute = "objectGUID"         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-objectguid
	ObjectSID          Attribute = "objectSid"          // https://learn.microsoft.com/en-us/windows/win32/adschema/a-objectsid
	WhenChanged        Attribute = "whenChanged"        // https://learn.microsoft.com/en-us/windows/win32/adschema/a-whenchanged
	WhenCreated        Attribute = "whenCreated"        // https://learn.microsoft.com/en-us/windows/win32/adschema/a-whencreated
)

// Computer attributes
const (
	DNSHostName                Attribute = "dNSHostName"                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-dnshostname
	OperatingSystem            Attribute = "operatingSystem"            // https://learn.microsoft.com/en-us/windows/win32/adschema/a-operatingsystem
	OperatingSystemServicePack Attribute = "operatingSystemServicePack" // https://learn.microsoft.com/en-us/windows/win32/adschema/a-operatingsystemservicepack
	OperatingSystemVersion     Attribute = "operatingSystemVersion"     // https://learn.microsoft.com/en-us/windows/win32/adschema/a-operatingsystemversion
	ServicePrincipalName       Attribute = "servicePrincipalName"       // https://learn.microsoft.com/en-us/windows/win32/adschema/a-serviceprincipalname
)

// Group attributes
const (
	GroupMember Attribute = "member"    // https://learn.microsoft.com/en-us/windows/win32/adschema/a-member
	GroupType   Attribute = "groupType" // https://learn.microsoft.com/en-us/windows/win32/adschema/a-grouptype
	ManagedBy   Attribute = "managedBy" // https://learn.microsoft.com/en-us/windows/win32/adschema/a-managedby
)

// Organizational Unit attributes
const (
	OrganizationName Attribute = "o"  // https://learn.microsoft.com/en-us/windows/win32/adschema/a-o
	OU               Attribute = "ou" // https://learn.microsoft.com/en-us/windows/win32/adschema/a-ou
)

// User attributes
const (
	AccountExpires             Attribute = "accountExpires"             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-accountexpires
	AdminCount                 Attribute = "adminCount"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-admincount
	AltSecurityIdentities      Attribute = "altSecurityIdentities"      // https://learn.microsoft.com/en-us/windows/win32/adschema/a-altsecurityidentities
	BadPasswordTime            Attribute = "badPasswordTime"            // https://learn.microsoft.com/en-us/windows/win32/adschema/a-badpasswordtime
	BadPwdCount                Attribute = "badPwdCount"                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-badpwdcount
	City                       Attribute = "l"                          // https://learn.microsoft.com/en-us/windows/win32/adschema/a-l
	CodePage                   Attribute = "codePage"                   // https://learn.microsoft.com/en-us/windows/win32/adschema/a-codepage
	Country                    Attribute = "c"                          // https://learn.microsoft.com/en-us/windows/win32/adschema/a-c
	CountryCode                Attribute = "countryCode"                // https://learn.microsoft.com/en-us/windows/win32/adschema/a-countrycode
	Department                 Attribute = "department"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-department
	DSCorePropagationData      Attribute = "dSCorePropagationData"      // https://learn.microsoft.com/en-us/windows/win32/adschema/a-dscorepropagationdata
	Division                   Attribute = "division"                   // https://learn.microsoft.com/en-us/windows/win32/adschema/a-division
	EmployeeID                 Attribute = "employeeID"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-employeeid
	EmployeeNumber             Attribute = "employeeNumber"             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-employeenumber
	GivenName                  Attribute = "givenName"                  // https://learn.microsoft.com/en-us/windows/win32/adschema/a-givenname
	InstanceType               Attribute = "instanceType"               // https://learn.microsoft.com/en-us/windows/win32/adschema/a-instancetype
	LastLogoff                 Attribute = "lastLogoff"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-lastlogoff
	Mail                       Attribute = "mail"                       // https://learn.microsoft.com/en-us/windows/win32/adschema/a-mail
	Manager                    Attribute = "manager"                    // https://learn.microsoft.com/en-us/windows/win32/adschema/a-manager
	MemberOf                   Attribute = "memberOf"                   // https://learn.microsoft.com/en-us/windows/win32/adschema/a-memberof
	Mobile                     Attribute = "mobile"                     // https://learn.microsoft.com/en-us/windows/win32/adschema/a-mobile
	Name                       Attribute = "name"                       // https://learn.microsoft.com/en-us/windows/win32/adschema/a-name
	PhysicalDeliveryOfficeName Attribute = "physicalDeliveryOfficeName" // https://learn.microsoft.com/en-us/windows/win32/adschema/a-physicaldeliveryofficename
	PostalCode                 Attribute = "postalCode"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-postalcode
	PwdLastSet                 Attribute = "pwdLastSet"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-pwdlastset
	ReplPropertyMetaData       Attribute = "replPropertyMetaData"       // https://learn.microsoft.com/en-us/windows/win32/adschema/a-replpropertymetadata
	SAMAccountName             Attribute = "sAMAccountName"             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-samaccountname
	SAMAccountType             Attribute = "sAMAccountType"             // https://learn.microsoft.com/en-us/windows/win32/adschema/a-samaccounttype
	SN                         Attribute = "sn"                         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-sn
	State                      Attribute = "st"                         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-st
	StreetAddress              Attribute = "streetAddress"              // https://learn.microsoft.com/en-us/windows/win32/adschema/a-streetaddress
	TelephoneNumber            Attribute = "telephoneNumber"            // https://learn.microsoft.com/en-us/windows/win32/adschema/a-telephonenumber
	Title                      Attribute = "title"                      // https://learn.microsoft.com/en-us/windows/win32/adschema/a-title
	UserAccountControl         Attribute = "userAccountControl"         // https://learn.microsoft.com/en-us/windows/win32/adschema/a-useraccountcontrol
	UserPrincipalName          Attribute = "userPrincipalName"          // https://learn.microsoft.com/en-us/windows/win32/adschema/a-userprincipalname
	USNChanged                 Attribute = "uSNChanged"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-usnchanged
	USNCreated                 Attribute = "uSNCreated"                 // https://learn.microsoft.com/en-us/windows/win32/adschema/a-usncreated
)

var (
	DefaultUserAttributes = &[]Attribute{
		AccountExpires, AdminCount, AltSecurityIdentities,
		BadPasswordTime, BadPwdCount,
		City, CodePage, CommonName, Country, CountryCode,
		Department, DistinguishedName, DSCorePropagationData, DisplayName, Division,
		EmployeeID, EmployeeNumber,
		GivenName,
		InstanceType,
		LastLogoff, LastLogon, LastLogonTimestamp,
		Manager, Mail, MemberOf, Mobile,
		Name,
		ObjectCategory, ObjectClass, ObjectGUID, ObjectSID,
		PhysicalDeliveryOfficeName, PostalCode, PwdLastSet,
		ReplPropertyMetaData,
		SAMAccountName, SAMAccountType, SN, StreetAddress,
		TelephoneNumber, Title,
		UserAccountControl, UserPrincipalName, USNChanged, USNCreated,
		WhenChanged, WhenCreated,
	}

	MinimalUserAttributes = &[]Attribute{
		AltSecurityIdentities,
		CommonName,
		DisplayName, DistinguishedName,
		LastLogoff, LastLogon, LastLogonTimestamp,
		MemberOf,
		Name,
		ObjectClass,
		SAMAccountName,
		UserAccountControl,
	}
)

// LDAPObjectClass holds possible LDAP object class constants
type LDAPObjectClass string

// Enum values for LDAP object classes
const (
	ObjectClassUser   LDAPObjectClass = "user"
	ObjectClassGroup  LDAPObjectClass = "group"
	ObjectClassPerson LDAPObjectClass = "person"
)

// Enum values for SAMAccountType
// https://learn.microsoft.com/en-us/windows/win32/adschema/a-samaccounttype
const (
	SAM_DOMAIN_OBJECT             int = 0x00000000 // A domain object.
	SAM_GROUP_OBJECT              int = 0x10000000 // A group object.
	SAM_NON_SECURITY_GROUP_OBJECT int = 0x10000001 // A non-security group object.
	SAM_ALIAS_OBJECT              int = 0x20000000 // An alias object.
	SAM_NON_SECURITY_ALIAS_OBJECT int = 0x20000001 // A non-security alias object.
	SAM_USER_OBJECT               int = 0x30000000 // A user object.
	SAM_NORMAL_USER_ACCOUNT       int = 0x30000000 // A normal user account.
	SAM_MACHINE_ACCOUNT           int = 0x30000001 // A machine account.
	SAM_TRUST_ACCOUNT             int = 0x30000002 // A trust account.
	SAM_APP_BASIC_GROUP           int = 0x40000000 // An application basic group.
	SAM_APP_QUERY_GROUP           int = 0x40000001 // An application query group.
	SAM_ACCOUNT_TYPE_MAX          int = 0x7FFFFFFF // The maximum value for a SAM account type.
)

// Enum values for UserAccountControl
// https://learn.microsoft.com/en-us/windows/win32/api/iads/ne-iads-ads_user_flag_enum
const (
	ADS_UF_SCRIPT                                 int = 0x0001    // The logon script will be run.
	ADS_UF_ACCOUNTDISABLED                        int = 0x0002    // The account is disabled.
	ADS_UF_HOMEDIR_REQUIRED                       int = 0x0008    // A home directory is required.
	ADS_UF_LOCKOUT                                int = 0x0010    // The account is currently locked out.
	ADS_UF_PASSWD_NOTREQD                         int = 0x0020    // No password is required.
	ADS_UF_PASSWD_CANT_CHANGE                     int = 0x0040    // The user cannot change the password.
	ADS_UF_ENCRYPTED_TEXT_PASSWORD_ALLOWED        int = 0x0080    // The user can send an encrypted password.
	ADS_UF_TEMP_DUPLICATE_ACCOUNT                 int = 0x0100    // This is an account for users whose primary account is in another domain. This account provides user access to this domain, but not to any domain that trusts this domain. Also known as a local user account.
	ADS_UF_NORMAL_ACCOUNT                         int = 0x0200    // This is a default account type that represents a typical user.
	ADS_UF_INTERDOMAIN_TRUST_ACCOUNT              int = 0x0800    // This is a trust account for a system domain that trusts other domains.
	ADS_UF_WORKSTATION_TRUST_ACCOUNT              int = 0x1000    // This is a computer account for a computer that is a member of this domain.
	ADS_UF_SERVER_TRUST_ACCOUNT                   int = 0x2000    // This is a computer account for a system backup domain controller that is a member of this domain.
	ADS_UF_DONT_EXPIRE_PASSWD                     int = 0x10000   // The password for this account will never expire.
	ADS_UF_MNS_LOGON_ACCOUNT                      int = 0x20000   // This is an MNS logon account.
	ADS_UF_SMARTCARD_REQUIRED                     int = 0x40000   // The user must log on using a smart card.
	ADS_UF_TRUSTED_FOR_DELEGATION                 int = 0x80000   // The service account (user or computer account), under which a service runs, is trusted for Kerberos delegation. Any such service can impersonate a client requesting the service.
	ADS_UF_NOT_DELEGATED                          int = 0x100000  // The security context of the user will not be delegated to a service even if the service account is set as trusted for Kerberos delegation.
	ADS_UF_USE_DES_KEY_ONLY                       int = 0x200000  // Restrict this principal to use only Data Encryption Standard (DES) encryption types for keys.
	ADS_UF_DONT_REQUIRE_PREAUTH                   int = 0x400000  // This account does not require Kerberos pre-authentication for logon.
	ADS_UF_PASSWORD_EXPIRED                       int = 0x800000  // The user password has expired. This flag is created by the system using data from the Pwd-Last-Set attribute and the domain policy.
	ADS_UF_TRUSTED_TO_AUTHENTICATE_FOR_DELEGATION int = 0x1000000 // The account is enabled for delegation. This is a security-sensitive setting; accounts with this option enabled should be strictly controlled. This setting enables a service running under the account to assume a client identity and authenticate as that user to other remote servers on the network.
)
