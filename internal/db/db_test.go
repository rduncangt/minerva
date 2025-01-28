package db

import (
	"testing"
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
