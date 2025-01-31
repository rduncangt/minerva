package db

import (
	"database/sql"
	"fmt"
	"minerva/internal/geo"

	_ "github.com/lib/pq" // PostgreSQL driver
)

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
// InsertLogEntry inserts a new log entry into the log_data table, including new fields.
func InsertLogEntry(db *sql.DB, timestamp, sourceIP, destinationIP, protocol, action, reason string, sourcePort, destinationPort, packetLength, ttl int) error {
	insertSQL := `
    INSERT INTO log_data (
        timestamp, source_ip, destination_ip, protocol, source_port, destination_port, action, reason, packet_length, ttl
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    ON CONFLICT (timestamp, source_ip, destination_ip, protocol, source_port, destination_port) DO NOTHING;`

	_, err := db.Exec(insertSQL, timestamp, sourceIP, destinationIP, protocol, sourcePort, destinationPort, action, reason, packetLength, ttl)
	if err != nil {
		return fmt.Errorf("failed to insert log entry: %w", err)
	}
	return nil
}

// DBHandler is a wrapper around *sql.DB that implements GeoDataHandler.
type DBHandler struct {
	DB *sql.DB
}

// IsIPInGeoTable checks whether the given IP address exists in the ip_geo table.
func (h *DBHandler) IsIPInGeoTable(ip string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ip_geo WHERE ip_address = $1)`
	err := h.DB.QueryRow(query, ip).Scan(&exists)
	return exists, err
}

// InsertOrUpdateGeoData inserts or updates geolocation data for an IP address.
func (h *DBHandler) InsertOrUpdateGeoData(ip string, geoData *geo.GeoData) error {
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

	_, err := h.DB.Exec(insertSQL, ip, geoData.Country, geoData.Region, geoData.City, geoData.ISP)
	if err != nil {
		return fmt.Errorf("failed to insert or update geolocation data: %w", err)
	}
	return nil
}
