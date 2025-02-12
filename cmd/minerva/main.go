package main

import (
	"flag"
	"fmt"
	"log"
	"minerva/internal/config"
	"minerva/internal/db"
	"minerva/internal/geo"
	"minerva/internal/input"
	"minerva/internal/parser"
	"minerva/internal/progress"
	"os"
	"strconv"
	"sync"
	"time"
)

const maxGeoQueriesPerMinute = 40

func main() {
	reverseFlag := flag.Bool("r", false, "Process logs in oldest-first order")
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.Println("Starting log processing...")

	// Load configuration from file.
	conf, err := config.LoadConfig("minerva_config.toml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Allow overriding the database name with an environment variable.
	if dbName := os.Getenv("MINERVA_DB_NAME"); dbName != "" {
		conf.Database.Name = dbName
	}

	// Connect to the database using config values.
	dbPort := strconv.Itoa(conf.Database.Port)
	database, err := db.Connect(conf.Database.Host, dbPort, conf.Database.User, conf.Database.Password, conf.Database.Name)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	dbHandler := &db.Handler{DB: database}

	// Read input logs from stdin.
	lines, err := input.ReadLines(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	totalLines := len(lines)

	// Optionally reverse lines.
	if !*reverseFlag {
		lines = input.ReverseLines(lines)
	}

	// Set up statistics and progress tracker.
	stats := &progress.Stats{}
	prog := progress.NewProgress(int64(totalLines), stats)

	// Channels to move data through pipeline.
	logChan := make(chan string, 10000)
	geoChan := make(chan string, 10000)
	doneChan := make(chan struct{})

	// We’ll keep track of IPs we’ve already queued for geo so we don’t re-queue them.
	var seenIPs sync.Map

	//
	// 1) Pre-filter logs
	//    - If line is invalid → stats.IncrementMalformed()
	//    - Else if flagged → send to logChan
	//    - Else increment benign
	//
	go func() {
		for _, line := range lines {
			stats.IncrementLinesRead()

			if !parser.IsValidLine(line) {
				stats.IncrementMalformed()
				continue
			}
			if parser.IsFlaggedLog(line) {
				stats.IncrementFlagged()
				logChan <- line
			} else {
				stats.IncrementBenign()
			}
		}
		close(logChan)
	}()

	//
	// 2) Worker pool to insert logs in DB and dispatch new IP lookups.
	//
	workerCount := 20
	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for line := range logChan {
				timestamp, srcIP, dstIP, spt, dpt, proto, action, reason, packetLength, ttl := parser.ExtractFields(line)

				if dstIP == "" {
					// Additional malformed check
					prog.BufferMessage(fmt.Sprintf("Skipping malformed log line: %s", line))
					stats.IncrementMalformed()
					continue
				}

				// Insert into database
				if rowInserted, err := db.InsertLogEntry(database, timestamp, srcIP, dstIP, proto, action, reason, spt, dpt, packetLength, ttl); err != nil {
					stats.IncrementErrors()
					prog.BufferMessage(fmt.Sprintf("Insert error for DST=%q: %v", dstIP, err))
				} else if rowInserted > 0 {
					stats.IncrementInserted()
				}

				// Check for IP lookups
				if srcIP != "" {
					if _, loaded := seenIPs.LoadOrStore(srcIP, struct{}{}); !loaded {
						exists, err := dbHandler.IsIPInGeoTable(srcIP)
						if err != nil {
							stats.IncrementErrors()
							prog.BufferMessage(fmt.Sprintf("DB error checking IP: %v", err))
						} else if !exists {
							stats.IncrementGeoQueued()
							geoChan <- srcIP
						} // IP not seen yet, queue it
					}
				}

				prog.IncrementProcessed()
				prog.DisplayIfNeeded(2 * time.Second) // Show updates periodically
			}
		}()
	}

	//
	// 3) Single goroutine to handle geo lookups with throttling.
	//
	var geoWG sync.WaitGroup
	geoWG.Add(1)

	go func() {
		defer geoWG.Done()
		ticker := time.NewTicker(time.Minute / time.Duration(maxGeoQueriesPerMinute))
		defer ticker.Stop()

		for ip := range geoChan {
			<-ticker.C

			err := geo.ProcessIP(dbHandler, ip)

			// Decrement from the “in queue” count
			stats.DecrementGeoQueued()

			if err != nil {
				stats.IncrementGeoErrors()
				prog.BufferMessage(fmt.Sprintf("Geo lookup failed for IP=%s: %v", ip, err))
				continue
			}
			stats.IncrementGeoCompleted()
		}
	}()

	// Close geoChan when DB workers finish
	go func() {
		wg.Wait()
		close(geoChan)
	}()

	// Signal main doneChan when geo lookups finish
	go func() {
		geoWG.Wait()
		close(doneChan)
	}()

	// Start periodic progress display until everything is done
	prog.StartPeriodicDisplay(5*time.Second, doneChan)
}
