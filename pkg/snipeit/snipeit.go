/*
# SnipeIT

This package initializes all the methods for functions which interact with the SnipeIT API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/snipeit/snipeit.go
package snipeit

import (
	"fmt"
	"strings"

	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	BaseURL = fmt.Sprintf("https://%s/api/v1", "%s") // https://snipe-it.readme.io/reference/api-overview

)

const (
	Assets           = "%s/hardware"             // https://snipe-it.readme.io/reference#hardware
	Fields           = "%s/fields"               // https://snipe-it.readme.io/reference/fields-1
	FieldSets        = "%s/fieldsets"            // https://snipe-it.readme.io/reference/fieldsets
	Companies        = "%s/companies"            // https://snipe-it.readme.io/reference#companies
	Locations        = "%s/locations"            // https://snipe-it.readme.io/reference#locations
	Accessories      = "%s/accessories"          // https://snipe-it.readme.io/reference#accessories
	Consumables      = "%s/consumables"          // https://snipe-it.readme.io/reference#consumables
	Components       = "%s/components"           // https://snipe-it.readme.io/reference#components
	Users            = "%s/users"                // https://snipe-it.readme.io/reference#users
	StatusLabels     = "%s/statuslabels"         // https://snipe-it.readme.io/reference#status-labels
	Models           = "%s/models"               // https://snipe-it.readme.io/reference#models
	Licenses         = "%s/licenses"             // https://snipe-it.readme.io/reference#licenses
	Categories       = "%s/categories"           // https://snipe-it.readme.io/reference#categories
	Manufacturers    = "%s/manufacturers"        // https://snipe-it.readme.io/reference#manufacturers
	Suppliers        = "%s/suppliers"            // https://snipe-it.readme.io/reference#suppliers
	AssetMaintenance = "%s/hardware/maintenance" // https://snipe-it.readme.io/reference#maintenances
	Departments      = "%s/departments"          // https://snipe-it.readme.io/reference#departments
	Groups           = "%s/groups"               // https://snipe-it.readme.io/reference#groups
	Settings         = "%s/settings"             // https://snipe-it.readme.io/reference#settings
	Reports          = "%s/reports"              // https://snipe-it.readme.io/reference#reports
)

type Client struct {
	BaseURL    string
	HTTPClient *requests.Client
	Logger     *log.Logger
}

func NewClient(verbosity int) *Client {

	url := config.GetEnv("SNIPEIT_URL", "snipeit_url")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.Trim(url, "./")

	BaseURL = fmt.Sprintf(BaseURL, url)
	token := config.GetEnv("SNIPEIT_TOKEN", "snipeit_token")

	headers := requests.Headers{
		"Authorization": "Bearer " + token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	log := log.NewLogger("{snipeit}", verbosity)

	// https://snipe-it.readme.io/reference/api-throttling
	rl := ratelimit.NewRateLimiter(120)

	return &Client{
		BaseURL:    BaseURL,
		HTTPClient: requests.NewClient(nil, headers, rl),
		Logger:     log,
	}
}
