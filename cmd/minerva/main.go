package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// IPEntry represents information about a suspicious IP log entry.
type IPEntry struct {
	Timestamp       string   `json:"timestamp"`
	SourceIP        string   `json:"source_ip"`
	DestinationIP   string   `json:"destination_ip,omitempty"`
	SourcePort      string   `json:"source_port,omitempty"`
	DestinationPort string   `json:"destination_port,omitempty"`
	Protocol        string   `json:"protocol,omitempty"`
	Geolocation     *GeoData `json:"geolocation,omitempty"`
}

// GeoData represents geolocation information for an IP.
type GeoData struct {
	Country string `json:"country,omitempty"`
	Region  string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
	ISP     string `json:"isp,omitempty"`
}

// ReverseSlice reverses the order of lines in a slice.
func ReverseSlice(lines []string) []string {
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}

// isExternalIP checks whether an IP address is outside private ranges.
func isExternalIP(ip string) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false // Invalid IP address
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(parsedIP) {
			return false // IP is private
		}
	}

	return true // IP is external
}

// getGeolocation fetches geolocation data for an IP address.
func getGeolocation(ip string) (*GeoData, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geolocation for IP %s: %w", ip, err)
	}
	defer resp.Body.Close()

	body := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for IP %s: %w", ip, err)
	}

	var geo GeoData
	if err := json.Unmarshal(body, &geo); err != nil {
		return nil, fmt.Errorf("failed to parse geolocation data for IP %s: %w", ip, err)
	}

	return &geo, nil
}

// isSuspiciousLog checks if a log line indicates a potential threat.
func isSuspiciousLog(line string) bool {
	suspiciousReasons := []string{
		"POLICY-INPUT-GEN-DISCARD",
		"PORTSCAN",
		"INTRUSION-DETECTED",
		"MALFORMED-PACKET",
	}

	// Look for action=DROP and any suspicious reason
	return strings.Contains(line, "action=DROP") &&
		func() bool {
			for _, reason := range suspiciousReasons {
				if strings.Contains(line, reason) {
					return true
				}
			}
			return false
		}()
}

func main() {
	// Add flags for limiting the number of outputs and processing order
	limitFlag := flag.Int("limit", -1, "Limit the number of results (-1 for no limit)")
	reverseFlag := flag.Bool("r", false, "Process logs in oldest-first order (reverse default latest-first behavior)")
	flag.Parse()
	limit := *limitFlag
	reverse := *reverseFlag

	// Initialize logger
	log.SetOutput(os.Stderr)
	log.Println("Starting log processing...")

	// Read input lines from stdin
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading stdin: %v", err)
	}

	// Reverse the lines if the reverse flag is NOT set (default: latest-first)
	if !reverse {
		lines = ReverseSlice(lines)
	}

	// Regex patterns for various fields
	timestampRegex := regexp.MustCompile(`^\S+`) // First word as timestamp
	srcIPRegex := regexp.MustCompile(`SRC=(\d{1,3}(?:\.\d{1,3}){3})`)
	dstIPRegex := regexp.MustCompile(`DST=(\d{1,3}(?:\.\d{1,3}){3})`)
	sptRegex := regexp.MustCompile(`SPT=(\d+)`)
	dptRegex := regexp.MustCompile(`DPT=(\d+)`)
	protoRegex := regexp.MustCompile(`PROTO=(\w+)`)

	// Slice to store results
	var results []IPEntry

	// Process lines
	count := 0
	for _, line := range lines {
		// Check if the log line is suspicious
		if !isSuspiciousLog(line) {
			continue
		}

		// Extract fields
		timestamp := timestampRegex.FindString(line)
		srcIP := srcIPRegex.FindStringSubmatch(line)
		dstIP := dstIPRegex.FindStringSubmatch(line)
		spt := sptRegex.FindStringSubmatch(line)
		dpt := dptRegex.FindStringSubmatch(line)
		proto := protoRegex.FindStringSubmatch(line)

		// Only process if source IP is valid and external
		if len(srcIP) > 1 && isExternalIP(srcIP[1]) {
			entry := IPEntry{
				Timestamp:       timestamp,
				SourceIP:        srcIP[1],
				DestinationIP:   ifExists(dstIP),
				SourcePort:      ifExists(spt),
				DestinationPort: ifExists(dpt),
				Protocol:        ifExists(proto),
			}

			// Fetch geolocation data
			geo, err := getGeolocation(srcIP[1])
			if err != nil {
				log.Printf("Warning: %v", err)
			} else {
				entry.Geolocation = geo
			}

			results = append(results, entry)
			count++

			// Break if the limit is reached
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	// Output aggregated JSON data
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate JSON output: %v", err)
	}
	fmt.Println(string(jsonData))

	log.Println("Log processing completed.")
}

// ifExists safely extracts the first element from a regex match or returns empty string.
func ifExists(match []string) string {
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
