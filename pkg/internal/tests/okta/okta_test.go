/*
# Okta - Test

This package runs tests for functions which interact with the Okta API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/internal/tests/okta/okta_test.go
package okta_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/okta"
)

// setupTestServer returns a new test server and a cleanup function
func setupTestServer(t *testing.T, path string, response string) (*httptest.Server, func()) {
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
func setupTestClient(serverURL string) *okta.Client {
	client := okta.NewClient(log.DEBUG)

	client.BaseURL = serverURL

	return client
}
