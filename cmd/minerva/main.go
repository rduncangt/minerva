package main

import (
	"flag"
	"fmt"
	"log"
	"minerva/internal/config" // <-- new import
	"minerva/internal/db"
	"minerva/internal/geo"
	"minerva/internal/input"
	"minerva/internal/parser"
	"minerva/internal/progress"
	"os"
	"strconv" // <-- new import for port conversion
	"sync"
	"sync/atomic"
	"time"
)

const maxGeoQueriesPerMinute = 40

type PipelineSummary struct {
	TotalLines       int `json:"total_lines"`
	SuspiciousEvents int `json:"suspicious_events"`
	RecordsInserted  int `json:"records_inserted"`
	NewIPs           int `json:"new_ips"`
}

func main() {
	reverseFlag := flag.Bool("r", false, "Process logs in oldest-first order")
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.Println("Starting log processing...")

	// Load configuration (new step)
	conf, err := config.LoadConfig("minerva_config.toml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to the database (original comment preserved, but now using config)
	dbPort := strconv.Itoa(conf.Database.Port)
	database, err := db.Connect(conf.Database.Host, dbPort, conf.Database.User, conf.Database.Password, conf.Database.Name)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	dbHandler := &db.DBHandler{DB: database}

	// Read input logs
	lines, err := input.ReadLines(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	if !*reverseFlag {
		lines = input.ReverseLines(lines)
	}

	// Stats and progress tracking
	stats := &progress.Stats{}
	totalLines := len(lines)
	prog := progress.NewProgress(totalLines, stats)

	// Channels and worker variables
	logChan := make(chan string, 10000)
	geoChan := make(chan string, 100)
	doneChan := make(chan struct{})

	var seenIPs sync.Map
	var insertSuccesses int64

	// Pre-filter logs for events of interest
	go func() {
		for _, line := range lines {
			if parser.IsSuspiciousLog(line) {
				stats.IncrementEventLines()
				logChan <- line
			} else {
				stats.IncrementSkippedLines()
			}
		}
		close(logChan)
	}()

	// Worker pool to handle log processing
	workerCount := 20
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range logChan {
				timestamp, srcIP, dstIP, spt, dpt, proto, action, reason, packetLength, ttl := parser.ExtractFields(line)

				if dstIP == "" {
					prog.BufferMessage(fmt.Sprintf("Skipping malformed log line: %s", line))
					stats.IncrementSkippedLines()
					continue
				}

				if err := db.InsertLogEntry(database, timestamp, srcIP, dstIP, proto, action, reason, spt, dpt, packetLength, ttl); err != nil {
					stats.IncrementErrors()
					prog.BufferMessage(fmt.Sprintf("Insert error for DST=%q: %v", dstIP, err))
				} else {
					atomic.AddInt64(&insertSuccesses, 1)
					stats.IncrementAlreadyInDB()
				}

				if _, loaded := seenIPs.LoadOrStore(srcIP, struct{}{}); !loaded {
					exists, err := dbHandler.IsIPInGeoTable(srcIP)
					if err != nil {
						stats.IncrementErrors()
						prog.BufferMessage(fmt.Sprintf("DB error checking IP: %v", err))
					} else if !exists {
						stats.IncrementNewIPs()
						geoChan <- srcIP
					}
				}

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

	// Close channels when workers finish
	go func() {
		wg.Wait()
		close(doneChan)
		close(geoChan)
	}()

	// Periodic progress display
	prog.StartPeriodicDisplay(5*time.Second, doneChan)
}
