package lenel_s2

import (
	"fmt"
	"time"
)

func (c *Client) ListAllUsers() (*[]*Person, error) {
	url := c.BuildURL(NetBoxAPI)
	cache_key := fmt.Sprintf("%s_%s", url, NetboxCommands.People.SearchPersonData)

	var cache People
	if c.GetCache(cache_key, &cache) {
		return cache.People, nil
	}

	payload := c.BuildRequest(
		NetboxCommands.People.SearchPersonData,
		struct {
			AllPartitions bool `xml:"ALLPARTITIONS"`
			StartFromKey  int  `xml:"STARTFROMKEY"`
			NumToReturn   int  `xml:"NUMTORETURN"`
		}{
			AllPartitions: true,
			StartFromKey:  0,
			NumToReturn:   1000,
		},
	)

	users, err := doPaginated[People](c, "POST", url, payload)
	if err != nil {
		return nil, err
	}

	// Cache and return our successfully fetched list of users.
	c.SetCache(cache_key, users, 5*time.Minute)
	return (*users).People, nil
}

func (c *Client) GetPerson(personID string) (*Person, error) {
	url := c.BuildURL(NetBoxAPI)
	cache_key := fmt.Sprintf("%s_%s_%s", url, NetboxCommands.People.GetPerson, personID)

	var cache Person
	if c.GetCache(cache_key, &cache) {
		return &cache, nil
	}

	payload := c.BuildRequest(
		NetboxCommands.People.GetPerson,
		struct {
			PersonID string `xml:"PERSONID"`
		}{
			PersonID: personID,
		},
	)

	person, err := do[Person](c, "POST", url, payload)
	if err != nil {
		return nil, err
	}

	c.SetCache(cache_key, person, 5*time.Minute)
	return &person, nil
}
