/*
# Lenel S2

This package initializes all the methods for functions which interact with the Lenel S2 API:
https://developer.okta.com/docs/api/

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/lenel_s2/lenel_s2.go
package lenel_s2

import (
	"bufio"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

const (
	NetBoxAPI = "%s/nbws/goforms/nbapi"
)

var (
	BaseURL = fmt.Sprintf("http://%s", "%s") // https://
)

// BuildURL builds a URL for a given resource and identifiers.
func (c *Client) BuildURL(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseURL)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s/%s", url, id)
	}
	return url
}

// BuildRequest creates a consistent XML request struct.
func (c *Client) BuildRequest(name string, params interface{}) NetboxCommand {
	return NetboxCommand{
		SessionID: c.Session.ID,
		Command: Command{
			Name:       name,
			Num:        "1",
			DateFormat: "tzoffset",
			Params:     params,
		},
	}
}

// UseCache() enables caching for the next method call.
func (c *Client) UseCache() *Client {
	c.Cache.Enabled = true
	return c
}

/*
 * SetCache stores an S2 API response in the cache
 */
func (c *Client) SetCache(key string, value interface{}, duration time.Duration) {
	// Convert value to a byte slice and cache it
	data, err := json.Marshal(value)
	if err != nil {
		c.Log.Error("Error marshalling cache data:", err)
		return
	}
	c.Cache.Set(key, data, duration)
}

/*
 * GetCache retrieves an S2 API response from the cache
 */
func (c *Client) GetCache(key string, target interface{}) bool {
	data, found := c.Cache.Get(key)
	if !found {
		return false
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		c.Log.Error("Error unmarshalling cache data:", err)
		return false
	}
	return true
}

/*
 * # Create a new S2 SessionID based on the credentials provided
 * <COMMAND name="Login" num="1" dateformat="tzoffset">
 */
