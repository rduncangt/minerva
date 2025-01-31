package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var apiURL = "http://ip-api.com/json" // Default URL for geolocation API

// Data represents the geolocation information for an IP address.
type Data struct {
	Country string `json:"country"`
	Region  string `json:"regionName"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
}

// client is a reusable HTTP client with a timeout.
var client = &http.Client{Timeout: 10 * time.Second}

// FetchGeolocation retrieves geolocation data for the given IP address by querying the geolocation API.
func FetchGeolocation(ip string) (*Data, error) {
	// Construct the request URL with the given IP
	url := fmt.Sprintf("%s/%s", apiURL, ip)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geolocation data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var geoData Data
	if err := json.NewDecoder(resp.Body).Decode(&geoData); err != nil {
		return nil, fmt.Errorf("failed to decode geolocation data: %w", err)
	}
	return &geoData, nil
}
