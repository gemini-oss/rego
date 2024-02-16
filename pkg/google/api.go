/*
# Google Discovery API - Tests

This package contains all of the allowed scopes for the Google Workspace API:
https://developers.google.com/

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/google/api.go
package google

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gemini-oss/rego/pkg/common/requests"
)

var (
	//go:embed json/google_directory.json
	directoryJSON embed.FS

	//go:embed json/google_endpoints.json
	endpointsJSON embed.FS

	//go:embed json/google_scopes.json
	scopesJSON embed.FS
)

type DirectoryList struct {
	DiscoveryVersion string          `json:"discoveryVersion,omitempty"` // "v1"
	Items            []DirectoryItem `json:"items,omitempty"`            // List of Google API's
	Kind             string          `json:"kind,omitempty"`             // "discovery#directoryList"
}

type DirectoryItem struct {
	Description       string `json:"description,omitempty"`       // Lets you access information about other Google Workspace services"
	DiscoveryRestUrl  string `json:"discoveryRestUrl,omitempty"`  // https://www.googleapis.com/discovery/v1/apis/admin/directory_v1/rest
	DocumentationLink string `json:"documentationLink,omitempty"` // https://developers.google.com/admin-sdk/directory/
	ID                string `json:"id,omitempty"`                // "admin:directory_v1"
	Icons             Icon   `json:"icons,omitempty"`             // Icons for the API
	Kind              string `json:"kind,omitempty"`              // "discovery#directoryItem"
	Name              string `json:"name,omitempty"`              // "admin"
	Preferred         bool   `json:"preferred,omitempty"`         // true
	Title             string `json:"title,omitempty"`             // "Admin SDK"
	Version           string `json:"version,omitempty"`           // "directory_v1"
}

type Icon struct {
	X16 string `json:"x16,omitempty"`
	X32 string `json:"x32,omitempty"`
}

type Endpoints []Endpoint

type Endpoint struct {
	BasePath    string     `json:"basePath,omitempty"`
	BaseUrl     string     `json:"baseUrl,omitempty"`
	Description string     `json:"description,omitempty"`
	Name        string     `json:"name,omitempty"`
	Revision    string     `json:"revision,omitempty"`
	Title       string     `json:"title,omitempty"`
	Version     string     `json:"version,omitempty"`
	Auth        AuthDetail `json:"auth,omitempty"`
}

type AuthDetail struct {
	Oauth2 Oauth2Scopes `json:"oauth2,omitempty"`
}

type Oauth2Scopes struct {
	Scopes map[string]ScopeDetail `json:"scopes,omitempty"`
}

type ScopeDetail struct {
	Description string   `json:"description,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

type AllowedScopes map[string]ScopeDetail

/*
# FetchGoogleAPIScopes
*/
func FetchDirectoryEndpoints() (*DirectoryList, *Endpoints, error) {
	headers := requests.Headers{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	httpClient := requests.NewClient(nil, headers, nil)
	resp, body, err := httpClient.DoRequest("GET", "https://www.googleapis.com/discovery/v1/apis/", nil, nil)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("failed to fetch Google API scopes: %s", string(body))
	}

	var googleAPIs DirectoryList
	err = json.Unmarshal(body, &googleAPIs)
	if err != nil {
		return nil, nil, err
	}

	Endpoints := &Endpoints{}

	for _, item := range googleAPIs.Items {
		fmt.Println(item.DiscoveryRestUrl)
		switch item.DiscoveryRestUrl {
		case "https://realtimebidding.googleapis.com/$discovery/rest?version=v1alpha":
		case "https://poly.googleapis.com/$discovery/rest?version=v1":
		default:
			resp, body, err := httpClient.DoRequest("GET", item.DiscoveryRestUrl, nil, nil)
			if err != nil {
				return nil, nil, err
			}
			if resp.StatusCode != 200 {
				return nil, nil, fmt.Errorf("failed to fetch API details for %s: %s", item.Name, string(body))
			}

			endpoint := Endpoint{}
			err = json.Unmarshal(body, &endpoint)
			if err != nil {
				return nil, nil, err
			}

			*Endpoints = append(*Endpoints, endpoint)
		}
	}

	SaveEndpoints(Endpoints)
	return &googleAPIs, Endpoints, nil
}

