package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GeoData represents geolocation information for an IP.
type GeoData struct {
	Country string `json:"country,omitempty"`
	Region  string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
	ISP     string `json:"isp,omitempty"`
}

// FetchGeolocation fetches geolocation data for an IP address.
func FetchGeolocation(ip string) (*GeoData, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geolocation for IP %s: %w", ip, err)
	}
	defer resp.Body.Close()

	var geo GeoData
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return nil, fmt.Errorf("failed to parse geolocation data for IP %s: %w", ip, err)
	}

	return &geo, nil
}
