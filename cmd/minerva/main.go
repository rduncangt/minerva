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
	"sync"
	"time"
)

func main() {
	// Command-line flags
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

	// Initialize cache and channels
	geoCache := make(map[string]*geo.Data)
	lines, err := input.ReadLines(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	if !*reverseFlag {
		lines = input.ReverseLines(lines)
	}

	logChan := make(chan string, 100)
	summaryChan := make(chan map[string]interface{}, 100)
	apiLimiter := time.NewTicker(time.Minute / 45)
	var wg sync.WaitGroup

	// Worker pool to process logs
	workerCount := 5
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for line := range logChan {
				if !parser.IsSuspiciousLog(line) {
					continue
				}

				// Extract fields from the log line
				timestamp, srcIP, dstIP, spt, dpt, proto := parser.ExtractFields(line)
				if srcIP == "" {
					continue
				}

				// Check if IP is already in cache or database
				if _, cached := geoCache[srcIP]; cached {
					continue
				}
				exists, err := db.IsIPInDatabase(database, srcIP)
				if err != nil {
					log.Printf("Error checking IP in database: %v", err)
					continue
				}
				if exists {
					continue
				}

				// Fetch and cache geolocation data with API rate limiting
				<-apiLimiter.C
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

				// Add to summary
				summaryChan <- map[string]interface{}{
					"date":           timestamp,
					"source_ip":      srcIP,
					"ports_targeted": fmt.Sprintf("%s:%s", spt, dpt),
					"geolocation":    geoData.Country,
				}
			}
		}(i)
	}

	// Feed log lines to workers
	go func() {
		for _, line := range lines {
			logChan <- line
		}
		close(logChan)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(summaryChan)
	}()

	// Collect and write summary output
	var summary []map[string]interface{}
	for entry := range summaryChan {
		summary = append(summary, entry)
	}

	if err := output.WriteIPSummaryTable(summary, os.Stdout); err != nil {
		log.Fatalf("Error writing IP summary table: %v", err)
	}

	// Log statistics and execution time
	executionTime := time.Since(startTime)
	log.Printf("Execution time: %v", executionTime)
	log.Printf("Total rows: %d", len(lines))
	log.Println("Log processing completed.")
}
