package main

import (
	"flag"
	"fmt"
	"log"
	"minerva/internal/db"
	"minerva/internal/geo"
	"minerva/internal/input"
	"minerva/internal/output"
	"minerva/internal/parser"
	"os"
	"time"
)

func main() {
	// Command-line flags
	limitFlag := flag.Int("limit", -1, "Limit the number of results (-1 for no limit)")
	reverseFlag := flag.Bool("r", false, "Process logs in oldest-first order (reverse default latest-first behavior)")
	flag.Parse()

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
	defer database.Close()

	// Initialize cache and summary
	geoCache := make(map[string]*geo.Data)
	lines, err := input.ReadLines(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	if !*reverseFlag {
		lines = input.ReverseLines(lines)
	}

	// Statistics
	totalRows := len(lines)
	duplicateCount := 0
	uniqueIPs := make(map[string]bool)

	// Process lines
	var summary []map[string]interface{}
	count := 0

	for _, line := range lines {
		if !parser.IsSuspiciousLog(line) {
			continue
		}

		// Extract fields from the log line
		timestamp, srcIP, dstIP, spt, dpt, proto := parser.ExtractFields(line)
		if srcIP == "" {
			continue
		}

		// Check if the IP is already processed
		if _, cached := geoCache[srcIP]; cached {
			duplicateCount++
			continue
		}

		// Check if IP is in the database
		exists, err := db.IsIPInDatabase(database, srcIP)
		if err != nil {
			log.Printf("Error checking IP in database: %v", err)
			continue
		}
		if exists {
			duplicateCount++
			continue
		}

		// Fetch and cache geolocation data
		geoData, err := geo.FetchGeolocation(srcIP)
		if err != nil {
			log.Printf("Warning: failed to fetch geolocation for %s: %v", srcIP, err)
			continue
		}
		geoCache[srcIP] = geoData

		// Insert the new record into the database
		err = db.InsertIPData(database, map[string]interface{}{
			"timestamp":        timestamp,
			"source_ip":        srcIP,
			"destination_ip":   dstIP,
			"protocol":         proto,
			"source_port":      spt,
			"destination_port": dpt,
			"geolocation":      geoData,
		})
		if err != nil {
			log.Printf("Error inserting data for IP %s: %v", srcIP, err)
			continue
		}

		// Add to summary data
		uniqueIPs[srcIP] = true
		summary = append(summary, map[string]interface{}{
			"date":           timestamp,
			"source_ip":      srcIP,
			"frequency":      1, // Can be updated based on future needs
			"ports_targeted": fmt.Sprintf("%s:%s", spt, dpt),
			"log_level":      "INFO",          // Placeholder
			"action_taken":   "Processed",     // Placeholder
			"geolocation":    geoData.Country, // Example field
			"notes":          "",              // Placeholder
		})

		// Enforce limit if specified
		count++
		if *limitFlag > 0 && count >= *limitFlag {
			break
		}
	}

	// Generate output
	if err := output.WriteIPSummaryTable(summary, os.Stdout); err != nil {
		log.Fatalf("Error writing IP summary table: %v", err)
	}

	// Log statistics and execution time
	executionTime := time.Since(startTime)
	log.Printf("Execution time: %v", executionTime)
	log.Printf("Total rows: %d", totalRows)
	log.Printf("Duplicate rows: %d", duplicateCount)
	log.Printf("Unique IPs: %d", len(uniqueIPs))
	log.Println("Log processing completed.")
}
