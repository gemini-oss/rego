/*
# Backupify - Users

This package initializes all the methods for functions which interact with Datto's Backupify WebUI:
https://www.backupify.com/

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/backupify/users.go
package backupify

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UserClient for chaining methods
type UserClient struct {
	*Client
}

// Entry point for export-related operations
func (c *Client) Users() *UserClient {
	return &UserClient{
		Client: c,
	}
}

// GetAllUsers() retrieves all users from Backupify.
func (c *UserClient) GetAllUsers() (*Users, error) {
	url := c.BuildURL(customerServices)
	cache_key := fmt.Sprintf("%s_%s", url, string(c.AppType))
	c.Log.Println("Getting all users from Backupify...")

	var cache Users
	if c.GetCache(cache_key, &cache) {
		return &cache, nil
	}

	userPayload := UserPayload{
		Draw: "1",
		Columns: []Column{
			{
				Data:       "name",
				Name:       "",
				Searchable: true,
				Orderable:  true,
				Search: Search{
					Value: "",
					Regex: false,
				},
			},
			{
				Data:       "email",
				Name:       "",
				Searchable: true,
				Orderable:  true,
				Search: Search{
					Value: "",
					Regex: false,
				},
			},
			{
				Data:       "latestSnap",
				Name:       "",
				Searchable: true,
				Orderable:  true,
				Search: Search{
					Value: "",
					Regex: false,
				},
			},
			{
				Data:       "usedBytes",
				Name:       "",
				Searchable: true,
				Orderable:  true,
				Search: Search{
					Value: "",
					Regex: false,
				},
			},
		},
		Order: []Order{
			{
				Column: "1",
				Dir:    "asc",
			},
		},
		Start:  0,
		Length: 75,
		Search: Search{
			Value: "",
			Regex: false,
		},
		AppType: c.AppType,
	}

	var allUsers Users
	for {
		c.Log.Printf("Getting users %d-%d from Backupify %s...", userPayload.Start, userPayload.Start+userPayload.Length-1, c.AppType)
		users, err := do[Users](c.Client, "POST", url, nil, userPayload)
		if err != nil {
			c.Log.Fatal(err)
		}

		remainingUsers := users.RecordsTotal - userPayload.Length
		if remainingUsers < userPayload.Length {
			userPayload.Length = remainingUsers
		}
		if userPayload.Start <= users.RecordsTotal {
			userPayload.Start += userPayload.Length
		} else {
			allUsers.Draw = users.Draw
			allUsers.RecordsTotal = users.RecordsTotal
			allUsers.RecordsFiltered = users.RecordsFiltered
			break
		}
		allUsers.Data = append(allUsers.Data, users.Data...)
	}
	c.convertUserBytes(&allUsers, false)

	c.SetCache(cache_key, allUsers, 6*time.Hour)
	return &allUsers, nil
}

// Initialize a map to count users and sum storage by the first letter of their email
func (c *UserClient) UserStorageReport(users *Users) map[string]UserCounts {

	userCountsByLetter := make(map[string]UserCounts)
	for email, user := range users.Map() {
		firstLetter := strings.ToUpper(string(email[0])) // Ensure the letter is uppercase
		stats := userCountsByLetter[firstLetter]
		stats.Count++
		stats.TotalStorage += user.UsedBytesFloat
		userCountsByLetter[firstLetter] = stats
	}

	return userCountsByLetter
}

func (c *UserClient) convertUserBytes(users *Users, useBinary bool) {
	var wg sync.WaitGroup
	var kilobyte float64
	if !useBinary {
		kilobyte = 1000 // Decimal unit (powers of 1000)
	}

	// Cascading definitions properly reflect the choice of kilobyte
	megabyte = kilobyte * kilobyte
	gigabyte = megabyte * kilobyte
	terabyte = gigabyte * kilobyte

	for _, user := range users.Data {
		wg.Add(1)
		go func(user *User) {
			defer wg.Done()

			// Extract the numeric part from the string (before the first space)
			usedBytes, err := strconv.ParseFloat(user.UsedBytes[:strings.Index(user.UsedBytes, " ")], 64)
			if err != nil {
				fmt.Printf("Error converting used bytes for user %s: %v\n", user.Name, err)
				return
			}

			// Calculate the UsedBytesFloat based on the unit found and the predefined variables
			switch {
			case strings.Contains(user.UsedBytes, "bytes"):
				user.UsedBytesFloat = usedBytes
			case strings.Contains(user.UsedBytes, "KB"):
				user.UsedBytesFloat = usedBytes * kilobyte
			case strings.Contains(user.UsedBytes, "MB"):
				user.UsedBytesFloat = usedBytes * megabyte
			case strings.Contains(user.UsedBytes, "GB"):
				user.UsedBytesFloat = usedBytes * gigabyte
			case strings.Contains(user.UsedBytes, "TB"):
				user.UsedBytesFloat = usedBytes * terabyte
			}

			fmt.Printf("Converted %s to %.2f bytes for user %s\n", user.UsedBytes, user.UsedBytesFloat, user.Name)
		}(user)
	}

	wg.Wait()
}

func (c *UserClient) filterUsersBySize(users *Users, size float64) *Users {
	var filteredUsers Users
	for _, user := range users.Data {
		if user.UsedBytesFloat > size {
			filteredUsers.Data = append(filteredUsers.Data, user)
			c.Log.Println("User:", user.Name, "has used "+user.UsedBytes+" of storage")
		}
	}
	return &filteredUsers
}
