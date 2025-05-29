/*
# Active Directory

This package initializes all the methods which interact with {Active Directory/LDAP}:
- https://docs.microsoft.com/en-us/windows/win32/ad/active-directory-schema

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/active_directory/active_directory.go
package active_directory

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/config"
	"github.com/gemini-oss/rego/pkg/common/generics"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/starstruct"

	"github.com/go-ldap/ldap/v3"
)

const (
	LDAPPort  = "389" // Default LDAP port for pure TCP connection
	LDAPSPort = "636" // Default LDAPS port for SSL connection
)

var (
	reGen = regexp.MustCompile(`^(\d{14})(?:\.(\d+))?([Zz]|[+-]\d{4})$`) // reGen is a regular expression for parsing LDAP GeneralizedTime
)

// BuildURL builds a URL for a given resource and identifiers. TODO: This is not correct
func (c *Client) BuildDN(endpoint string, identifiers ...interface{}) string {
	url := fmt.Sprintf(endpoint, c.BaseDN)
	for _, id := range identifiers {
		url = fmt.Sprintf("%s,%v", url, id)
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

	// Setup TLS with LDAP CA
	ca := config.GetEnv("AD_LDAP_CA")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(ca))

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}

	encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		log.Fatal("REGO_ENCRYPTION_KEY is not set")
	}

	cache, err := cache.NewCache(encryptionKey, "rego_cache_active_directory.gob", 1000000)
	if err != nil {
		panic(err)
	}

	l, err := ldap.DialURL(server, ldap.DialWithTLSConfig(tlsConfig))
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
	// []Attribute needs to be []string
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
	sr, err := c.LDAP.SearchWithPaging(searchRequest, 1000)
	if err != nil {
		return nil, fmt.Errorf(ldap.LDAPResultCodeMap[err.(*ldap.Error).ResultCode])
	}

	results, err := unmarshalEntries[E](sr.Entries, attributes)
	if err != nil {
		return *new(T), err
	}
	return T(results), nil
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

func unmarshalEntries[E any](entries []*ldap.Entry, attributes *[]Attribute) ([]E, error) {

	entry, isPtr := generics.DerefGeneric[E]()

	// ldapTag → struct-field-index
	fieldMap := make(map[string]int, entry.NumField())
	for i := 0; i < entry.NumField(); i++ {
		if tag := entry.Field(i).Tag.Get("ldap"); tag != "" {
			fieldMap[tag] = i
		}
	}

	results := make([]E, len(entries))
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup

	var (
		mu   sync.Mutex
		errs []error
	)
	addErr := func(e error) {
		mu.Lock()
		errs = append(errs, e)
		mu.Unlock()
	}

	for i := range entries {
		sem <- struct{}{}
		wg.Add(1)

		go func(i int) {
			defer func() {
				<-sem
				wg.Done()
			}()

			dst := reflect.New(entry).Elem()

			for _, a := range *attributes {
				name := string(a)

				vals := entries[i].GetAttributeValues(name)
				if name == "dn" {
					vals = []string{entries[i].DN}
				}
				if len(vals) == 0 {
					continue
				}

				idx, ok := fieldMap[name]
				if !ok {
					continue
				}
				if err := unmarshalAttribute(dst.Field(idx), vals); err != nil {
					addErr(fmt.Errorf("entry %d field %s: %w", i, entry.Field(idx).Name, err))
					return
				}
			}

			// store result on success
			if isPtr {
				results[i] = dst.Addr().Interface().(E)
			} else {
				results[i] = dst.Interface().(E)
			}
		}(i)
	}

	wg.Wait()

	if len(errs) > 0 {
		return results, errors.Join(errs...)
	}
	return results, nil
}

// unmarshalAttribute assigns LDAP attribute values to a Field Value (fv)
func unmarshalAttribute(fv reflect.Value, vals []string) error {
	if len(vals) == 0 {
		return nil
	}

	// Dereference pointers and interfaces
	if fv.Kind() == reflect.Pointer || fv.Kind() == reflect.Interface {
		var err error
		fv, err = starstruct.DerefPointers(fv)
		if err != nil {
			return err
		}
	}

	// time.Time
	timeType := reflect.TypeOf(time.Time{})
	switch {
	case fv.Type() == timeType: // time.Time
		t, err := parseLDAPTime(vals[0])
		if err != nil {
			return err
		}
		fv.Set(reflect.ValueOf(t))
		return nil

	case fv.Kind() == reflect.Slice && fv.Type().Elem() == timeType: // []time.Time
		times := make([]time.Time, 0, len(vals))
		for _, v := range vals {
			t, err := parseLDAPTime(v)
			if err != nil {
				return err
			}
			times = append(times, t)
		}
		sort.Slice(times, func(i, j int) bool { return times[i].After(times[j]) })
		fv.Set(reflect.ValueOf(times))
		return nil
	}

	// []byte
	if fv.Kind() == reflect.Slice && fv.Type().Elem().Kind() == reflect.Uint8 {
		fv.SetBytes([]byte(vals[0]))
		return nil
	}

	// Generic Slice
	if fv.Kind() == reflect.Slice {
		n := len(vals)
		slice := reflect.MakeSlice(fv.Type(), n, n)
		for i, s := range vals {
			if err := unmarshalAttribute(slice.Index(i), []string{s}); err != nil {
				return err
			}
		}
		fv.Set(slice)
		return nil
	}

	// Skip other structs for now
	if fv.Kind() == reflect.Struct {
		return nil
	}

	// Primitives
	val := vals[0]
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("bool %q: %w", val, err)
		}
		fv.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("int %q: %w", val, err)
		}
		fv.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("uint %q: %w", val, err)
		}
		fv.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("float %q: %w", val, err)
		}
		fv.SetFloat(f)

	default:
		return fmt.Errorf("unsupported kind %s", fv.Kind())
	}

	return nil
}

// parseLDAPTime understands the two formats that Active Directory hands out:
//
//   - Windows FILETIME  – an integer number of 100-ns “ticks” since 1601-01-01.
//     0 or 9223372036854775807 (max int64) means “never/unset”.
//   - GeneralizedTime   – "YYYYMMDDHHmmSS(.fraction)(Z|±HHMM)".
//
// It always returns a UTC value; a zero Time means “unset”.
func parseLDAPTime(raw string) (time.Time, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, nil
	}

	// 1) FILETIME ----------------------------------------------------------
	if v, err := strconv.ParseUint(s, 10, 64); err == nil {
		if v == 0 || v == math.MaxInt64 {
			return time.Time{}, nil // unset / never
		}
		const ticksPerSec = int64(10_000_000)
		const winToUnixSec = int64(11_644_473_600) // seconds 1601-01-01 → 1970-01-01

		secs := int64(v / uint64(ticksPerSec))
		nanos := int64(v%uint64(ticksPerSec)) * 100

		secs -= winToUnixSec

		// time.Unix insists that nsec be in [0,1e9) with the same sign as sec.
		if secs < 0 && nanos > 0 {
			secs++
			nanos -= 1_000_000_000
		}
		return time.Unix(secs, nanos).UTC(), nil
	}

	// 2) GeneralizedTime ---------------------------------------------------
	match := reGen.FindStringSubmatch(s)
	if match == nil {
		return time.Time{}, fmt.Errorf("unsupported time format %q", s)
	}

	base, frac, tz := match[1], match[2], match[3]

	t, err := time.Parse("20060102150405", base)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse generalizedTime %q: %w", s, err)
	}

	// fraction → nanoseconds
	if frac != "" {
		if len(frac) > 9 { // trim >ns precision
			frac = frac[:9]
		}
		for len(frac) < 9 { // right-pad to ns
			frac += "0"
		}
		ns, _ := strconv.Atoi(frac)
		t = t.Add(time.Duration(ns) * time.Nanosecond)
	}

	// timezone offset
	if strings.EqualFold(tz, "Z") {
		return t.UTC(), nil
	}
	sign := 1
	if tz[0] == '-' {
		sign = -1
	}
	hh, _ := strconv.Atoi(tz[1:3])
	mm, _ := strconv.Atoi(tz[3:])
	offset := sign * ((hh * 60) + mm)
	return t.Add(time.Duration(-offset) * time.Minute).UTC(), nil
}
