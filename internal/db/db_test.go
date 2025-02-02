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

func TestInsertLogEntry(t *testing.T) {
	db, err := Connect("localhost", "5432", "minerva_user", "secure_password", "minerva_test")
	if err != nil {
		t.Fatalf("Failed to connect to the test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("TRUNCATE TABLE log_data RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

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
			err := InsertLogEntry(db, tc.timestamp, tc.sourceIP, tc.destIP, tc.protocol, tc.action, tc.reason, 12345, 80, tc.packetLen, tc.ttl)
			if (err != nil) != tc.expectErr {
				t.Errorf("Test %q: expected error: %v, got: %v", tc.name, tc.expectErr, err)
			}
		})
	}
}

func TestGeoDataInsertion(t *testing.T) {
	db, err := Connect("localhost", "5432", "minerva_user", "secure_password", "minerva_test")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("TRUNCATE TABLE ip_geo RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	handler := &DBHandler{DB: db}

	geoData := &geo.GeoData{
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
