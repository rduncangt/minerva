package db

import (
	"minerva/internal/geo"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		port      string
		user      string
		password  string
		dbname    string
		expectErr bool
	}{
		{
			name:      "Valid connection",
			host:      "localhost",
			port:      "5432",
			user:      "minerva_user",
			password:  "secure_password",
			dbname:    "minerva_test",
			expectErr: false,
		},
		{
			name:      "Invalid credentials",
			host:      "localhost",
			port:      "5432",
			user:      "invalid_user",
			password:  "invalid_password",
			dbname:    "test_db",
			expectErr: true,
		},
		{
			name:      "Invalid host",
			host:      "invalid_host",
			port:      "5432",
			user:      "minerva_user",
			password:  "secure_password",
			dbname:    "minerva_test",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, err := Connect(test.host, test.port, test.user, test.password, test.dbname)
			if (err != nil) != test.expectErr {
				t.Errorf("Test %q: expected error: %v, got: %v", test.name, test.expectErr, err)
			}
			if db != nil {
				defer db.Close()
			}
		})
	}
}

func TestInsertIPData(t *testing.T) {
	// Setup: Connect to the test database
	db, err := Connect("localhost", "5432", "minerva_user", "secure_password", "minerva_test")
	if err != nil {
		t.Fatalf("Failed to connect to the test database: %v", err)
	}
	defer db.Close()

	// Clear the ip_data table before running the test
	_, err = db.Exec("TRUNCATE TABLE ip_data RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	// Define test cases
	testCases := []struct {
		name      string
		entry     map[string]interface{}
		expectErr bool
	}{
		{
			name: "Valid entry",
			entry: map[string]interface{}{
				"timestamp":        time.Now(),
				"source_ip":        "192.0.2.1",
				"destination_ip":   "203.0.113.5",
				"protocol":         "TCP",
				"source_port":      12345,
				"destination_port": 80,
				"geolocation": &geo.Data{
					Country: "United States",
					Region:  "California",
					City:    "San Francisco",
					ISP:     "Mock ISP",
				},
			},
			expectErr: false,
		},
		{
			name: "Duplicate entry",
			entry: map[string]interface{}{
				"timestamp":        time.Now(),
				"source_ip":        "192.0.2.1",
				"destination_ip":   "203.0.113.5",
				"protocol":         "TCP",
				"source_port":      12345,
				"destination_port": 80,
				"geolocation": &geo.Data{
					Country: "United States",
					Region:  "California",
					City:    "San Francisco",
					ISP:     "Mock ISP",
				},
			},
			expectErr: false,
		},
		{
			name: "Missing required field",
			entry: map[string]interface{}{
				"timestamp":        time.Now(),
				"source_ip":        "192.0.2.1",
				"protocol":         "TCP",
				"source_port":      12345,
				"destination_port": 80,
				"geolocation": &geo.Data{
					Country: "United States",
					Region:  "California",
					City:    "San Francisco",
					ISP:     "Mock ISP",
				},
			},
			expectErr: true, // Missing destination_ip
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := InsertIPData(db, tc.entry)
			if (err != nil) != tc.expectErr {
				t.Errorf("Test %q: expected error: %v, got: %v", tc.name, tc.expectErr, err)
			}

			// If the test case is valid, verify the data was inserted
			if !tc.expectErr {
				var exists bool
				query := `SELECT EXISTS(
					SELECT 1 FROM ip_data 
					WHERE source_ip = $1 AND destination_ip = $2
					  AND protocol = $3 AND source_port = $4
					  AND destination_port = $5
				)`
				err = db.QueryRow(
					query,
					tc.entry["source_ip"],
					tc.entry["destination_ip"],
					tc.entry["protocol"],
					tc.entry["source_port"],
					tc.entry["destination_port"],
				).Scan(&exists)
				if err != nil {
					t.Fatalf("Failed to query database: %v", err)
				}
				if !exists {
					t.Errorf("Test %q: expected row to exist but it does not", tc.name)
				}
			}
		})
	}
}
