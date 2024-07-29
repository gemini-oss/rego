// pkg/common/requests/error_handling.go
package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// RequestError represents an API request error.
type RequestError struct {
	StatusCode  int    `json:"status_code"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	Message     string `json:"message"`
	RawResponse string `json:"raw_response"`
}

// Error returns a string representation of the RequestError.
func (e *RequestError) Error() string {
	return fmt.Sprintf("Request Error: StatusCode=%d, Method=%s, URL=%s, Message=%s",
		e.StatusCode, e.Method, e.URL, e.Message)
}

// handleErrorResponse processes the HTTP error response.
func (c *Client) handleErrorResponse(resp *http.Response, body []byte) *RequestError {
	reqError := &RequestError{
		StatusCode:  resp.StatusCode,
		Method:      resp.Request.Method,
		URL:         resp.Request.URL.String(),
		RawResponse: string(body),
	}

	contentType := resp.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/json"):
		c.parseJSONError(body, reqError)
	case strings.Contains(contentType, "text/plain"):
		reqError.Message = string(body)
	default:
		reqError.Message = fmt.Sprintf("Unexpected error (Status: %d)", resp.StatusCode)
	}

	if c.Log != nil {
		c.Log.Error(reqError.Error())
	}
	return reqError
}

func (c *Client) parseJSONError(body []byte, reqError *RequestError) {
	var jsonError struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(body, &jsonError); err == nil {
		if jsonError.Message != "" {
			reqError.Message = jsonError.Message
		} else if jsonError.Error != "" {
			reqError.Message = jsonError.Error
		}
	}

	if reqError.Message == "" {
		reqError.Message = "Unknown JSON error response"
	}
}
