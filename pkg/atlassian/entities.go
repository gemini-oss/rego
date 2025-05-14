/*
# Atlassian - Entities [Structs]

This package contains many structs for handling responses from the Atlassian APIs:

:Copyright: (c) 2025 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/atlassian/entities.go
package atlassian

import (
	"strings"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ### Atlassian Client Entities
// ---------------------------------------------------------------------
type Client struct {
	BaseURL string           // BaseURL is the base URL for Atlassian API requests.
	CloudID string           // CloudID is the Atlassian Cloud ID.
	HTTP    *requests.Client // HTTPClient is the client used to make HTTP requests.
	Error   *Error           // Error is the error response from the last request made by the client.
	Log     *log.Logger      // Log is the logger used to log messages.
	Cache   *cache.Cache     // Cache is the cache used to store responses from the Atlassian API.
}

type Error struct {
}

type ErrorCause struct {
}

type Links struct {
}

type Link struct {
}

type Hints struct {
}

/*
 * AtlassianPage
 * @param Self string
 * @param NextPage string
 * @param Paged bool
 */
type AtlassianPage struct {
	Self          string   `json:"self"`
	NextPageLink  string   `json:"next"`
	NextPageToken string   `json:"next_page_token"`
	Paged         bool     `json:"paged"`
	Links         []string `json:"links"`
}

func (p *AtlassianPage) HasNextPage(links []string) bool {
	for _, link := range links {
		rawLink := strings.Split(link, ";")[0]
		rawLink = strings.Trim(rawLink, "<>")

		if strings.Contains(link, `rel="self"`) {
			p.Self = rawLink
		}
		if strings.Contains(link, `rel="next"`) {
			p.NextPageLink = rawLink
			p.Paged = true
			return true
		}
	}
	return false
}

func (p *AtlassianPage) NextPage(links []string) string {
	if p.HasNextPage(links) {
		return p.NextPageLink
	}
	return ""
}

// END OF ATLASSIAN CLIENT ENTITIES
// ---------------------------------------------------------------------

// ### Atlassian Client Configuration Options
// ---------------------------------------------------------------------

type clientConfig struct {
	useSandbox bool
	orgID      string
	baseURL    string
	token      string
}

type Option func(*clientConfig)

// WithSandbox sets the client to use sandbox credentials
func WithSandbox() Option {
	return func(cfg *clientConfig) {
		cfg.useSandbox = true
	}
}

// WithCustomOrgID overrides the default environment variable key for Org Name
func WithCustomOrgID(key string) Option {
	return func(cfg *clientConfig) {
		cfg.orgID = key
	}
}

// WithCustomBase overrides the default environment variable key for Base URL
func WithCustomBaseURL(key string) Option {
	return func(cfg *clientConfig) {
		cfg.baseURL = key
	}
}

// WithCustomToken overrides the default environment variable key for API Token
func WithCustomToken(key string) Option {
	return func(cfg *clientConfig) {
		cfg.token = key
	}
}

// END OF ATLASSIAN CLIENT CONFIGURATION OPTIONS
// ---------------------------------------------------------------------
