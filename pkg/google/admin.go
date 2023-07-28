/*
# Google Workspace - Admin

This package initializes all the methods for functions which interact with the Google Admin API:
https://developers.google.com/admin-sdk/reference-overview

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/drive.go
package google

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	ss "github.com/gemini-oss/rego/pkg/common/starstruct"
)

var (
	AdminDirectory           = fmt.Sprintf("%s/admin/directory/v1", AdminBaseURL)                            // https://developers.google.com/admin-sdk/reference-overview
	DirectoryASPS            = fmt.Sprintf("%s/users/%s/asps", AdminDirectory, "%s")                         // https://developers.google.com/admin-sdk/directory/reference/rest/v1/asps
	DirectoryChannels        = fmt.Sprintf("%s/channels", AdminDirectory)                                    // https://developers.google.com/admin-sdk/directory/reference/rest/v1/channels
	DirectoryChromeOSDevices = fmt.Sprintf("%s/customer/%s/devices/chromeos", AdminDirectory, "%s")          // https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices
	DirectoryCustomers       = fmt.Sprintf("%s/customers", AdminDirectory)                                   // https://developers.google.com/admin-sdk/directory/reference/rest/v1/customers
	DirectoryDomains         = fmt.Sprintf("%s/domains", AdminDirectory)                                     // https://developers.google.com/admin-sdk/directory/reference/rest/v1/domains
	DirectoryGroups          = fmt.Sprintf("%s/groups", AdminDirectory)                                      // https://developers.google.com/admin-sdk/directory/reference/rest/v1/groups
	DirectoryMembers         = fmt.Sprintf("%s/groups/%s/members", AdminDirectory, "%s")                     // https://developers.google.com/admin-sdk/directory/reference/rest/v1/members
	DirectoryMobileDevices   = fmt.Sprintf("%s/customer/%s/devices/mobile", AdminDirectory, "%s")            // https://developers.google.com/admin-sdk/directory/reference/rest/v1/mobiledevices
	DirectoryOrgUnits        = fmt.Sprintf("%s/customer/%s/orgunits", AdminDirectory, "%s")                  // https://developers.google.com/admin-sdk/directory/reference/rest/v1/orgunits
	DirectoryPrivileges      = fmt.Sprintf("%s/privileges", AdminDirectory)                                  // https://developers.google.com/admin-sdk/directory/reference/rest/v1/privileges
	DirectoryResources       = fmt.Sprintf("%s/customer/%s/resources", AdminDirectory, "%s")                 // https://developers.google.com/admin-sdk/directory/reference/rest/v1/resources
	DirectoryRoleAssignments = fmt.Sprintf("%s/customer/%s/roleassignments", AdminDirectory, "%s")           // https://developers.google.com/admin-sdk/directory/reference/rest/v1/roleassignments
	DirectoryRoles           = fmt.Sprintf("%s/customer/%s/roles", AdminDirectory, "%s")                     // https://developers.google.com/admin-sdk/directory/reference/rest/v1/roles
	DirectorySchemas         = fmt.Sprintf("%s/schemas", AdminDirectory)                                     // https://developers.google.com/admin-sdk/directory/reference/rest/v1/schemas
	DirectoryTokens          = fmt.Sprintf("%s/tokens", AdminDirectory)                                      // https://developers.google.com/admin-sdk/directory/reference/rest/v1/tokens
	DirectoryUsers           = fmt.Sprintf("%s/users", AdminDirectory)                                       // https://developers.google.com/admin-sdk/directory/reference/rest/v1/users
	AdminReports             = fmt.Sprintf("%s/admin/reports/v1", AdminBaseURL)                              // https://developers.google.com/admin-sdk/reports/reference/rest
	ReportsActivities        = fmt.Sprintf("%s/activity/users/%s/applications/%s", AdminReports, "%s", "%s") // https://developers.google.com/admin-sdk/reports/reference/rest/v1/activities
	ReportsChannels          = fmt.Sprintf("%s/channels", AdminReports)                                      // https://developers.google.com/admin-sdk/reports/reference/rest/v1/channels
	ReportsCustomerUsage     = fmt.Sprintf("%s/customerUsageReports", AdminReports)                          // https://developers.google.com/admin-sdk/reports/reference/rest/v1/customerUsageReports
	ReportsEntityUsage       = fmt.Sprintf("%s/entityUsageReports", AdminReports)                            // https://developers.google.com/admin-sdk/reports/reference/rest/v1/entityUsageReports
	ReportsUserUsage         = fmt.Sprintf("%s/userUsageReport", AdminReports)                               // https://developers.google.com/admin-sdk/reports/reference/rest/v1/userUsageReport
)

/*
 * Query Parameters for Admin Reports
 * https://developers.google.com/admin-sdk/reports/reference/rest/v1/activities/list#query-parameters
 */
