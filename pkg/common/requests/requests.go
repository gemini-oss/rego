// pkg/common/requests/requests.go
package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	rl "github.com/gemini-oss/rego/pkg/common/ratelimit"
	ss "github.com/gemini-oss/rego/pkg/common/starstruct"
)

type Headers map[string]string

const (
	Atom              = "application/atom+xml"
	CSS               = "text/css"
	Excel             = "application/vnd.ms-excel"
	FormURLEncoded    = "application/x-www-form-urlencoded"
	GIF               = "image/gif"
	HTML              = "text/html"
	JPEG              = "image/jpeg"
	JavaScript        = "text/javascript"
	JSON              = "application/json"
	MP3               = "audio/mpeg"
	MP4               = "video/mp4"
	MPEG              = "video/mpeg"
	MultipartFormData = "multipart/form-data"
	OctetStream       = "application/octet-stream"
	PDF               = "application/pdf"
	PNG               = "image/png"
	Plain             = "text/plain"
	RSS               = "application/rss+xml"
	WAV               = "audio/wav"
	XML               = "application/xml"
	ZIP               = "application/zip"
)

/*
 * Client
 * @param httpClient *http.Client
 * @param headers Headers
 */
type Client struct {
	httpClient  *http.Client
	Headers     Headers
	RateLimiter *rl.RateLimiter
}

/*
 * NewClient
 * @param headers Headers
 * @return *Client
 */
func NewClient(c *http.Client, headers Headers, rateLimiter *rl.RateLimiter) *Client {
	if c != nil {
		return &Client{
			httpClient:  c,
			Headers:     headers,
			RateLimiter: rateLimiter,
		}
	}
	return &Client{
		httpClient:  &http.Client{},
		Headers:     headers,
		RateLimiter: rateLimiter,
	}
}

// UpdateHeaders changes the headers for the HTTP client
func (c *Client) UpdateContentType(contentType string) {
	c.Headers["Content-Type"] = contentType
}

/*
 * Paginator
 * @param Self string
 * @param NextPage string
 * @param Paged bool
 */
type Paginator struct {
	Self          string `json:"self"`
	NextPageLink  string `json:"next"`
	NextPageToken string `json:"next_page_token"`
	Paged         bool   `json:"paged"`
}

/*
 * DecodeJSON
 * @param body []byte
 * @param result interface{}
 * @return error
 */
func DecodeJSON(body []byte, result interface{}) error {
	return json.Unmarshal(body, result)
}

func (c *Client) CreateRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func SetQueryParams(req *http.Request, query interface{}) {
	if query == nil {
		return
	}

	q := req.URL.Query()
	parameters := ss.StructToMap(query)

	for key, value := range parameters {
		q.Add(key, value)
	}

	req.URL.RawQuery = q.Encode()
}

func SetJSONPayload(req *http.Request, data interface{}) error {
	if data == nil {
		return nil
	}
	// p := ss.StructToMap(data)
	// payload, err := json.Marshal(p)
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	req.Body = io.NopCloser(strings.NewReader(string(payload)))
	req.ContentLength = int64(len(payload))
	return nil
}

func SetFormURLEncodedPayload(req *http.Request, data interface{}) error {
	if data == nil {
		return nil
	}

	// Convert data to URL-encoded form
	formData := url.Values{}
	parameters := ss.StructToMap(data)
	for key, value := range parameters {
		formData.Add(key, value)
	}

	req.Body = io.NopCloser(strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", FormURLEncoded)
	req.ContentLength = int64(len(formData.Encode()))
	return nil
}

func (c *Client) DoRequest(method string, url string, query interface{}, data interface{}) (*http.Response, []byte, error) {

	if c.RateLimiter != nil {
		c.RateLimiter.Wait()
	}

	req, err := c.CreateRequest(method, url)
	if err != nil {
		return nil, nil, err
	}

	SetQueryParams(req, query)

	switch c.Headers["Content-Type"] {
	case FormURLEncoded, fmt.Sprintf("%s; charset=utf-8", FormURLEncoded):
		err = SetFormURLEncodedPayload(req, data)
	case MultipartFormData:
	case JSON, fmt.Sprintf("%s; charset=utf-8", JSON):
		err = SetJSONPayload(req, data)
	default:
	}

	if err != nil {
		return nil, nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Update rate limiter if headers are present
	if c.RateLimiter != nil {
		c.RateLimiter.UpdateFromHeaders(resp.Header)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return nil, body, fmt.Errorf(string(body))
	case http.StatusUnauthorized:
		return nil, body, fmt.Errorf(string(body))
	case http.StatusForbidden:
		return nil, body, fmt.Errorf(string(body))
	case http.StatusNotFound:
		return nil, body, fmt.Errorf(string(body))
	case http.StatusTooManyRequests:
		fmt.Println(string(body))
		return nil, body, fmt.Errorf(string(body))
	default:
		return resp, body, nil
	}
}

/*
 * PaginatedRequest
 * @param method string
 * @param url string
 * @param query interface{}
 * @return []json.RawMessage
 * @return error
 */
func (c *Client) PaginatedRequest(method string, url string, query interface{}, payload interface{}) ([]json.RawMessage, error) {
	var results []json.RawMessage

	if c.RateLimiter != nil {
		c.RateLimiter.Wait()
	}

	// Initial request
	resp, body, err := c.DoRequest(method, url, query, nil)
	if err != nil {
		return results, err
	}

	// Decode JSON array to raw messages
	var page []json.RawMessage
	err = DecodeJSON(body, &page)
	if err != nil {
		// If it's not an array, try to unmarshal as a single object
		var singleObject json.RawMessage
		err = json.Unmarshal(body, &singleObject)
		if err != nil {
			// Return an error if it's neither an object nor an array
			return results, fmt.Errorf("decoding response: %w", err)
		}
		// If it's an object, add it to the results as a single-item slice
		results = append(results, singleObject)
	} else {
		// If it's an array, add it to the results
		results = append(results, page...)
	}

	// Pagination
	p := &Paginator{}
	for p.HasNextPage(resp.Header.Values("Link")) {
		if c.RateLimiter != nil {
			c.RateLimiter.Wait()
		}

		// Request next page
		resp, body, err = c.DoRequest("GET", p.NextPageLink, nil, nil)
		if err != nil {
			return results, err
		}

		// Decode JSON array to raw messages
		newPage := []json.RawMessage{}
		err = DecodeJSON(body, &newPage)
		if err != nil {
			// If it's not an array, try to unmarshal as a single object
			var singleObject json.RawMessage
			err = json.Unmarshal(body, &singleObject)
			if err != nil {
				// Return an error if it's neither an object nor an array
				return results, fmt.Errorf("decoding response: %w", err)
			}
			// If it's an object, add it to the results as a single-item slice
			results = append(results, singleObject)
		} else {
			// If it's an array, add it to the results
			results = append(results, page...)
		}
	}

	return results, nil
}

/*
 * HasNextPage
 * @param links []string
 * @return bool
 */
func (p *Paginator) HasNextPage(links []string) bool {
	for _, link := range links {
		rawLink := strings.Split(link, ";")[0]
		rawLink = strings.Trim(rawLink, "<>")

		if strings.Contains(link, `rel="self"`) {
			p.Self = rawLink
		}
		if strings.Contains(link, `rel="next"`) {
			p.NextPageLink = rawLink
			return true
		}
	}
	return false
}
