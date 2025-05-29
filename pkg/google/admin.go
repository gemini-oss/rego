/*
# Google Workspace - Admin

This package initializes all the methods for functions which interact with the Google Admin API:
https://developers.google.com/admin-sdk/reference-overview

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/drive.go
package google

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	ss "github.com/gemini-oss/rego/pkg/common/starstruct"
)

var (
	AdminDirectory           = fmt.Sprintf("%s/admin/directory/v1", AdminBaseURL)                             // https://developers.google.com/admin-sdk/reference-overview
	DirectoryASPS            = fmt.Sprintf("%s/users/%s/asps", AdminDirectory, "%s")                          // https://developers.google.com/admin-sdk/directory/reference/rest/v1/asps
	DirectoryChannels        = fmt.Sprintf("%s/channels", AdminDirectory)                                     // https://developers.google.com/admin-sdk/directory/reference/rest/v1/channels
	DirectoryChromeOSDevices = fmt.Sprintf("%s/customer/%s/devices/chromeos", AdminDirectory, "%s")           // https://developers.google.com/admin-sdk/directory/reference/rest/v1/chromeosdevices
	DirectoryCustomers       = fmt.Sprintf("%s/customers/%s", AdminDirectory, "%s")                           // https://developers.google.com/admin-sdk/directory/reference/rest/v1/customers
	DirectoryDomains         = fmt.Sprintf("%s/domains", AdminDirectory)                                      // https://developers.google.com/admin-sdk/directory/reference/rest/v1/domains
	DirectoryGroups          = fmt.Sprintf("%s/groups", AdminDirectory)                                       // https://developers.google.com/admin-sdk/directory/reference/rest/v1/groups
	DirectoryMembers         = fmt.Sprintf("%s/groups/%s/members", AdminDirectory, "%s")                      // https://developers.google.com/admin-sdk/directory/reference/rest/v1/members
	DirectoryMobileDevices   = fmt.Sprintf("%s/customer/%s/devices/mobile", AdminDirectory, "%s")             // https://developers.google.com/admin-sdk/directory/reference/rest/v1/mobiledevices
	DirectoryOrgUnits        = fmt.Sprintf("%s/customer/%s/orgunits", AdminDirectory, "%s")                   // https://developers.google.com/admin-sdk/directory/reference/rest/v1/orgunits
	DirectoryPrivileges      = fmt.Sprintf("%s/privileges", AdminDirectory)                                   // https://developers.google.com/admin-sdk/directory/reference/rest/v1/privileges
	DirectoryResources       = fmt.Sprintf("%s/customer/%s/resources", AdminDirectory, "%s")                  // https://developers.google.com/admin-sdk/directory/reference/rest/v1/resources
	DirectoryRoleAssignments = fmt.Sprintf("%s/customer/%s/roleassignments", AdminDirectory, "%s")            // https://developers.google.com/admin-sdk/directory/reference/rest/v1/roleassignments
	DirectoryRoles           = fmt.Sprintf("%s/customer/%s/roles", AdminDirectory, "%s")                      // https://developers.google.com/admin-sdk/directory/reference/rest/v1/roles
	DirectorySchemas         = fmt.Sprintf("%s/schemas", AdminDirectory)                                      // https://developers.google.com/admin-sdk/directory/reference/rest/v1/schemas
	DirectoryTokens          = fmt.Sprintf("%s/tokens", AdminDirectory)                                       // https://developers.google.com/admin-sdk/directory/reference/rest/v1/tokens
	DirectoryUsers           = fmt.Sprintf("%s/users", AdminDirectory)                                        // https://developers.google.com/admin-sdk/directory/reference/rest/v1/users
	AdminDirectoryBeta       = fmt.Sprintf("%s/admin/directory/v1.1beta1", AdminBaseURL)                      // https://support.google.com/chrome/a/answer/9681204?ref_topic=9301744
	DirectoryChromeBrowsers  = fmt.Sprintf("%s/customer/%s/devices/chromebrowsers", AdminDirectoryBeta, "%s") // https://support.google.com/chrome/a/answer/9681204?ref_topic=9301744
	AdminReports             = fmt.Sprintf("%s/admin/reports/v1", AdminBaseURL)                               // https://developers.google.com/admin-sdk/reports/reference/rest
	ReportsActivities        = fmt.Sprintf("%s/activity/users/%s/applications/%s", AdminReports, "%s", "%s")  // https://developers.google.com/admin-sdk/reports/reference/rest/v1/activities
	ReportsChannels          = fmt.Sprintf("%s/channels", AdminReports)                                       // https://developers.google.com/admin-sdk/reports/reference/rest/v1/channels
	ReportsCustomerUsage     = fmt.Sprintf("%s/customerUsageReports", AdminReports)                           // https://developers.google.com/admin-sdk/reports/reference/rest/v1/customerUsageReports
	ReportsEntityUsage       = fmt.Sprintf("%s/entityUsageReports", AdminReports)                             // https://developers.google.com/admin-sdk/reports/reference/rest/v1/entityUsageReports
	ReportsUserUsage         = fmt.Sprintf("%s/userUsageReport", AdminReports)                                // https://developers.google.com/admin-sdk/reports/reference/rest/v1/userUsageReport
)

// AdminClient for chaining methods
type AdminClient struct {
	*Client
}

// Entry point for admin-related operations
func (c *Client) Admin() *AdminClient {
	ac := &AdminClient{
		Client: c,
	}

	// https://developers.google.com/admin-sdk/directory/v1/limits
	ac.HTTP.RateLimiter.Available = 2400
	ac.HTTP.RateLimiter.Limit = 2400
	ac.HTTP.RateLimiter.Interval = 1 * time.Minute
	ac.HTTP.RateLimiter.Log.Verbosity = c.Log.Verbosity

	return ac
}

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

func (q *ReportsQuery) SetPageToken(token string) {
	q.PageToken = token
}

/*
 * List all Roles in the domain with pagination support
 * /admin/directory/v1/customer/{customer}/roles
 * https://developers.google.com/admin-sdk/directory/v1/reference/roles/list
 */
