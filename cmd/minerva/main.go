package main

import (
	"flag"
	"log"
	"minerva/internal/db"
	"minerva/internal/geo"
	"minerva/internal/input"
	"minerva/internal/output"
	"minerva/internal/parser"
	"os"
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

	// Process lines
	var results []map[string]interface{}
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
				log.Printf("Skipping geolocation lookup for IP: %s (already in database)", srcIP)
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

		// Add to results for JSON output
		results = append(results, entry)

		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	// Write results to JSON
	err = output.WriteJSONOutput(results, os.Stdout)
	if err != nil {
		log.Fatalf("Error writing JSON output: %v", err)
	}

	log.Println("Log processing completed.")
}
