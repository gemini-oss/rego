/*
# Google Workspace - Sheets

This package initializes all the methods for functions which interact with the Google Sheets API:
https://developers.google.com/sheets/api/reference/rest

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/sheets.go
package google

import (
	"encoding/json"
	"fmt"
)

var (
	SheetsBaseURL          = "https://sheets.googleapis.com/v4"
	Sheets                 = fmt.Sprintf("%s/spreadsheets", SheetsBaseURL)             // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets
	SheetByID              = fmt.Sprintf("%s/%s", Sheets, "%s")                        // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/get
	SheetValuesRange       = fmt.Sprintf("%s/%s/values/%s", Sheets, "%s", "%s")        // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get
	SheetValuesBatchGet    = fmt.Sprintf("%s/%s/values:batchGet", Sheets, "%s")        // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchGet
	SheetValuesBatchUpdate = fmt.Sprintf("%s/%s/values:batchUpdate", Sheets, "%s")     // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchUpdate
	SheetValuesAppend      = fmt.Sprintf("%s/%s/values/%s:append", Sheets, "%s", "%s") // https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/append
)

/*
 * Query Parameters for Sheet Values
 */
type SheetValueQuery struct {
	ValueInputOption             string // How the input data should be interpreted. Accepted values are: RAW or USER_ENTERED. The default is USER_ENTERED.
	IncludeValuesInResponse      bool   // Determines if the update response should include the values of the cells that were updated. By default, responses do not include the updated values. If the range to write was larger than the range actually written, the response includes all values in the requested range (excluding trailing empty rows and columns).
	ResponseValueRenderOption    string // Determines how values in the response should be rendered. The default render option is FORMATTED_VALUE.
	ResponseDateTimeRenderOption string // Determines how dates, times, and durations in the response should be rendered. This is ignored if responseValueRenderOption is FORMATTED_VALUE. The default dateTime render option is SERIAL_NUMBER.
}

/*
 * # Set Sheet Value Defaults
 * - Sets default values for ValueRange if they are not defined
 */
func VerifySheetValueRange(vr *ValueRange) error {
	if vr.Range == "" {
		vr.Range = "A:Z"
	}
	if vr.MajorDimension == "" {
		vr.MajorDimension = "ROWS"
	}
	if vr.Values == nil {
		return fmt.Errorf("ValueRange.Values cannot be empty")
	}
	return nil
}

/*
 * # Spreadsheet: Create
 * - Creates a new spreadsheet, with basic properties.
 *   - https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/create
 */
func (c *Client) CreateSpreadsheet() (*Spreadsheet, error) {
	url := Sheets

	_, body, err := c.HTTPClient.DoRequest("POST", url, nil, nil)
	if err != nil {
		return nil, err
	}
	c.Logger.Debugf("Request Body: %s", string(body))

	spreadsheet := &Spreadsheet{}
	err = json.Unmarshal(body, &spreadsheet)
	if err != nil {
		return nil, err
	}

	return spreadsheet, nil
}

/*
 * # Spreadsheet Values: Update
 * - Sets/Replaces values in a range of a spreadsheet. The caller must specify the spreadsheet ID, range, and a valueInputOption
 *   - https://sheets.googleapis.com/v4/spreadsheets/{spreadsheetId}/values/{range}
 *   - https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/update
 */
func (c *Client) UpdateSpreadsheet(spreadsheetID string, vr *ValueRange) error {

	q := SheetValueQuery{
		ValueInputOption: "RAW",
	}

	// Check Value paramters
	err := VerifySheetValueRange(vr)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/values/%s", Sheets, spreadsheetID, vr.Range)

	// Prepare request
	res, body, err := c.HTTPClient.DoRequest("PUT", url, q, &vr)
	if err != nil {
		c.Logger.Panic(err)
		return err
	}
	c.Logger.Println("Response Status: ", res.Status)
	c.Logger.Debug("Response Body: ", string(body))

	return nil
}

/*
 * # Spreadsheet Values: Append
 * - Appends values within the range of a spreadsheet. The caller must specify the spreadsheet ID, range, and a valueInputOption
 *   - https://sheets.googleapis.com/v4/spreadsheets/{spreadsheetId}/values/{range}
 *   - https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/update
 */
func (c *Client) AppendSpreadsheet(spreadsheetID string, vr *ValueRange) error {

	q := SheetValueQuery{
		ValueInputOption: "RAW",
	}

	// Check Value paramters
	err := VerifySheetValueRange(vr)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/values/%s:append", Sheets, spreadsheetID, vr.Range)

	// Prepare request
	res, body, err := c.HTTPClient.DoRequest("POST", url, q, vr)
	if err != nil {
		return err
	}
	c.Logger.Println("Response Status: ", res.Status)
	c.Logger.Debug("Response Body: ", string(body))

	return nil
}

/*
 * # Format Header and AutoSize
 * - Sets the header row to bold and green, and auto-sizes all columns
 */
func (c *Client) FormatHeaderAndAutoSize(spreadsheetId string, rows int, columns int) error {
	url := fmt.Sprintf("%s/%s:batchUpdate", Sheets, spreadsheetId)

	format := &SheetBatchRequest{}

	// Set the header row to bold and green
	format.Requests = append(format.Requests, &SheetRequest{
		RepeatCell: &RepeatCellRequest{
			Range: &GridRange{
				SheetID:          0,
				StartRowIndex:    0,
				EndRowIndex:      1,
				StartColumnIndex: 0,
				EndColumnIndex:   columns,
			},
			Cell: &CellData{
				UserEnteredFormat: &CellFormat{
					BackgroundColor: &Color{
						Alpha: 1.0,
						Red:   (182.0 / 255.0),
						Green: (215.0 / 255.0),
						Blue:  (168.0 / 255.0),
					},
					TextFormat: &TextFormat{
						FontSize: 12,
						Bold:     true,
					},
				},
			},
			Fields: "userEnteredFormat(backgroundColor,textFormat)",
		},
	})

	// Add a filter view for the header row
	format.Requests = append(format.Requests, &SheetRequest{
		SetBasicFilter: &SetBasicFilterRequest{
			Filter: &BasicFilter{
				Range: &GridRange{
					SheetID:          0,
					StartRowIndex:    0,
					EndRowIndex:      rows,
					StartColumnIndex: 0,
					EndColumnIndex:   columns,
				},
			},
		},
	})

	// Auto resize all columns
	format.Requests = append(format.Requests, &SheetRequest{
		AutoResizeDimensions: &AutoResizeDimensionsRequest{
			Dimensions: &DimensionRange{
				SheetID:    0,
				Dimension:  "COLUMNS",
				StartIndex: 0,
				EndIndex:   columns,
			},
		},
	})

	// Execute the batchUpdate request
	resp, body, err := c.HTTPClient.DoRequest("POST", url, nil, format)
	if err != nil {
		return err
	}
	c.Logger.Println("Response Status: ", resp.Status)
	c.Logger.Debug("Response Body: ", string(body))

	return nil
}
