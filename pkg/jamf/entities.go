// pkg/jamf/entities.go
package jamf

import (
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ### Jamf Client Structs
// ---------------------------------------------------------------------
// Credentials for Jamf Pro
type Credentials struct {
	Username string
	Password string
	Token    JamfToken
}

type JamfToken struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

type Client struct {
	BaseURL    string           // Base URL for the Jamf Pro API.
	ClassicURL string           // Base URL for the Jamf Pro Classic API.
	HTTP       *requests.Client // HTTP client for making requests to the Jamf Pro API.
	Log        *log.Logger      // Logger for the Jamf Pro client.
	Cache      *cache.Cache     // Cache for the Jamf Pro client.
}

// END OF JAMF CLIENT STRUCTS
//---------------------------------------------------------------------

// ### Jamf Generic Structs
// ---------------------------------------------------------------------
type JamfProperty struct {
	ID   interface{} `json:"id,omitempty" xml:"id,omitempty"`     // ID of the object.
	Name string      `json:"name,omitempty" xml:"name,omitempty"` // Name of the object.
}

// END OF JAMF GENERIC STRUCTS
//---------------------------------------------------------------------

// ### Jamf Device Structs
// ---------------------------------------------------------------------
type Inventory struct {
	Computers     Computers     `json:"computers"`
	MobileDevices MobileDevices `json:"mobile_devices"`
}

// Response structure for the Jamf Pro API for computers
type Computers struct {
	Results    *[]*Computer `json:"results"`    // List of computers.
	TotalCount int          `json:"totalCount"` // Total number of computers.
}

// Total() [Computers] returns the total number of computers in generic functions
func (c Computers) Total() int {
	return c.TotalCount
}

// Append() [Computers] Appends the results of Computers in generic functions to an existing list
func (c Computers) Append(result interface{}) {
	more, ok := result.(*Computers)
	if !ok {
		return
	}
	*c.Results = append(*c.Results, *more.Results...)
}

// Response structure for the Jamf Pro API for mobile devices
type MobileDevices struct {
	Results    *[]*MobileDevice `json:"results"`    // List of mobile devices.
	TotalCount int              `json:"totalCount"` // Total number of mobile devices.
}

// Total() [MobileDevices] returns the total number of mobile devices in generics
func (c MobileDevices) Total() int {
	return c.TotalCount
}

// Append() [MobileDevices] Appends the results of Mobile Devices in generic functions to an existing list
func (m MobileDevices) Append(result interface{}) {
	more, ok := result.(*MobileDevices)
	if !ok {
		return
	}
	*m.Results = append(*m.Results, *more.Results...)
}

// MobileDevice represents the details of a mobile device.
type MobileDevice struct {
	ID                     string `json:"id"`                     // Unique identifier for the mobile device.
	ManagementID           string `json:"managementId"`           // Management identifier for the mobile device.
	Model                  string `json:"model"`                  // Model of the mobile device.
	ModelIdentifier        string `json:"modelIdentifier"`        // Model identifier for the mobile device.
	Name                   string `json:"name"`                   // Name of the mobile device.
	PhoneNumber            string `json:"phoneNumber"`            // Phone number associated with the mobile device.
	SerialNumber           string `json:"serialNumber"`           // Serial number of the mobile device.
	SoftwareUpdateDeviceID string `json:"softwareUpdateDeviceId"` // Software update device ID.
	Type                   string `json:"type"`                   // Type of the mobile device (e.g., iOS).
	UDID                   string `json:"udid"`                   // Unique Device Identifier.
	Username               string `json:"username"`               // Username associated with the mobile device.
	WifiMacAddress         string `json:"wifiMacAddress"`         // WiFi MAC address of the mobile device.
}

// Computer represents the details of a computer.
type Computer struct {
	Applications          *[]*Application          `json:"applications,omitempty"`          // List of applications installed on the computer.
	Attachments           *[]*Attachment           `json:"attachments,omitempty"`           // List of attachments.
	Certificates          *[]*Certificate          `json:"certificates,omitempty"`          // List of certificates installed on the computer.
	ConfigurationProfiles *[]*ConfigurationProfile `json:"configurationProfiles,omitempty"` // List of configuration profiles applied to the computer.
	ContentCaching        *ContentCaching          `json:"contentCaching,omitempty"`        // Content caching details.
	DiskEncryption        *DiskEncryption          `json:"diskEncryption,omitempty"`        // Disk encryption details.
	Fonts                 *[]*Font                 `json:"fonts,omitempty"`                 // List of fonts installed on the computer.
	General               *General                 `json:"general,omitempty"`               // General information about the computer.
	GroupMemberships      *[]*GroupMembership      `json:"groupMemberships,omitempty"`      // List of group memberships.
	Hardware              *Hardware                `json:"hardware,omitempty"`              // Hardware details of the computer.
	IBeacons              *[]*JamfProperty         `json:"ibeacons,omitempty"`              // iBeacons associated with the computer.
	ID                    interface{}              `json:"id"`                              // Unique identifier for the computer.
	LicensedSoftware      *[]*JamfProperty         `json:"licensedSoftware,omitempty"`      // List of licensed software.
	LocalUserAccounts     *[]*LocalUserAccount     `json:"localUserAccounts,omitempty"`     // List of local user accounts on the computer.
	OperatingSystem       *OperatingSystem         `json:"operatingSystem,omitempty"`       // Operating system details.
	PackageReceipts       *PackageReceipts         `json:"packageReceipts,omitempty"`       // Information about package receipts.
	Plugins               *[]*Plugin               `json:"plugins,omitempty"`               // List of plugins installed on the computer.
	Printers              *[]*Printer              `json:"printers,omitempty"`              // List of printers configured on the computer.
	Purchasing            *Purchasing              `json:"purchasing,omitempty"`            // Purchasing information.
	Security              *ComputerSecurity        `json:"security,omitempty"`              // Security settings and information.
	Services              *[]*Service              `json:"services,omitempty"`              // List of services on the computer.
	SoftwareUpdates       *[]*SoftwareUpdate       `json:"softwareUpdates,omitempty"`       // List of software updates.
	Storage               *Storage                 `json:"storage,omitempty"`               // Storage details.
	UDID                  string                   `json:"udid"`                            // Unique Device Identifier.
	UserAndLocation       *UserAndLocation         `json:"userAndLocation,omitempty"`       // User and location information.
}

// Application represents the details of an application installed on the computer.
type Application struct {
	BundleID          string `json:"bundleId,omitempty"`          // Bundle identifier of the application.
	ExternalVersionID string `json:"externalVersionId,omitempty"` // External version identifier.
	MacAppStore       bool   `json:"macAppStore,omitempty"`       // Indicates if the application is from Mac App Store.
	Name              string `json:"name,omitempty"`              // Name of the application.
	Path              string `json:"path,omitempty"`              // Installation path of the application.
	SizeMegabytes     int    `json:"sizeMegabytes,omitempty"`     // Size of the application in megabytes.
	UpdateAvailable   bool   `json:"updateAvailable,omitempty"`   // Indicates if an update is available.
	Version           string `json:"version,omitempty"`           // Version of the application.
}

// Attachment represents an attachment in the inventory.
type Attachment struct {
	FileType  string `json:"fileType,omitempty"`  // Type of the file.
	ID        string `json:"id,omitempty"`        // ID of the attachment.
	Name      string `json:"name,omitempty"`      // Name of the attachment.
	SizeBytes int    `json:"sizeBytes,omitempty"` // Size of the attachment in bytes.
}

type Certificate struct {
	CertificateStatus string `json:"certificateStatus,omitempty"` // Status of the certificate.
	CommonName        string `json:"commonName,omitempty"`        // Common name of the certificate.
	ExpirationDate    string `json:"expirationDate,omitempty"`    // Expiration date of the certificate.
	Identity          bool   `json:"identity,omitempty"`          // Indicates if the certificate is an identity certificate.
	IssuedDate        string `json:"issuedDate,omitempty"`        // Issued date of the certificate.
	LifecycleStatus   string `json:"lifecycleStatus,omitempty"`   // Lifecycle status of the certificate.
	SerialNumber      string `json:"serialNumber,omitempty"`      // Serial number of the certificate.
	Sha1Fingerprint   string `json:"sha1Fingerprint,omitempty"`   // SHA1 fingerprint of the certificate.
	SubjectName       string `json:"subjectName,omitempty"`       // Subject name of the certificate.
}

// ContentCaching represents content caching information of a computer.
type ContentCaching struct {
	Activated                           bool               `json:"activated,omitempty"`                           // Indicates if caching is activated.
	Active                              bool               `json:"active,omitempty"`                              // Indicates if caching is currently active.
	ActualCacheBytesUsed                int64              `json:"actualCacheBytesUsed,omitempty"`                // Actual bytes used by the cache.
	Address                             string             `json:"address,omitempty"`                             // Address of the caching server.
	Addresses                           []string           `json:"addresses,omitempty"`                           // List of addresses related to the caching server.
	Alerts                              interface{}        `json:"alerts,omitempty"`                              // Alerts related to caching. Can either be an array or object assigned to CacheAlert
	CacheBytesFree                      int64              `json:"cacheBytesFree,omitempty"`                      // Free bytes in the cache.
	CacheBytesLimit                     int64              `json:"cacheBytesLimit,omitempty"`                     // Limit of cache bytes.
	CacheBytesUsed                      int64              `json:"cacheBytesUsed,omitempty"`                      // Bytes used in the cache.
	CacheDetails                        []CacheDetail      `json:"cacheDetails,omitempty"`                        // Details about the cache.
	CacheStatus                         string             `json:"cacheStatus,omitempty"`                         // Status of the cache.
	ComputerContentCachingInformationID string             `json:"computerContentCachingInformationId,omitempty"` // Information ID for content caching.
	DataMigrationCompleted              bool               `json:"dataMigrationCompleted,omitempty"`              // Indicates if data migration is completed.
	DataMigrationError                  DataMigrationError `json:"dataMigrationError,omitempty"`                  // Data migration error details.
	DataMigrationProgressPercentage     int                `json:"dataMigrationProgressPercentage,omitempty"`     // Progress percentage of data migration.
	Details                             CacheDetail        `json:"details,omitempty"`                             // Details of the content caching. Seems redundant, but it helps with struct recursion.
	GUID                                string             `json:"guid,omitempty"`                                // GUID of the caching server.
	Healthy                             bool               `json:"healthy,omitempty"`                             // Indicates if the caching server is healthy.
	MaxCachePressureLast1HourPercentage int                `json:"maxCachePressureLast1HourPercentage,omitempty"` // Max cache pressure in the last hour.
	Parents                             []ContentCaching   `json:"parents,omitempty"`                             // Parent caching servers, recursively using the same structure.
	ParentID                            string             `json:"contentCachingParentId,omitempty"`              // Parent ID of the content caching event.
	ParentDetailsID                     string             `json:"contentCachingParentDetailsId,omitempty"`       // Parent details ID of the content caching event.
	PersonalCacheBytesFree              int64              `json:"personalCacheBytesFree,omitempty"`              // Free bytes in the personal cache.
	PersonalCacheBytesLimit             int64              `json:"personalCacheBytesLimit,omitempty"`             // Limit of personal cache bytes.
	PersonalCacheBytesUsed              int64              `json:"personalCacheBytesUsed,omitempty"`              // Used bytes in the personal cache.
	Port                                int                `json:"port,omitempty"`                                // Port number for caching.
	PublicAddress                       string             `json:"publicAddress,omitempty"`                       // Public address for caching.
	RegistrationError                   string             `json:"registrationError,omitempty"`                   // Registration error message.
	RegistrationResponseCode            int                `json:"registrationResponseCode,omitempty"`            // Response code for registration.
	RegistrationStarted                 string             `json:"registrationStarted,omitempty"`                 // Start time of registration.
	RegistrationStatus                  string             `json:"registrationStatus,omitempty"`                  // Status of registration.
	RestrictedMedia                     bool               `json:"restrictedMedia,omitempty"`                     // Indicates if media is restricted.
	ServerGuid                          string             `json:"serverGuid,omitempty"`                          // GUID of the server.
	StartupStatus                       string             `json:"startupStatus,omitempty"`                       // Startup status of the caching.
	TetheratorStatus                    string             `json:"tetheratorStatus,omitempty"`                    // Tetherator status.
	TotalBytesAreSince                  string             `json:"totalBytesAreSince,omitempty"`                  // Total bytes are calculated since this time.
	TotalBytesDropped                   int64              `json:"totalBytesDropped,omitempty"`                   // Total bytes dropped.
	TotalBytesImported                  int64              `json:"totalBytesImported,omitempty"`                  // Total bytes imported.
	TotalBytesReturnedToChildren        int64              `json:"totalBytesReturnedToChildren,omitempty"`        // Total bytes returned to children.
	TotalBytesReturnedToClients         int64              `json:"totalBytesReturnedToClients,omitempty"`         // Total bytes returned to clients.
	TotalBytesReturnedToPeers           int64              `json:"totalBytesReturnedToPeers,omitempty"`           // Total bytes returned to peers.
	TotalBytesStoredFromOrigin          int64              `json:"totalBytesStoredFromOrigin,omitempty"`          // Total bytes stored from origin.
	TotalBytesStoredFromParents         int64              `json:"totalBytesStoredFromParents,omitempty"`         // Total bytes stored from parents.
	TotalBytesStoredFromPeers           int64              `json:"totalBytesStoredFromPeers,omitempty"`           // Total bytes stored from peers.
	Version                             string             `json:"version,omitempty"`                             // Version of the caching server.
}

// CacheAlert represents an alert related to caching.
type CacheAlert struct {
	Addresses            []string `json:"addresses,omitempty"`                   // List of addresses related to the caching server.
	CacheBytesLimit      int64    `json:"cacheBytesLimit,omitempty"`             // Limit of cache bytes for the alert.
	ClassName            string   `json:"className,omitempty"`                   // The class name of the alert.
	ID                   string   `json:"contentCachingParentAlertId,omitempty"` // Unique identifier of the alert.
	PathPreventingAccess string   `json:"pathPreventingAccess,omitempty"`        // Path that is preventing access.
	PostDate             string   `json:"postDate,omitempty"`                    // Post date of the alert.
	ReservedVolumeBytes  int64    `json:"reservedVolumeBytes,omitempty"`         // Reserved volume bytes.
	Resource             string   `json:"resource,omitempty"`                    // Resource associated with the alert.
}

// CacheDetail represents details of the cache.
type CacheDetail struct {
	ComputerContentCachingCacheDetailsID string `json:"computerContentCachingCacheDetailsId,omitempty"` // ID of the cache detail.
	CategoryName                         string `json:"categoryName,omitempty"`                         // Name of the category.
	DiskSpaceBytesUsed                   int64  `json:"diskSpaceBytesUsed,omitempty"`                   // Disk space used in bytes.
}

// DataMigrationError represents details of a data migration error.
type DataMigrationError struct {
	Code     int        `json:"code,omitempty"`     // Error code.
	Domain   string     `json:"domain,omitempty"`   // Error domain.
	UserInfo []KeyValue `json:"userInfo,omitempty"` // Additional user info.
}

// KeyValue represents a key-value pair.
type KeyValue struct {
	Key   string `json:"key,omitempty"`   // Key of the user info.
	Value string `json:"value,omitempty"` // Value of the user info.
}

// DataMigrationInfo represents additional information for a data migration error.
type DataMigrationInfo struct {
	Key   string `json:"key"`   // Key of the additional information.
	Value string `json:"value"` // Value of the additional information.
}

// DiskEncryption represents details of disk encryption on the computer.
type DiskEncryption struct {
	BootPartitionEncryptionDetails      BootPartitionDetails `json:"bootPartitionEncryptionDetails"`      // Details of the boot partition encryption.
	IndividualRecoveryKeyValidityStatus string               `json:"individualRecoveryKeyValidityStatus"` // Validity status of the individual recovery key.
	InstitutionalRecoveryKeyPresent     bool                 `json:"institutionalRecoveryKeyPresent"`     // Indicates if institutional recovery key is present.
	DiskEncryptionConfigurationName     string               `json:"diskEncryptionConfigurationName"`     // Name of the disk encryption configuration.
	FileVault2EnabledUserNames          []string             `json:"fileVault2EnabledUserNames"`          // List of usernames with FileVault 2 enabled.
	FileVault2EligibilityMessage        string               `json:"fileVault2EligibilityMessage"`        // Eligibility message for FileVault 2.
}

type BootPartitionDetails struct {
	PartitionName              string `json:"partitionName"`              // Name of the partition.
	PartitionFileVault2State   string `json:"partitionFileVault2State"`   // FileVault 2 state of the partition.
	PartitionFileVault2Percent int    `json:"partitionFileVault2Percent"` // FileVault 2 percent of the partition.
}

// Font represents details of a font installed on the computer.
type Font struct {
	Name    string `json:"name,omitempty"`    // Name of the font.
	Path    string `json:"path,omitempty"`    // Path to the font.
	Version string `json:"version,omitempty"` // Version of the font.
}

// General information about the computer.
type General struct {
	AssetTag                             string               `json:"assetTag"`                             // Asset tag of the computer.
	Barcode1                             string               `json:"barcode1"`                             // First barcode value.
	Barcode2                             string               `json:"barcode2"`                             // Second barcode value.
	DistributionPoint                    string               `json:"distributionPoint"`                    // Name of the distribution point.
	EnrolledViaAutomatedDeviceEnrollment bool                 `json:"enrolledViaAutomatedDeviceEnrollment"` // Indicates if enrolled via automated device enrollment.
	EnrollmentMethod                     EnrollmentMethod     `json:"enrollmentMethod"`                     // Method of enrollment.
	ExtensionAttributes                  []ExtensionAttribute `json:"extensionAttributes"`                  // List of extension attributes.
	InitialEntryDate                     string               `json:"initialEntryDate"`                     // Date of initial entry.
	ItunesStoreAccountActive             bool                 `json:"itunesStoreAccountActive"`             // Indicates if iTunes Store account is active.
	JamfBinaryVersion                    string               `json:"jamfBinaryVersion"`                    // Version of the Jamf binary.
	LastContactTime                      string               `json:"lastContactTime"`                      // Time of last contact.
	LastEnrolledDate                     string               `json:"lastEnrolledDate"`                     // Date of last enrollment.
	LastIpAddress                        string               `json:"lastIpAddress"`                        // Last known IP address.
	LastReportedIp                       string               `json:"lastReportedIp"`                       // Last reported IP address.
	LastCloudBackupDate                  string               `json:"lastCloudBackupDate"`                  // Date of last cloud backup.
	ManagementID                         string               `json:"managementId"`                         // Management ID.
	MDMCapable                           MDMCapable           `json:"mdmCapable"`                           // MDM capability information.
	MdmProfileExpiration                 string               `json:"mdmProfileExpiration"`                 // Expiration of the MDM profile.
	Name                                 string               `json:"name"`                                 // Name of the computer.
	Platform                             string               `json:"platform"`                             // Platform of the computer (e.g., Mac).
	RemoteManagement                     RemoteManagement     `json:"remoteManagement"`                     // Remote management information.
	ReportDate                           string               `json:"reportDate"`                           // Date of report.
	Site                                 Site                 `json:"site"`                                 // Site information.
	Supervised                           bool                 `json:"supervised"`                           // Indicates if the device is supervised.
	UserApprovedMDM                      bool                 `json:"userApprovedMdm"`                      // Indicates if MDM is user-approved.
	DeclarativeDeviceManagementEnabled   bool                 `json:"declarativeDeviceManagementEnabled"`   // Indicates if declarative device management is enabled.
}

// GroupMembership represents the membership details of a computer in a group.
type GroupMembership struct {
	GroupID    string `json:"groupId,omitempty"`    // Unique identifier of the group.
	GroupName  string `json:"groupName,omitempty"`  // Name of the group.
	SmartGroup bool   `json:"smartGroup,omitempty"` // Indicates if the group is a smart group.
}

// EnrollmentMethod represents the method of enrollment of a computer.
type EnrollmentMethod struct {
	ID         string `json:"id"`         // Identifier of the enrollment method.
	ObjectName string `json:"objectName"` // Name of the object associated with the enrollment.
	ObjectType string `json:"objectType"` // Type of the object associated with the enrollment.
}

// ExtensionAttribute represents the extension attributes in UserAndLocation and Purchasing.
type ExtensionAttribute struct {
	DefinitionID string   `json:"definitionId"` // Unique identifier of the definition.
	Name         string   `json:"name"`         // Name of the attribute.
	Description  string   `json:"description"`  // Description of the attribute.
	Enabled      bool     `json:"enabled"`      // Indicates if the attribute is enabled.
	MultiValue   bool     `json:"multiValue"`   // Indicates if the attribute has multiple values.
	Values       []string `json:"values"`       // List of values for the attribute.
	DataType     string   `json:"dataType"`     // Data type of the attribute.
	Options      []string `json:"options"`      // List of options for the attribute.
	InputType    string   `json:"inputType"`    // Input type of the attribute.
}

// MDMCapable represents MDM capability information of a computer.
type MDMCapable struct {
	Capable      bool     `json:"capable"`      // Indicates if the computer is MDM capable.
	CapableUsers []string `json:"capableUsers"` // List of users capable of MDM.
}

// RemoteManagement represents remote management information of a computer.
type RemoteManagement struct {
	Managed            bool   `json:"managed"`            // Indicates if the computer is managed.
	ManagementUsername string `json:"managementUsername"` // Username for management.
}

// Site represents site information of a Jamf object.
type Site struct {
	*JamfProperty
}

// Hardware represents the hardware details of a computer in the inventory.
type Hardware struct {
	Make                   string               `json:"make"`                   // Manufacturer of the hardware.
	Model                  string               `json:"model"`                  // Model of the hardware.
	ModelIdentifier        string               `json:"modelIdentifier"`        // Identifier for the model.
	SerialNumber           string               `json:"serialNumber"`           // Serial number of the hardware.
	ProcessorSpeedMhz      int                  `json:"processorSpeedMhz"`      // Processor speed in MHz.
	ProcessorCount         int                  `json:"processorCount"`         // Number of processors.
	CoreCount              int                  `json:"coreCount"`              // Number of cores.
	ProcessorType          string               `json:"processorType"`          // Type of processor.
	ProcessorArchitecture  string               `json:"processorArchitecture"`  // Processor architecture.
	BusSpeedMhz            int                  `json:"busSpeedMhz"`            // Bus speed in MHz.
	CacheSizeKilobytes     int                  `json:"cacheSizeKilobytes"`     // Cache size in Kilobytes.
	NetworkAdapterType     string               `json:"networkAdapterType"`     // Primary network adapter type.
	MacAddress             string               `json:"macAddress"`             // MAC address.
	AltNetworkAdapterType  string               `json:"altNetworkAdapterType"`  // Alternate network adapter type.
	AltMacAddress          string               `json:"altMacAddress"`          // Alternate MAC address.
	TotalRamMegabytes      int                  `json:"totalRamMegabytes"`      // Total RAM in Megabytes.
	OpenRamSlots           int                  `json:"openRamSlots"`           // Number of open RAM slots.
	BatteryCapacityPercent int                  `json:"batteryCapacityPercent"` // Battery capacity as a percentage.
	SmcVersion             string               `json:"smcVersion"`             // SMC version.
	NicSpeed               string               `json:"nicSpeed"`               // Network interface card speed.
	OpticalDrive           string               `json:"opticalDrive"`           // Optical drive type.
	BootRom                string               `json:"bootRom"`                // Boot ROM version.
	BleCapable             bool                 `json:"bleCapable"`             // Indicates if Bluetooth Low Energy is supported.
	SupportsIosAppInstalls bool                 `json:"supportsIosAppInstalls"` // Indicates if iOS app installs are supported.
	AppleSilicon           bool                 `json:"appleSilicon"`           // Indicates if the device has Apple Silicon.
	ExtensionAttributes    []ExtensionAttribute `json:"extensionAttributes"`    // List of extension attributes.
}

// LocalUserAccount represents a local user account on the computer.
type LocalUserAccount struct {
	Admin                          bool   `json:"admin,omitempty"`                          // Indicates if the user is an admin.
	AzureActiveDirectoryID         string `json:"azureActiveDirectoryId,omitempty"`         // Azure Active Directory ID.
	ComputerAzureActiveDirectoryID string `json:"computerAzureActiveDirectoryId,omitempty"` // Computer's Azure Active Directory ID.
	FileVault2Enabled              bool   `json:"fileVault2Enabled,omitempty"`              // Indicates if FileVault2 is enabled.
	FullName                       string `json:"fullName,omitempty"`                       // Full name of the user.
	HomeDirectory                  string `json:"homeDirectory,omitempty"`                  // Path to the home directory.
	HomeDirectorySizeMb            int    `json:"homeDirectorySizeMb,omitempty"`            // Size of the home directory in MB.
	PasswordHistoryDepth           int    `json:"passwordHistoryDepth,omitempty"`           // Depth of password history.
	PasswordMaxAge                 int    `json:"passwordMaxAge,omitempty"`                 // Maximum age of the password.
	PasswordMinComplexCharacters   int    `json:"passwordMinComplexCharacters,omitempty"`   // Minimum number of complex characters in password.
	PasswordMinLength              int    `json:"passwordMinLength,omitempty"`              // Minimum length of the password.
	PasswordRequireAlphanumeric    bool   `json:"passwordRequireAlphanumeric,omitempty"`    // Indicates if password requires alphanumeric characters.
	Uid                            string `json:"uid,omitempty"`                            // User ID.
	UserAccountType                string `json:"userAccountType,omitempty"`                // Type of the user account.
	UserAzureActiveDirectoryID     string `json:"userAzureActiveDirectoryId,omitempty"`     // User's Azure Active Directory ID.
	UserGuid                       string `json:"userGuid,omitempty"`                       // User GUID.
	Username                       string `json:"username,omitempty"`                       // Username.
}

// OperatingSystem represents information about the operating system of the computer.
type OperatingSystem struct {
	ActiveDirectoryStatus    string               `json:"activeDirectoryStatus,omitempty"`    // Status of Active Directory binding.
	Build                    string               `json:"build,omitempty"`                    // Build version of the operating system.
	FileVault2Status         string               `json:"fileVault2Status,omitempty"`         // Status of FileVault2 encryption.
	Name                     string               `json:"name,omitempty"`                     // Name of the operating system.
	RapidSecurityResponse    string               `json:"rapidSecurityResponse,omitempty"`    // Rapid Security Response status.
	SoftwareUpdateDeviceID   string               `json:"softwareUpdateDeviceId,omitempty"`   // Software Update Device ID.
	SupplementalBuildVersion string               `json:"supplementalBuildVersion,omitempty"` // Supplemental build version of the operating system.
	Version                  string               `json:"version,omitempty"`                  // Version of the operating system.
	ExtensionAttributes      []ExtensionAttribute `json:"extensionAttributes,omitempty"`      // List of extension attributes.
}

// PackageReceipts represents the package receipts on the computer.
type PackageReceipts struct {
	Cached                  []string `json:"cached,omitempty"`                  // List of packages cached.
	InstalledByInstallerSwu []string `json:"installedByInstallerSwu,omitempty"` // List of packages installed by InstallerSwu.
	InstalledByJamfPro      []string `json:"installedByJamfPro,omitempty"`      // List of packages installed by Jamf Pro.
}

// Plugin represents a plugin installed on the computer.
type Plugin struct {
	Name    string `json:"name,omitempty"`    // Name of the plugin.
	Path    string `json:"path,omitempty"`    // Path to the plugin.
	Version string `json:"version,omitempty"` // Version of the plugin.
}

// Printer represents a printer in the inventory.
type Printer struct {
	Name     string `json:"name"`     // Name of the printer.
	Type     string `json:"type"`     // Type/model of the printer.
	URI      string `json:"uri"`      // URI for the printer.
	Location string `json:"location"` // Physical location of the printer.
}

// Purchasing represents the purchasing information of a computer.
type Purchasing struct {
	AppleCareID         string               `json:"appleCareId"`         // AppleCare ID.
	ExtensionAttributes []ExtensionAttribute `json:"extensionAttributes"` // List of extension attributes.
	LeaseDate           string               `json:"leaseDate"`           // Date of the lease.
	Leased              bool                 `json:"leased"`              // Indicates if the computer is leased.
	LifeExpectancy      int                  `json:"lifeExpectancy"`      // Expected life expectancy in years.
	PoDate              string               `json:"poDate"`              // Purchase order date.
	PoNumber            string               `json:"poNumber"`            // Purchase order number.
	PurchasePrice       string               `json:"purchasePrice"`       // Purchase price.
	Purchased           bool                 `json:"purchased"`           // Indicates if the computer is purchased.
	PurchasingAccount   string               `json:"purchasingAccount"`   // Account used for purchasing.
	PurchasingContact   string               `json:"purchasingContact"`   // Contact for purchasing.
	Vendor              string               `json:"vendor"`              // Vendor from where the computer is purchased.
	WarrantyDate        string               `json:"warrantyDate"`        // Date of warranty expiration.
}

// Security represents the security settings of the computer.
type ComputerSecurity struct {
	ActivationLockEnabled bool   `json:"activationLockEnabled,omitempty"` // Indicates if activation lock is enabled.
	AutoLoginDisabled     bool   `json:"autoLoginDisabled,omitempty"`     // Indicates if auto-login is disabled.
	BootstrapTokenAllowed bool   `json:"bootstrapTokenAllowed,omitempty"` // Indicates if bootstrap token is allowed.
	ExternalBootLevel     string `json:"externalBootLevel,omitempty"`     // Level of external boot allowed.
	FirewallEnabled       bool   `json:"firewallEnabled,omitempty"`       // Indicates if the firewall is enabled.
	GatekeeperStatus      string `json:"gatekeeperStatus,omitempty"`      // Status of Gatekeeper.
	RecoveryLockEnabled   bool   `json:"recoveryLockEnabled,omitempty"`   // Indicates if recovery lock is enabled.
	RemoteDesktopEnabled  bool   `json:"remoteDesktopEnabled,omitempty"`  // Indicates if remote desktop is enabled.
	SecureBootLevel       string `json:"secureBootLevel,omitempty"`       // Level of secure boot.
	SipStatus             string `json:"sipStatus,omitempty"`             // Status of System Integrity Protection.
	XprotectVersion       string `json:"xprotectVersion,omitempty"`       // Version of XProtect.
}

// Service represents a service in the inventory.
type Service struct {
	Name string `json:"name"` // Name of the service.
}

// SoftwareUpdate represents a software update available for the computer.
type SoftwareUpdate struct {
	Name        string `json:"name,omitempty"`        // Name of the software update.
	PackageName string `json:"packageName,omitempty"` // Package name of the software update.
	Version     string `json:"version,omitempty"`     // Version of the software update.
}

// Storage represents the storage details of a computer.
type Storage struct {
	BootDriveAvailableSpaceMegabytes int    `json:"bootDriveAvailableSpaceMegabytes"` // Available space in megabytes on the boot drive.
	Disks                            []Disk `json:"disks"`                            // List of disks in the computer.
}

// Disk represents a disk in a computer.
type Disk struct {
	Device        string      `json:"device"`        // Identifier of the disk device.
	ID            string      `json:"id"`            // Unique identifier for the disk.
	Model         string      `json:"model"`         // Model of the disk.
	Partitions    []Partition `json:"partitions"`    // Partitions on the disk.
	Revision      string      `json:"revision"`      // Revision number of the disk.
	SerialNumber  string      `json:"serialNumber"`  // Serial number of the disk.
	SizeMegabytes int         `json:"sizeMegabytes"` // Total size of the disk in megabytes.
	SmartStatus   string      `json:"smartStatus"`   // S.M.A.R.T status of the disk.
	Type          string      `json:"type"`          // Type of the disk (e.g., SSD, HDD).
}

// Partition represents a partition on a disk.
type Partition struct {
	AvailableMegabytes        int    `json:"availableMegabytes"`        // Available space in megabytes on the partition.
	FileVault2ProgressPercent int    `json:"fileVault2ProgressPercent"` // Progress percentage of FileVault2 encryption.
	FileVault2State           string `json:"fileVault2State"`           // State of FileVault2 encryption on the partition.
	LvmManaged                bool   `json:"lvmManaged"`                // Indicates if the partition is managed by Logical Volume Management.
	Name                      string `json:"name"`                      // Name of the partition.
	PartitionType             string `json:"partitionType"`             // Type of the partition (e.g., BOOT, DATA).
	PercentUsed               int    `json:"percentUsed"`               // Percentage of space used on the partition.
	SizeMegabytes             int    `json:"sizeMegabytes"`             // Total size of the partition in megabytes.
}

// UserAndLocation represents user and location information of a computer.
type UserAndLocation struct {
	Username            string               `json:"username"`            // Username associated with the computer.
	Realname            string               `json:"realname"`            // Real name of the user.
	Email               string               `json:"email"`               // Email address of the user.
	Position            string               `json:"position"`            // Position or title of the user.
	Phone               string               `json:"phone"`               // Phone number of the user.
	DepartmentID        string               `json:"departmentId"`        // Department ID.
	BuildingID          string               `json:"buildingId"`          // Building ID.
	Room                string               `json:"room"`                // Room number or name.
	ExtensionAttributes []ExtensionAttribute `json:"extensionAttributes"` // List of extension attributes.
}

// END OF JAMF DEVICE STRUCTS
//---------------------------------------------------------------------

// ### Jamf User Structs
// ---------------------------------------------------------------------
// Response structure for the Jamf Pro API for Configuration Profiles
type Users struct {
	List *[]*User `json:"users"` // List of JSS users
}

// User represents the details of a JSS User.
type User struct {
	*JamfProperty
	CustomPhotoURL       string                `json:"custom_photo_url,omitempty" xml:"custom_photo_url,omitempty"`               // Custom photo URL of the user.
	Email                string                `json:"email,omitempty" xml:"email,omitempty"`                                     // Email of the user.
	EmailAddress         string                `json:"email_address,omitempty" xml:"email_address,omitempty"`                     // Email address of the user.
	EnableCustomPhotoURL bool                  `json:"enable_custom_photo_url,omitempty" xml:"enable_custom_photo_url,omitempty"` // Indicates if custom photo URL is enabled.
	ExtensionAttributes  []*ExtensionAttribute `json:"extension_attributes,omitempty" xml:"extension_attributes,omitempty"`       // Extension attributes
	FullName             string                `json:"full_name,omitempty" xml:"full_name,omitempty"`                             // Full name of the user.
	LDAPServer           []*JamfProperty       `json:"ldap_server,omitempty" xml:"ldap_server,omitempty"`                         // LDAP server information.
	Links                []*UserLink           `json:"links,omitempty" xml:"links,omitempty"`                                     // Links associated with the user.
	ManagedAppleID       string                `json:"managed_apple_id,omitempty" xml:"managed_apple_id,omitempty"`               // Managed Apple ID.
	PhoneNumber          string                `json:"phone_number,omitempty" xml:"phone_number,omitempty"`                       // Phone number of the user.
	Position             string                `json:"position,omitempty" xml:"position,omitempty"`                               // Position of the user.
	Sites                []*Site               `json:"sites,omitempty" xml:"sites,omitempty"`                                     // Sites
	UserGroups           []*UserGroup          `json:"user_groups,omitempty" xml:"user_groups,omitempty"`                         // Groups the user belongs to.
}

// UserLink represents a link associated with a user.
type UserLink struct {
	Computer          []*Computer `json:"computer,omitempty" xml:"computer,omitempty"`                         // Computer information.
	TotalVPPCodeCount int         `json:"total_vpp_code_count,omitempty" xml:"total_vpp_code_count,omitempty"` // Total VPP code count.
}

// UserGroup represents a group that a user belongs to.
type UserGroup struct {
	*JamfProperty
	IsSmart bool `json:"is_smart,omitempty" xml:"is_smart,omitempty"` // Indicates if the group is a smart group.
}

// END OF JAMF USER STRUCTS
//---------------------------------------------------------------------

// ### Jamf Management Structs
// ---------------------------------------------------------------------
// ManagementResponse represents a generic response for device management operations.
type ManagementResponse struct {
	DeviceID         string             `json:"deviceId,omitempty"`          // The unique identifier of the device.
	CommandUUID      string             `json:"commandUuid,omitempty"`       // The UUID of the command issued to the device.
	UnprocessedUDIDs *UDIDsNotProcessed `json:"udidsNotProcessed,omitempty"` // UDIDs that were not processed, if any.
}

// UDIDsNotProcessed represents a list of UDIDs that were not processed.
type UDIDsNotProcessed struct {
	UDIDs []string `json:"udids"` // List of UDIDs that were not processed.
}

// END OF JAMF MANAGEMENT STRUCTS
//---------------------------------------------------------------------

// ### Jamf {Configuration Profile, Policy} Structs
// ---------------------------------------------------------------------
// Response structure for the Jamf Pro API for Configuration Profiles
type OSXConfigurationProfiles struct {
	List *[]*OSXConfigurationProfile `json:"os_x_configuration_profiles"` // List of configuration profiles.
}

// OSXConfigurationProfile represents the details of a configuration profile.
type OSXConfigurationProfile struct {
	*JamfProperty
	Details struct {
		General     *ConfigurationProfile `json:"general,omitempty" xml:"general,omitempty"`           // General configuration details.
		Scope       *Scope                `json:"scope,omitempty" xml:"scope,omitempty"`               // Scope of the configuration.
		SelfService *SelfService          `json:"self_service,omitempty" xml:"self_service,omitempty"` // Self-service related configurations.
	} `json:"os_x_configuration_profile,omitempty" xml:"os_x_configuration_profile,omitempty"` // Configuration profile details.
}

type ConfigurationProfile struct {
	*JamfProperty
	Category           *Category `json:"category,omitempty" xml:"category,omitempty"`                       // Category information.
	Description        string    `json:"description,omitempty" xml:"description,omitempty"`                 // Description of the profile.
	DisplayName        string    `json:"displayName,omitempty"`                                             // Display name of the configuration profile.
	DistributionMethod string    `json:"distribution_method,omitempty" xml:"distribution_method,omitempty"` // Distribution method.
	LastInstalled      string    `json:"lastInstalled,omitempty"`                                           // Last installed date of the configuration profile.
	Level              string    `json:"level,omitempty" xml:"level,omitempty"`                             // Level of the configuration.
	Payloads           string    `json:"payloads,omitempty" xml:"payloads,omitempty"`                       // Payloads
	ProfileIdentifier  string    `json:"profileIdentifier,omitempty"`                                       // Profile identifier of the configuration profile.
	RedeployOnUpdate   string    `json:"redeploy_on_update,omitempty" xml:"redeploy_on_update,omitempty"`   // Redeployment criteria.
	Removable          bool      `json:"removable,omitempty"`                                               // Indicates if the profile is removable.
	Site               *Site     `json:"site,omitempty" xml:"site,omitempty"`                               // Site information.
	Username           string    `json:"username,omitempty"`                                                // Username associated with the configuration profile.
	UserRemovable      bool      `json:"user_removable,omitempty" xml:"user_removable,omitempty"`           // Whether user can remove the profile.
	UUID               string    `json:"uuid,omitempty" xml:"uuid,omitempty"`                               // Universal Unique Identifier.
}

// Category represents category information of the {configuration profile, policy}.
type Category struct {
	*JamfProperty
}

// Scope represents the scope of the {configuration profile, policy}.
type Scope struct {
	AllComputers    bool             `json:"all_computers,omitempty" xml:"all_computers,omitempty"`       // If all computers are included.
	AllJSSUsers     bool             `json:"all_jss_users,omitempty" xml:"all_jss_users,omitempty"`       // If all JSS users are included.
	Buildings       interface{}      `json:"buildings,omitempty" xml:"buildings,omitempty"`               // Buildings
	ComputerGroups  []*ComputerGroup `json:"computer_groups,omitempty" xml:"computer_groups,omitempty"`   // Computer groups.
	Computers       []*Computer      `json:"computers,omitempty" xml:"computers,omitempty"`               // Computers
	Departments     interface{}      `json:"departments,omitempty" xml:"departments,omitempty"`           // Departments
	Exclusions      *Exclusions      `json:"exclusions,omitempty" xml:"exclusions,omitempty"`             // Exclusions from the scope.
	IBeacons        interface{}      `json:"ibeacons,omitempty" xml:"ibeacons,omitempty"`                 // iBeacons
	JSSUserGroups   interface{}      `json:"jss_user_groups,omitempty" xml:"jss_user_groups,omitempty"`   // JSS user groups
	JSSUsers        interface{}      `json:"jss_users,omitempty" xml:"jss_users,omitempty"`               // JSS users
	Limitations     *Limitations     `json:"limitations,omitempty" xml:"limitations,omitempty"`           // Limitations in the scope.
	NetworkSegments interface{}      `json:"network_segments,omitempty" xml:"network_segments,omitempty"` // Network segments
	Users           []*User          `json:"users,omitempty" xml:"users,omitempty"`                       // Users
	UserGroups      interface{}      `json:"user_groups,omitempty" xml:"user_groups,omitempty"`           // User groups
}

// Limitations represents limitations within the scope of the {configuration profile, policy}.
type Limitations struct {
	IBeacons        []*JamfProperty `json:"ibeacons,omitempty" xml:"ibeacons,omitempty"`                 // iBeacons
	NetworkSegments interface{}     `json:"network_segments,omitempty" xml:"network_segments,omitempty"` // Network segments
	UserGroups      []*UserGroup    `json:"user_groups,omitempty" xml:"user_groups,omitempty"`           // User groups
	Users           []*User         `json:"users,omitempty" xml:"users,omitempty"`                       // Users
}

// Exclusions represents exclusions from the scope of the {configuration profile, policy}.
type Exclusions struct {
	Buildings       interface{}      `json:"buildings,omitempty" xml:"buildings,omitempty"`               // Buildings
	ComputerGroups  []*ComputerGroup `json:"computer_groups,omitempty" xml:"computer_groups,omitempty"`   // Computer groups.
	Computers       []*Computer      `json:"computers,omitempty" xml:"computers,omitempty"`               // Computers
	Departments     interface{}      `json:"departments,omitempty" xml:"departments,omitempty"`           // Departments
	IBeacons        []*JamfProperty  `json:"ibeacons,omitempty" xml:"ibeacons,omitempty"`                 // iBeacons
	JSSUserGroups   []*UserGroup     `json:"jss_user_groups,omitempty" xml:"jss_user_groups,omitempty"`   // JSS user groups
	JSSUsers        []*User          `json:"jss_users,omitempty" xml:"jss_users,omitempty"`               // JSS users
	NetworkSegments interface{}      `json:"network_segments,omitempty" xml:"network_segments,omitempty"` // Network segments
	UserGroups      []*UserGroup     `json:"user_groups,omitempty" xml:"user_groups,omitempty"`           // User groups
	Users           []*User          `json:"users,omitempty" xml:"users,omitempty"`                       // Users
}

// ComputerGroup represents a single computer group.
type ComputerGroup struct {
	*JamfProperty
}

// SelfService represents self-service configurations.
type SelfService struct {
	FeatureOnMainPage           bool        `json:"feature_on_main_page,omitempty" xml:"feature_on_main_page,omitempty"`                       // If featured on the main page.
	ForceUsersToViewDescription bool        `json:"force_users_to_view_description,omitempty" xml:"force_users_to_view_description,omitempty"` // If users are forced to view description.
	InstallButtonText           string      `json:"install_button_text,omitempty" xml:"install_button_text,omitempty"`                         // Text on the install button.
	Notification                interface{} `json:"notification,omitempty" xml:"notification,omitempty"`                                       // Notification settings
	NotificationMessage         interface{} `json:"notification_message,omitempty" xml:"notification_message,omitempty"`                       // Notification message
	NotificationSubject         string      `json:"notification_subject,omitempty" xml:"notification_subject,omitempty"`                       // Notification subject.
	RemovalDisallowed           string      `json:"removal_disallowed,omitempty" xml:"removal_disallowed,omitempty"`                           // Removal policy.
	Security                    *Security   `json:"security,omitempty" xml:"security,omitempty"`                                               // Security settings.
	SelfServiceCategories       interface{} `json:"self_service_categories,omitempty" xml:"self_service_categories,omitempty"`                 // Self-service categories
	SelfServiceDescription      interface{} `json:"self_service_description,omitempty" xml:"self_service_description,omitempty"`               // Self-service description
	SelfServiceDisplayName      string      `json:"self_service_display_name,omitempty" xml:"self_service_display_name,omitempty"`             // Display name in self-service.
	SelfServiceIcon             interface{} `json:"self_service_icon,omitempty" xml:"self_service_icon,omitempty"`                             // Self-service icon
}

// Security represents security configurations in self-service.
type Security struct {
	RemovalDisallowed string `json:"removal_disallowed,omitempty" xml:"removal_disallowed,omitempty"` // Removal policy.
}

// END OF JAMF {CONFIGURATION PROFILE, POLICY} STRUCTS
//---------------------------------------------------------------------

// ### Jamf History Structs
// ---------------------------------------------------------------------
// History represents the complete history and configuration of a computer
type History struct {
	ComputerHistory ComputerHistory `json:"computer_history,omitempty"`
}

// ComputerHistory contains the main sections of computer history
type ComputerHistory struct {
	Audits                  []HistoryRecord         `json:"audits,omitempty"`                     // List of audits performed on the computer
	CasperImagingLogs       []HistoryRecord         `json:"casper_imaging_logs,omitempty"`        // Logs of Casper imaging operations
	CasperRemoteLogs        []HistoryRecord         `json:"casper_remote_logs,omitempty"`         // Logs of Casper remote operations
	Commands                Commands                `json:"commands,omitempty"`                   // Information about completed, pending, and failed commands
	ComputerUsageLogs       []HistoryRecord         `json:"computer_usage_logs,omitempty"`        // Logs of computer usage events
	General                 HistoryRecordGeneral    `json:"general,omitempty"`                    // General information about the computer
	MacAppStoreApplications MacAppStoreApplications `json:"mac_app_store_applications,omitempty"` // Information about installed Mac App Store applications
	PolicyLogs              []PolicyLog             `json:"policy_logs,omitempty"`                // Logs of policy executions
	UserLocation            []UserLocation          `json:"user_location,omitempty"`              // User location information
}

// Record represents common fields of a a historical event
type HistoryRecord struct {
	DateTime      string `json:"date_time,omitempty"`       // Date and time of the log entry
	DateTimeEpoch int64  `json:"date_time_epoch,omitempty"` // Date and time in epoch format
	DateTimeUTC   string `json:"date_time_utc,omitempty"`   // Date and time in UTC
	Event         string `json:"event,omitempty"`           // Type of event logged
	Username      string `json:"username,omitempty"`        // Username associated with the event
	Status        string `json:"status,omitempty"`          // Status of the event
}

// HistoryRecordApp represents a single application installation from the Mac App Store
type HistoryRecordApp struct {
	Deployed        string `json:"deployed,omitempty"`          // Time when the app deployment was attempted
	DeployedEpoch   int64  `json:"deployed_epoch,omitempty"`    // Deployment attempt time in epoch format
	DeployedUTC     string `json:"deployed_utc,omitempty"`      // Deployment attempt time in UTC
	LastUpdate      string `json:"last_update,omitempty"`       // Time of the last update
	LastUpdateEpoch int64  `json:"last_update_epoch,omitempty"` // Last update time in epoch format
	LastUpdateUTC   string `json:"last_update_utc,omitempty"`   // Last update time in UTC
	Name            string `json:"name,omitempty"`              // Name of the application
	Status          string `json:"status,omitempty"`            // Status of the failed installation
	SizeMB          int    `json:"size_mb,omitempty"`           // Size of the application in megabytes
	Version         string `json:"version,omitempty"`           // Version of the application
}

// HistoryRecordGeneral contains general information about the computer
type HistoryRecordGeneral struct {
	ID           int    `json:"id,omitempty"`            // Identifier of the computer
	MacAddress   string `json:"mac_address,omitempty"`   // MAC address of the computer
	Name         string `json:"name,omitempty"`          // Name of the computer
	SerialNumber string `json:"serial_number,omitempty"` // Serial number of the computer
	UDID         string `json:"udid,omitempty"`          // Unique Device Identifier
}

// Commands represents the status of various commands executed on the computer
type Commands struct {
	Completed []Command `json:"completed,omitempty"` // List of completed commands
	Failed    []Command `json:"failed,omitempty"`    // List of failed commands
	Pending   []Command `json:"pending,omitempty"`   // List of pending commands
}

// Command represents a single command executed on the computer
type Command struct {
	Completed      string `json:"completed,omitempty"`       // Completion time of the command
	CompletedEpoch int64  `json:"completed_epoch,omitempty"` // Completion time in epoch format
	CompletedUTC   string `json:"completed_utc,omitempty"`   // Completion time in UTC
	Failed         string `json:"failed,omitempty"`          // Time when the command failed
	FailedEpoch    int64  `json:"failed_epoch,omitempty"`    // Failure time in epoch format
	FailedUTC      string `json:"failed_utc,omitempty"`      // Failure time in UTC
	Issued         string `json:"issued,omitempty"`          // Time when the command was issued
	IssuedEpoch    int64  `json:"issued_epoch,omitempty"`    // Issue time in epoch format
	IssuedUTC      string `json:"issued_utc,omitempty"`      // Issue time in UTC
	LastPush       string `json:"last_push,omitempty"`       // Time of the last push notification
	LastPushEpoch  int64  `json:"last_push_epoch,omitempty"` // Last push time in epoch format
	LastPushUTC    string `json:"last_push_utc,omitempty"`   // Last push time in UTC
	Name           string `json:"name,omitempty"`            // Name of the command
	Status         string `json:"status,omitempty"`          // Status of the command
	Username       string `json:"username,omitempty"`        // Username associated with the command
}

// MacAppStoreApplications contains information about installed Mac App Store applications
type MacAppStoreApplications struct {
	Failed    []HistoryRecordApp `json:"failed,omitempty"`    // List of failed application installations
	Installed []HistoryRecordApp `json:"installed,omitempty"` // List of installed applications
	Pending   []HistoryRecordApp `json:"pending,omitempty"`   // List of pending application installations
}

// PolicyLog represents a single policy execution log
type PolicyLog struct {
	DateCompleted      string `json:"date_completed,omitempty"`       // Date and time when the policy was completed
	DateCompletedEpoch int64  `json:"date_completed_epoch,omitempty"` // Completion date and time in epoch format
	DateCompletedUTC   string `json:"date_completed_utc,omitempty"`   // Completion date and time in UTC
	PolicyID           int    `json:"policy_id,omitempty"`            // Identifier of the policy
	PolicyName         string `json:"policy_name,omitempty"`          // Name of the policy
	Status             string `json:"status,omitempty"`               // Status of the policy execution
	Username           string `json:"username,omitempty"`             // Username associated with the policy execution
}

// UserLocation represents the location information of a user
type UserLocation struct {
	*HistoryRecord
	Building     string `json:"building,omitempty"`      // Building where the user is located
	Department   string `json:"department,omitempty"`    // Department of the user
	EmailAddress string `json:"email_address,omitempty"` // Email address of the user
	FullName     string `json:"full_name,omitempty"`     // Full name of the user
	PhoneNumber  string `json:"phone_number,omitempty"`  // Phone number of the user
	Position     string `json:"position,omitempty"`      // Position or job title of the user
	Room         string `json:"room,omitempty"`          // Room where the user is located
}

// END OF JAMF COMPUTER HISTORY STRUCTS
//---------------------------------------------------------------------

// ### Jamf Web Admin Structs
// ---------------------------------------------------------------------
// JamfUsers represents a list of user accounts from the Jamf Admin WebUI
type JamfUsers []JamfUser

// JamfUser represents a single user account record.
type JamfUser struct {
	ID                        int    `json:"userId,omitempty"`                    // The unique identifier of the user.
	Username                  string `json:"username,omitempty"`                  // The user's login name.
	RealName                  string `json:"realname,omitempty"`                  // The user's full name.
	Password                  string `json:"password,omitempty"`                  // The user's password.
	PasswordModified          int64  `json:"passwordModified,omitempty"`          // Timestamp (in seconds) when the password was last modified.
	Email                     string `json:"email,omitempty"`                     // The user's email address.
	Phone                     string `json:"phone,omitempty"`                     // The user's phone number.
	LDAPServerID              string `json:"ldapServerId,omitempty"`              // The LDAP server used for authentication.
	DistinguishedName         string `json:"distinguishedName,omitempty"`         // The user's distinguished name in LDAP.
	SiteID                    int    `json:"siteId,omitempty"`                    // The site ID associated with the user.
	AccessLevel               string `json:"accessLevel,omitempty"`               // The user's access level (e.g., "Full Access", "Limited Access").
	PrivilegeLevel            string `json:"privilegeLevel,omitempty"`            // The user's privilege level (e.g., "ADMINISTRATOR", "CUSTOM").
	LastPasswordChange        string `json:"lastPasswordChange,omitempty"`        // The date and time when the password was last changed.
	ChangePasswordOnNextLogin bool   `json:"changePasswordOnNextLogin,omitempty"` // Indicates if the user must change the password upon next login.
	FailedLoginAttempts       int    `json:"failedLoginAttempts,omitempty"`       // The number of failed login attempts.
	ForcePasswordChange       bool   `json:"forcePasswordChange,omitempty"`       // Indicates if the user must change the password upon next login.
	Expires                   string `json:"expires,omitempty"`                   // The expiration date of the account in ISO 8601 format.
	WebAdmin                  bool   `json:"webAdmin,omitempty"`                  // Indicates if the user has web administration privileges.
	AccountStatus             string `json:"accountStatus,omitempty"`             // The status of the user account (e.g., "Enabled", "Disabled").
	Disabled                  bool   `json:"disabled,omitempty"`                  // Indicates if the user account is disabled.
}

// END OF JAMF WEB ADMIN STRUCTS
//---------------------------------------------------------------------

// ### Enums
// --------------------------------------------------------------------
// Inteded for Device Query parameters, `Sections` serves as a namespace for valid Computer Detail section constants.
type Sections struct {
	General               string
	DiskEncryption        string
	Purchasing            string
	Applications          string
	Storage               string
	UserAndLocation       string
	ConfigurationProfiles string
	Printers              string
	Services              string
	Hardware              string
	LocalUserAccounts     string
	Certificates          string
	Attachments           string
	Plugins               string
	PackageReceipts       string
	Fonts                 string
	Security              string
	OperatingSystem       string
	LicensedSoftware      string
	IBeacons              string
	SoftwareUpdates       string
	ExtensionAttributes   string
	ContentCaching        string
	GroupMemberships      string
}

// Section is an instance of the Sections struct, where we assign the constants.
var Section = Sections{
	General:               "GENERAL",
	DiskEncryption:        "DISK_ENCRYPTION",
	Purchasing:            "PURCHASING",
	Applications:          "APPLICATIONS",
	Storage:               "STORAGE",
	UserAndLocation:       "USER_AND_LOCATION",
	ConfigurationProfiles: "CONFIGURATION_PROFILES",
	Printers:              "PRINTERS",
	Services:              "SERVICES",
	Hardware:              "HARDWARE",
	LocalUserAccounts:     "LOCAL_USER_ACCOUNTS",
	Certificates:          "CERTIFICATES",
	Attachments:           "ATTACHMENTS",
	Plugins:               "PLUGINS",
	PackageReceipts:       "PACKAGE_RECEIPTS",
	Fonts:                 "FONTS",
	Security:              "SECURITY",
	OperatingSystem:       "OPERATING_SYSTEM",
	LicensedSoftware:      "LICENSED_SOFTWARE",
	IBeacons:              "IBEACONS",
	SoftwareUpdates:       "SOFTWARE_UPDATES",
	ExtensionAttributes:   "EXTENSION_ATTRIBUTES",
	ContentCaching:        "CONTENT_CACHING",
	GroupMemberships:      "GROUP_MEMBERSHIPS",
}

// Inteded for Device Query parameters, `SortOptions serves as a namespace for valid sort criteria constants.
type SortOptions struct {
	GeneralName                          string
	UDID                                 string
	ID                                   string
	GeneralAssetTag                      string
	GeneralJamfBinaryVersion             string
	GeneralLastContactTime               string
	GeneralLastEnrolledDate              string
	GeneralLastCloudBackupDate           string
	GeneralReportDate                    string
	GeneralRemoteManagementUsername      string
	GeneralMDMCertificateExpiration      string
	GeneralPlatform                      string
	HardwareMake                         string
	HardwareModel                        string
	OperatingSystemBuild                 string
	OperatingSystemSupplementalBuild     string
	OperatingSystemRapidSecurityResponse string
	OperatingSystemName                  string
	OperatingSystemVersion               string
	UserAndLocationRealname              string
	PurchasingLifeExpectancy             string
	PurchasingWarrantyDate               string
}

// Sort is an instance of the SortOptions struct, where we assign the constants.
var Sort = SortOptions{
	GeneralName:                          "general.name",
	UDID:                                 "udid",
	ID:                                   "id",
	GeneralAssetTag:                      "general.assetTag",
	GeneralJamfBinaryVersion:             "general.jamfBinaryVersion",
	GeneralLastContactTime:               "general.lastContactTime",
	GeneralLastEnrolledDate:              "general.lastEnrolledDate",
	GeneralLastCloudBackupDate:           "general.lastCloudBackupDate",
	GeneralReportDate:                    "general.reportDate",
	GeneralRemoteManagementUsername:      "general.remoteManagement.managementUsername",
	GeneralMDMCertificateExpiration:      "general.mdmCertificateExpiration",
	GeneralPlatform:                      "general.platform",
	HardwareMake:                         "hardware.make",
	HardwareModel:                        "hardware.model",
	OperatingSystemBuild:                 "operatingSystem.build",
	OperatingSystemSupplementalBuild:     "operatingSystem.supplementalBuildVersion",
	OperatingSystemRapidSecurityResponse: "operatingSystem.rapidSecurityResponse",
	OperatingSystemName:                  "operatingSystem.name",
	OperatingSystemVersion:               "operatingSystem.version",
	UserAndLocationRealname:              "userAndLocation.realname",
	PurchasingLifeExpectancy:             "purchasing.lifeExpectancy",
	PurchasingWarrantyDate:               "purchasing.warrantyDate",
}
