// pkg/common/requests/error_handling.go
package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
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
	case strings.Contains(contentType, JSON):
		c.parseJSONError(body, reqError)
	case strings.Contains(contentType, HTML):
		c.parseHTMLError(body, reqError)
	case strings.Contains(contentType, Plain):
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

	if reqError.RawResponse != "" {
		reqError.Message = fmt.Sprintf("%s: %s", reqError.Message, reqError.RawResponse)
	}

	if reqError.Message == "" {
		reqError.Message = "Unknown JSON error response"
	}
}

func (c *Client) parseHTMLError(body []byte, reqError *RequestError) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		reqError.Message = http.StatusText(reqError.StatusCode)
		return
	}

	blockTags := map[string]bool{
		"title": true, "h1": true, "h2": true, "h3": true,
		"p": true, "li": true, "div": true, "section": true,
	}

	var segments []string
	var collectText func(*html.Node) string
	collectText = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return strings.TrimSpace(n.Data)
		}
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return ""
		}
		var parts []string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if t := collectText(c); t != "" {
				parts = append(parts, t)
			}
		}
		return strings.Join(parts, " ")
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && blockTags[n.Data] {
			if txt := collectText(n); txt != "" {
				segments = append(segments, txt)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if len(segments) == 0 {
		reqError.Message = http.StatusText(reqError.StatusCode)
		return
	}

	msg := strings.Join(segments, " | ")

	// ▸ 1) remove any spaces *before* punctuation
	reSpaceBeforePunct := regexp.MustCompile(`\s+([.,;:!?])`)
	msg = reSpaceBeforePunct.ReplaceAllString(msg, `$1`)

	// ▸ 2) collapse runs of spaces into a single space
	reMultiSpace := regexp.MustCompile(`\s{2,}`)
	msg = reMultiSpace.ReplaceAllString(msg, " ")

	reqError.Message = msg
}
