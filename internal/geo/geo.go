package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var apiURL = "http://ip-api.com/json" // Default URL for geolocation API

type GeoData struct {
	Country string `json:"country"`
	Region  string `json:"regionName"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
}

func FetchGeolocation(ip string) (*GeoData, error) {
	// Construct the request URL with the given IP
	url := fmt.Sprintf("%s/%s", apiURL, ip)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var geoData GeoData
	if err := json.NewDecoder(resp.Body).Decode(&geoData); err != nil {
		return nil, err
	}
	return &geoData, nil
}
