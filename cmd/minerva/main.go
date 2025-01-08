package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

// IPEntry represents information about a suspicious IP log entry.
type IPEntry struct {
	Timestamp       string
	SourceIP        string
	DestinationIP   string
	SourcePort      string
	DestinationPort string
	Protocol        string
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
	// Regex patterns for various fields
	timestampRegex := regexp.MustCompile(`^\S+`) // First word as timestamp
	srcIPRegex := regexp.MustCompile(`SRC=(\d{1,3}(?:\.\d{1,3}){3})`)
	dstIPRegex := regexp.MustCompile(`DST=(\d{1,3}(?:\.\d{1,3}){3})`)
	sptRegex := regexp.MustCompile(`SPT=(\d+)`)
	dptRegex := regexp.MustCompile(`DPT=(\d+)`)
	protoRegex := regexp.MustCompile(`PROTO=(\w+)`)

	// Map to store unique entries
	uniqueEntries := make(map[string]IPEntry)

	// Scan input line by line
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

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
			uniqueEntries[srcIP[1]] = entry
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Output structured data
	for _, entry := range uniqueEntries {
		fmt.Printf("Time: %s | SRC: %s | DST: %s | SPT: %s | DPT: %s | PROTO: %s\n",
			entry.Timestamp, entry.SourceIP, entry.DestinationIP,
			entry.SourcePort, entry.DestinationPort, entry.Protocol)
	}
}

// ifExists safely extracts the first element from a regex match or returns empty string.
func ifExists(match []string) string {
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
