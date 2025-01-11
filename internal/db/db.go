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

func InsertIPData(db *sql.DB, entry map[string]interface{}) error {
	// Validate required fields
	requiredFields := []string{"timestamp", "source_ip", "destination_ip", "protocol", "source_port", "destination_port", "geolocation"}
	for _, field := range requiredFields {
		if _, ok := entry[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Assert the type of geolocation
	geoData, ok := entry["geolocation"].(*geo.GeoData)
	if !ok {
		return fmt.Errorf("geolocation field is not of type *geo.GeoData")
	}

	// SQL query with ON CONFLICT to skip duplicates
	insertSQL := `
    INSERT INTO ip_data (
        timestamp, source_ip, destination_ip, protocol,
        source_port, destination_port, country, region, city, isp
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    ON CONFLICT ON CONSTRAINT unique_entry DO NOTHING;` // Avoid duplicates

	// Execute the query
	_, err := db.Exec(
		insertSQL,
		entry["timestamp"],
		entry["source_ip"],
		entry["destination_ip"],
		entry["protocol"],
		entry["source_port"],
		entry["destination_port"],
		geoData.Country,
		geoData.Region,
		geoData.City,
		geoData.ISP,
	)

	if err != nil {
		// Log only errors
		return fmt.Errorf("failed to insert data into database: %w", err)
	}

	return nil
}

func IsIPInDatabase(db *sql.DB, ip string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ip_data WHERE source_ip = $1 LIMIT 1)`
	err := db.QueryRow(query, ip).Scan(&exists)
	return exists, err
}
