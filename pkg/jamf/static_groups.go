/*
# Jamf - Static Computer Groups

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/get_v1-computer-groups
- https://developer.jamf.com/jamf-pro/reference/get_v2-computer-groups-static-groups-id
- https://developer.jamf.com/jamf-pro/reference/post_v2-computer-groups-static-groups
- https://developer.jamf.com/jamf-pro/reference/put_v2-computer-groups-static-groups-id
- https://developer.jamf.com/jamf-pro/reference/delete_v2-computer-groups-static-groups-id
- https://developer.jamf.com/jamf-pro/reference/updatecomputergroupbyid (Classic API)

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/static_groups.go
package jamf

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/gemini-oss/rego/pkg/common/log"
)

var (
	// StaticComputerGroups points to v2 static groups endpoint (requires newer Jamf version)
	StaticComputerGroups = fmt.Sprintf("%s/computer-groups/static-groups", V2) // /api/v2/computer-groups/static-groups

	// ClassicComputerGroups points to Classic API computer groups endpoint
	ClassicComputerGroups = fmt.Sprintf("%s/computergroups", "%s") // /computergroups
)

// StaticGroupClient for chaining methods
type StaticGroupClient struct {
	baseClient *Client
	query      StaticGroupQuery
	log        *log.Logger
}

// Entry point for static group-related operations
func (c *Client) StaticGroups() *StaticGroupClient {
	return &StaticGroupClient{
		baseClient: c,
		query:      StaticGroupQuery{}, // Default query parameters
		log:        c.Log,
	}
}

/*
 * Query parameters for Static Computer Groups
 */
type StaticGroupQuery struct {
	Page     int    `url:"page,omitempty"`      // Page number for pagination (0-based index)
	PageSize int    `url:"page-size,omitempty"` // Number of records per page. Default is 100.
	Sort     string `url:"sort,omitempty"`      // Sorting criteria in the format: property:asc/desc. Default sort is name:asc.
	Filter   string `url:"filter,omitempty"`    // RSQL query string used for filtering the static computer groups collection.
}

/*
 * Check if the StaticGroupQuery is empty
 */
func (q *StaticGroupQuery) IsEmpty() bool {
	return q.Page == 0 &&
		q.PageSize == 0 &&
		q.Sort == "" &&
		q.Filter == ""
}

/*
 * Validate the query parameters for Jamf static groups
 */
func (q *StaticGroupQuery) ValidateQuery() error {
	if q.Page < 0 {
		return fmt.Errorf("page must be greater than or equal to 0")
	}

	if q.PageSize < 0 {
		return fmt.Errorf("page size must be greater than or equal to 0")
	}

	return nil
}

// ### Chainable StaticGroupClient Methods
// ---------------------------------------------------------------------
func (sgc *StaticGroupClient) Page(page int) *StaticGroupClient {
	sgc.query.Page = page
	return sgc
}

func (sgc *StaticGroupClient) PageSize(pageSize int) *StaticGroupClient {
	sgc.query.PageSize = pageSize
	return sgc
}

func (sgc *StaticGroupClient) Sort(sort string) *StaticGroupClient {
	sgc.query.Sort = sort
	return sgc
}

func (sgc *StaticGroupClient) Filter(filter string) *StaticGroupClient {
	sgc.query.Filter = filter
	return sgc
}

// END OF CHAINABLE METHODS
//---------------------------------------------------------------------

/*
 * # List All Static Computer Groups
 * /api/v1/computer-groups
 * - https://developer.jamf.com/jamf-pro/reference/get_v1-computer-groups
 *
 * Note: This endpoint returns both smart and static groups. We filter for static groups only (isSmart=false).
 */
func (sgc *StaticGroupClient) ListAll() ([]V1ComputerGroup, error) {
	// Use the ComputerGroups constant from devices.go which points to /api/v1/computer-groups
	url := sgc.baseClient.BuildURL(ComputerGroups)

	var cache []V1ComputerGroup
	if sgc.baseClient.GetCache(url, &cache) {
		return cache, nil
	}

	// Get all computer groups (both smart and static)
	// Note: The v1 endpoint returns a raw array, not wrapped in an object
	allGroups, err := do[[]V1ComputerGroup](sgc.baseClient, "GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list computer groups: %w", err)
	}

	// Filter for static groups only (where IsSmart = false)
	staticGroups := make([]V1ComputerGroup, 0)
	for _, group := range allGroups {
		if !group.IsSmart {
			staticGroups = append(staticGroups, group)
		}
	}

	sgc.baseClient.SetCache(url, staticGroups, 5*time.Minute)
	return staticGroups, nil
}

/*
 * # Get Static Computer Group by ID
 * /api/v2/computer-groups/static-groups/{id}
 * - https://developer.jamf.com/jamf-pro/reference/get_v2-computer-groups-static-groups-id
 */
