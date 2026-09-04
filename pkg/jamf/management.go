/*
# Jamf - Management

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/devices.go
package jamf

import (
	"fmt"
)

var (
	ManagementFramework = fmt.Sprintf("%s/jamf-management-framework", V1) // /api/v1/jamf-management-framework
	V1_MDM              = fmt.Sprintf("%s/mdm", V1)                       // /api/v1/mdm
	RenewProfile        = fmt.Sprintf("%s/renew-profile", V1_MDM)         // /api/v1/mdm/renew-profile
	V2_MDM              = fmt.Sprintf("%s/mdm", V2)                       // /api/v2/mdm
	V2_MDM_Commands     = fmt.Sprintf("%s/commands", V2_MDM)              // /api/v2/mdm/commands
)

/*
 * # Renew MDM Profile
 * /api/v1/mdm/renew-profile
 * - https://developer.jamf.com/jamf-pro/reference/post_v1-mdm-renew-profile
 */
func (c *Client) RenewMDMProfile(udids []string) (*ManagementResponse, error) {
	url := c.BuildURL(RenewProfile)

	payload := map[string][]string{
		"udids": udids,
	}

	mr, err := do[*ManagementResponse](c, "POST", url, nil, payload)
	if err != nil {
		return nil, err
	}

	return mr, nil
}

/*
 * # Repair Jamf Management Framework
 * /api/v1/jamf-management-framework/redeploy/{id}
 * - https://developer.jamf.com/jamf-pro/reference/post_v1-jamf-management-framework-redeploy-id
 */
func (c *Client) RepairManagementFramework(id string) (string, error) {
	url := c.BuildURL(fmt.Sprintf("%s/redeploy/%s", ManagementFramework, id))

	mf, err := do[any](c, "POST", url, nil, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", mf), nil
}

/*
 * # Send MDM Command
 * /api/v2/mdm/commands
 * - https://developer.jamf.com/jamf-pro/reference/post_v2-mdm-commands
 * Generic method to send any MDM command to one or more devices.
 * Returns an array of command responses (one per device targeted).
 */
func (c *Client) SendMDMCommand(request *MDMCommandRequest) ([]MDMCommandResponse, error) {
	url := c.BuildURL(V2_MDM_Commands)

	resp, err := do[[]MDMCommandResponse](c, "POST", url, nil, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

/*
 * # Lock Device
 * /api/v2/mdm/commands
 * - https://developer.jamf.com/jamf-pro/reference/post_v2-mdm-commands
 * Sends a DEVICE_LOCK MDM command to lock a device.
 * For computers, a 6-digit passcode is required to unlock.
 * For mobile devices, an optional message and phone number can be displayed on the lock screen.
 */
func (c *Client) LockDevice(managementID string, clientType string, passcode string, message string, phoneNumber string) ([]MDMCommandResponse, error) {
	request := &MDMCommandRequest{
		ClientData: []MDMClientData{
			{
				ManagementID: managementID,
				ClientType:   clientType,
			},
		},
		CommandData: MDMCommandData{
			CommandType: CommandType.DeviceLock,
			Pin:         passcode,
			Message:     message,
			PhoneNumber: phoneNumber,
		},
	}

	return c.SendMDMCommand(request)
}

/*
 * # Lock Computer
 * /api/v2/mdm/commands
 * - https://developer.jamf.com/jamf-pro/reference/post_v2-mdm-commands
 * Convenience method to lock a computer. Requires a 6-digit passcode.
 */
func (c *Client) LockComputer(managementID string, passcode string, message string) ([]MDMCommandResponse, error) {
	return c.LockDevice(managementID, ClientType.Computer, passcode, message, "")
}

/*
 * # Lock Mobile Device
 * /api/v2/mdm/commands
 * - https://developer.jamf.com/jamf-pro/reference/post_v2-mdm-commands
 * Convenience method to lock a mobile device. Optionally displays a message and phone number.
 */
func (c *Client) LockMobileDevice(managementID string, message string, phoneNumber string) ([]MDMCommandResponse, error) {
	return c.LockDevice(managementID, ClientType.MobileDevice, "", message, phoneNumber)
}

/*
 * # Get Lock Status by Command UUID
 * GET /api/v2/mdm/commands?filter=uuid=={commandUUID}
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-mdm-commands
 * Retrieves the status of a DEVICE_LOCK command by its UUID.
 * The UUID is returned from LockComputer/LockMobileDevice/LockDevice.
 */
func (c *Client) GetLockStatusByUUID(commandUUID string) (*MDMCommands, error) {
	url := c.BuildURL(V2_MDM_Commands)

	query := &MDMCommandQuery{
		Filter: fmt.Sprintf("uuid==\"%s\"", commandUUID),
	}

	resp, err := do[*MDMCommands](c, "GET", url, query, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

/*
 * # Get Lock Status by Management ID
 * GET /api/v2/mdm/commands?filter=managementId=={managementID};command==DEVICE_LOCK
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-mdm-commands
 * Retrieves ALL DEVICE_LOCK commands for a specific device by management ID.
 */
func (c *Client) GetLockStatus(managementID string) (*MDMCommands, error) {
	url := c.BuildURL(V2_MDM_Commands)

	query := &MDMCommandQuery{
		Filter: fmt.Sprintf("managementId==\"%s\";command==\"%s\"", managementID, CommandType.DeviceLock),
		Sort:   "dateSent:desc",
	}

	resp, err := do[*MDMCommands](c, "GET", url, query, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

/*
 * # Get MDM Command Status
 * GET /api/v2/mdm/commands/{uuid}
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-mdm-commands-id
 * Retrieves the status of a specific MDM command by its UUID.
 */
func (c *Client) GetMDMCommandStatus(commandUUID string) (*MDMCommandResponse, error) {
	url := c.BuildURL(fmt.Sprintf("%s/%s", V2_MDM_Commands, commandUUID))

	resp, err := do[*MDMCommandResponse](c, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

/*
 * # List MDM Commands
 * GET /api/v2/mdm/commands
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-mdm-commands
 * Lists MDM commands with optional filtering and pagination.
 * Use MDMCommandQuery to filter by deviceId, status, commandType, etc.
 */
func (c *Client) ListMDMCommands(query *MDMCommandQuery) (*MDMCommands, error) {
	url := c.BuildURL(V2_MDM_Commands)

	resp, err := do[*MDMCommands](c, "GET", url, query, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

/*
 * # Get Device MDM Commands
 * GET /api/v2/mdm/commands?filter=managementId=={managementId}
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-mdm-commands
 * Convenience method to list all MDM commands for a specific device.
 */
func (c *Client) GetDeviceMDMCommands(managementID string) (*MDMCommands, error) {
	query := &MDMCommandQuery{
		Filter: fmt.Sprintf("managementId==\"%s\"", managementID),
		Sort:   "dateSent:desc",
	}

	return c.ListMDMCommands(query)
}

/*
 * # Get Pending MDM Commands
 * GET /api/v2/mdm/commands?filter=status==PENDING
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-mdm-commands
 * Convenience method to list all pending MDM commands, optionally for a specific device.
 */
func (c *Client) GetPendingMDMCommands(managementID string) (*MDMCommands, error) {
	filter := fmt.Sprintf("status==\"%s\"", CommandStatus.Pending)
	if managementID != "" {
		filter = fmt.Sprintf("managementId==\"%s\";status==\"%s\"", managementID, CommandStatus.Pending)
	}

	query := &MDMCommandQuery{
		Filter: filter,
		Sort:   "dateSent:desc",
	}

	return c.ListMDMCommands(query)
}
