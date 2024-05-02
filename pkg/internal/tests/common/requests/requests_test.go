// pkg/internal/tests/common/requests/requests_test.go
package requests_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockHTTPClient(responseBody string, statusCode int, err error) *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if err != nil {
				return nil, err
			}
			return &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
				Header:     make(http.Header),
			}, nil
		}),
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		client      *http.Client
		headers     requests.Headers
		rateLimiter *ratelimit.RateLimiter
		wantNil     bool
	}{
		{"With Custom Client", &http.Client{}, requests.Headers{"Content-Type": requests.JSON}, nil, false},
		{"With Nil Client", nil, requests.Headers{"Content-Type": requests.JSON}, nil, false},
		{"With Rate Limiter", nil, requests.Headers{"Content-Type": requests.JSON}, &ratelimit.RateLimiter{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requests.NewClient(tt.client, tt.headers, tt.rateLimiter)
			if (got == nil) != tt.wantNil {
				t.Errorf("NewClient() = %v, want nil: %v", got, tt.wantNil)
			}
		})
	}
}

func TestDecodeJSON(t *testing.T) {
	type SampleStruct struct {
		Field string `json:"field"`
	}

	tests := []struct {
		name    string
		body    []byte
		result  interface{}
		wantErr bool
	}{
		{"Valid JSON", []byte(`{"field":"value"}`), &SampleStruct{}, false},
		{"Invalid JSON", []byte(`{"field":}`), &SampleStruct{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result SampleStruct
			err := requests.DecodeJSON(tt.body, &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetQueryParams(t *testing.T) {
	type QueryStruct struct {
		Param1 string `json:"param1"`
		Param2 int    `json:"param2"`
	}

	tests := []struct {
		name        string
		request     *http.Request
		queryParams interface{}
		expectedURL string
	}{
		{
			"Single Query Param",
			httptest.NewRequest("GET", "http://gemini.com", nil),
			QueryStruct{Param1: "test"},
			"http://gemini.com?param1=test",
		},
		{
			"Multiple Query Params",
			httptest.NewRequest("GET", "http://gemini.com", nil),
			QueryStruct{Param1: "test", Param2: 123},
			"http://gemini.com?param1=test&param2=123",
		},
		{
			"Empty Query Params",
			httptest.NewRequest("GET", "http://gemini.com", nil),
			QueryStruct{},
			"http://gemini.com",
		},
		{
			"Nil Query Params",
			httptest.NewRequest("GET", "http://gemini.com", nil),
			nil,
			"http://gemini.com",
		},
		{
			"Multiple Values in a Single Query Param",
			httptest.NewRequest("GET", "http://gemini.com", nil),
			struct {
				Param1 []string `json:"param1"`
			}{Param1: []string{"value1", "value2"}},
			"http://gemini.com?param1=value1&param1=value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests.SetQueryParams(tt.request, tt.queryParams)
			if tt.request.URL.String() != tt.expectedURL {
				t.Errorf("SetQueryParams() URL = %v, want %v", tt.request.URL, tt.expectedURL)
			}
		})
	}
}

func TestSetJSONPayload(t *testing.T) {
	type PayloadStruct struct {
		Field1 string   `json:"field1"`
		Field2 int      `json:"field2"`
		Field3 []string `json:"field3"`
	}

	tests := []struct {
		name         string
		payload      interface{}
		wantErr      bool
		expectedBody string
	}{
		{
			"Valid Payload",
			PayloadStruct{Field1: "test", Field2: 123, Field3: []string{"one", "two"}},
			false,
			`{"field1":"test","field2":123,"field3":["one","two"]}`,
		},
		{
			"Nil Payload",
			nil,
			false,
			"",
		},
		{
			"Empty Payload",
			PayloadStruct{},
			false,
			`{"field1":"","field2":0,"field3":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://gemini.com", nil)
			err := requests.SetJSONPayload(req, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetJSONPayload() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Read body from the request
			body, _ := io.ReadAll(req.Body)
			bodyString := string(body)
			if bodyString != tt.expectedBody {
				t.Errorf("SetJSONPayload() body = %v, want %v", bodyString, tt.expectedBody)
			}
		})
	}
}

func TestSetFormURLEncodedPayload(t *testing.T) {
	type FormDataStruct struct {
		Field1 string   `url:"field1"`
		Field2 int      `url:"field2"`
		Field3 []string `url:"field3"`
	}

	tests := []struct {
		name         string
		formData     interface{}
		wantErr      bool
		expectedBody string
		expectHeader bool
	}{
		{
			"Valid Form Data",
			FormDataStruct{Field1: "test", Field2: 123, Field3: []string{"one", "two"}},
			false,
			"field1=test&field2=123&field3%5B%5D=one&field3%5B%5D=two",
			true,
		},
		{
			"Nil Form Data",
			nil,
			false,
			"",
			false,
		},
		{
			"Empty Form Data",
			FormDataStruct{},
			false,
			"",   // Expecting an empty body
			true, // Expecting Content-Type header to be set even for empty form data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://gemini.com", nil)
			err := requests.SetFormURLEncodedPayload(req, tt.formData)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetFormURLEncodedPayload() error = %v, wantErr %v", err, tt.wantErr)
			}

			body, _ := io.ReadAll(req.Body)
			bodyString := string(body)
			if bodyString != tt.expectedBody {
				t.Errorf("SetFormURLEncodedPayload() body = %v, want %v", bodyString, tt.expectedBody)
			}

			contentType := req.Header.Get("Content-Type")
			if tt.expectHeader && contentType != requests.FormURLEncoded {
				t.Errorf("SetFormURLEncodedPayload() Content-Type = %v, want %v", contentType, requests.FormURLEncoded)
			}
			if !tt.expectHeader && contentType != "" {
				t.Errorf("SetFormURLEncodedPayload() unexpected Content-Type = %v", contentType)
			}
		})
	}
}

func TestDoRequest(t *testing.T) {
	// Mock HTTP client for simulating responses
	mockClient := mockHTTPClient("mock response", http.StatusOK, nil)

	tests := []struct {
		name       string
		method     string
		url        string
		query      interface{}
		data       interface{}
		wantErr    bool
		wantBody   string
		wantStatus int
	}{
		{
			"GET Valid Request",
			"GET",
			"http://gemini.com",
			nil,
			nil,
			false,
			"mock response",
			http.StatusOK,
		},
		{
			"POST Valid Request",
			"POST",
			"http://gemini.com",
			nil,
			map[string]interface{}{"field1": "value1"},
			false,
			"mock response",
			http.StatusOK,
		},
		{
			"Invalid Method",
			"INVALID",
			"http://gemini.com",
			nil,
			nil,
			true,
			"",
			0,
		},
		{
			"Invalid URL",
			"GET",
			":",
			nil,
			nil,
			true,
			"",
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := requests.NewClient(mockClient, requests.Headers{"Content-Type": requests.JSON}, nil)
			resp, body, err := client.DoRequest(tt.method, tt.url, tt.query, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("DoRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if string(body) != tt.wantBody {
					t.Errorf("DoRequest() body = %s, want %s", string(body), tt.wantBody)
				}
				if resp.StatusCode != tt.wantStatus {
					t.Errorf("DoRequest() status code = %v, want %v", resp.StatusCode, tt.wantStatus)
				}
			}
		})
	}
}

func TestRetryLogic(t *testing.T) {
	rateLimitStatusCode := http.StatusTooManyRequests
	normalStatusCode := http.StatusOK
	rateLimitedResponse := "Rate limited"
	normalResponse := "Success"

	// Counter to track the number of requests made
	var requestCount int

	// Mock HTTP client to simulate rate-limited responses
	mockClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			requestCount++
			if requestCount <= 3 { // Simulate being rate-limited for the first three requests
				return &http.Response{
					StatusCode: rateLimitStatusCode,
					Body:       io.NopCloser(bytes.NewBufferString(rateLimitedResponse)),
					Header:     make(http.Header),
				}, nil
			}
			// Return a normal response on the fourth request
			return &http.Response{
				StatusCode: normalStatusCode,
				Body:       io.NopCloser(bytes.NewBufferString(normalResponse)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	c := requests.NewClient(mockClient, nil, ratelimit.NewRateLimiter(log.NewLogger("{requests_test}", log.TRACE), 3))

	_, body, err := c.DoRequest("GET", "http://gemini.com", nil, nil)
	if err != nil {
		t.Fatalf("DoRequest() error: %v", err)
	}

	responseBody := string(body)
	if responseBody != normalResponse {
		t.Errorf("DoRequest() expected body %s, got %s", normalResponse, responseBody)
	}

	if requestCount != 4 { // 3 retries + 1 successful request
		t.Errorf("DoRequest() expected 4 total requests, got %d", requestCount)
	}
}
