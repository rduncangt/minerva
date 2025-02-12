package db

import (
	"database/sql"
	"fmt"
	"minerva/internal/geo"
	"testing"
	"time"
)

// Constants for test database parameters.
const (
	testHost     = "localhost"
	testPort     = "5432"
	testUser     = "minerva_user"
	testPassword = "secure_password"
	testDBName   = "minerva_test"
)

// truncateTable truncates the specified table in the database.
func truncateTable(t *testing.T, db *sql.DB, table string) {
	t.Helper()
	_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table))
	if err != nil {
		t.Fatalf("Failed to truncate table %s: %v", table, err)
	}
}

// TestConnect verifies the Connect function with various parameters.
// NOTE: These tests assume that a test database named "minerva_test" exists
// and that the schema is properly configured.
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
			host:      testHost,
			port:      testPort,
			user:      testUser,
			password:  testPassword,
			dbname:    testDBName,
			expectErr: false,
		},
		{
			name:      "Invalid credentials",
			host:      testHost,
			port:      testPort,
			user:      "invalid_user",
			password:  "invalid_password",
			dbname:    "test_db",
			expectErr: true,
		},
		{
			name:      "Invalid host",
			host:      "invalid_host",
			port:      testPort,
			user:      testUser,
			password:  testPassword,
			dbname:    testDBName,
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

func TestInsertLogEntry(t *testing.T) {
	db, err := Connect(testHost, testPort, testUser, testPassword, testDBName)
	if err != nil {
		t.Fatalf("Failed to connect to the test database: %v", err)
	}
	defer db.Close()

	// Truncate the log_data table before testing.
	truncateTable(t, db, "log_data")

	testCases := []struct {
		name      string
		timestamp string
		sourceIP  string
		destIP    string
		protocol  string
		action    string
		reason    string
		packetLen int
		ttl       int
		expectErr bool
	}{
		{
			name:      "Valid entry",
			timestamp: time.Now().Format(time.RFC3339),
			sourceIP:  "192.0.2.1",
			destIP:    "203.0.113.5",
			protocol:  "TCP",
			action:    "ALLOW",
			reason:    "Routine test",
			packetLen: 128,
			ttl:       64,
			expectErr: false,
		},
		{
			name:      "Missing destination IP",
			timestamp: time.Now().Format(time.RFC3339),
			sourceIP:  "192.0.2.1",
			destIP:    "",
			protocol:  "TCP",
			action:    "DROP",
			reason:    "No destination",
			packetLen: 64,
			ttl:       128,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := InsertLogEntry(db, tc.timestamp, tc.sourceIP, tc.destIP, tc.protocol, tc.action, tc.reason, 12345, 80, tc.packetLen, tc.ttl)
			if (err != nil) != tc.expectErr {
				t.Errorf("Test %q: expected error: %v, got: %v", tc.name, tc.expectErr, err)
			}
		})
	}
}

func TestGeoDataInsertion(t *testing.T) {
	db, err := Connect(testHost, testPort, testUser, testPassword, testDBName)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Truncate the ip_geo table before testing.
	truncateTable(t, db, "ip_geo")

	handler := &Handler{DB: db}

	geoData := &geo.Data{
		Country: "United States",
		Region:  "California",
		City:    "San Francisco",
		ISP:     "Mock ISP",
	}

	t.Run("Insert new geolocation data", func(t *testing.T) {
		err := handler.InsertOrUpdateGeoData("192.0.2.1", geoData)
		if err != nil {
			t.Fatalf("Failed to insert geolocation data: %v", err)
		}
	})

	t.Run("Update existing geolocation data", func(t *testing.T) {
		geoData.City = "Los Angeles"
		err := handler.InsertOrUpdateGeoData("192.0.2.1", geoData)
		if err != nil {
			t.Fatalf("Failed to update geolocation data: %v", err)
		}
	})
}
