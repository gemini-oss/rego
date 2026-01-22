package jira

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

   // "github.com/gemini-oss/rego/pkg/common/log"
)

// Client represents the Jira API client
type Client struct {
    HTTP     *http.Client
    Log      *log.Logger
    BaseURL  string
    Username string
    Token    string
}

// Issue represents a Jira ticket payload
type Issue struct {
    Fields struct {
        Project struct {
            Key string `json:"key"`
        } `json:"project"`
        Summary     string `json:"summary"`
        Description string `json:"description"`
        Issuetype   struct {
            Name string `json:"name"`
        } `json:"issuetype"`
        Assignee struct {
            Name string `json:"name,omitempty"`
        } `json:"assignee,omitempty"`
    } `json:"fields"`
}

// Response represents the Jira API response
type Response struct {
    ID   string `json:"id"`
    Key  string `json:"key"`
    Self string `json:"self"`
}

// NewClient creates a new Jira client instance
func NewClient(baseURL, username, token string, logger *log.Logger) *Client {
    return &Client{
        HTTP:     &http.Client{},
        Log:      logger,
        BaseURL:  baseURL,
        Username: username,
        Token:    token,
    }
}

// CreateTicket creates a new Jira ticket
func (j *Client) CreateTicket(summary, description, projectKey, issueType, assignee string) (string, error) {
    url := fmt.Sprintf("%s/rest/api/2/issue", j.BaseURL)

    issue := Issue{
        Fields: struct {
            Project struct {
                Key string `json:"key"`
            } `json:"project"`
            Summary     string `json:"summary"`
            Description string `json:"description"`
            Issuetype   struct {
                Name string `json:"name"`
            } `json:"issuetype"`
            Assignee struct {
                Name string `json:"name,omitempty"`
            } `json:"assignee,omitempty"`
        }{
            Project: struct {
                Key string `json:"key"`
            }{
                Key: projectKey,
            },
            Summary:     summary,
            Description: description,
            Issuetype: struct {
                Name string `json:"name"`
            }{
                Name: issueType,
            },
        },
    }

    if assignee != "" {
        issue.Fields.Assignee.Name = assignee
    }

    payload, err := json.Marshal(issue)
    if err != nil {
        return "", fmt.Errorf("marshaling Jira issue: %w", err)
    }

    req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(payload))
    if err != nil {
        return "", fmt.Errorf("creating request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.SetBasicAuth(j.Username, j.Token)

    resp, err := j.HTTP.Do(req)
    if err != nil {
        return "", fmt.Errorf("sending request: %w", err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("reading response: %w", err)
    }

    j.Log.Printf("Jira Response Status: %s", resp.Status)
    j.Log.Printf("Jira Response Body: %s", string(respBody))

    var jiraResp Response
    if err := json.Unmarshal(respBody, &jiraResp); err != nil {
        return "", fmt.Errorf("unmarshalling response: %w", err)
    }

    return jiraResp.Key, nil
}
