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

// Response structure for the Jamf Pro API for mobile devices
type MobileDevices struct {
	Results    *[]*MobileDevice `json:"results"`    // List of mobile devices.
	TotalCount int              `json:"totalCount"` // Total number of mobile devices.
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
	IBeacons              *[]*IBeacon              `json:"ibeacons,omitempty"`              // iBeacons associated with the computer.
	ID                    string                   `json:"id"`                              // Unique identifier for the computer.
	LicensedSoftware      *[]*LicensedSoftware     `json:"licensedSoftware,omitempty"`      // List of licensed software.
	LocalUserAccounts     *[]*LocalUserAccount     `json:"localUserAccounts,omitempty"`     // List of local user accounts on the computer.
	OperatingSystem       *OperatingSystem         `json:"operatingSystem,omitempty"`       // Operating system details.
	PackageReceipts       *PackageReceipts         `json:"packageReceipts,omitempty"`       // Information about package receipts.
	Plugins               *[]*Plugin               `json:"plugins,omitempty"`               // List of plugins installed on the computer.
	Printers              *[]*Printer              `json:"printers,omitempty"`              // List of printers configured on the computer.
	Purchasing            *Purchasing              `json:"purchasing,omitempty"`            // Purchasing information.
	Security              *Security                `json:"security,omitempty"`              // Security settings and information.
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
	Id        string `json:"id,omitempty"`        // ID of the attachment.
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

type ConfigurationProfile struct {
	DisplayName       string `json:"displayName,omitempty"`       // Display name of the configuration profile.
	ID                string `json:"id,omitempty"`                // Identifier of the configuration profile.
	LastInstalled     string `json:"lastInstalled,omitempty"`     // Last installed date of the configuration profile.
	ProfileIdentifier string `json:"profileIdentifier,omitempty"` // Profile identifier of the configuration profile.
	Removable         bool   `json:"removable,omitempty"`         // Indicates if the profile is removable.
	Username          string `json:"username,omitempty"`          // Username associated with the configuration profile.
}

// ContentCaching represents content caching information of a computer.
type ContentCaching struct {
	Activated                           bool               `json:"activated,omitempty"`                           // Indicates if caching is activated.
	Active                              bool               `json:"active,omitempty"`                              // Indicates if caching is currently active.
	ActualCacheBytesUsed                int64              `json:"actualCacheBytesUsed,omitempty"`                // Actual bytes used by the cache.
	Address                             string             `json:"address,omitempty"`                             // Address of the caching server.
	Addresses                           []string           `json:"addresses,omitempty"`                           // List of addresses related to the caching server.
	Alerts                              []CacheAlert       `json:"alerts,omitempty"`                              // Alerts related to caching.
	CacheBytesFree                      int64              `json:"cacheBytesFree,omitempty"`                      // Free bytes in the cache.
	CacheBytesLimit                     int64              `json:"cacheBytesLimit,omitempty"`                     // Limit of cache bytes.
	CacheBytesUsed                      int64              `json:"cacheBytesUsed,omitempty"`                      // Bytes used in the cache.
	CacheDetails                        []CacheDetail      `json:"cacheDetails,omitempty"`                        // Details about the cache.
	CacheStatus                         string             `json:"cacheStatus,omitempty"`                         // Status of the cache.
	ComputerContentCachingInformationID string             `json:"computerContentCachingInformationId,omitempty"` // Information ID for content caching.
	DataMigrationCompleted              bool               `json:"dataMigrationCompleted,omitempty"`              // Indicates if data migration is completed.
	DataMigrationError                  DataMigrationError `json:"dataMigrationError,omitempty"`                  // Data migration error details.
	DataMigrationProgressPercentage     int                `json:"dataMigrationProgressPercentage,omitempty"`     // Progress percentage of data migration.
	Details                             []CacheDetail      `json:"details,omitempty"`                             // Details of the content caching. Seems redundant, but it helps with struct recursion.
	GUID                                string             `json:"guid,omitempty"`                                // GUID of the caching server.
	Healthy                             bool               `json:"healthy,omitempty"`                             // Indicates if the caching server is healthy.
	MaxCachePressureLast1HourPercentage int                `json:"maxCachePressureLast1HourPercentage,omitempty"` // Max cache pressure in the last hour.
	Parents                             []ContentCaching   `json:"parents,omitempty"`                             // Parent caching servers, recursively using the same structure.
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
	CacheBytesLimit      int64  `json:"cacheBytesLimit,omitempty"`      // Limit of cache bytes for the alert.
	ClassName            string `json:"className,omitempty"`            // The class name of the alert.
	PathPreventingAccess string `json:"pathPreventingAccess,omitempty"` // Path that is preventing access.
	PostDate             string `json:"postDate,omitempty"`             // Post date of the alert.
	ReservedVolumeBytes  int64  `json:"reservedVolumeBytes,omitempty"`  // Reserved volume bytes.
	Resource             string `json:"resource,omitempty"`             // Resource associated with the alert.
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
	MdmCapable                           MdmCapable           `json:"mdmCapable"`                           // MDM capability information.
	MdmProfileExpiration                 string               `json:"mdmProfileExpiration"`                 // Expiration of the MDM profile.
	Name                                 string               `json:"name"`                                 // Name of the computer.
	Platform                             string               `json:"platform"`                             // Platform of the computer (e.g., Mac).
	RemoteManagement                     RemoteManagement     `json:"remoteManagement"`                     // Remote management information.
	ReportDate                           string               `json:"reportDate"`                           // Date of report.
	Site                                 Site                 `json:"site"`                                 // Site information.
	Supervised                           bool                 `json:"supervised"`                           // Indicates if the device is supervised.
	UserApprovedMdm                      bool                 `json:"userApprovedMdm"`                      // Indicates if MDM is user-approved.
	DeclarativeDeviceManagementEnabled   bool                 `json:"declarativeDeviceManagementEnabled"`   // Indicates if declarative device management is enabled.
}

// GroupMembership represents the membership details of a computer in a group.
type GroupMembership struct {
	GroupId    string `json:"groupId,omitempty"`    // Unique identifier of the group.
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

// MdmCapable represents MDM capability information of a computer.
type MdmCapable struct {
	Capable      bool     `json:"capable"`      // Indicates if the computer is MDM capable.
	CapableUsers []string `json:"capableUsers"` // List of users capable of MDM.
}

// RemoteManagement represents remote management information of a computer.
type RemoteManagement struct {
	Managed            bool   `json:"managed"`            // Indicates if the computer is managed.
	ManagementUsername string `json:"managementUsername"` // Username for management.
}

// Site represents the site information of a computer.
type Site struct {
	ID   string `json:"id"`   // Identifier of the site.
	Name string `json:"name"` // Name of the site.
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

// IBeacon represents an iBeacon associated with the computer.
type IBeacon struct {
	Name string `json:"name,omitempty"` // Name of the iBeacon.
}

// LicensedSoftware represents licensed software installed on the computer.
type LicensedSoftware struct {
	Id   string `json:"id,omitempty"`   // ID of the licensed software.
	Name string `json:"name,omitempty"` // Name of the licensed software.
}

// LocalUserAccount represents a local user account on the computer.
type LocalUserAccount struct {
	Admin                          bool   `json:"admin,omitempty"`                          // Indicates if the user is an admin.
	AzureActiveDirectoryId         string `json:"azureActiveDirectoryId,omitempty"`         // Azure Active Directory ID.
	ComputerAzureActiveDirectoryId string `json:"computerAzureActiveDirectoryId,omitempty"` // Computer's Azure Active Directory ID.
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
	UserAzureActiveDirectoryId     string `json:"userAzureActiveDirectoryId,omitempty"`     // User's Azure Active Directory ID.
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
	SoftwareUpdateDeviceId   string               `json:"softwareUpdateDeviceId,omitempty"`   // Software Update Device ID.
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
type Security struct {
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
