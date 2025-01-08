package geo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchGeolocation(t *testing.T) {
	// Mock API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure the requested path includes the IP
		expectedPath := "/192.0.2.1"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"country":"United States","regionName":"CA","city":"San Francisco","isp":"MockISP"}`))
	}))
	defer mockServer.Close()

	// Replace API URL with mock server URL
	originalURL := apiURL
	apiURL = mockServer.URL
	defer func() { apiURL = originalURL }()

	// Test case
	ip := "192.0.2.1"
	geoData, err := FetchGeolocation(ip)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if geoData.Country != "United States" {
		t.Errorf("Expected country to be 'United States', got %q", geoData.Country)
	}
	if geoData.Region != "CA" {
		t.Errorf("Expected region to be 'CA', got %q", geoData.Region)
	}
	if geoData.City != "San Francisco" {
		t.Errorf("Expected city to be 'San Francisco', got %q", geoData.City)
	}
	if geoData.ISP != "MockISP" {
		t.Errorf("Expected ISP to be 'MockISP', got %q", geoData.ISP)
	}
}

func TestFetchGeolocation_EdgeCases(t *testing.T) {
	// Mock API server for different cases
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/192.0.2.1":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"country":"United States","regionName":"CA","city":"San Francisco","isp":"MockISP"}`))
		case "/500-error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/malformed-json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"country":`)) // Incomplete JSON
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Replace API URL with mock server URL
	originalURL := apiURL
	apiURL = mockServer.URL
	defer func() { apiURL = originalURL }()

	tests := []struct {
		ip        string
		expectErr bool
	}{
		{"192.0.2.1", false},
		{"500-error", true},
		{"malformed-json", true},
		{"", true},
		{"invalid-ip", true},
	}

	for _, test := range tests {
		_, err := FetchGeolocation(test.ip)
		if (err != nil) != test.expectErr {
			t.Errorf("FetchGeolocation(%q) error = %v, expectErr = %v", test.ip, err, test.expectErr)
		}
	}
}
