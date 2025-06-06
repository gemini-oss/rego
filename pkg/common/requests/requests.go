// pkg/common/requests/requests.go
package requests

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	rl "github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/retry"
	ss "github.com/gemini-oss/rego/pkg/common/starstruct"
)

type Headers map[string]string

const (
	All               = "*/*"                               // RFC-7231 (https://www.rfc-editor.org/rfc/rfc7231.html)
	Atom              = "application/atom+xml"              // RFC-4287 (https://www.rfc-editor.org/rfc/rfc4287.html)
	CSS               = "text/css"                          // RFC-2318 (https://www.rfc-editor.org/rfc/rfc2318.html)
	Excel             = "application/vnd.ms-excel"          // Proprietary
	FormURLEncoded    = "application/x-www-form-urlencoded" // RFC-1866 (https://www.rfc-editor.org/rfc/rfc1866.html)
	GIF               = "image/gif"                         // RFC-2046 (https://www.rfc-editor.org/rfc/rfc2046.html)
	HTML              = "text/html"                         // RFC-2854 (https://www.rfc-editor.org/rfc/rfc2854.html)
	JPEG              = "image/jpeg"                        // RFC-2045 (https://www.rfc-editor.org/rfc/rfc2045.html)
	JavaScript        = "text/javascript"                   // RFC-9239 (https://www.rfc-editor.org/rfc/rfc9239.html)
	JSON              = "application/json"                  // RFC-8259 (https://www.rfc-editor.org/rfc/rfc8259.html)
	MP3               = "audio/mpeg"                        // RFC-3003 (https://www.rfc-editor.org/rfc/rfc3003.html)
	MP4               = "video/mp4"                         // RFC-4337 (https://www.rfc-editor.org/rfc/rfc4337.html)
	MPEG              = "video/mpeg"                        // RFC-4337 (https://www.rfc-editor.org/rfc/rfc4337.html)
	MultipartFormData = "multipart/form-data"               // RFC-7578 (https://www.rfc-editor.org/rfc/rfc7578.html)
	OctetStream       = "application/octet-stream"          // RFC-2046 (https://www.rfc-editor.org/rfc/rfc2046.html)
	PDF               = "application/pdf"                   // RFC-3778 (https://www.rfc-editor.org/rfc/rfc3778.html)
	PNG               = "image/png"                         // RFC-2083 (https://www.rfc-editor.org/rfc/rfc2083.html)
	Plain             = "text/plain"                        // RFC-2046 (https://www.rfc-editor.org/rfc/rfc2046.html)
	RSS               = "application/rss+xml"               // RFC-7303 (https://www.rfc-editor.org/rfc/rfc4287.html)
	WAV               = "audio/wav"                         // RFC-2361 (https://www.rfc-editor.org/rfc/rfc2361.html)
	XML               = "application/xml"                   // RFC-7303 (https://www.rfc-editor.org/rfc/rfc7303.html)
	YAML              = "application/yaml"                  // RFC-9512 (https://www.rfc-editor.org/rfc/rfc9512.html)
	ZIP               = "application/zip"                   // RFC-1951 (https://www.rfc-editor.org/rfc/rfc1951.html)
)

var (
	l = log.NewLogger("{requests}", log.DEBUG)
)

/*
 * Client
 * @param httpClient *http.Client
 * @param headers Headers
 */
type Client struct {
	httpClient  *http.Client
	BodyType    string
	Cache       *cache.Cache
	Headers     Headers
	Log         *log.Logger
	RateLimiter *rl.RateLimiter
}

/*
 * NewClient
 * @param headers Headers
 * @return *Client
 */
func NewClient(options ...interface{}) *Client {
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		l.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_requests.gob", 1000000)
	if err != nil {
		panic(err)
	}

	client := &Client{
		httpClient:  &http.Client{},
		Cache:       cache,
		Headers:     Headers{},
		Log:         l,
		RateLimiter: nil,
	}

	for _, option := range options {
		switch opt := option.(type) {
		case *http.Client:
			client.httpClient = opt
		case Headers:
			client.Headers = opt
		case *rl.RateLimiter:
			client.RateLimiter = opt
		}
	}

	if client.RateLimiter == nil {
		client.RateLimiter = rl.NewRateLimiter(100)
	}

	return client
}

// UpdateHeaders changes the headers for the HTTP client
func (c *Client) UpdateAcceptType(contentType string) {
	c.Headers["Accept"] = contentType
}

// UpdateHeaders changes the headers for the HTTP client
func (c *Client) UpdateContentType(contentType string) {
	c.Headers["Content-Type"] = contentType
}