func Login(baseURL string) (*NetboxResponse[any], error) {
	url := fmt.Sprintf(NetBoxAPI, baseURL)

	// Prepare the credentials for Basic Auth
	creds := Credentials{
		Username: config.GetEnv("S2_USERNAME"),
		Password: config.GetEnv("S2_PASSWORD"),
	}
	if len(creds.Username) == 0 {
		return nil, fmt.Errorf("S2_USERNAME environment variable is not set")
	}
	if len(creds.Password) == 0 {
		return nil, fmt.Errorf("S2_PASSWORD environment variable is not set")
	}

	headers := requests.Headers{
		"Content-Type": requests.XML,
	}

	payload := NetboxCommand{
		Command: Command{
			Name:       "Login",
			Num:        "1",
			DateFormat: "tzoffset",
			Params: struct {
				Username string `xml:"USERNAME"`
				Password string `xml:"PASSWORD"`
			}{
				Username: creds.Username,
				Password: creds.Password,
			},
		},
	}

	hc := requests.NewClient(nil, headers, nil)
	hc.BodyType = requests.XML
	_, body, err := hc.DoRequest(context.Background(), "POST", url, nil, payload)
	if err != nil {
		return nil, err
	}

	session := &NetboxResponse[any]{}
	err = xml.Unmarshal(body, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Logout ends the current session
func (c *Client) Logout() error {
	payload := NetboxCommand{
		SessionID: c.Session.ID,
		Command: Command{
			Name:       "Logout",
			Num:        "1",
			DateFormat: "tzoffset",
		},
	}
	_, err := do[struct{}](c, "POST", "", payload)
	return err
}

/*
  - # Generate S2 Client
  - @param logger *log.Logger
  - @return *Client
  - Example:

```go

	s := s2.NewClient(log.DEBUG)

```
*/
func NewClient(baseURL string, verbosity int) *Client {
	log := log.NewLogger("{lenel_s2}", verbosity)

	url := config.GetEnv("S2_URL")
	if len(url) == 0 {
		log.Fatal("S2_URL is not set")
	}
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")
	url = fmt.Sprintf(BaseURL, url)

	session, err := Login(url)
	if err != nil {
		log.Fatalf("Failed to login to S2: %v", err)
	}
	if session == nil {
		log.Fatal("Received nil S2 session response")
	}
	if session.ID == "" {
		log.Fatal("Failed to retrieve session ID")
	}

	headers := requests.Headers{
		//"Cookie":       fmt.Sprintf(".sessionId=%s", session.ID),
		"Content-Type": requests.XML,
	}

	httpClient := requests.NewClient(nil, headers, nil)
	httpClient.BodyType = requests.XML

	// Look into `Functional Options` patterns for a better way to handle this (and other clients while we're at it)
	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_lenel_s2.gob", 1000000)
	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		BaseURL: url,
		Session: session,
		HTTP:    httpClient,
		Log:     log,
		Cache:   cache,
	}
}

// PaginatedResponse defines methods needed for handling pagination.
type PaginatedResponse interface {
	// NextToken returns the token to be used for the next request.
	NextToken() string
	// Append merges a new page of results with the existing result set.
	Append(resp PaginatedResponse) PaginatedResponse

	SetCommand(nb *Client, cmd *NetboxCommand) NetboxCommand
}

func do[T any](c *Client, method string, url string, payload NetboxCommand) (T, error) {
	var result NetboxResponse[T]
	var empty T

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, body, err := c.HTTP.DoRequest(ctx, method, url, nil, payload)
	if err != nil {
		return empty, err
	}

	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	if err := xml.Unmarshal(body, &result); err != nil {
		return empty, fmt.Errorf("XML unmarshalling error: %w", err)
	}

	if result.Response.Code != "SUCCESS" {
		return empty, fmt.Errorf("error: %s", result.Response.Error)
	}

	return *result.Response.Details, nil
}

// doPaginated runs a paginated API request that returns a type implementing PaginatedResponse.
func doPaginated[T PaginatedResponse](c *Client, method string, url string, payload NetboxCommand) (*T, error) {
	var results T

	pageToken := ""
	cmd := &payload

	for {
		r, err := do[T](c, method, url, *cmd)
		if err != nil {
			return nil, err
		}

		// Merge the current page with previous pages.
		results = results.Append(r).(T)

		// Retrieve the token for the next page.
		pageToken = r.NextToken()
		if pageToken == "" {
			break
		}

		// Set the page token into the query for the next request.
		*cmd = r.SetCommand(c, cmd)
	}

	return &results, nil
}

// StreamResult holds the collected data and any error from a streaming operation
type StreamResult[T any] struct {
	Data  []T
	Error error
}

// StreamOption is a functional option for configuring streaming
type StreamOption func(*streamConfig)

type streamConfig struct {
	ctx           context.Context
	siemForwarder func(Event) error
	heartbeat     func() // Called when heartbeat is received
	streamParams  any    // Parameters built by the builder
}

// WithContext sets a custom context for the stream
func WithContext(ctx context.Context) StreamOption {
	return func(cfg *streamConfig) {
		cfg.ctx = ctx
	}
}

// WithEventStreamBuilder uses a builder to configure the stream parameters
func WithEventParameters(params *StreamEventsParams) StreamOption {
	return func(cfg *streamConfig) {
		cfg.streamParams = params
	}
}

// WithSIEMForwarder sets a function to forward events to a SIEM
func WithSIEMForwarder(forwarder func(Event) error) StreamOption {
	return func(cfg *streamConfig) {
		cfg.siemForwarder = forwarder
	}
}

// WithHeartbeat sets a function to be called when a heartbeat is received
func WithHeartbeat(heartbeat func()) StreamOption {
	return func(cfg *streamConfig) {
		cfg.heartbeat = heartbeat
	}
}

// doStream processes streaming API requests and collects all data
func doStream[T Event](ctx context.Context, c *Client, method string, url string, payload NetboxCommand, processFunc func(T) bool, heartbeat func()) ([]T, error) {
	// Channel for collecting results
	resultChan := make(chan StreamResult[T], 1)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer signal.Stop(sigChan) // Stop receiving signals when function exits

	// Create a done channel to clean up the goroutine
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-sigChan:
			c.Log.Print("Interrupt received in StreamEvents, stopping...")
			cancel()
		case <-done:
			// Clean exit when the function returns
			return
		}
	}()

	// Run the stream processing in a goroutine
	go func() {
		var collected []T
		c.Log.Debug("Starting stream processing")

		streamErr := c.HTTP.DoStream(ctx, method, url, nil, payload, func(decoder interface{}) error {
			// For S2 multipart streams, we should get a raw reader
			reader, ok := decoder.(io.Reader)
			if !ok {
				return fmt.Errorf("expected io.Reader for multipart stream, got %T", decoder)
			}

			c.Log.Debug("Got io.Reader, creating scanner")

			// Create a scanner to read the multipart stream line by line
			scanner := bufio.NewScanner(reader)
			var xmlBuffer strings.Builder
			inXML := false
			lineCount := 0

			for scanner.Scan() {
				line := scanner.Text()
				lineCount++

				// Log raw line in trace mode
				c.Log.Trace(fmt.Sprintf("Line %d: %s", lineCount, line))

				// Check if this line has XML content followed by boundary
				if strings.Contains(line, "<NETBOX>") && strings.Contains(line, "--Boundary") {
					// Extract XML part before the boundary
					xmlPart := strings.Split(line, "--Boundary")[0]
					c.Log.Debug("XML:", xmlPart)

					if err := parseStreamEvent(xmlPart, &collected, processFunc, c, heartbeat); err != nil {
						if err == io.EOF {
							c.Log.Println("Stream stopped by processFunc")
							return nil
						}
						return err
					}
					continue
				}

				// Skip pure boundary lines
				if line == "--Boundary" {
					c.Log.Debug("Skipping pure boundary line")
					continue
				}

				// Skip headers
				if strings.HasPrefix(line, "Content-Type:") || strings.HasPrefix(line, "Content-Length:") {
					c.Log.Debug("Skipping header line:", line)
					continue
				}

				// Detect start of XML (without boundary)
				if strings.Contains(line, "<NETBOX>") {
					c.Log.Debug("Found start of XML at line", lineCount)
					xmlBuffer.Reset()
					xmlBuffer.WriteString(line)
					xmlBuffer.WriteString("\n")
					inXML = true

					// Check if it's a single-line response
					if strings.Contains(line, "</NETBOX>") {
						inXML = false
						xmlData := xmlBuffer.String()
						c.Log.Debug("Found complete single-line XML:", xmlData)

						if err := parseStreamEvent(xmlData, &collected, processFunc, c, heartbeat); err != nil {
							if err == io.EOF {
								c.Log.Println("Stream stopped by processFunc")
								return nil
							}
							return err
						}
					}
				} else if inXML {
					xmlBuffer.WriteString(line)
					xmlBuffer.WriteString("\n")

					// Check for end of XML
					if strings.Contains(line, "</NETBOX>") {
						inXML = false
						xmlData := xmlBuffer.String()
						c.Log.Debug("Found complete multi-line XML:", xmlData)

						if err := parseStreamEvent(xmlData, &collected, processFunc, c, heartbeat); err != nil {
							if err == io.EOF {
								c.Log.Println("Stream stopped by processFunc")
								return nil
							}
							return err
						}
					}
				} else if line != "" {
					// Log any non-empty lines we're not processing
					c.Log.Trace("Skipping non-XML line:", line)
				}

				// Check for context cancellation
				select {
				case <-ctx.Done():
					c.Log.Println("Stream cancelled by context")
					return ctx.Err()
				default:
				}
			}

			c.Log.Debug("Scanner finished after", lineCount, "lines")

			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading stream: %w", err)
			}

			return nil
		})

		// Send the result
		resultChan <- StreamResult[T]{
			Data:  collected,
			Error: streamErr,
		}
	}()

	// Wait for the result
	result := <-resultChan

	// Handle results
	if result.Error != nil && result.Error != context.Canceled {
		c.Log.Warningf("Stream error: %v\n", result.Error)
	}
	c.Log.Println("Stream completed. Collected items:", len(result.Data))

	// Always return collected data, even if there was an error
	return result.Data, result.Error
}

