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
	NewIPs              int `json:"new_ips"` // If you'd like to output how many new IPs were geolocated
}

func main() {
	reverseFlag := flag.Bool("r", false, "Process logs in oldest-first order")
	flag.Parse()

	// Log to stderr so it's separate from any JSON on stdout.
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

	// Create a stats object to hold counters (from progress package)
	stats := &progress.Stats{}
	totalLines := len(lines)

	// Create a Progress tracker with the Stats reference
	prog := progress.NewProgress(totalLines, stats)

	// Channels for log lines and geo lookups
	logChan := make(chan string, 10000)
	geoChan := make(chan string, 100)
	doneChan := make(chan struct{})

	// In-memory set of IPs encountered during this run
	// so we don't re-check the same IP multiple times.
	var seenIPs sync.Map // key: string (IP), value: struct{}{}

	// For the final JSON summary, keep track of successful inserts
	// and how many brand-new IPs are found (optional).
	var insertSuccesses int64
	var newIPs int64

	// Pre-filter suspicious logs
	go func() {
		for _, line := range lines {
			if parser.IsSuspiciousLog(line) {
				stats.IncrementSuspicious() // track suspicious lines
				logChan <- line
			} else {
				// Optionally track non-suspicious lines:
				// stats.IncrementNonRelevant()
			}
		}
		close(logChan)
	}()

	// Worker pool
	workerCount := 20
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range logChan {
				// Extract fields and insert log record
				timestamp, srcIP, dstIP, spt, dpt, proto, action, reason, packetLength, ttl :=
					parser.ExtractFields(line)

				if err := db.InsertLogEntry(database, timestamp, srcIP, dstIP, proto, action, reason, spt, dpt, packetLength, ttl); err == nil {
					atomic.AddInt64(&insertSuccesses, 1)
				} else {
					log.Printf("Error inserting log entry: %v", err)
				}

				// Now check if we've seen this IP yet *in this run*:
				// If not, we do a DB check to see if it's already in ip_geo.
				// Only if truly new, enqueue for geolocation.
				if _, loaded := seenIPs.LoadOrStore(srcIP, struct{}{}); !loaded {
					exists, err := dbHandler.IsIPInGeoTable(srcIP)
					if err != nil {
						log.Printf("Error checking IP in DB: %v", err)
					} else if !exists {
						// brand-new IP => queue for geolocation
						atomic.AddInt64(&newIPs, 1) // track newly discovered IPs
						geoChan <- srcIP
					}
					// If it exists, do nothing further for geolocation
				}

				// Update progress
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
			geo.ProcessIP(dbHandler, ip) // Actually do the geolocation & DB update
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
		NewIPs:              int(atomic.LoadInt64(&newIPs)),
	}

	if err := json.NewEncoder(os.Stdout).Encode(summary); err != nil {
		log.Printf("Failed to encode summary JSON: %v", err)
	}
}
