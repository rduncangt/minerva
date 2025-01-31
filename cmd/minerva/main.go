package main

import (
	"flag"
	"log"
	"minerva/internal/db"
	"minerva/internal/geo"
	"minerva/internal/input"
	"minerva/internal/parser"
	"minerva/internal/progress"
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

	// Create a database handler for geolocation operations
	dbHandler := &db.DBHandler{DB: database}

	// Read and reverse input lines if necessary
	lines, err := input.ReadLines(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	if !*reverseFlag {
		lines = input.ReverseLines(lines)
	}

	// Initialize progress tracker
	totalLines := len(lines)
	prog := progress.NewProgress(totalLines)

	logChan := make(chan string, 10000)
	doneChan := make(chan struct{})

	// Pre-filter and feed log lines to workers
	go func() {
		for _, line := range lines {
			if parser.IsSuspiciousLog(line) {
				logChan <- line
			}
		}
		close(logChan)
	}()

	// Worker pool for log processing
	workerCount := 20
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for line := range logChan {
				timestamp, srcIP, dstIP, spt, dpt, proto := parser.ExtractFields(line)
				if srcIP == "" {
					continue
				}

				// Insert log data
				err := db.InsertLogEntry(database, timestamp, srcIP, dstIP, proto, spt, dpt)
				if err != nil {
					log.Printf("Error inserting log entry: %v", err)
				}

				// Queue IP for geolocation lookup (handled by a separate process)
				geo.ProcessIP(dbHandler, srcIP)

				// Update progress
				prog.Increment()
				prog.DisplayIfNeeded(1 * time.Second)
			}
		}(i)
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// Start periodic progress updates
	prog.StartPeriodicDisplay(5*time.Second, doneChan)

	// Log statistics and execution time
	executionTime := time.Since(startTime)
	log.Printf("Execution time: %v", executionTime)
	log.Printf("Total rows processed: %d", totalLines)
	log.Println("Log processing completed.")
}