/*
# ReadDiscoveryDirectory

	Reads the Google API discovery directory from a local file
*/
func ReadDiscoveryDirectory() (*DirectoryList, *Endpoints, error) {

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error determining root path: %s\n", err)
		return nil, nil, err
	}

	filePath := filepath.Join(rootPath, "..", "json", "google_directory.json")

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return nil, nil, err
	}
	defer file.Close()

	var googleAPIs DirectoryList
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&googleAPIs)
	if err != nil {
		return nil, nil, err
	}

	headers := requests.Headers{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	httpClient := requests.NewClient(nil, headers, nil)

	Endpoints := &Endpoints{}

	for _, item := range googleAPIs.Items {
		fmt.Println(item.DiscoveryRestUrl)
		switch item.DiscoveryRestUrl {
		// Skip this API as it's not available
		case "https://realtimebidding.googleapis.com/$discovery/rest?version=v1alpha":
		case "https://poly.googleapis.com/$discovery/rest?version=v1":
		default:
			resp, body, err := httpClient.DoRequest("GET", item.DiscoveryRestUrl, nil, nil)
			if err != nil {
				return nil, nil, err
			}
			if resp.StatusCode != 200 {
				return nil, nil, fmt.Errorf("failed to fetch API details for %s: %s", item.Name, string(body))
			}

			endpoint := Endpoint{}
			err = json.Unmarshal(body, &endpoint)
			if err != nil {
				return nil, nil, err
			}

			*Endpoints = append(*Endpoints, endpoint)
		}
	}

	SaveEndpoints(Endpoints)
	return &googleAPIs, Endpoints, nil
}

/*
# SaveEndpoints

	Saves Google API endpoints to a JSON file
*/
func SaveEndpoints(data interface{}) error {
	_, err := endpointsJSON.ReadFile("json/google_endpoints.json")
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return err
	}

	// Get the absolute path of the currently running file
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}
	executableDir := filepath.Dir(executablePath)

	var packageRoot string
	switch testing.Short() {
	case true:
		packageRoot, _ = filepath.Abs(filepath.Join(executableDir, "json"))
	case false:
		packageRoot, _ = filepath.Abs(filepath.Join(executableDir, "..", "json"))
	}

	json, _ := json.MarshalIndent(data, "", "  ")
	_ = os.WriteFile(filepath.Join(packageRoot, "google_endpoints.json"), json, 0644)

	return nil
}

/*
# Organize Scopes from Google API Endpoints

	Organizes the scopes from the Google API endpoints into a map of scopes by service
*/
func OrganizeScopes() error {
	endpoints := &Endpoints{}
	file, err := endpointsJSON.ReadFile("json/google_endpoints.json")
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return err
	}
	_ = json.Unmarshal([]byte(file), &endpoints)

	allowedScopes := &AllowedScopes{}
	for _, endpoint := range *endpoints {
		as := ScopeDetail{
			Description: endpoint.Description,
			Scopes:      []string{},
		}
		for scope := range endpoint.Auth.Oauth2.Scopes {
			as.Scopes = append(as.Scopes, scope)
		}
		// If the endpoint is already in the map, append the scopes, and remove duplicates
		if val, ok := (*allowedScopes)[endpoint.Title]; ok {
			as.Scopes = append(val.Scopes, as.Scopes...)
			as.Scopes = DedupeScopes(as.Scopes)
		}
		(*allowedScopes)[endpoint.Title] = as
	}

	executablePath, err := os.Executable()
	if err != nil {
		return err
	}
	executableDir := filepath.Dir(executablePath)

	var packageRoot string
	switch testing.Short() {
	case true:
		packageRoot, _ = filepath.Abs(filepath.Join(executableDir, "json"))
	case false:
		packageRoot, _ = filepath.Abs(filepath.Join(executableDir, "..", "json"))
	}

	json, _ := json.MarshalIndent(allowedScopes, "", "  ")
	err = os.WriteFile(filepath.Join(packageRoot, "google_scopes.json"), json, 0644)
	if err != nil {
		return err
	}

	return nil
}

/*
# DedupeScopes

	Removes duplicate scopes from a slice
*/
func DedupeScopes(slice []string) []string {
	encountered := make(map[string]bool)
	result := []string{}

	for _, value := range slice {
		if !encountered[value] {
			encountered[value] = true
			result = append(result, value)
		}
	}

	return result
}

/*
# LoadScopes

	Loads scopes from a JSON file
*/
func LoadScopes(service string) ([]string, error) {
	file, err := scopesJSON.ReadFile("json/google_scopes.json")
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return nil, err
	}

	as := &AllowedScopes{}
	err = json.Unmarshal([]byte(file), &as)
	if err != nil {
		return nil, err
	}

	return (*as)[service].Scopes, nil
}
