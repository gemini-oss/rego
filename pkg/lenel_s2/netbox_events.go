package lenel_s2

import (
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
		}{
			AfterLogID: logID,
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

func (c *Client) StreamEvents() (any, error) {
	url := c.BuildURL(NetBoxAPI)
	cache_key := fmt.Sprintf("%s_%s", url, NetboxCommands.Events.StreamEvents)

	var cache Events
	if c.GetCache(cache_key, &cache) {
		return &cache, nil
	}

	payload := c.BuildRequest(
		NetboxCommands.Events.StreamEvents,
		nil,
	)

	events, err := do[Events](c, "POST", url, payload)
	if err != nil {
		return nil, err
	}

	c.SetCache(cache_key, events, 5*time.Minute)
	return &events, nil
}
