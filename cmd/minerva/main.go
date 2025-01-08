package main

import (
	"flag"
	"log"
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

		geoData, err := geo.FetchGeolocation(srcIP)
		if err != nil {
			log.Printf("Warning: %v", err)
		}

		results = append(results, map[string]interface{}{
			"timestamp":        timestamp,
			"source_ip":        srcIP,
			"destination_ip":   dstIP,
			"source_port":      spt,
			"destination_port": dpt,
			"protocol":         proto,
			"geolocation":      geoData,
		})

		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	// Write results to JSON
	err = output.WriteJSONOutput(results, os.Stdout) // Updated to pass os.Stdout as the writer
	if err != nil {
		log.Fatalf("Error writing JSON output: %v", err)
	}

	log.Println("Log processing completed.")
}
