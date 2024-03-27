/*
# Active Directory

This package initializes all the methods which interact with {Active Directory/LDAP}:
- https://docs.microsoft.com/en-us/windows/win32/ad/active-directory-schema

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/active_directory/active_directory.go
package active_directory

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/log"

	"github.com/go-ldap/ldap/v3"
)

const (
	LDAPPort  = "389" // Default LDAP port for pure TCP connection
	LDAPSPort = "636" // Default LDAPS port for SSL connection
)

// BuildURL builds a URL for a given resource and identifiers. TODO: This is not correct
func (c *Client) BuildDN(endpoint string, identifiers ...string) string {
	url := fmt.Sprintf(endpoint, c.BaseDN)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s,%s", url, id)
	}
	return url
}

/*
  - # Generate {Active Directory,LDAP} Client
  - @param logger *log.Logger
  - @return *Client
  - Example:

```go

	a := active_directory.NewClient(log.DEBUG)

```
*/
func NewClient(verbosity int) *Client {
	log := log.NewLogger("{active_directory}", verbosity)

	url := config.GetEnv("AD_LDAP_SERVER")
	if url == "" {
		log.Fatal("AD_LDAP_SERVER is not set")
	}

	port := config.GetEnv("AD_PORT")
	if len(port) == 0 {
		log.Warning("AD_PORT is not set, using default")
	}
	server := fmt.Sprintf("%s:%s", url, port)

	if port == LDAPSPort {
		log.Debug("Using LDAPS")
		server = fmt.Sprintf("ldaps://%s", server)
	} else {
		log.Debug("Using LDAP")
		server = fmt.Sprintf("ldap://%s", server)
	}

	baseDN := config.GetEnv("AD_BASE_DN")
	if len(baseDN) == 0 {
		log.Warning("AD_BASE_DN is not set, using default")
	}

	username := config.GetEnv("AD_USERNAME")
	if len(username) == 0 {
		log.Fatal("AD_USERNAME is not set")
	}

	password := config.GetEnv("AD_PASSWORD")
	if len(password) == 0 {
		log.Fatal("AD_PASSWORD is not set")
	}

	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "/tmp/rego_cache_active_directory.gob", 1000000)
	if err != nil {
		panic(err)
	}

	l, err := ldap.DialURL(server)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to LDAP server: %v", err))
	}

	// Bind to the server
	err = l.Bind(username, password)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to bind to LDAP server: %v", err))
	}

	return &Client{
		Server:   server,
		BaseDN:   baseDN,
		username: username,
		password: password,
		LDAP:     l,
		Log:      log,
		Cache:    cache,
	}
}

/*
 * Perform a generic request to the Active Directory Server
 */
func do[T Slice[E], E any](c *Client, filter string, attributes *[]Attribute) (T, error) {
	attr := ConvertAttributes(attributes)

	// Prepare the LDAP search request
	searchRequest := ldap.NewSearchRequest(
		c.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1000000,
		0,
		false,
		filter,
		attr,
		nil,
	)

	// Execute the LDAP search
	// sr, err := c.LDAP.Search(searchRequest)
	sr, err := c.LDAP.SearchWithPaging(searchRequest, 1000)
	if err != nil {
		return *new(T), fmt.Errorf(ldap.LDAPResultCodeMap[err.(*ldap.Error).ResultCode])
	}

	// Process each LDAP entry
	var results T = make([]E, 0, len(sr.Entries))
	for _, entry := range sr.Entries {
		// Create a new instance of element type using reflection
		itemType := reflect.TypeOf((*new(E))).Elem()
		item := reflect.New(itemType).Elem()

		// Map LDAP entry to struct
		for _, attr := range *attributes {
			fieldName := string(attr)
			ldapValues := entry.GetAttributeValues(fieldName)

			if fieldName == "dn" {
				// Set the DN field
				item.FieldByName("DN").Set(reflect.ValueOf(entry.DN))
				continue
			}

			// Iterate over the fields of the struct and set the values
			for i := 0; i < item.NumField(); i++ {
				structField := item.Type().Field(i)
				if structField.Tag.Get("ldap") == fieldName {
					fieldVal := item.Field(i)

					// Handle different types of fields
					switch fieldVal.Kind() {
					case reflect.String:
						// Set the first value for fields of type 'string'
						if len(ldapValues) > 0 && fieldVal.IsValid() && fieldVal.CanSet() {
							fieldVal.Set(reflect.ValueOf(ldapValues[0]))
						}
					case reflect.Slice:
						// Assign the entire slice for fields of type '[]string'
						if fieldVal.Type().Elem().Kind() == reflect.String && fieldVal.IsValid() && fieldVal.CanSet() {
							fieldVal.Set(reflect.ValueOf(ldapValues))
						}
					}
				}
			}
		}

		// Append the mapped item to results
		results = append(results, item.Addr().Interface().(E))
	}
	return results, nil
}

/*
 * SetCache stores an Active Directory response in the cache
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
 * GetCache retrieves an Active Directory response from the cache
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
