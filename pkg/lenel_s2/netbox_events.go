package lenel_s2

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"time"
)

func (c *Client) GetAccessHistory(logID int) (*AccessHistory, error) {
	url := c.BuildURL(NetBoxAPI)
	cache_key := fmt.Sprintf("%s_%s_%v", url, "GetAccessHistory", logID)

	var cache AccessHistory
	if c.GetCache(cache_key, &cache) {
		return &cache, nil
	}

	payload := c.BuildRequest(
		"GetAccessHistory",
		struct {
			AfterLogID int `xml:"AFTERLOGID"`
			MaxRecords int `xml:"MAXRECORDS"`
		}{
			AfterLogID: logID,
			MaxRecords: 1000,
		},
	)

	ah, err := doPaginated[*AccessHistory](c, "POST", url, payload)
	if err != nil {
		return nil, err
	}

	c.SetCache(cache_key, ah, 5*time.Minute)
	return *ah, nil
}

func (c *Client) GetCardAccessDetails(ac *AccessCard) (*AccessHistory, error) {
	url := c.BuildURL(NetBoxAPI)
	cache_key := fmt.Sprintf("%s_%s_%s_%s", url, NetboxCommands.History.GetCardAccessDetails, ac.EncodedNum, ac.Format)

	var cache AccessHistory
	if c.GetCache(cache_key, &cache) {
		return &cache, nil
	}

	payload := c.BuildRequest(
		NetboxCommands.History.GetCardAccessDetails,
		struct {
			EncodedNum string `xml:"ENCODEDNUM"`
			CardFormat string `xml:"CARDFORMAT"`
		}{
			EncodedNum: ac.EncodedNum,
			CardFormat: ac.Format,
		},
	)

	cah, err := doPaginated[*AccessHistory](c, "POST", url, payload)
	if err != nil {
		return nil, err
	}

	c.SetCache(cache_key, cah, 5*time.Minute)
	return *cah, nil
}

func (c *Client) GetEventHistory(logID int) (any, error) {
	url := c.BuildURL(NetBoxAPI)
	cache_key := fmt.Sprintf("%s_%s_%v", url, NetboxCommands.History.GetEventHistory, logID)

	var cache AccessHistory
	if c.GetCache(cache_key, &cache) {
		return &cache, nil
	}

	payload := c.BuildRequest(
		NetboxCommands.History.GetEventHistory,
		struct {
			STARTDTTM string `xml:"STARTDTTM"`
		}{
			STARTDTTM: "2016-10-01 01:01:01",
		},
	)

	ah, err := doPaginated[*AccessHistory](c, "POST", url, payload)
	if err != nil {
		return nil, err
	}

	c.SetCache(cache_key, ah, 5*time.Minute)
	return *ah, nil
}

// StreamEvents captures all events with optional context and SIEM forwarding
func (c *Client) StreamEvents(opts ...StreamOption) ([]Event, error) {
	// Default configuration
	cfg := &streamConfig{
		ctx: context.Background(),
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	url := c.BuildURL(NetBoxAPI)
	c.Log.Println("Starting StreamEvents to URL:", url)

	c.HTTP.Headers["Connection"] = "keep-alive"
	defer delete(c.HTTP.Headers, "Connection")

	payload := c.BuildRequest(
		NetboxCommands.Events.StreamEvents,
		cfg.streamParams,
	)

	// Log the request payload
	payloadXML, _ := xml.Marshal(payload)
	c.Log.Debug("Stream request payload:", string(payloadXML))

	// Process each event as it arrives
	processFunc := func(event Event) bool {
		c.Log.Println("Event received:", event.DescName, "at", event.CDT)
		c.Log.Debug("Event details:", fmt.Sprintf("%+v", event))

		// Forward to SIEM if forwarder provided
		if cfg.siemForwarder != nil {
			c.Log.Debug("Forwarding event to SIEM")
			if err := cfg.siemForwarder(event); err != nil {
				c.Log.Error("Failed to forward event to SIEM:", err)
				// Continue processing other events even if one fails
			}
		}

		return true // Continue streaming
	}

	// doStream always returns collected events, even on context cancellation
	events, err := doStream(cfg.ctx, c, "POST", url, payload, processFunc, cfg.heartbeat)

	// Log context cancellation but still return collected events
	if errors.Is(err, context.Canceled) {
		c.Log.Println("StreamEvents cancelled by context, returning", len(events), "collected events")
	}

	return events, err
}