type ReportsQuery struct {
	ActorIpAddress                 string `url:"actorIpAddress,omitempty"`                 // The Internet Protocol (IP) Address of host where the event was performed.
	CustomerId                     string `url:"customerId,omitempty"`                     // The unique ID of the customer to retrieve data for.
	EndTime                        string `url:"endTime,omitempty"`                        // Sets the end of the range of time shown in the report.
	EventName                      string `url:"eventName,omitempty"`                      // The name of the event being queried by the API.
	Filters                        string `url:"filters,omitempty"`                        // The filters query string is a comma-separated list composed of event parameters manipulated by relational operators.
	MaxResults                     int    `url:"maxResults,omitempty"`                     // Determines how many activity records are shown on each response page.
	OrgUnitId                      string `url:"orgUnitId,omitempty"`                      // ID of the organizational unit to report on.
	PageToken                      string `url:"pageToken,omitempty"`                      // The token to specify next page.
	StartTime                      string `url:"startTime,omitempty"`                      // Sets the beginning of the range of time shown in the report.
	GroupIdFilter                  string `url:"groupIdFilter,omitempty"`                  // Comma separated group ids (obfuscated) on which user activities are filtered.
	RoleId                         string `url:"roleId,omitempty"`                         // ID of the role to report on.
	UserKey                        string `url:"userKey,omitempty"`                        // Represents the profile id or the user email for which the data should be filtered.
	IncludeIndirectRoleAssignments bool   `url:"includeIndirectRoleAssignments,omitempty"` // Whether to include indirect role assignments.
}

/*
 * List all Roles in the domain with pagination support
 * /admin/directory/v1/customer/{customer}/roles
 * https://developers.google.com/admin-sdk/directory/v1/reference/roles/list
 */
func (c *Client) ListAllRoles(customerId string) (*Roles, error) {
	c.Logger.Println("Getting all roles...")
	roles := &Roles{}

	var url string
	switch customerId {
	case "":
		url = fmt.Sprintf(DirectoryRoles, "my_customer")
	default:
		url = fmt.Sprintf(DirectoryRoles, customerId)
	}

	res, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &roles)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

/*
 * Get a Role by ID
 * /admin/directory/v1/customer/{customer}/roles/{roleId}
 * https://developers.google.com/admin-sdk/directory/v1/reference/roles/get
 */
