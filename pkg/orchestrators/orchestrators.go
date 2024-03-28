/*
# Orchestrators

This package contains some functions involving practical examples of multi-service orchestration.

:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/orchestrators/orchestrators.go
package orchestrators

import (
	"github.com/gemini-oss/rego/pkg/active_directory"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/google"
	"github.com/gemini-oss/rego/pkg/jamf"
	"github.com/gemini-oss/rego/pkg/okta"
	"github.com/gemini-oss/rego/pkg/snipeit"
)

type Client struct {
	Log             *log.Logger
	ActiveDirectory *active_directory.Client
	Google          *google.Client
	Jamf            *jamf.Client
	Okta            *okta.Client
	SnipeIT         *snipeit.Client
}

/*
 * Orchestrate the following:
 * Generate a report of all users and their roles in Okta
 * Save the report to a Google Sheet
 * Format the sheet
 */
func (c *Client) OktaRoleReportToGoogleSheet() error {
	roleReports, err := c.Okta.GenerateRoleReport()
	if err != nil {
		return err
	}

	sheet, err := c.Google.CreateSpreadsheet()
	if err != nil {
		return err
	}

	vr := &google.ValueRange{
		Range:          "A:Z",
		MajorDimension: "ROWS",
	}
	headers := []string{"ID", "Name", "Email", "Status", "Role ID", "Role Label", "Role AssignmentType", "Last Login"}
	vr.Values = append(vr.Values, headers)

	for _, report := range *roleReports {
		for _, user := range *report.Users {
			vr.Values = append(vr.Values, []string{user.ID, user.Profile.Email, user.Profile.Login, user.Status, report.Role.ID, report.Role.Label, report.Role.AssignmentType, user.LastLogin.String()})
		}
	}

	rows := len(vr.Values)
	columns := len(headers)

	err = c.Google.UpdateSpreadsheet(sheet.SpreadsheetID, vr)
	if err != nil {
		return err
	}

	err = c.Google.FormatHeaderAndAutoSize(sheet.SpreadsheetID, rows, columns)
	if err != nil {
		return err
	}

	c.Log.Println("Okta role report saved to Google Sheet.")
	c.Log.Println("Spreadsheet URL: ", sheet.SpreadsheetURL)

	return nil
}

/*
 * Orchestrate the following:
 * Generate a report of all members of a group in Active Directory
 * Save the report to a Google Sheet
 * Format the sheet
 */
func (c *Client) ADReportToGoogleSheet(group string) error {
	users, err := c.ActiveDirectory.MemberOf(group)
	if err != nil {
		return err
	}

	sheet, err := c.Google.CreateSpreadsheet()
	if err != nil {
		return err
	}

	vr := &google.ValueRange{
		Range:          "A:Z",
		MajorDimension: "ROWS",
	}
	headers := []string{"SAM Account Name", "First", "Last", "User Principal Name (UPN)", "Display Name", "Group"}
	vr.Values = append(vr.Values, headers)

	for _, user := range *users {
		vr.Values = append(vr.Values, []string{user.SAMAccountName, user.GivenName, user.SN, user.UserPrincipalName, user.DisplayName, group})
	}

	rows := len(vr.Values)
	columns := len(headers)

	err = c.Google.UpdateSpreadsheet(sheet.SpreadsheetID, vr)
	if err != nil {
		return err
	}

	err = c.Google.FormatHeaderAndAutoSize(sheet.SpreadsheetID, rows, columns)
	if err != nil {
		return err
	}

	c.Log.Println("Okta role report saved to Google Sheet.")
	c.Log.Println("Spreadsheet URL: ", sheet.SpreadsheetURL)

	return nil
}