func (sgc *StaticGroupClient) GetByID(id string) (*StaticComputerGroup, error) {
	url := sgc.baseClient.BuildURL(StaticComputerGroups, id)

	var cache StaticComputerGroup
	if sgc.baseClient.GetCache(url, &cache) {
		return &cache, nil
	}

	group, err := do[StaticComputerGroup](sgc.baseClient, "GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get static computer group %s: %w", id, err)
	}

	sgc.baseClient.SetCache(url, group, 5*time.Minute)
	return &group, nil
}

/*
 * # Create Static Computer Group
 * /api/v2/computer-groups/static-groups
 * - https://developer.jamf.com/jamf-pro/reference/post_v2-computer-groups-static-groups
 */
func (sgc *StaticGroupClient) Create(group *StaticComputerGroupRequest) (*StaticComputerGroupResponse, error) {
	url := sgc.baseClient.BuildURL(StaticComputerGroups)

	if group == nil {
		return nil, fmt.Errorf("group cannot be nil")
	}

	if group.Name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	created, err := do[StaticComputerGroupResponse](sgc.baseClient, "POST", url, nil, group)
	if err != nil {
		return nil, fmt.Errorf("failed to create static computer group: %w", err)
	}

	return &created, nil
}

/*
 * # Update Static Computer Group
 * /api/v2/computer-groups/static-groups/{id}
 * - https://developer.jamf.com/jamf-pro/reference/put_v2-computer-groups-static-groups-id
 */
func (sgc *StaticGroupClient) Update(id string, group *StaticComputerGroupRequest) (*StaticComputerGroupResponse, error) {
	url := sgc.baseClient.BuildURL(StaticComputerGroups, id)

	if group == nil {
		return nil, fmt.Errorf("group cannot be nil")
	}

	if group.Name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	updated, err := do[StaticComputerGroupResponse](sgc.baseClient, "PUT", url, nil, group)
	if err != nil {
		return nil, fmt.Errorf("failed to update static computer group %s: %w", id, err)
	}

	return &updated, nil
}

/*
 * # Delete Static Computer Group
 * /api/v2/computer-groups/static-groups/{id}
 * - https://developer.jamf.com/jamf-pro/reference/delete_v2-computer-groups-static-groups-id
 */
func (sgc *StaticGroupClient) Delete(id string) error {
	url := sgc.baseClient.BuildURL(StaticComputerGroups, id)

	_, err := do[any](sgc.baseClient, "DELETE", url, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete static computer group %s: %w", id, err)
	}

	return nil
}

/*
 * # Add Computers to Static Group
 * /JSSResource/computergroups/id/{id}
 * - https://developer.jamf.com/jamf-pro/reference/updatecomputergroupbyid
 *
 * This function uses the Classic API approach:
 * 1. GET the current group (with its computer list)
 * 2. Add the new computer IDs to the existing computers
 * 3. PUT the updated group back
 */
func (sgc *StaticGroupClient) AddComputers(groupID string, computerIDs []string) error {
	if len(computerIDs) == 0 {
		return fmt.Errorf("at least one computer ID is required")
	}

	// Convert groupID from string to int
	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		return fmt.Errorf("invalid group ID %s: must be a number", groupID)
	}

	// Step 1: GET the current group using Classic API
	url := sgc.baseClient.BuildClassicURL(ClassicComputerGroups, "id", groupIDInt)

	group, err := do[ClassicComputerGroup](sgc.baseClient, "GET", url, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to get computer group %s: %w", groupID, err)
	}

	// Step 2: Build the list of computer IDs to add
	// Create a map of existing computer IDs for quick lookup
	existingComputers := make(map[int]bool)
	for _, computer := range group.Computers {
		existingComputers[computer.ID] = true
	}

	// Add new computers to the group (skip duplicates)
	for _, computerIDStr := range computerIDs {
		computerID, err := strconv.Atoi(computerIDStr)
		if err != nil {
			sgc.log.Warning(fmt.Sprintf("Skipping invalid computer ID %s: must be a number", computerIDStr))
			continue
		}

		if !existingComputers[computerID] {
			group.Computers = append(group.Computers, ClassicComputerGroupMember{
				ID: computerID,
			})
		}
	}

	// Step 3: PUT the updated group back using Classic API
	// Wrap the group in an XML root element
	groupBody := struct {
		XMLName xml.Name `xml:"computer_group"`
		ClassicComputerGroup
	}{
		ClassicComputerGroup: group,
	}

	_, err = do[any](sgc.baseClient, "PUT", url, nil, groupBody)
	if err != nil {
		return fmt.Errorf("failed to add computers to static group %s: %w", groupID, err)
	}

	return nil
}

/*
 * # Remove Computer from Static Group
 * /api/v2/computer-groups/static-groups/{id}/computers/{computerId}
 * - https://developer.jamf.com/jamf-pro/reference/delete_v2-computer-groups-static-groups-id-computers-computerid
 */
func (sgc *StaticGroupClient) RemoveComputer(groupID string, computerID string) error {
	url := sgc.baseClient.BuildURL(StaticComputerGroups, groupID, "computers", computerID)

	_, err := do[any](sgc.baseClient, "DELETE", url, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to remove computer %s from static group %s: %w", computerID, groupID, err)
	}

	return nil
}

