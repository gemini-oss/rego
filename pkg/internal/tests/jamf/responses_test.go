// pkg/internal/tests/jamf/responses_test.go
package jamf_test

var (
	ComputerGroups = map[string]string{
		"/api/v1/computer-groups": `[
				{
					"id": "1",
					"name": "All Managed Computers",
					"smartGroup": true
				},
				{
					"id": "2",
					"name": "Special Computers",
					"smartGroup": false
				},
				{
					"id": "3",
					"name": "Old Computers",
					"smartGroup": true
				}
			]`,
	}
	ComputersInventory = map[string]string{
		"/api/v1/computers-inventory": `{
				"totalCount": 3,
				"results": [
				  {
					"id": "1",
					"udid": "123",
					"general": {
					  "name": "Boalime",
					  "lastIpAddress": "247.185.82.186",
					  "lastReportedIp": "247.185.82.186",
					  "jamfBinaryVersion": "9.27",
					  "platform": "Mac",
					  "barcode1": "5 12345 678900",
					  "barcode2": "5 12345 678900",
					  "assetTag": "304822",
					  "remoteManagement": {
						"managed": true,
						"managementUsername": "rootname"
					  },
					  "supervised": true,
					  "mdmCapable": {
						"capable": true,
						"capableUsers": [
						  "admin",
						  "rootadmin"
						]
					  },
					  "reportDate": "2018-10-31T18:04:13Z",
					  "lastContactTime": "2018-10-31T18:04:13Z",
					  "lastCloudBackupDate": "2018-10-31T18:04:13Z",
					  "lastEnrolledDate": "2018-10-31T18:04:13Z",
					  "mdmProfileExpiration": "2018-10-31T18:04:13Z",
					  "initialEntryDate": "2018-10-31",
					  "distributionPoint": "distribution point name",
					  "enrollmentMethod": {
						"id": "1",
						"objectName": "user@domain.com",
						"objectType": "User-initiated - no invitation"
					  },
					  "site": {
						"id": "1",
						"name": "Eau Claire"
					  },
					  "itunesStoreAccountActive": true,
					  "enrolledViaAutomatedDeviceEnrollment": true,
					  "userApprovedMdm": true,
					  "declarativeDeviceManagementEnabled": true,
					  "extensionAttributes": [
						{
						  "definitionId": "23",
						  "name": "Some Attribute",
						  "description": "Some Attribute defines how much Foo impacts Bar.",
						  "enabled": true,
						  "multiValue": true,
						  "values": [
							"foo",
							"bar"
						  ],
						  "dataType": "STRING",
						  "options": [
							"foo",
							"bar"
						  ],
						  "inputType": "TEXT"
						}
					  ],
					  "managementId": "73226fb6-61df-4c10-9552-eb9bc353d507"
					},
					"diskEncryption": {
					  "bootPartitionEncryptionDetails": {
						"partitionName": "main",
						"partitionFileVault2State": "VALID",
						"partitionFileVault2Percent": 100
					  },
					  "individualRecoveryKeyValidityStatus": "VALID",
					  "institutionalRecoveryKeyPresent": true,
					  "diskEncryptionConfigurationName": "Test configuration",
					  "fileVault2EnabledUserNames": [
						"admin"
					  ],
					  "fileVault2EligibilityMessage": "Not a boot partition"
					},
					"purchasing": {
					  "leased": true,
					  "purchased": true,
					  "poNumber": "53-1",
					  "poDate": "2019-01-01",
					  "vendor": "Example Vendor",
					  "warrantyDate": "2019-01-01",
					  "appleCareId": "abcd",
					  "leaseDate": "2019-01-01",
					  "purchasePrice": "$500",
					  "lifeExpectancy": 5,
					  "purchasingAccount": "admin",
					  "purchasingContact": "true",
					  "extensionAttributes": [
						{
						  "definitionId": "23",
						  "name": "Some Attribute",
						  "description": "Some Attribute defines how much Foo impacts Bar.",
						  "enabled": true,
						  "multiValue": true,
						  "values": [
							"foo",
							"bar"
						  ],
						  "dataType": "STRING",
						  "options": [
							"foo",
							"bar"
						  ],
						  "inputType": "TEXT"
						}
					  ]
					},
					"applications": [
					  {
						"name": "Microsoft Word",
						"path": "/usr/local/app",
						"version": "1.0.0",
						"macAppStore": true,
						"sizeMegabytes": 25,
						"bundleId": "1",
						"updateAvailable": false,
						"externalVersionId": "1"
					  }
					],
					"storage": {
					  "bootDriveAvailableSpaceMegabytes": 3072,
					  "disks": [
						{
						  "id": "170",
						  "device": "disk0",
						  "model": "APPLE HDD TOSHIBA MK5065GSXF",
						  "revision": "5",
						  "serialNumber": "a8598f013366",
						  "sizeMegabytes": 262144,
						  "smartStatus": "OK",
						  "type": "false",
						  "partitions": [
							{
							  "name": "Foo",
							  "sizeMegabytes": 262144,
							  "availableMegabytes": 131072,
							  "partitionType": "BOOT",
							  "percentUsed": 25,
							  "fileVault2State": "VALID",
							  "fileVault2ProgressPercent": 45,
							  "lvmManaged": true
							}
						  ]
						}
					  ]
					},
					"userAndLocation": {
					  "username": "Madison Anderson",
					  "realname": "13-inch MacBook",
					  "email": "email@com.pl",
					  "position": "IT Team Lead",
					  "phone": "123-456-789",
					  "departmentId": "1",
					  "buildingId": "1",
					  "room": "5",
					  "extensionAttributes": [
						{
						  "definitionId": "23",
						  "name": "Some Attribute",
						  "description": "Some Attribute defines how much Foo impacts Bar.",
						  "enabled": true,
						  "multiValue": true,
						  "values": [
							"foo",
							"bar"
						  ],
						  "dataType": "STRING",
						  "options": [
							"foo",
							"bar"
						  ],
						  "inputType": "TEXT"
						}
					  ]
					},
					"configurationProfiles": [
					  {
						"id": "1",
						"username": "username",
						"lastInstalled": "2018-10-31T18:04:13Z",
						"removable": true,
						"displayName": "Displayed profile",
						"profileIdentifier": "0ae590fe-9b30-11ea-bb37-0242ac130002"
					  }
					],
					"printers": [
					  {
						"name": "My Printer",
						"type": "XYZ 1122",
						"uri": "ipp://10.0.0.5",
						"location": "7th floor"
					  }
					],
					"services": [
					  {
						"name": "SomeService"
					  }
					],
					"hardware": {
					  "make": "Apple",
					  "model": "13-inch MacBook Pro (Mid 2012)",
					  "modelIdentifier": "MacBookPro9,2",
					  "serialNumber": "C02ZC2QYLVDL",
					  "processorSpeedMhz": 2100,
					  "processorCount": 2,
					  "coreCount": 2,
					  "processorType": "Intel Core i5",
					  "processorArchitecture": "i386",
					  "busSpeedMhz": 2133,
					  "cacheSizeKilobytes": 3072,
					  "networkAdapterType": "Foo",
					  "macAddress": "6A:2C:4B:B7:65:B5",
					  "altNetworkAdapterType": "Bar",
					  "altMacAddress": "82:45:58:44:dc:01",
					  "totalRamMegabytes": 4096,
					  "openRamSlots": 0,
					  "batteryCapacityPercent": 85,
					  "smcVersion": "2.2f38",
					  "nicSpeed": "N/A",
					  "opticalDrive": "MATSHITA DVD-R UJ-8A8",
					  "bootRom": "MBP91.00D3.B08",
					  "bleCapable": false,
					  "supportsIosAppInstalls": false,
					  "appleSilicon": false,
					  "extensionAttributes": [
						{
						  "definitionId": "23",
						  "name": "Some Attribute",
						  "description": "Some Attribute defines how much Foo impacts Bar.",
						  "enabled": true,
						  "multiValue": true,
						  "values": [
							"foo",
							"bar"
						  ],
						  "dataType": "STRING",
						  "options": [
							"foo",
							"bar"
						  ],
						  "inputType": "TEXT"
						}
					  ]
					},
					"localUserAccounts": [
					  {
						"uid": "501",
						"userGuid": "844F1177-0CF5-40C6-901F-38EDD9969C1C",
						"username": "jamf",
						"fullName": "John Jamf",
						"admin": true,
						"homeDirectory": "/Users/jamf",
						"homeDirectorySizeMb": 131072,
						"fileVault2Enabled": true,
						"userAccountType": "LOCAL",
						"passwordMinLength": 4,
						"passwordMaxAge": 5,
						"passwordMinComplexCharacters": 5,
						"passwordHistoryDepth": 5,
						"passwordRequireAlphanumeric": true,
						"computerAzureActiveDirectoryId": "1",
						"userAzureActiveDirectoryId": "1",
						"azureActiveDirectoryId": "ACTIVATED"
					  }
					],
					"certificates": [
					  {
						"commonName": "jamf.com",
						"identity": true,
						"expirationDate": "2030-10-31T18:04:13Z",
						"username": "test",
						"lifecycleStatus": "ACTIVE",
						"certificateStatus": "ISSUED",
						"subjectName": "CN=jamf.com",
						"serialNumber": "40f3d9fb",
						"sha1Fingerprint": "ed361458724d06082b2314acdb82e1f586f085f5",
						"issuedDate": "2022-05-23T14:54:10Z"
					  }
					],
					"attachments": [
					  {
						"id": "1",
						"name": "Attachment.pdf",
						"fileType": "application/pdf",
						"sizeBytes": 1024
					  }
					],
					"plugins": [
					  {
						"name": "plugin name",
						"version": "1.02",
						"path": "/Applications/"
					  }
					],
					"packageReceipts": {
					  "installedByJamfPro": [
						"com.jamf.protect.JamfProtect"
					  ],
					  "installedByInstallerSwu": [
						"com.apple.pkg.Core"
					  ],
					  "cached": [
						"com.jamf.protect.JamfProtect"
					  ]
					},
					"fonts": [
					  {
						"name": "font name",
						"version": "1.02",
						"path": "/Applications/"
					  }
					],
					"security": {
					  "sipStatus": "ENABLED",
					  "gatekeeperStatus": "APP_STORE_AND_IDENTIFIED_DEVELOPERS",
					  "xprotectVersion": "1.2.3",
					  "autoLoginDisabled": false,
					  "remoteDesktopEnabled": true,
					  "activationLockEnabled": true,
					  "recoveryLockEnabled": true,
					  "firewallEnabled": true,
					  "secureBootLevel": "FULL_SECURITY",
					  "externalBootLevel": "ALLOW_BOOTING_FROM_EXTERNAL_MEDIA",
					  "bootstrapTokenAllowed": true,
					  "bootstrapTokenEscrowedStatus": "ESCROWED"
					},
					"operatingSystem": {
					  "name": "Mac OS X",
					  "version": "10.9.5",
					  "build": "13A603",
					  "supplementalBuildVersion": "13A953",
					  "rapidSecurityResponse": "(a)",
					  "activeDirectoryStatus": "Not Bound",
					  "fileVault2Status": "ALL_ENCRYPTED",
					  "softwareUpdateDeviceId": "J132AP",
					  "extensionAttributes": [
						{
						  "definitionId": "23",
						  "name": "Some Attribute",
						  "description": "Some Attribute defines how much Foo impacts Bar.",
						  "enabled": true,
						  "multiValue": true,
						  "values": [
							"foo",
							"bar"
						  ],
						  "dataType": "STRING",
						  "options": [
							"foo",
							"bar"
						  ],
						  "inputType": "TEXT"
						}
					  ]
					},
					"licensedSoftware": [
					  {
						"id": "1",
						"name": "Microsoft Word"
					  }
					],
					"ibeacons": [
					  {
						"name": "room A"
					  }
					],
					"softwareUpdates": [
					  {
						"name": "BEdit",
						"version": "1.15.2",
						"packageName": "com.apple.pkg.AdditionalEssentials"
					  }
					],
					"extensionAttributes": [
					  {
						"definitionId": "23",
						"name": "Some Attribute",
						"description": "Some Attribute defines how much Foo impacts Bar.",
						"enabled": true,
						"multiValue": true,
						"values": [
						  "foo",
						  "bar"
						],
						"dataType": "STRING",
						"options": [
						  "foo",
						  "bar"
						],
						"inputType": "TEXT"
					  }
					],
					"contentCaching": {
					  "computerContentCachingInformationId": "1",
					  "parents": [
						{
						  "contentCachingParentId": "1",
						  "address": "SomeAddress",
						  "alerts": {
							"contentCachingParentAlertId": "1",
							"addresses": [],
							"className": "SomeClass",
							"postDate": "2018-10-31T18:04:13Z"
						  },
						  "details": {
							"contentCachingParentDetailsId": "1",
							"acPower": true,
							"cacheSizeBytes": 0,
							"capabilities": {
							  "contentCachingParentCapabilitiesId": "1",
							  "imports": true,
							  "namespaces": true,
							  "personalContent": true,
							  "queryParameters": true,
							  "sharedContent": true,
							  "prioritization": true
							},
							"portable": true,
							"localNetwork": [
							  {
								"contentCachingParentLocalNetworkId": "1",
								"speed": 5000,
								"wired": true
							  }
							]
						  },
						  "guid": "CD1E1291-4AF9-4468-B5D5-0F780C13DB2F",
						  "healthy": true,
						  "port": 0,
						  "version": "1"
						}
					  ],
					  "alerts": [
						{
						  "cacheBytesLimit": 0,
						  "className": "SomeClass",
						  "pathPreventingAccess": "/some/path",
						  "postDate": "2018-10-31T18:04:13Z",
						  "reservedVolumeBytes": 0,
						  "resource": "SomeResource"
						}
					  ],
					  "activated": false,
					  "active": false,
					  "actualCacheBytesUsed": 0,
					  "cacheDetails": [
						{
						  "computerContentCachingCacheDetailsId": "1",
						  "categoryName": "SomeCategory",
						  "diskSpaceBytesUsed": 0
						}
					  ],
					  "cacheBytesFree": 23353884672,
					  "cacheBytesLimit": 0,
					  "cacheStatus": "OK",
					  "cacheBytesUsed": 0,
					  "dataMigrationCompleted": false,
					  "dataMigrationProgressPercentage": 0,
					  "dataMigrationError": {
						"code": 0,
						"domain": "SomeDomain",
						"userInfo": [
						  {
							"key": "foo",
							"value": "bar"
						  }
						]
					  },
					  "maxCachePressureLast1HourPercentage": 0,
					  "personalCacheBytesFree": 23353884672,
					  "personalCacheBytesLimit": 0,
					  "personalCacheBytesUsed": 0,
					  "port": 0,
					  "publicAddress": "SomeAddress",
					  "registrationError": "NOT_ACTIVATED",
					  "registrationResponseCode": 403,
					  "registrationStarted": "2018-10-31T18:04:13Z",
					  "registrationStatus": "CONTENT_CACHING_FAILED",
					  "restrictedMedia": false,
					  "serverGuid": "CD1E1291-4AF9-4468-B5D5-0F780C13DB2F",
					  "startupStatus": "FAILED",
					  "tetheratorStatus": "CONTENT_CACHING_DISABLED",
					  "totalBytesAreSince": "2018-10-31T18:04:13Z",
					  "totalBytesDropped": 0,
					  "totalBytesImported": 0,
					  "totalBytesReturnedToChildren": 0,
					  "totalBytesReturnedToClients": 0,
					  "totalBytesReturnedToPeers": 0,
					  "totalBytesStoredFromOrigin": 0,
					  "totalBytesStoredFromParents": 0,
					  "totalBytesStoredFromPeers": 0
					},
					"groupMemberships": [
					  {
						"groupId": "1",
						"groupName": "groupOne",
						"smartGroup": true
					  }
					]
				  }
				]
			}
		`,
	}
	ComputersInventoryDetail = map[string]string{
		"/api/v1/computers-inventory-detail/1": `{
			"id": "1",
			"udid": "123",
			"general": {
			  "name": "Boalime",
			  "lastIpAddress": "247.185.82.186",
			  "lastReportedIp": "247.185.82.186",
			  "jamfBinaryVersion": "9.27",
			  "platform": "Mac",
			  "barcode1": "5 12345 678900",
			  "barcode2": "5 12345 678900",
			  "assetTag": "304822",
			  "remoteManagement": {
				"managed": true,
				"managementUsername": "rootname"
			  },
			  "supervised": true,
			  "mdmCapable": {
				"capable": true,
				"capableUsers": [
				  "admin",
				  "rootadmin"
				]
			  },
			  "reportDate": "2018-10-31T18:04:13Z",
			  "lastContactTime": "2018-10-31T18:04:13Z",
			  "lastCloudBackupDate": "2018-10-31T18:04:13Z",
			  "lastEnrolledDate": "2018-10-31T18:04:13Z",
			  "mdmProfileExpiration": "2018-10-31T18:04:13Z",
			  "initialEntryDate": "2018-10-31",
			  "distributionPoint": "distribution point name",
			  "enrollmentMethod": {
				"id": "1",
				"objectName": "user@domain.com",
				"objectType": "User-initiated - no invitation"
			  },
			  "site": {
				"id": "1",
				"name": "Eau Claire"
			  },
			  "itunesStoreAccountActive": true,
			  "enrolledViaAutomatedDeviceEnrollment": true,
			  "userApprovedMdm": true,
			  "declarativeDeviceManagementEnabled": true,
			  "extensionAttributes": [
				{
				  "definitionId": "23",
				  "name": "Some Attribute",
				  "description": "Some Attribute defines how much Foo impacts Bar.",
				  "enabled": true,
				  "multiValue": true,
				  "values": [
					"foo",
					"bar"
				  ],
				  "dataType": "STRING",
				  "options": [
					"foo",
					"bar"
				  ],
				  "inputType": "TEXT"
				}
			  ],
			  "managementId": "73226fb6-61df-4c10-9552-eb9bc353d507"
			},
			"diskEncryption": {
			  "bootPartitionEncryptionDetails": {
				"partitionName": "main",
				"partitionFileVault2State": "VALID",
				"partitionFileVault2Percent": 100
			  },
			  "individualRecoveryKeyValidityStatus": "VALID",
			  "institutionalRecoveryKeyPresent": true,
			  "diskEncryptionConfigurationName": "Test configuration",
			  "fileVault2EnabledUserNames": [
				"admin"
			  ],
			  "fileVault2EligibilityMessage": "Not a boot partition"
			},
			"purchasing": {
			  "leased": true,
			  "purchased": true,
			  "poNumber": "53-1",
			  "poDate": "2019-01-01",
			  "vendor": "Example Vendor",
			  "warrantyDate": "2019-01-01",
			  "appleCareId": "abcd",
			  "leaseDate": "2019-01-01",
			  "purchasePrice": "$500",
			  "lifeExpectancy": 5,
			  "purchasingAccount": "admin",
			  "purchasingContact": "true",
			  "extensionAttributes": [
				{
				  "definitionId": "23",
				  "name": "Some Attribute",
				  "description": "Some Attribute defines how much Foo impacts Bar.",
				  "enabled": true,
				  "multiValue": true,
				  "values": [
					"foo",
					"bar"
				  ],
				  "dataType": "STRING",
				  "options": [
					"foo",
					"bar"
				  ],
				  "inputType": "TEXT"
				}
			  ]
			},
			"applications": [
			  {
				"name": "Microsoft Word",
				"path": "/usr/local/app",
				"version": "1.0.0",
				"macAppStore": true,
				"sizeMegabytes": 25,
				"bundleId": "1",
				"updateAvailable": false,
				"externalVersionId": "1"
			  }
			],
			"storage": {
			  "bootDriveAvailableSpaceMegabytes": 3072,
			  "disks": [
				{
				  "id": "170",
				  "device": "disk0",
				  "model": "APPLE HDD TOSHIBA MK5065GSXF",
				  "revision": "5",
				  "serialNumber": "a8598f013366",
				  "sizeMegabytes": 262144,
				  "smartStatus": "OK",
				  "type": "false",
				  "partitions": [
					{
					  "name": "Foo",
					  "sizeMegabytes": 262144,
					  "availableMegabytes": 131072,
					  "partitionType": "BOOT",
					  "percentUsed": 25,
					  "fileVault2State": "VALID",
					  "fileVault2ProgressPercent": 45,
					  "lvmManaged": true
					}
				  ]
				}
			  ]
			},
			"userAndLocation": {
			  "username": "Madison Anderson",
			  "realname": "13-inch MacBook",
			  "email": "email@com.pl",
			  "position": "IT Team Lead",
			  "phone": "123-456-789",
			  "departmentId": "1",
			  "buildingId": "1",
			  "room": "5",
			  "extensionAttributes": [
				{
				  "definitionId": "23",
				  "name": "Some Attribute",
				  "description": "Some Attribute defines how much Foo impacts Bar.",
				  "enabled": true,
				  "multiValue": true,
				  "values": [
					"foo",
					"bar"
				  ],
				  "dataType": "STRING",
				  "options": [
					"foo",
					"bar"
				  ],
				  "inputType": "TEXT"
				}
			  ]
			},
			"configurationProfiles": [
			  {
				"id": "1",
				"username": "username",
				"lastInstalled": "2018-10-31T18:04:13Z",
				"removable": true,
				"displayName": "Displayed profile",
				"profileIdentifier": "0ae590fe-9b30-11ea-bb37-0242ac130002"
			  }
			],
			"printers": [
			  {
				"name": "My Printer",
				"type": "XYZ 1122",
				"uri": "ipp://10.0.0.5",
				"location": "7th floor"
			  }
			],
			"services": [
			  {
				"name": "SomeService"
			  }
			],
			"hardware": {
			  "make": "Apple",
			  "model": "13-inch MacBook Pro (Mid 2012)",
			  "modelIdentifier": "MacBookPro9,2",
			  "serialNumber": "C02ZC2QYLVDL",
			  "processorSpeedMhz": 2100,
			  "processorCount": 2,
			  "coreCount": 2,
			  "processorType": "Intel Core i5",
			  "processorArchitecture": "i386",
			  "busSpeedMhz": 2133,
			  "cacheSizeKilobytes": 3072,
			  "networkAdapterType": "Foo",
			  "macAddress": "6A:2C:4B:B7:65:B5",
			  "altNetworkAdapterType": "Bar",
			  "altMacAddress": "82:45:58:44:dc:01",
			  "totalRamMegabytes": 4096,
			  "openRamSlots": 0,
			  "batteryCapacityPercent": 85,
			  "smcVersion": "2.2f38",
			  "nicSpeed": "N/A",
			  "opticalDrive": "MATSHITA DVD-R UJ-8A8",
			  "bootRom": "MBP91.00D3.B08",
			  "bleCapable": false,
			  "supportsIosAppInstalls": false,
			  "appleSilicon": false,
			  "extensionAttributes": [
				{
				  "definitionId": "23",
				  "name": "Some Attribute",
				  "description": "Some Attribute defines how much Foo impacts Bar.",
				  "enabled": true,
				  "multiValue": true,
				  "values": [
					"foo",
					"bar"
				  ],
				  "dataType": "STRING",
				  "options": [
					"foo",
					"bar"
				  ],
				  "inputType": "TEXT"
				}
			  ]
			},
			"localUserAccounts": [
			  {
				"uid": "501",
				"userGuid": "844F1177-0CF5-40C6-901F-38EDD9969C1C",
				"username": "jamf",
				"fullName": "John Jamf",
				"admin": true,
				"homeDirectory": "/Users/jamf",
				"homeDirectorySizeMb": 131072,
				"fileVault2Enabled": true,
				"userAccountType": "LOCAL",
				"passwordMinLength": 4,
				"passwordMaxAge": 5,
				"passwordMinComplexCharacters": 5,
				"passwordHistoryDepth": 5,
				"passwordRequireAlphanumeric": true,
				"computerAzureActiveDirectoryId": "1",
				"userAzureActiveDirectoryId": "1",
				"azureActiveDirectoryId": "ACTIVATED"
			  }
			],
			"certificates": [
			  {
				"commonName": "jamf.com",
				"identity": true,
				"expirationDate": "2030-10-31T18:04:13Z",
				"username": "test",
				"lifecycleStatus": "ACTIVE",
				"certificateStatus": "ISSUED",
				"subjectName": "CN=jamf.com",
				"serialNumber": "40f3d9fb",
				"sha1Fingerprint": "ed361458724d06082b2314acdb82e1f586f085f5",
				"issuedDate": "2022-05-23T14:54:10Z"
			  }
			],
			"attachments": [
			  {
				"id": "1",
				"name": "Attachment.pdf",
				"fileType": "application/pdf",
				"sizeBytes": 1024
			  }
			],
			"plugins": [
			  {
				"name": "plugin name",
				"version": "1.02",
				"path": "/Applications/"
			  }
			],
			"packageReceipts": {
			  "installedByJamfPro": [
				"com.jamf.protect.JamfProtect"
			  ],
			  "installedByInstallerSwu": [
				"com.apple.pkg.Core"
			  ],
			  "cached": [
				"com.jamf.protect.JamfProtect"
			  ]
			},
			"fonts": [
			  {
				"name": "font name",
				"version": "1.02",
				"path": "/Applications/"
			  }
			],
			"security": {
			  "sipStatus": "ENABLED",
			  "gatekeeperStatus": "APP_STORE_AND_IDENTIFIED_DEVELOPERS",
			  "xprotectVersion": "1.2.3",
			  "autoLoginDisabled": false,
			  "remoteDesktopEnabled": true,
			  "activationLockEnabled": true,
			  "recoveryLockEnabled": true,
			  "firewallEnabled": true,
			  "secureBootLevel": "FULL_SECURITY",
			  "externalBootLevel": "ALLOW_BOOTING_FROM_EXTERNAL_MEDIA",
			  "bootstrapTokenAllowed": true,
			  "bootstrapTokenEscrowedStatus": "ESCROWED"
			},
			"operatingSystem": {
			  "name": "Mac OS X",
			  "version": "10.9.5",
			  "build": "13A603",
			  "supplementalBuildVersion": "13A953",
			  "rapidSecurityResponse": "(a)",
			  "activeDirectoryStatus": "Not Bound",
			  "fileVault2Status": "ALL_ENCRYPTED",
			  "softwareUpdateDeviceId": "J132AP",
			  "extensionAttributes": [
				{
				  "definitionId": "23",
				  "name": "Some Attribute",
				  "description": "Some Attribute defines how much Foo impacts Bar.",
				  "enabled": true,
				  "multiValue": true,
				  "values": [
					"foo",
					"bar"
				  ],
				  "dataType": "STRING",
				  "options": [
					"foo",
					"bar"
				  ],
				  "inputType": "TEXT"
				}
			  ]
			},
			"licensedSoftware": [
			  {
				"id": "1",
				"name": "Microsoft Word"
			  }
			],
			"ibeacons": [
			  {
				"name": "room A"
			  }
			],
			"softwareUpdates": [
			  {
				"name": "BEdit",
				"version": "1.15.2",
				"packageName": "com.apple.pkg.AdditionalEssentials"
			  }
			],
			"extensionAttributes": [
			  {
				"definitionId": "23",
				"name": "Some Attribute",
				"description": "Some Attribute defines how much Foo impacts Bar.",
				"enabled": true,
				"multiValue": true,
				"values": [
				  "foo",
				  "bar"
				],
				"dataType": "STRING",
				"options": [
				  "foo",
				  "bar"
				],
				"inputType": "TEXT"
			  }
			],
			"contentCaching": {
			  "computerContentCachingInformationId": "1",
			  "parents": [
				{
				  "contentCachingParentId": "1",
				  "address": "SomeAddress",
				  "alerts": {
					"contentCachingParentAlertId": "1",
					"addresses": [],
					"className": "SomeClass",
					"postDate": "2018-10-31T18:04:13Z"
				  },
				  "details": {
					"contentCachingParentDetailsId": "1",
					"acPower": true,
					"cacheSizeBytes": 0,
					"capabilities": {
					  "contentCachingParentCapabilitiesId": "1",
					  "imports": true,
					  "namespaces": true,
					  "personalContent": true,
					  "queryParameters": true,
					  "sharedContent": true,
					  "prioritization": true
					},
					"portable": true,
					"localNetwork": [
					  {
						"contentCachingParentLocalNetworkId": "1",
						"speed": 5000,
						"wired": true
					  }
					]
				  },
				  "guid": "CD1E1291-4AF9-4468-B5D5-0F780C13DB2F",
				  "healthy": true,
				  "port": 0,
				  "version": "1"
				}
			  ],
			  "alerts": [
				{
				  "cacheBytesLimit": 0,
				  "className": "SomeClass",
				  "pathPreventingAccess": "/some/path",
				  "postDate": "2018-10-31T18:04:13Z",
				  "reservedVolumeBytes": 0,
				  "resource": "SomeResource"
				}
			  ],
			  "activated": false,
			  "active": false,
			  "actualCacheBytesUsed": 0,
			  "cacheDetails": [
				{
				  "computerContentCachingCacheDetailsId": "1",
				  "categoryName": "SomeCategory",
				  "diskSpaceBytesUsed": 0
				}
			  ],
			  "cacheBytesFree": 23353884672,
			  "cacheBytesLimit": 0,
			  "cacheStatus": "OK",
			  "cacheBytesUsed": 0,
			  "dataMigrationCompleted": false,
			  "dataMigrationProgressPercentage": 0,
			  "dataMigrationError": {
				"code": 0,
				"domain": "SomeDomain",
				"userInfo": [
				  {
					"key": "foo",
					"value": "bar"
				  }
				]
			  },
			  "maxCachePressureLast1HourPercentage": 0,
			  "personalCacheBytesFree": 23353884672,
			  "personalCacheBytesLimit": 0,
			  "personalCacheBytesUsed": 0,
			  "port": 0,
			  "publicAddress": "SomeAddress",
			  "registrationError": "NOT_ACTIVATED",
			  "registrationResponseCode": 403,
			  "registrationStarted": "2018-10-31T18:04:13Z",
			  "registrationStatus": "CONTENT_CACHING_FAILED",
			  "restrictedMedia": false,
			  "serverGuid": "CD1E1291-4AF9-4468-B5D5-0F780C13DB2F",
			  "startupStatus": "FAILED",
			  "tetheratorStatus": "CONTENT_CACHING_DISABLED",
			  "totalBytesAreSince": "2018-10-31T18:04:13Z",
			  "totalBytesDropped": 0,
			  "totalBytesImported": 0,
			  "totalBytesReturnedToChildren": 0,
			  "totalBytesReturnedToClients": 0,
			  "totalBytesReturnedToPeers": 0,
			  "totalBytesStoredFromOrigin": 0,
			  "totalBytesStoredFromParents": 0,
			  "totalBytesStoredFromPeers": 0
			},
			"groupMemberships": [
			  {
				"groupId": "1",
				"groupName": "groupOne",
				"smartGroup": true
			  }
			]
		  }
		`,
	}
	MobileDevices = map[string]string{
		"/api/v2/mobile-devices": `{
			"totalCount": 3,
				"results": [
				  {
					"id": "1",
					"name": "iPad",
					"serialNumber": "DMQVGC0DHLA0",
					"wifiMacAddress": "C4:84:66:92:78:00",
					"udid": "0dad565fb40b010a9e490440188063a378721069",
					"phoneNumber": "651-555-5555 Ext111",
					"model": "iPad 5th Generation (Wi-Fi)",
					"modelIdentifier": "iPad6,11",
					"username": "admin",
					"type": "ios",
					"managementId": "73226fb6-61df-4c10-9552-eb9bc353d507",
					"softwareUpdateDeviceId": "J132AP"
				  },
				  {
					"id": "2",
					"name": "iPhone",
					"serialNumber": "FMDK2D0DFK34",
					"wifiMacAddress": "E8:80:2E:6F:4A:09",
					"udid": "5b6a7b9db01c821b3e672a6ebd8a987654321abc",
					"phoneNumber": "651-555-5555 Ext112",
					"model": "iPhone 11 Pro",
					"modelIdentifier": "iPhone12,3",
					"username": "user1",
					"type": "ios",
					"managementId": "3fa26br7-72ef-1c20-8652-fb8cc353e507",
					"softwareUpdateDeviceId": "N104AP"
				  },
				  {
					"id": "3",
					"name": "iPhone SE",
					"serialNumber": "GMRW3D9DTYX9",
					"wifiMacAddress": "D4:61:9D:F2:77:14",
					"udid": "7c8b9d0fb21e3ab9e5d874101234567890abcdef",
					"phoneNumber": "651-555-5555 Ext113",
					"model": "iPhone SE (2nd generation)",
					"modelIdentifier": "iPhone12,8",
					"username": "user2",
					"type": "ios",
					"managementId": "5h79jk8l-45hv-9t60-5643-km67nv83qp90",
					"softwareUpdateDeviceId": "N122AP"
				  }
				]
			  }`,
	}
)
