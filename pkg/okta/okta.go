/*
# Okta

This package initializes all the methods for functions which interact with the Okta API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/okta/okta.go
package okta

import (
	"fmt"
	"strings"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	BaseURL = fmt.Sprintf("https://%s.%s.com/api/v1", "%s", "%s") // https://developer.okta.com/docs/api/#versioning
)

const (
	OktaApps       = "%s/apps"         // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/
	OktaGroups     = "%s/groups"       // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/
	OktaGroupRules = "%s/groups/rules" // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/GroupRule/
	OktaDevices    = "%s/devices"      // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Device/
	OktaUsers      = "%s/users"        // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/
	OktaIAM        = "%s/iam"          // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleAssignment/
	OktaRoles      = "%s/iam/roles"    // https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Role/
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%s", url, id)
	}
	return url
}

// UseCache() enables caching for the next method call.
func (c *Client) UseCache() *Client {
	c.Cache.Enabled = true
	return c
}

/*
  - # Generate Okta Client
  - @param logger *log.Logger
  - @return *Client
  - Example:

```go

	o := okta.NewClient(log.DEBUG)

```
*/
func NewClient(verbosity int) *Client {

	org_name := config.GetEnv("OKTA_ORG_NAME", "yourOktaDomain")
	//org_name := config.GetEnv("OKTA_SANDBOX_ORG_NAME", "yourOktaDomain")
	org_name = strings.TrimPrefix(org_name, "https://")
	org_name = strings.TrimPrefix(org_name, "http://")
	org_name = strings.TrimSuffix(org_name, ".okta.com")

	base := config.GetEnv("OKTA_BASE_URL", "okta.com")
	//base := config.GetEnv("OKTA_SANDBOX_BASE_URL", "oktapreview.com")
	base = strings.Trim(base, "./")
	base = strings.TrimSuffix(base, ".com")

	token := config.GetEnv("OKTA_API_TOKEN", "oktaApiKey")
	//token := config.GetEnv("OKTA_SANDBOX_API_TOKEN", "oktaApiKey")
	BaseURL := fmt.Sprintf(BaseURL, org_name, base)

	headers := requests.Headers{
		"Authorization": "SSWS " + token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	// Look into `Functional Options` patterns for a better way to handle this (and othe clients while we're at it)
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY", "32-byte-long-encryption-key-1234"))
	cache, err := cache.NewCache(encryptionKey, "/tmp/rego_cache_okta.gob")
	if err != nil {
		panic(err)
	}

	// https://developer.okta.com/docs/reference/rl-best-practices/
	rl := ratelimit.NewRateLimiter()
	rl.UsesReset = true

	return &Client{
		BaseURL: BaseURL,
		HTTP:    requests.NewClient(nil, headers, rl),
		Log:     log.NewLogger("{okta}", verbosity),
		Cache:   cache,
	}
}
