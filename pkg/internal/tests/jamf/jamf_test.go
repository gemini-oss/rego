// pkg/internal/tests/jamf/jamf_test.go
package jamf_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/jamf"
)

func setupTestServer(t *testing.T, responseMap map[string]string) (*httptest.Server, func()) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		response, ok := responseMap[r.URL.Path]
		if !ok {
			t.Errorf("No mock response for path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte(response))
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	return server, func() { server.Close() }
}

func setupTestClient(serverURL string) *jamf.Client {
	client := jamf.NewClient(log.DEBUG) // Assuming NewClient setups the client for testing
	client.BaseURL = serverURL + "/api"
	client.ClassicURL = serverURL + "/JSSResource"
	return client
}
