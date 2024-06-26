/*
# Slack

This package initializes all the methods for functions which interact with the Slack Web API:
https://api.slack.com/web

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/slack.go
package slack

import (
	"fmt"

	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

const (
	BaseURL = "https://slack.com/api" // https://slack.com/api/METHOD_FAMILY.method?pretty=1
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...interface{}) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%v", url, id)
	}
	return url
}

/*
  - # Generate Slack Client
  - @param log *log.Logger
  - @return *Client
  - Example:

```go

	startServer1 := flag.Bool("startServer1", false, "Start server 1")
	flag.Parse()

	s := slack.NewClient(log.DEBUG)

	if *startServer1 {
		handlers := server.Handler{
			"/slack/events":           s.EventHandler,
			"/slack/command/userlist": s.CommandHandler,
		}
		go server.StartServer("127.0.0.1:8080", handlers)
	}

```
*/
func NewClient(verbosity int) *Client {
	log := log.NewLogger("{slack}", verbosity)

	token := config.GetEnv("SLACK_API_TOKEN")
	if len(token) == 0 {
		log.Fatal("SLACK_API_TOKEN is not set.")
	}

	signingSecret := config.GetEnv("SLACK_SIGNING_SECRET")
	if len(signingSecret) == 0 {
		log.Fatal("SLACK_SIGNING_SECRET is not set.")
	}

	headers := requests.Headers{
		"Authorization": "Bearer " + token,
		"Accept":        requests.JSON,
		"Content-Type":  fmt.Sprintf("%s; charset=utf-8", requests.JSON),
	}
	httpClient := requests.NewClient(nil, headers, nil)
	httpClient.BodyType = requests.JSON

	return &Client{
		BaseURL:       BaseURL,
		HTTP:          httpClient,
		Log:           log,
		Token:         token,
		SigningSecret: signingSecret,
	}
}
