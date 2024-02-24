/*
# Orchestrators

This package contains some functions involving practical examples of multi-service orchestration.

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/orchestrators/orchestrators.go
package orchestrators

import (
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/google"
	"github.com/gemini-oss/rego/pkg/jamf"
	"github.com/gemini-oss/rego/pkg/okta"
	"github.com/gemini-oss/rego/pkg/snipeit"
)

type Client struct {
	Log     *log.Logger
	Google  *google.Client
	Jamf    *jamf.Client
	Okta    *okta.Client
	SnipeIT *snipeit.Client
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
	headers := []string{"ID", "Name", "Email", "Status", "Role", "Last Login"}
	vr.Values = append(vr.Values, headers)

	for _, report := range roleReports {
		for _, user := range report.Users {
			vr.Values = append(vr.Values, []string{user.ID, user.Profile.Email, user.Profile.Login, user.Status, report.Role.ID, user.LastLogin.String()})
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
