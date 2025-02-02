package main

import (
	"encoding/json"
	"flag"
	"log"
	"minerva/internal/db"
	"minerva/internal/geo"
	"minerva/internal/input"
	"minerva/internal/parser"
	"minerva/internal/progress"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const maxGeoQueriesPerMinute = 40 // Throttle below API limit of 45 per minute

// PipelineSummary holds the info we'll print as JSON
type PipelineSummary struct {
	TotalLines          int `json:"total_lines"`
	SuspiciousLines     int `json:"suspicious_lines"`
	SuccessfullyInserts int `json:"successful_inserts"`
}

func main() {
	reverseFlag := flag.Bool("r", false, "Process logs in oldest-first order")
	flag.Parse()

	// Log to stderr so it's separate from JSON we write to stdout.
	log.SetOutput(os.Stderr)
	log.Println("Starting log processing...")

	startTime := time.Now()

	// Connect to DB
	database, err := db.Connect("localhost", "5432", "minerva_user", "secure_password", "minerva")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	dbHandler := &db.DBHandler{DB: database}

	// Read input
	lines, err := input.ReadLines(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	if !*reverseFlag {
		lines = input.ReverseLines(lines)
	}

	// Create a stats object to hold counters
	stats := &progress.Stats{}
	totalLines := len(lines)

	// Create a Progress tracker with the stats reference
	prog := progress.NewProgress(totalLines, stats)

	logChan := make(chan string, 10000)
	geoChan := make(chan string, 100)
	doneChan := make(chan struct{})

	// Pre-filter suspicious logs
	go func() {
		for _, line := range lines {
			if parser.IsSuspiciousLog(line) {
				// Increment the suspicious counter instead of a local variable
				stats.IncrementSuspicious()
				logChan <- line
			} else {
				// If you’d like to track non-suspicious logs, uncomment:
				// stats.IncrementNonRelevant()
			}
		}
		close(logChan)
	}()

	// Worker pool
	workerCount := 20
	var wg sync.WaitGroup

	// We'll introduce a new stat for successful inserts. Add it to the Stats struct if you like:
	var insertSuccesses int64 // atomic counter for inserts

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range logChan {
				timestamp, srcIP, dstIP, spt, dpt, proto, action, reason, packetLength, ttl :=
					parser.ExtractFields(line)

				// Attempt insert
				if err := db.InsertLogEntry(database, timestamp, srcIP, dstIP, proto, action, reason, spt, dpt, packetLength, ttl); err == nil {
					atomic.AddInt64(&insertSuccesses, 1) // increment success
				} else {
					// Log insertion errors but continue
					log.Printf("Error inserting log entry: %v", err)
				}

				// Queue IP for geolocation
				geoChan <- srcIP

				// Track overall line progress
				prog.Increment()
				prog.DisplayIfNeeded(1 * time.Second)
			}
		}()
	}

	// Geo lookups with throttling
	go func() {
		ticker := time.NewTicker(time.Minute / time.Duration(maxGeoQueriesPerMinute))
		defer ticker.Stop()
		for ip := range geoChan {
			<-ticker.C
			geo.ProcessIP(dbHandler, ip)
		}
	}()

	// Close channels once workers finish
	go func() {
		wg.Wait()
		close(doneChan)
		close(geoChan)
	}()

	// Periodic progress
	prog.StartPeriodicDisplay(5*time.Second, doneChan)

	// Final logs
	executionTime := time.Since(startTime)
	log.Printf("Execution time: %v", executionTime)
	log.Printf("Total rows processed: %d", totalLines)
	log.Println("Log processing completed.")

	// Output JSON summary to stdout for the test
	summary := PipelineSummary{
		TotalLines:          totalLines,
		SuspiciousLines:     int(atomic.LoadInt64(&stats.SuspiciousLines)),
		SuccessfullyInserts: int(atomic.LoadInt64(&insertSuccesses)),
	}

	if err := json.NewEncoder(os.Stdout).Encode(summary); err != nil {
		log.Printf("Failed to encode summary JSON: %v", err)
	}
}