// parseStreamEvent handles a single XML event response from the stream
func parseStreamEvent[T any](xmlData string, collected *[]T, processFunc func(T) bool, c *Client, heartbeat func()) error {
	c.Log.Debug("Parsing XML event response of length:", len(xmlData))

	var result NetboxEventResponse[T]
	if err := xml.Unmarshal([]byte(xmlData), &result); err != nil {
		c.Log.Debug("Failed to parse XML:", xmlData)
		c.Log.Debug("Parse error:", err)
		return nil // Skip malformed chunks
	}

	// Log parsed structure
	c.Log.Debug(fmt.Sprintf("Parsed event response - Command: %s, Code: %s, APIError: %d, HasEvent: %v",
		result.Response.Command,
		result.Response.Code,
		result.Response.APIError,
		result.Response.Event != nil))

	// Skip heartbeat messages
	// Heartbeats have no CODE and no EVENT (just empty RESPONSE tag with command attribute)
	if result.Response.Code == "" && result.Response.Event == nil && result.Response.APIError == 0 {
		c.Log.Debug("Event heartbeat received (empty response)")
		if heartbeat != nil {
			heartbeat()
		}
		return nil
	}

	// Check for API errors
	if result.Response.APIError != 0 {
		return fmt.Errorf("API error %d: %s", result.Response.APIError, result.Response.Error)
	}

	// Check for command failures
	if result.Response.Code == "FAIL" {
		if result.Response.Error != "" {
			return fmt.Errorf("command failed: %s", result.Response.Error)
		}
		return fmt.Errorf("command failed with unknown error")
	}

	// Process successful event responses
	if result.Response.Event != nil {
		item := *result.Response.Event
		*collected = append(*collected, item)
		c.Log.Println(fmt.Sprintf("Collected event #%d", len(*collected)))

		// Call the process function for real-time handling
		if processFunc != nil && !processFunc(item) {
			c.Log.Println("Stream stopped by processFunc")
			return io.EOF // Use EOF to signal intentional stop
		}
	} else if result.Response.Code == "SUCCESS" && result.Response.Event == nil {
		// Success but no event - could be initial connection confirmation
		c.Log.Debug("SUCCESS response with no event")
	} else {
		// Log why we're not collecting this response
		c.Log.Debug("Not collecting response - Code:", result.Response.Code, "HasEvent:", result.Response.Event != nil)
	}

	return nil
}
