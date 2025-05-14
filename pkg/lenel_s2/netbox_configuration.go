package lenel_s2

import (
	"fmt"
	"time"
)

func (c *Client) ListAllUDFs() (*[]*UDF, error) {
	url := c.BuildURL(NetBoxAPI)
	cache_key := fmt.Sprintf("%s_%s", url, NetboxCommands.Configuration.GetUDFLists)

	var cache UDFLists
	if c.GetCache(cache_key, &cache) {
		return &cache.UserDefinedFields, nil
	}

	payload := c.BuildRequest(NetboxCommands.Configuration.GetUDFLists, nil)

	udfl, err := do[UDFLists](c, "POST", url, payload)
	if err != nil {
		return nil, err
	}

	c.SetCache(cache_key, udfl, 5*time.Minute)
	return &udfl.UserDefinedFields, nil
}