func (c *AdminClient) MyCustomer() (*Customer, error) {
	url := c.BuildURL(DirectoryCustomers, &Customer{ID: "my_customer"})

	var cache Customer
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	c.Log.Println("Getting ID of current client...")
	customer, err := do[Customer](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, customer, 1*time.Hour)
	return &customer, nil
}

/*
 * List all Roles in the domain with pagination support
 * /admin/directory/v1/customer/{customer}/roles
 * https://developers.google.com/admin-sdk/directory/v1/reference/roles/list
 */
func (c *AdminClient) ListAllRoles(customer *Customer) (*Roles, error) {
	url := c.BuildURL(DirectoryRoles, customer)

	var cache Roles
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	c.Log.Println("Getting all roles...")
	roles, err := do[Roles](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, roles, 5*time.Minute)
	return &roles, nil
}

/*
 * Get a Role by ID
 * /admin/directory/v1/customer/{customer}/roles/{roleId}
 * https://developers.google.com/admin-sdk/directory/v1/reference/roles/get
 */
func (c *AdminClient) GetRole(roleId string, customer *Customer) (*Role, error) {
	url := c.BuildURL(DirectoryRoles, customer, roleId)

	var cache Role
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	c.Log.Println("Getting role...")
	role, err := do[Role](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, role, 5*time.Minute)
	return &role, nil
}

/*
 * Get all Role's Assignments within a target customer/domain
 * /admin/directory/v1/customer/{customer}/roleassignments
 * https://developers.google.com/admin-sdk/directory/v1/reference/roleAssignments/list
 */
func (c *AdminClient) ListAllRoleAssignments(customer *Customer) (*RoleAssignment, error) {
	url := c.BuildURL(DirectoryRoleAssignments, customer)

	var cache RoleAssignment
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	c.Log.Println("Getting all role assignments...")
	roleAssignments, err := do[RoleAssignment](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, roleAssignments, 5*time.Minute)
	return &roleAssignments, nil
}

/*
 * Get assignments for a targeted Role
 * /admin/directory/v1/customer/{customer}/roleassignments
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/roleAssignments/list#query-parameters
 */
func (c *AdminClient) GetAssignmentsForRole(roleId string, customer *Customer) (*RoleAssignment, error) {
	url := c.BuildURL(DirectoryRoleAssignments, customer)
	cacheKey := fmt.Sprintf("%s_%s", url, roleId)

	var cache RoleAssignment
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	q := ReportsQuery{
		RoleId: roleId,
	}

	c.Log.Println("Getting role's assignment...")
	roleAssignment, err := do[RoleAssignment](c.Client, "GET", url, q, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, roleAssignment, 5*time.Minute)
	return &roleAssignment, nil
}

/*
 * Create user list from a role's assignments
 * /admin/directory/v1/customer/{customer}/roleassignments
 * https://developers.google.com/admin-sdk/directory/v1/reference/roleAssignments/list
 */
