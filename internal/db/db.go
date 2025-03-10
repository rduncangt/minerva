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
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// InsertLogEntry inserts a new log entry into the log_data table, including new fields.
func InsertLogEntry(db *sql.DB, timestamp, sourceIP, destinationIP, protocol, action, reason string,
	sourcePort, destinationPort, packetLength, ttl int) (rowsInserted int64, err error) {

	// Basic validation to enforce mandatory fields.
	if timestamp == "" || timestamp == "unknown" {
		return 0, fmt.Errorf("invalid timestamp")
	}
	if destinationIP == "" || destinationIP == "unknown" {
		return 0, fmt.Errorf("invalid destination IP")
	}

	insertSQL := `
        INSERT INTO log_data (
            timestamp, source_ip, destination_ip, protocol,
            source_port, destination_port, action, reason,
            packet_length, ttl
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (timestamp, source_ip, destination_ip, protocol, source_port, destination_port)
        DO NOTHING;
    `
	result, errExec := db.Exec(insertSQL,
		timestamp,
		sourceIP,
		destinationIP,
		protocol,
		sourcePort,
		destinationPort,
		action,
		reason,
		packetLength,
		ttl,
	)
	if errExec != nil {
		return 0, fmt.Errorf("failed to insert log entry: %w", errExec)
	}

	// Get the number of affected rows
	rowsInserted, errRowsAffected := result.RowsAffected()
	if errRowsAffected != nil {
		return 0, fmt.Errorf("failed to retrieve affected row count: %w", errRowsAffected)
	}
	return rowsInserted, nil
}

// Handler is a wrapper around *sql.DB that implements GeoDataHandler.
type Handler struct {
	DB *sql.DB
}

// IsIPInGeoTable checks whether the given IP address exists in the ip_geo table.
func (h *Handler) IsIPInGeoTable(ip string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ip_geo WHERE ip_address = $1)`
	err := h.DB.QueryRow(query, ip).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check geo table for IP %s: %w", ip, err)
	}
	return exists, nil
}

// InsertOrUpdateGeoData inserts or updates geolocation data for an IP address.
func (h *Handler) InsertOrUpdateGeoData(ip string, geoData *geo.Data) error {
	insertSQL := `
    INSERT INTO ip_geo (
        ip_address, country, region, city, isp, latitude, longitude, last_updated
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
    ON CONFLICT (ip_address) DO UPDATE SET 
        country = EXCLUDED.country,
        region = EXCLUDED.region,
        city = EXCLUDED.city,
        isp = EXCLUDED.isp,
        latitude = EXCLUDED.latitude,
        longitude = EXCLUDED.longitude,
        last_updated = NOW();`

	_, err := h.DB.Exec(insertSQL, ip, geoData.Country, geoData.Region, geoData.City, geoData.ISP, geoData.Latitude, geoData.Longitude)
	if err != nil {
		return fmt.Errorf("failed to insert or update geolocation data for IP %s: %w", ip, err)
	}
	return nil
}
