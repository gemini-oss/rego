/*
# Google - Test

This package runs tests for functions which interact with the Google Workspace API:
https://developers.google.com/workspace

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/google/google_test.go
package google_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/google"
)

// setupTestServer returns a new test server and a cleanup function
func SetupTestServer(t *testing.T, path string, response string) (*httptest.Server, func()) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != path {
			t.Errorf("Expected path `%s`, got `%s`", path, r.URL.String())
		}
		w.Write([]byte(response))
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	return server, func() { server.Close() }
}

// setupTestClient returns a new Okta client with test server URL
func SetupTestClient(serverURL string) *google.Client {
	client, _ := google.NewClient(
		google.AuthCredentials{
			Type:   google.SERVICE_ACCOUNT,
			CICD:   true,
			Scopes: []string{"https://www.googleapis.com/auth/userinfo.email"},
		},
		log.DEBUG,
	)

	client.BaseURL = serverURL

	return client
}

func TestNewClient(t *testing.T) {
	ac := google.AuthCredentials{
		Type:   google.SERVICE_ACCOUNT,
		CICD:   true,
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email"},
	}

	c, err := google.NewClient(ac, log.DEBUG)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if c == nil {
		t.Fatalf("Expected client to be created, got nil")
	}

	if c.BaseURL != google.BaseURL {
		t.Fatalf("Expected baseURL to be %v, got %v", google.BaseURL, c.BaseURL)
	}

	if len(c.Auth.Scopes) != 1 {
		t.Fatalf("Expected one scope, got %v", len(c.Auth.Scopes))
	}

	if c.Auth.Scopes[0] != "https://www.googleapis.com/auth/userinfo.email" {
		t.Fatalf("Expected scope to be 'https://www.googleapis.com/auth/userinfo.email', got %v", c.Auth.Scopes[0])
	}
}
