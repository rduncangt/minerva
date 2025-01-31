package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// GeoData represents the geolocation information for an IP address.
type GeoData struct {
	Country string
	Region  string
	City    string
	ISP     string
}

// Connect establishes a connection to the database and returns the *sql.DB instance.
func Connect(host, port, user, password, dbname string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// InsertLogEntry inserts a new log entry into the log_data table.
func InsertLogEntry(db *sql.DB, timestamp string, sourceIP, destinationIP, protocol string, sourcePort, destinationPort int) error {
	insertSQL := `
    INSERT INTO log_data (
        timestamp, source_ip, destination_ip, protocol, source_port, destination_port
    ) VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (timestamp, source_ip, destination_ip, protocol, source_port, destination_port) DO NOTHING;`

	_, err := db.Exec(insertSQL, timestamp, sourceIP, destinationIP, protocol, sourcePort, destinationPort)
	if err != nil {
		return fmt.Errorf("failed to insert log entry: %w", err)
	}
	return nil
}

// IsIPInGeoTable checks whether the given IP address exists in the ip_geo table.
func IsIPInGeoTable(db *sql.DB, ip string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ip_geo WHERE ip_address = $1)`
	err := db.QueryRow(query, ip).Scan(&exists)
	return exists, err
}

// InsertOrUpdateGeoData inserts or updates geolocation data for an IP address.
func InsertOrUpdateGeoData(db *sql.DB, ip string, geoData *GeoData) error {
	insertSQL := `
    INSERT INTO ip_geo (
        ip_address, country, region, city, isp, last_updated
    ) VALUES ($1, $2, $3, $4, $5, NOW())
    ON CONFLICT (ip_address) DO UPDATE SET 
        country = EXCLUDED.country,
        region = EXCLUDED.region,
        city = EXCLUDED.city,
        isp = EXCLUDED.isp,
        last_updated = NOW();`

	_, err := db.Exec(insertSQL, ip, geoData.Country, geoData.Region, geoData.City, geoData.ISP)
	if err != nil {
		return fmt.Errorf("failed to insert or update geolocation data: %w", err)
	}
	return nil
}