func (c *Client) GetRole(customerId string, roleId string) (*Role, error) {
	c.Logger.Println("Getting role...")
	role := &Role{}

	if customerId == "" {
		customerId = "my_customer"
	}

	url := fmt.Sprintf("%s/%s", DirectoryRoles, roleId)

	res, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

/*
 * Get all Role's Assignments within a target customer/domain
 * /admin/directory/v1/customer/{customer}/roleassignments
 * https://developers.google.com/admin-sdk/directory/v1/reference/roleAssignments/list
 */
func (c *Client) ListAllRoleAssignments(customerId string) (*RoleAssignment, error) {
	c.Logger.Println("Getting all role assignments...")
	roleAssignments := &RoleAssignment{}

	var url string
	switch customerId {
	case "":
		url = fmt.Sprintf(DirectoryRoleAssignments, "my_customer")
	default:
		url = fmt.Sprintf(DirectoryRoleAssignments, customerId)
	}

	res, body, err := c.HTTPClient.DoRequest("GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, roleAssignments)
	if err != nil {
		return nil, err
	}

	return roleAssignments, nil
}

/*
 * Get assignments for a targeted Role
 * /admin/directory/v1/customer/{customer}/roleassignments
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/roleAssignments/list#query-parameters
 */
func (c *Client) GetAssignmentsForRole(customerId string, roleId string) (*RoleAssignment, error) {
	c.Logger.Println("Getting role's assignment...")
	roleAssignment := &RoleAssignment{}

	q := ReportsQuery{
		RoleId: roleId,
	}

	var url string
	switch customerId {
	case "":
		url = fmt.Sprintf(DirectoryRoleAssignments, "my_customer")
	default:
		url = fmt.Sprintf(DirectoryRoleAssignments, customerId)
	}

	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &roleAssignment)
	if err != nil {
		return nil, err
	}

	return roleAssignment, nil
}

/*
 * Create user list from a role's assignments
 * /admin/directory/v1/customer/{customer}/roleassignments
 * https://developers.google.com/admin-sdk/directory/v1/reference/roleAssignments/list
 */
func (c *Client) GetUsersFromRoleAssignments(sem chan struct{}, roleAssignments []RoleAssignment) ([]*User, error) {
	// Make a channel for the users and their errors
	userChannel := make(chan *User, len(roleAssignments))
	userErrChannel := make(chan error)

	var userWg sync.WaitGroup
	for _, assignment := range roleAssignments {
		assign := assignment
		userWg.Add(1)
		go func(assign RoleAssignment) {
			defer userWg.Done()
			sem <- struct{}{} // Acquire a token
			user, err := c.GetUser(assign.AssignedTo)
			if err != nil {
				userErrChannel <- err
				return
			}
			userChannel <- user
			<-sem // Release the token
		}(assign)
	}

	userWg.Wait()
	close(userChannel)
	close(userErrChannel)

	for err := range userErrChannel {
		return nil, err
	}

	users := []*User{}
	for user := range userChannel {
		users = append(users, user)
	}

	return users, nil
}

/*
 * Get all Users assigned to a Role within a target customer/domain
 * /admin/directory/v1/customer/{customer}/roleassignments/{roleId}/users
 * https://developers.google.com/admin-sdk/directory/v1/reference/roleAssignments/list
 */
func (c *Client) GenerateRoleReport(customerId string, roleId string) ([]*RoleReport, error) {
	// If no customerId is provided, use 'my_customer' to represent the authenticated account
	if customerId == "" {
		customerId = "my_customer"
	}

	// Use a buffered channel as a semaphore to limit concurrent requests.
	// 10 is the maximum number of concurrent requests.
	sem := make(chan struct{}, 10)

	// WaitGroup to ensure all go routines complete their tasks.
	var wg sync.WaitGroup

	// Fetch all roles for the provided customer ID
	roles, err := c.ListAllRoles(customerId)
	if err != nil {
		return nil, err
	}

	// Buffered channel to hold the RoleReport result from each goroutine
	reportsChannel := make(chan *RoleReport, len(roles.Items))

	// Buffered channel to hold any errors that occur while generating the role reports
	reportsErrChannel := make(chan error)

	// Inner function to generate a RoleReport for a single role. This will be run in a separate goroutine.
	generateReport := func(role Role) {
		// Ensure the waitgroup counter is decremented when this function finishes
		defer wg.Done()

		roleAssignments, err := c.GetAssignmentsForRole(customerId, role.RoleID)
		if err != nil {
			reportsErrChannel <- err
			return
		}

		userList, err := c.GetUsersFromRoleAssignments(sem, roleAssignments.Items)
		if err != nil {
			reportsErrChannel <- err
			return
		}

		// Create a report for the role and send it to the reportsChannel
		roleReport := &RoleReport{
			Role:  &role,
			Users: userList,
		}
		reportsChannel <- roleReport
	}

	// If a specific role ID is provided, generate a report only for that role
	if roleId != "" {
		wg.Add(1)
		role, err := c.GetRole(customerId, roleId)
		if err != nil {
			return nil, err
		}
		go generateReport(*role)
	} else {
		for _, role := range roles.Items {
			// Ignore certain system roles
			switch role.RoleName {
			case "_GCDS_DIRECTORY_MANAGEMENT_ROLE", "_LDAP_USER_MANAGEMENT_SUPPORT_ROLE", "_LDAP_USER_MANAGEMENT_READONLY_ROLE", "_LDAP_PASSWORD_REBIND_ROLE", "_LDAP_GROUP_MANAGEMENT_READONLY_ROLE":
				continue
			default:
				// For each role, increment the waitgroup counter and launch a goroutine
				wg.Add(1)
				go generateReport(role)
			}
		}
	}

	// Wait for all the goroutines to finish, then close the result and error channels
	wg.Wait()
	close(reportsChannel)
	close(reportsErrChannel)

	if err, ok := <-reportsErrChannel; ok {
		return nil, err
	}

	roleReports := []*RoleReport{}
	for report := range reportsChannel {
		roleReports = append(roleReports, report)
	}

	return roleReports, nil
}

/*
 * Save a RoleReport to a new spreadsheet
 * @param reports: A slice of RoleReport structs
 * @return *Spreadsheet: A pointer to the newly created spreadsheet
 */
func (c *Client) SaveRoleReport(reports []*RoleReport) (*Spreadsheet, error) {

	// Create a ValueRange to hold the report data
	vr := &ValueRange{}
	headers := []string{"Name", "Email", "Role", "Last Login", "Org Unit Path", "Suspended", "Archived"}
	vr.Values = append(vr.Values, headers)
	for _, role := range reports {
		for _, user := range role.Users {
			vr.Values = append(vr.Values, []string{user.Name.FullName, user.PrimaryEmail, role.Role.RoleName, user.LastLoginTime, user.OrgUnitPath, strconv.FormatBool(user.Suspended), strconv.FormatBool(user.Archived)})
		}
	}

	c.Logger.Println("Creating new spreadsheet for role report")
	spreadsheet, err := c.CreateSpreadsheet()
	if err != nil {
		return nil, err
	}

	c.Logger.Println("Saving role report to spreadsheet")
	err = c.UpdateSpreadsheet(spreadsheet.SpreadsheetID, vr)
	if err != nil {
		return nil, err
	}

	c.Logger.Println("Formatting spreadsheet")
	rows := len(vr.Values)
	columns := len(headers)
	c.FormatHeaderAndAutoSize(spreadsheet.SpreadsheetID, rows, columns)

	return spreadsheet, nil
}

/*
 * Find the file ownership using the Reports API
 */
func (c *Client) GetFileOwnership(fileID string) (string, error) {
	c.Logger.Println("Getting file ownership...")
	c.Logger.Debug("fileID:", fileID)
	fileReport := &Report{}

	q := ReportsQuery{
		Filters:    fmt.Sprintf("doc_id==%s", fileID),
		MaxResults: 1,
	}
	c.Logger.Debug("query:", q)

	url := fmt.Sprintf(ReportsActivities, "all", "drive")
	c.Logger.Debug("url:", url)

	c.Logger.Println("Sending request...")
	res, body, err := c.HTTPClient.DoRequest("GET", url, q, nil)
	if err != nil {
		return "", err
	}
	c.Logger.Println("Response Status:", res.Status)
	c.Logger.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, fileReport)
	if err != nil {
		return "", err
	}
	c.Logger.Println(ss.PrettyJSON(fileReport))

	if len(fileReport.Items) == 0 {
		return "", fmt.Errorf("no events found for file %s", fileID)
	}
	for _, event := range fileReport.Items[0].Events {
		for i, param := range event.Parameters {
			if param.Name == "owner" {
				c.Logger.Println("Found owner!")
				c.Logger.Debug("owner:", event.Parameters[i].Value)
				return event.Parameters[i].Value, nil
			}
		}
	}

	return "", fmt.Errorf("no owner found for file %s", fileID)
}