// UpdateHeaders changes the payload body for the HTTP client
func (c *Client) UpdateBodyType(bodyType string) {
	c.BodyType = bodyType
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
	parameters, err := ss.ToMap(query, false)
	if err != nil {
		return
	}

	for key, value := range parameters {
		switch v := value.(type) {
		case []interface{}:
			for _, item := range v {
				q.Add(key, fmt.Sprintf("%v", item))
			}
		default:
			q.Add(key, fmt.Sprintf("%v", value))
		}
	}

	req.URL.RawQuery = q.Encode()
}

func SetJSONPayload(req *http.Request, data interface{}) error {
	if data == nil {
		return nil
	}
	// p, err := ss.ToMap(data, false)
	// if err != nil {
	// 	return err
	// }

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

	formData := url.Values{}
	parameters, err := ss.ToMap(data, false)
	if err != nil {
		return err
	}

	for key, value := range parameters {
		switch v := value.(type) {
		case []interface{}:
			arrayKey := fmt.Sprintf("%s[]", key)
			for _, item := range v {
				formData.Add(arrayKey, fmt.Sprintf("%v", item))
			}
		default:
			formData.Add(key, fmt.Sprintf("%v", value))
		}
	}

	req.Body = io.NopCloser(strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", FormURLEncoded)
	req.ContentLength = int64(len(formData.Encode()))
	return nil
}

func SetXMLPayload(req *http.Request, data interface{}) error {
	if data == nil {
		return nil
	}

	payload, err := xml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	req.Body = io.NopCloser(strings.NewReader(string(payload)))
	req.ContentLength = int64(len(payload))
	return nil
}

func (c *Client) DoRequest(ctx context.Context, method string, url string, query interface{}, data interface{}) (*http.Response, []byte, error) {
	realTime := retry.RealTime{}
	return c.doRetry(ctx, method, url, query, data, realTime)
}

func (c *Client) doRetry(ctx context.Context, method string, url string, query interface{}, data interface{}, time retry.Time) (*http.Response, []byte, error) {
	var resp *http.Response
	var body []byte
	err := retry.Retry(
		func() error {
			var reqErr error
			resp, body, reqErr = c.do(ctx, method, url, query, data)
			return reqErr
		},
		func(err error) bool {
			return err != nil && (resp == nil || IsRetryableStatusCode(resp.StatusCode))
		},
		time,
	)

	return resp, body, err
}

func (c *Client) do(ctx context.Context, method string, url string, query interface{}, data interface{}) (*http.Response, []byte, error) {
	// Validate HTTP method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"HEAD": true, "OPTIONS": true, "PATCH": true,
	}
	if _, valid := validMethods[method]; !valid {
		return nil, nil, fmt.Errorf("invalid HTTP method: %s", method)
	}

	req, err := c.CreateRequest(method, url)
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)

	SetQueryParams(req, query)

	if err := setPayload(req, data, c.BodyType); err != nil {
		return nil, nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, &RequestError{
			StatusCode: http.StatusInternalServerError,
			Method:     method,
			URL:        url,
			Message:    err.Error(),
		}
	}
	defer resp.Body.Close()

	// Update rate limiter if headers are present
	if c.RateLimiter != nil {
		c.RateLimiter.UpdateFromHeaders(resp.Header)
		c.RateLimiter.Wait()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response body: %w", err)
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return resp, body, nil
	case IsRedirectCode(resp.StatusCode):
		c.Log.Warning("Redirect status code encountered:", resp.StatusCode)
		return resp, body, c.handleErrorResponse(resp, body)
	case IsRetryableStatusCode(resp.StatusCode):
		c.Log.Warning("Retryable status code encountered:", resp.StatusCode)
		return resp, body, c.handleErrorResponse(resp, body)
	case IsNonRetryableCode(resp.StatusCode):
		c.Log.Error("Non-retryable status code encountered:", resp.StatusCode)
		return resp, body, c.handleErrorResponse(resp, body)
	case IsTemporaryErrorCode(resp.StatusCode):
		c.Log.Warning("Temporary error code encountered:", resp.StatusCode)
		return resp, body, c.handleErrorResponse(resp, body)
	default:
		c.Log.Error("Unexpected status code:", resp.StatusCode)
		return resp, body, c.handleErrorResponse(resp, body)
	}
}

func setPayload(req *http.Request, data interface{}, bodyType string) error {
	switch bodyType {
	case FormURLEncoded, fmt.Sprintf("%s; charset=utf-8", FormURLEncoded):
		return SetFormURLEncodedPayload(req, data)
	case JSON, fmt.Sprintf("%s; charset=utf-8", JSON):
		return SetJSONPayload(req, data)
	case XML, fmt.Sprintf("%s; charset=utf-8", XML):
		return SetXMLPayload(req, data)
	default:
		// No payload to set
		return nil
	}
}

func (c *Client) ExtractParam(u, parameter string) string {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return ""
	}

	queryParams := parsedURL.Query()

	paramValue := queryParams.Get(parameter)

	return paramValue
}
