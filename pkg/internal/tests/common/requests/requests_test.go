// pkg/internal/tests/common/requests/requests_test.go
package requests_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func setupMockServer(t *testing.T, method string, status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method && method != "INVALID" {
			t.Errorf("Unexpected method: got %v, want %v", r.Method, method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
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
	tests := []struct {
		name           string
		method         string
		responseStatus int
		responseBody   string
		query          interface{}
		data           interface{}
		wantErr        bool
	}{
		{
			name:           "GET Valid Request",
			method:         "GET",
			responseStatus: http.StatusOK,
			responseBody:   "mock response",
			wantErr:        false,
		},
		{
			name:           "POST Invalid Request",
			method:         "POST",
			responseStatus: http.StatusBadRequest,
			responseBody:   "Bad Request",
			data:           map[string]interface{}{"field1": "value1"},
			wantErr:        true,
		},
		{
			name:           "Invalid Method",
			method:         "INVALID",
			responseStatus: http.StatusOK,
			responseBody:   "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := setupMockServer(t, tt.method, tt.responseStatus, tt.responseBody)
			defer mockServer.Close()

			client := requests.NewClient(mockServer.Client(), requests.Headers{"Content-Type": requests.JSON}, nil)

			ctx := context.Background()
			resp, body, err := client.DoRequest(ctx, tt.method, mockServer.URL, tt.query, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("DoRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if string(body) != tt.responseBody {
					t.Errorf("DoRequest() body = %s, want %s", string(body), tt.responseBody)
				}
				if resp.StatusCode != tt.responseStatus {
					t.Errorf("DoRequest() status code = %v, want %v", resp.StatusCode, tt.responseStatus)
				}
			}
		})
	}
}

func TestRetryLogic(t *testing.T) {
	tests := []struct {
		name                string
		responseSequence    []int
		expectedRequests    int
		expectedFinalStatus int
		expectedError       bool
		timeout             time.Duration
	}{
		{
			name:                "Successful after retries",
			responseSequence:    []int{429, 429, 429, 200},
			expectedRequests:    4,
			expectedFinalStatus: 200,
			expectedError:       false,
			timeout:             5 * time.Second,
		},
		{
			name:                "Timeout before success",
			responseSequence:    []int{429, 429, 429, 200},
			expectedRequests:    2,
			expectedFinalStatus: 429,
			expectedError:       true,
			timeout:             1 * time.Second,
		},
		{
			name:                "Non-retryable error",
			responseSequence:    []int{400},
			expectedRequests:    1,
			expectedFinalStatus: 400,
			expectedError:       true,
			timeout:             5 * time.Second,
		},
		{
			name:                "Max retries reached",
			responseSequence:    []int{429, 429, 429, 429, 429, 429},
			expectedRequests:    5,
			expectedFinalStatus: 429,
			expectedError:       true,
			timeout:             10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestCount int
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requestCount >= len(tt.responseSequence) {
					t.Errorf("Unexpected request: count=%d", requestCount)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				status := tt.responseSequence[requestCount]
				requestCount++
				w.WriteHeader(status)
				w.Write([]byte(fmt.Sprintf("Response %d", requestCount)))
				time.Sleep(200 * time.Millisecond) // Add a slight delay to each response
			}))
			defer mockServer.Close()

			rateLimiter := ratelimit.NewRateLimiter(100, 1*time.Minute)
			client := requests.NewClient(mockServer.Client(), nil, rateLimiter)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			resp, _, err := client.DoRequest(ctx, "GET", mockServer.URL, nil, nil)

			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			} else if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if resp != nil && resp.StatusCode != tt.expectedFinalStatus {
				t.Errorf("Expected final status %d, got %d", tt.expectedFinalStatus, resp.StatusCode)
			}

			if requestCount != tt.expectedRequests {
				t.Errorf("Expected %d total requests, got %d", tt.expectedRequests, requestCount)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectedError  string
	}{
		{"JSON Error", http.StatusBadRequest, `{"error":"invalid input"}`, `{"error":"invalid input"}`},
		{"Plain Text Error", http.StatusInternalServerError, "Internal Server Error", "Internal Server Error"},
		{"Unexpected Error", http.StatusServiceUnavailable, "", "Unexpected error (Status: 503)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer mockServer.Close()

			client := requests.NewClient(mockServer.Client(), requests.Headers{"Content-Type": requests.JSON}, nil)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, _, err := client.DoRequest(ctx, "GET", mockServer.URL, nil, nil)

			if err == nil {
				t.Errorf("Expected an error, but got none")
				return
			}

			reqErr, ok := err.(*requests.RequestError)
			if !ok {
				t.Errorf("Expected error of type *requests.RequestError, got %T", err)
				return
			}

			if reqErr.StatusCode != tt.responseStatus {
				t.Errorf("Expected status code %d, got %d", tt.responseStatus, reqErr.StatusCode)
			}

			if reqErr.Message != tt.expectedError {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, reqErr.Message)
			}
		})
	}
}

