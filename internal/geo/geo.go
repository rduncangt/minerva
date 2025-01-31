package geo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var apiURL = "http://ip-api.com/json" // Default URL for geolocation API

// client is a reusable HTTP client with a timeout.
var client = &http.Client{Timeout: 10 * time.Second}

// GeoData represents geolocation information for an IP address.
type GeoData struct {
	Country string `json:"country"`
	Region  string `json:"regionName"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
}

// FetchGeolocation retrieves geolocation data for the given IP address by querying the geolocation API.
func FetchGeolocation(ip string) (*GeoData, error) {
	url := fmt.Sprintf("%s/%s", apiURL, ip)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geolocation data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var geoData GeoData
	if err := json.NewDecoder(resp.Body).Decode(&geoData); err != nil {
		return nil, fmt.Errorf("failed to decode geolocation data: %w", err)
	}
	return &geoData, nil
}

// GeoDataHandler defines methods for geolocation data handling.
type GeoDataHandler interface {
	IsIPInGeoTable(ip string) (bool, error)
	InsertOrUpdateGeoData(ip string, geoData *GeoData) error
}

// ProcessIP handles the full lifecycle of fetching and storing geolocation data for an IP.
func ProcessIP(handler GeoDataHandler, ip string) {
	// Check if the IP already exists in the ip_geo table
	exists, err := handler.IsIPInGeoTable(ip)
	if err != nil {
		log.Printf("Error checking IP in geo table: %v", err)
		return
	}
	if exists {
		return
	}

	// Fetch geolocation data
	geoData, err := FetchGeolocation(ip)
	if err != nil {
		log.Printf("Error fetching geolocation for IP %s: %v", ip, err)
		return
	}

	// Insert or update geolocation data
	err = handler.InsertOrUpdateGeoData(ip, geoData)
	if err != nil {
		log.Printf("Error inserting/updating geolocation data for IP %s: %v", ip, err)
	}
}
