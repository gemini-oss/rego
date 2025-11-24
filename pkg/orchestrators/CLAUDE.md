# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `orchestrators` package provides practical examples of multi-service orchestration using various Rego client libraries. It demonstrates how to combine different services (Okta, Active Directory, Google Sheets) to create automated workflows and reports, serving as both a reference implementation and a starting point for custom workflows.

## Architecture

### Core Components

1. **Client** (`orchestrators.go`):
   ```go
   type Client struct {
       Log             *log.Logger
       ActiveDirectory *active_directory.Client
       Google          *google.Client
       Jamf            *jamf.Client
       Okta            *okta.Client
       SnipeIT         *snipeit.Client
   }
   ```
   - Orchestrator client that aggregates multiple service clients
   - Each service client is optional (can be nil if not needed)
   - Uses common logger for consistent output

2. **Implemented Orchestrations**:
   - `OktaRoleReportToGoogleSheet()`: Generates Okta role reports in Google Sheets
   - `ADReportToGoogleSheet(group string)`: Creates AD group membership reports in Google Sheets

### Key Design Patterns

- **Service Aggregation**: Single client holds references to all service clients
- **Cross-Service Operations**: Combines data from one service with actions in another
- **Report Generation**: Automated spreadsheet creation with formatting
- **Error Propagation**: Consistent error handling across service boundaries
- **Timestamp Naming**: Reports include date in title for tracking

## Development Tasks

### Common Operations

1. **Setting Up the Orchestrator**:
   ```go
   // Initialize only the services you need
   orch := &orchestrators.Client{
       Log:    log.NewLogger(log.DEBUG),
       Okta:   oktaClient,
       Google: googleClient,
   }

   // Run orchestration
   err := orch.OktaRoleReportToGoogleSheet()
   ```

2. **Report Generation Pattern**:
   ```go
   func (c *Client) NewReportOrchestration() error {
       // 1. Fetch data from source service
       data, err := c.SourceService.GetData()
       if err != nil {
           return fmt.Errorf("failed to fetch data: %w", err)
       }

       // 2. Create Google Spreadsheet with timestamp
       sheet, err := c.Google.Sheets().CreateSpreadsheet(&google.Spreadsheet{
           Properties: &google.SpreadsheetProperties{
               Title: fmt.Sprintf("{Service} Report %s",
                   time.Now().Format("2006-01-02")),
           },
       })

       // 3. Prepare data with headers
       vr := &google.ValueRange{
           Range:          "A:Z",
           MajorDimension: "ROWS",
       }
       headers := []string{"Col1", "Col2", "Col3"}
       vr.Values = append(vr.Values, headers)

       // 4. Add data rows
       for _, item := range data {
           vr.Values = append(vr.Values, []string{
               item.Field1, item.Field2, item.Field3,
           })
       }

       // 5. Update and format sheet
       c.Google.Sheets().UpdateSpreadsheet(sheet.SpreadsheetID, vr)
       c.Google.Sheets().FormatHeaderAndAutoSize(
           sheet.SpreadsheetID,
           &sheet.Sheets[0],
           len(vr.Values),
           len(headers),
       )

       // 6. Log results
       c.Log.Printf("Report saved to: %s", sheet.SpreadsheetURL)
       return nil
   }
   ```

### Existing Orchestration Examples

#### OktaRoleReportToGoogleSheet
- Fetches role assignments from Okta using `GenerateRoleReport()`
- Creates spreadsheet with title format: `{Okta} Entitlement Review YYYY-MM-DD`
- Includes columns: ID, Name, Email, Status, Role ID, Role Label, Role AssignmentType, Last Login
- Automatically formats header row and auto-sizes columns
- Logs spreadsheet URL on completion

#### ADReportToGoogleSheet
- Takes AD group name as parameter
- Fetches group members using `MemberOf(group)`
- Creates spreadsheet with title format: `{Active Directory} Entitlement Review [GroupName] YYYY-MM-DD`
- Includes columns: SAM Account Name, First, Last, User Principal Name (UPN), Last Login, Enabled, Group
- Formats UserAccountControl as string for readability
- Returns spreadsheet URL in logs

## Important Notes

- **Error Handling**: Each orchestration should wrap errors with context
- **Nil Checks**: Always check if service clients are initialized before use
- **Logging**: Use the orchestrator's logger for consistent output
- **Timestamps**: Use `time.Now().Format("2006-01-02")` for consistent date formatting
- **Sheet Naming**: Include service name in braces for easy identification
- **Range**: Use "A:Z" for flexibility in column count
- **Headers**: Always include headers as first row for clarity

## Common Use Cases

1. **Compliance Reporting**:
   - Periodic Okta role reviews
   - AD group membership audits
   - Cross-system access verification

2. **Asset Management**:
   - Combine Jamf device data with SnipeIT inventory
   - Match AD users with assigned devices
   - Track software deployments across systems

3. **User Lifecycle**:
   - Provision accounts across Okta, AD, and Google
   - Deprovision with audit trail in Sheets
   - Track user status changes

4. **Security Auditing**:
   - Last login reports across systems
   - Disabled account tracking
   - Permission drift detection

## Implementation Examples

### Cross-System User Audit
```go
func (c *Client) CrossSystemUserAudit(email string) error {
    // Check user in each system
    oktaUser, _ := c.Okta.Users().GetUser(email)
    adUser, _ := c.ActiveDirectory.GetUser(email)
    googleUser, _ := c.Google.Users().Get(email)

    // Create comparison report
    // ... implementation
}
```

### Device Inventory Sync
```go
func (c *Client) SyncDeviceInventory() error {
    // Get all Jamf devices
    jamfDevices, _ := c.Jamf.Computers().ListAllComputers()

    // Update or create in SnipeIT
    for _, device := range jamfDevices {
        c.SnipeIT.Hardware().CreateOrUpdate(device)
    }
}
```

## Testing Orchestrations

```go
// Use mock clients for testing
mockGoogle := &mockGoogleClient{}
orch := &orchestrators.Client{
    Log:    log.NewLogger(log.DEBUG),
    Google: mockGoogle,
}

// Test orchestration logic without real API calls
```

## Best Practices

1. **Incremental Development**: Start with simple orchestrations, add complexity
2. **Error Context**: Always wrap errors with descriptive context
3. **Idempotency**: Design orchestrations to be safely re-runnable
4. **Progress Logging**: Log major steps for long-running operations
5. **Partial Failures**: Handle and report partial failures gracefully
6. **Rate Limiting**: Be aware of API limits when orchestrating bulk operations

## Future Expansion Ideas

- **Scheduled Reports**: Integration with cron/scheduler for periodic runs
- **Notification System**: Send Slack messages on completion/failure
- **Data Validation**: Cross-reference data for inconsistencies
- **Backup Orchestration**: Use Backupify to backup critical data
- **Approval Workflows**: Integrate with ticketing systems for changes
- **Metrics Collection**: Track orchestration performance and success rates
