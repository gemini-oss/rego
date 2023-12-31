/*
# Jamf

This package initializes all the methods for functions which interact with the Jamf API:
- https://developer.jamf.com/jamf-pro/reference/classic-api
- https://developer.jamf.com/jamf-pro/reference/jamf-pro-api

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/jamf/jamf.go
package jamf

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	BaseURL                 = fmt.Sprintf("https://%s/api/v1", "%s")      // https://developer.jamf.com/jamf-pro/reference/jamf-pro-api
	ClassicURL              = fmt.Sprintf("https://%s/JSSResource", "%s") // https://developer.jamf.com/jamf-pro/reference/classic-api
	JamfDevices             = fmt.Sprintf("%s/devices", "%s")             // https://developer.jamf.com/jamf-pro/reference/jamf-pro-api/devices
	JamfManagementFramework = fmt.Sprintf("%s/users", "%s")               // https://developer.jamf.com/jamf-pro/reference/jamf-pro-api/management-framework
	JameUsers               = fmt.Sprintf("%s/iam", "%s")                 // https://developer.jamf.com/jamf-pro/reference/jamf-pro-api/users
)

// Credentials for Jamf Pro
type Credentials struct {
	Username string
	Password string
	Token    JamfToken
}

type JamfToken struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

type Client struct {
	BaseURL    string
	ClassicURL string
	HTTP       *requests.Client
	Logger     *log.Logger
}

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%s", url, id)
	}
	return url
}

/*
 * # Create a new Jamf Token based on the credentials provided
 * /api/v1/auth/token
 * - https://developer.jamf.com/jamf-pro/reference/post_v1-auth-token
 */
func GetToken(baseURL string) (*JamfToken, error) {
	url := baseURL + "/auth/token"

	// Prepare the credentials for Basic Auth
	creds := Credentials{
		Username: config.GetEnv("JSS_USERNAME", ""),
		Password: config.GetEnv("JSS_PASSWORD", ""),
	}
	basicCreds := fmt.Sprintf("%s:%s", creds.Username, creds.Password)

	headers := requests.Headers{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(basicCreds)),
		"Content-Type":  "application/json",
	}

	hc := requests.NewClient(nil, headers)
	_, body, err := hc.DoRequest("POST", url, nil, nil)
	if err != nil {
		return nil, err
	}

	token := &JamfToken{}
	err = json.Unmarshal(body, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

/*
 * Create a new Jamf Client
 */
func NewClient(verbosity int) *Client {

	url := config.GetEnv("JSS_URL", "https://yourserver.jamfcloud.com")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")

	BaseURL := fmt.Sprintf(BaseURL, url)
	ClassicURL := fmt.Sprintf(ClassicURL, url)

	token, err := GetToken(BaseURL)
	if err != nil {
		panic(err)
	}

	headers := requests.Headers{
		"Authorization":             fmt.Sprintf("Bearer %s", token.Token),
		"Accept":                    "application/json, application/xml;q=0.9",
		"Cache-Control":             "no-store, no-cache, must-revalidate, max-age=0, post-check=0, pre-check=0",
		"Strict-Transport-Security": "max-age=31536000 ; includeSubDomains",
		"Content-Type":              "application/json",
	}

	return &Client{
		BaseURL:    BaseURL,
		ClassicURL: ClassicURL,
		HTTP:       requests.NewClient(nil, headers),
		Logger:     log.NewLogger("{jamf}", verbosity),
	}
}
