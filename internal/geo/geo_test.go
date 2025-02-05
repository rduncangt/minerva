package geo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchGeolocation(t *testing.T) {
	// Mock API server that returns a valid response for a known IP.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/192.0.2.1"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"country":"United States","regionName":"CA","city":"San Francisco","isp":"MockISP"}`))
	}))
	defer mockServer.Close()

	// Override the API URL using SetAPIURL.
	originalURL := apiURL
	SetAPIURL(mockServer.URL)
	defer SetAPIURL(originalURL)

	// Test case: valid IP response.
	ip := "192.0.2.1"
	geoData, err := FetchGeolocation(ip)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if geoData.Country != "United States" {
		t.Errorf("Expected country 'United States', got %q", geoData.Country)
	}
	if geoData.Region != "CA" {
		t.Errorf("Expected region 'CA', got %q", geoData.Region)
	}
	if geoData.City != "San Francisco" {
		t.Errorf("Expected city 'San Francisco', got %q", geoData.City)
	}
	if geoData.ISP != "MockISP" {
		t.Errorf("Expected ISP 'MockISP', got %q", geoData.ISP)
	}
}

func TestFetchGeolocation_EdgeCases(t *testing.T) {
	// Mock API server for various edge cases.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/192.0.2.1":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"country":"United States","regionName":"CA","city":"San Francisco","isp":"MockISP"}`))
		case "/500-error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/malformed-json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"country":`)) // Incomplete JSON.
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Override the API URL using SetAPIURL.
	originalURL := apiURL
	SetAPIURL(mockServer.URL)
	defer SetAPIURL(originalURL)

	tests := []struct {
		name      string
		ip        string
		expectErr bool
	}{
		{"ValidIP", "192.0.2.1", false},
		{"ServerError", "500-error", true},
		{"MalformedJSON", "malformed-json", true},
		{"EmptyIP", "", true},
		{"InvalidIP", "invalid-ip", true},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			_, err := FetchGeolocation(tc.ip)
			if (err != nil) != tc.expectErr {
				t.Errorf("FetchGeolocation(%q) error = %v, expectErr = %v", tc.ip, err, tc.expectErr)
			}
		})
	}
}