func (c *AdminClient) GetUsersFromRoleAssignments(sem chan struct{}, roleAssignments []RoleAssignment) ([]*User, error) {
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
			user, err := c.Users().GetUser(assign.AssignedTo)
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
func (c *AdminClient) GenerateRoleReport(roleId string, customer *Customer) ([]*RoleReport, error) {

	// Use a buffered channel as a semaphore to limit concurrent requests.
	// 10 is the maximum number of concurrent requests.
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))

	// WaitGroup to ensure all go routines complete their tasks.
	var wg sync.WaitGroup

	// Fetch all roles for the provided customer ID
	roles, err := c.ListAllRoles(customer)
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

		roleAssignments, err := c.GetAssignmentsForRole(role.RoleID, customer)
		if err != nil {
			c.Log.Println("Error getting role assignments:", err)
			reportsErrChannel <- err
			return
		}

		userList, err := c.GetUsersFromRoleAssignments(sem, roleAssignments.Items)
		if err != nil {
			c.Log.Println("Error getting role assignments:", err)
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
		role, err := c.GetRole(roleId, customer)
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
func (c *AdminClient) SaveRoleReport(reports []*RoleReport) (*Spreadsheet, error) {

	// Create a ValueRange to hold the report data
	vr := &ValueRange{}
	headers := []string{"Name", "Email", "Role", "Last Login", "Org Unit Path", "Suspended", "Archived"}
	vr.Values = append(vr.Values, headers)
	for _, role := range reports {
		for _, user := range role.Users {
			vr.Values = append(vr.Values, []string{user.Name.FullName, user.PrimaryEmail, role.Role.RoleName, user.LastLoginTime, user.OrgUnitPath, strconv.FormatBool(user.Suspended), strconv.FormatBool(user.Archived)})
		}
	}

	c.Log.Println("Creating new spreadsheet for role report")
	sheet := &Spreadsheet{
		Properties: &SpreadsheetProperties{
			Title: fmt.Sprintf("{Google} Entitlement Review %s", time.Now().Format("2006-01-02")),
		},
	}
	spreadsheet, err := c.Sheets().CreateSpreadsheet(sheet)
	if err != nil {
		return nil, err
	}

	c.Log.Println("Saving role report to spreadsheet")
	err = c.Sheets().UpdateSpreadsheet(spreadsheet.SpreadsheetID, vr)
	if err != nil {
		return nil, err
	}

	c.Log.Println("Formatting spreadsheet")
	rows := len(vr.Values)
	columns := len(headers)
	c.Sheets().FormatHeaderAndAutoSize(spreadsheet.SpreadsheetID, &spreadsheet.Sheets[0], rows, columns)

	return spreadsheet, nil
}

/*
 * Find the file ownership using the Reports API
 */
func (c *AdminClient) GetFileOwnership(fileID string) (string, error) {
	c.Log.Println("Getting file ownership...")
	c.Log.Debug("fileID:", fileID)
	fileReport := &Report{}

	q := ReportsQuery{
		Filters:    fmt.Sprintf("doc_id==%s", fileID),
		MaxResults: 1,
	}
	c.Log.Debug("query:", q)

	url := c.BuildURL(ReportsActivities, nil, "all", "drive")

	fileReport, err := do[*Report](c.Client, "GET", url, q, nil)
	if err != nil {
		return "", err
	}
	c.Log.Debug(ss.PrettyJSON(fileReport))

	if len(fileReport.Items) == 0 {
		return "", fmt.Errorf("no events found for file %s", fileID)
	}
	for _, event := range fileReport.Items[0].Events {
		for i, param := range event.Parameters {
			if param.Name == "owner" {
				c.Log.Println("Found owner!")
				c.Log.Debug("owner:", event.Parameters[i].Value)
				return event.Parameters[i].Value, nil
			}
		}
	}

	return "", fmt.Errorf("no owner found for file %s", fileID)
}

/*
 * Get Root Organization Unit of current customer
 * /admin/directory/v1/customer/{customerId}/orgunits/{orgUnitPath=**}
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/orgunits/get
 */
func (c *AdminClient) RootOU(customer *Customer) (*OrgUnit, error) {
	url := c.BuildURL(DirectoryOrgUnits, customer)
	cacheKey := fmt.Sprintf("%s_%s", url, "root")

	var cache OrgUnit
	if c.GetCache(cacheKey, &cache) {
		return &cache, nil
	}

	c.Log.Println("Getting ROOT org unit...")
	ou, err := do[*OrgUnit](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	ou, err = c.GetOU(customer, ou.OrganizationUnits[0].ParentID)
	if err != nil {
		return nil, err
	}

	c.SetCache(cacheKey, ou, 30*time.Minute)
	return ou, nil
}

/*
 * Get Organization Unit ID by Path
 * /admin/directory/v1/customer/{customerId}/orgunits/{orgUnitPath=**}
 * https://developers.google.com/admin-sdk/directory/reference/rest/v1/orgunits/get
 */
func (c *AdminClient) GetOU(customer *Customer, orgUnitPath string) (*OrgUnit, error) {
	url := c.BuildURL(DirectoryOrgUnits, customer, strings.TrimPrefix(orgUnitPath, "/"))

	var cache OrgUnit
	if c.GetCache(url, &cache) {
		return &cache, nil
	}

	c.Log.Println("Getting org unit...")
	ou, err := do[OrgUnit](c.Client, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}

	c.SetCache(url, ou, 5*time.Minute)
	return &ou, nil
}

/*
 * Clone Direct Chrome Policies (Non-Inherited) for Organization Units by Path
 * chromepolicy.googleapis.com/v1/{customer=customers/*}/policies/orgunits:batchModify
 * https://developers.google.com/chrome/policy/reference/rest/v1/customers.policies.orgunits/batchModify
 */
func (c *AdminClient) CloneOU(customer *Customer, sourcePath, targetPath string) error {
	url := c.BuildURL(DevicePolicies, customer, "orgunits:batchModify")

	schemas, err := c.Devices().ListAllDevicePolicySchemas(customer)
	if err != nil {
		return err
	}

	sourceOU, err := c.GetOU(customer, sourcePath)
	if err != nil {
		return err
	}

	sourcePolicies, err := c.Devices().ResolvePolicySchemas(customer, sourceOU)
	if err != nil {
		return err
	}

	targetOU, err := c.GetOU(customer, targetPath)
	if err != nil {
		return err
	}

	// We want to ensure we only clone policies directly applied to the source OU (non-inherited).
	userPayload := PolicyModificationRequests{
		Requests: createPolicyModificationRequests(*sourcePolicies.Users.Direct, targetOU, schemas),
	}

	devicePayload := PolicyModificationRequests{
		Requests: createPolicyModificationRequests(*sourcePolicies.Devices.Direct, targetOU, schemas),
	}

	if len(userPayload.Requests) > 0 {
		c.Log.Println("Cloning user policies...")
		_, err = do[any](c.Client, "POST", url, nil, userPayload)
		if err != nil {
			return err
		}
		c.Log.Printf("Successfully cloned user policies! (https://admin.google.com/ac/chrome/settings/user?ac_ouid=%s)", strings.TrimPrefix(targetOU.ID, "id:"))
	} else {
		c.Log.Println("No user policies to clone.")
	}

	if len(devicePayload.Requests) > 0 {
		c.Log.Println("Cloning device policies...")
		_, err = do[any](c.Client, "POST", url, nil, devicePayload)
		if err != nil {
			return err
		}
		c.Log.Printf("Successfully cloned chrome device policies! (https://admin.google.com/ac/chrome/settings/device?ac_ouid=%s)", strings.TrimPrefix(targetOU.ID, "id:"))
	} else {
		c.Log.Println("No device policies to clone.")
	}

	return nil
}

func createPolicyModificationRequests(policies []*ResolvedPolicy, targetOU *OrgUnit, schemas *PolicySchemas) []*PolicyModificationRequest {
	var requests []*PolicyModificationRequest
	for _, policy := range policies {
		var fields []string
		var updateMask string
		for _, schema := range *schemas.PolicySchemas {
			if policy.Value.PolicySchema == schema.SchemaName {
				for _, field := range schema.FieldDescriptions {
					_, exists := policy.Value.Value[field.Field]
					if exists {
						fields = append(fields, field.Field)
					}
				}
				updateMask = strings.Join(fields, ",")
				fields = nil
			}
		}
		requests = append(requests, &PolicyModificationRequest{
			PolicyTargetKey: PolicyTargetKey{
				TargetResource: fmt.Sprintf("orgunits/%s", strings.TrimPrefix(targetOU.ID, "id:")),
			},
			PolicyValue: policy.Value,
			UpdateMask:  updateMask,
		})
	}
	return requests
}