func TestStatusCodeChecks(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		checks     map[string]bool
	}{
		{
			"Redirect",
			http.StatusFound,
			map[string]bool{
				"IsRedirectCode":        true,
				"IsRetryableStatusCode": false,
				"IsNonRetryableCode":    false,
				"IsTemporaryErrorCode":  false,
			},
		},
		{
			"Retryable",
			http.StatusTooManyRequests,
			map[string]bool{
				"IsRedirectCode":        false,
				"IsRetryableStatusCode": true,
				"IsNonRetryableCode":    false,
				"IsTemporaryErrorCode":  false,
			},
		},
		{
			"Non-Retryable",
			http.StatusBadRequest,
			map[string]bool{
				"IsRedirectCode":        false,
				"IsRetryableStatusCode": false,
				"IsNonRetryableCode":    true,
				"IsTemporaryErrorCode":  false,
			},
		},
		{
			"Temporary Error",
			http.StatusInternalServerError,
			map[string]bool{
				"IsRedirectCode":        false,
				"IsRetryableStatusCode": true,
				"IsNonRetryableCode":    false,
				"IsTemporaryErrorCode":  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requests.IsRedirectCode(tt.statusCode); got != tt.checks["IsRedirectCode"] {
				t.Errorf("IsRedirectCode() = %v, want %v", got, tt.checks["IsRedirectCode"])
			}
			if got := requests.IsRetryableStatusCode(tt.statusCode); got != tt.checks["IsRetryableStatusCode"] {
				t.Errorf("IsRetryableStatusCode() = %v, want %v", got, tt.checks["IsRetryableStatusCode"])
			}
			if got := requests.IsNonRetryableCode(tt.statusCode); got != tt.checks["IsNonRetryableCode"] {
				t.Errorf("IsNonRetryableCode() = %v, want %v", got, tt.checks["IsNonRetryableCode"])
			}
			if got := requests.IsTemporaryErrorCode(tt.statusCode); got != tt.checks["IsTemporaryErrorCode"] {
				t.Errorf("IsTemporaryErrorCode() = %v, want %v", got, tt.checks["IsTemporaryErrorCode"])
			}
		})
	}
}

func TestClientMethods(t *testing.T) {
	client := requests.NewClient(nil, requests.Headers{"Content-Type": requests.JSON}, nil)

	t.Run("UpdateContentType", func(t *testing.T) {
		client.UpdateContentType(requests.FormURLEncoded)
		if client.Headers["Content-Type"] != requests.FormURLEncoded {
			t.Errorf("Expected Content-Type to be %s, got %s", requests.FormURLEncoded, client.Headers["Content-Type"])
		}
	})

	t.Run("UpdateBodyType", func(t *testing.T) {
		client.UpdateBodyType(requests.XML)
		if client.BodyType != requests.XML {
			t.Errorf("Expected BodyType to be %s, got %s", requests.XML, client.BodyType)
		}
	})

	t.Run("ExtractParam", func(t *testing.T) {
		url := "http://gemini.com?param1=value1&param2=value2"
		param := client.ExtractParam(url, "param1")
		if param != "value1" {
			t.Errorf("Expected extracted param to be 'value1', got '%s'", param)
		}
	})
}
