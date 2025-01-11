package main

import (
	"flag"
	"log"
	"minerva/internal/db"
	"minerva/internal/geo"
	"minerva/internal/input"
	"minerva/internal/parser"
	"os"
	"time"
)

func main() {
	// Flags for limiting the number of outputs and reversing input order
	limitFlag := flag.Int("limit", -1, "Limit the number of results (-1 for no limit)")
	reverseFlag := flag.Bool("r", false, "Process logs in oldest-first order (reverse default latest-first behavior)")
	flag.Parse()
	limit := *limitFlag
	reverse := *reverseFlag

	// Initialize logger
	log.SetOutput(os.Stderr)
	log.Println("Starting log processing...")

	// Start execution timer
	startTime := time.Now()

	// Initialize database connection
	database, err := db.Connect("localhost", "5432", "minerva_user", "secure_password", "minerva")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	// Initialize in-memory cache for geolocation lookups
	geoCache := make(map[string]*geo.GeoData)

	lines, err := input.ReadLines(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}

	// Reverse lines if not in oldest-first order
	if !reverse {
		lines = input.ReverseLines(lines)
	}

	// Statistics variables
	totalRows := len(lines)
	duplicateCount := 0
	uniqueIPs := make(map[string]bool)

	// Process lines
	count := 0
	for _, line := range lines {
		if !parser.IsSuspiciousLog(line) {
			continue
		}

		timestamp, srcIP, dstIP, spt, dpt, proto := parser.ExtractFields(line)
		if srcIP == "" {
			continue
		}

		// Check if IP exists in cache
		var geoData *geo.GeoData
		var cached bool
		if geoData, cached = geoCache[srcIP]; !cached {
			// Check if IP exists in the database
			exists, err := db.IsIPInDatabase(database, srcIP)
			if err != nil {
				log.Printf("Error checking IP in database: %v", err)
			}

			if exists {
				duplicateCount++ // Increment duplicate count
				continue
			}

			// Fetch geolocation data
			geoData, err = geo.FetchGeolocation(srcIP)
			if err != nil {
				log.Printf("Warning: %v", err)
				continue
			}

			// Cache the fetched data
			geoCache[srcIP] = geoData
		}

		// Mark IP as unique
		uniqueIPs[srcIP] = true

		entry := map[string]interface{}{
			"timestamp":        timestamp,
			"source_ip":        srcIP,
			"destination_ip":   dstIP,
			"source_port":      spt,
			"destination_port": dpt,
			"protocol":         proto,
			"geolocation":      geoData,
		}

		// Insert data into the database
		err = db.InsertIPData(database, entry)
		if err != nil {
			log.Printf("Error inserting data into database: %v", err)
		}

		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	// Calculate execution time
	executionTime := time.Since(startTime)

	// Log statistics
	log.Printf("Execution time: %v", executionTime)
	log.Printf("Total rows: %d", totalRows)
	log.Printf("Duplicate rows: %d", duplicateCount)
	log.Printf("Unique IPs: %d", len(uniqueIPs))

	log.Println("Log processing completed.")
}
