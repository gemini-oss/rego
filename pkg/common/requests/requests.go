// pkg/common/requests/requests.go
package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	ss "github.com/gemini-oss/rego/pkg/common/starstruct"
)

type Headers map[string]string

/*
 * Client
 * @param httpClient *http.Client
 * @param headers Headers
 */
type Client struct {
	httpClient *http.Client
	headers    Headers
}

/*
 * NewClient
 * @param headers Headers
 * @return *Client
 */
func NewClient(c *http.Client, headers Headers) *Client {
	if c != nil {
		return &Client{
			httpClient: c,
			headers:    headers,
		}
	}
	return &Client{
		httpClient: &http.Client{},
		headers:    headers,
	}
}

/*
 * Paginator
 * @param Self string
 * @param NextPage string
 * @param Paged bool
 */
type Paginator struct {
	Self     string `json:"self"`
	NextPage string `json:"next"`
	Paged    bool   `json:"paged"`
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

/*
 * DoRequest
 * @param method string
 * @param url string
 * @param query interface{}
 * @return *http.Response
 * @return []byte
 * @return error
 */
func (c *Client) DoRequest(method string, url string, query interface{}, data interface{}) (*http.Response, []byte, error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, err
	}

	// Set headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Set data
	if query != nil {
		switch method {
		case "GET":
			// Data will be treated as query parameters
			q := req.URL.Query()

			parameters := ss.StructToMap(query)

			for key, value := range parameters {
				q.Add(key, value)
			}

			req.URL.RawQuery = q.Encode()
		case "POST", "PUT", "PATCH":
			// Data will be treated as query parameters
			q := req.URL.Query()

			parameters := ss.StructToMap(query)

			for key, value := range parameters {
				q.Add(key, value)
			}

			req.URL.RawQuery = q.Encode()

			// Data will be treated as a JSON payload
			payload, err := json.Marshal(data)
			if err != nil {
				return nil, nil, fmt.Errorf("marshaling request body: %w", err)
			}
			req.Body = io.NopCloser(strings.NewReader(string(payload)))
			req.ContentLength = int64(len(payload))
		}
	} else if data != nil {
		switch method {
		case "POST", "PUT", "PATCH":
			// Data will be treated as a JSON payload
			payload, err := json.Marshal(data)
			if err != nil {
				return nil, nil, fmt.Errorf("marshaling request body: %w", err)
			}
			req.Body = io.NopCloser(strings.NewReader(string(payload)))
			req.ContentLength = int64(len(payload))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusForbidden:
		return nil, body, fmt.Errorf(string(body))
	case http.StatusNotFound:
		return nil, body, fmt.Errorf(string(body))
	case http.StatusTooManyRequests:
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
		// Request next page
		resp, body, err = c.DoRequest("GET", p.NextPage, nil, nil)
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
			p.NextPage = rawLink
			return true
		}
	}
	return false
}
