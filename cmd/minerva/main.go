package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

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
	// Regex to extract IP addresses
	srcIPRegex := regexp.MustCompile(`SRC=(\d{1,3}(?:\.\d{1,3}){3})`)
	uniqueIPs := make(map[string]bool)

	// Scan input line by line
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the log line is suspicious
		if !isSuspiciousLog(line) {
			continue
		}

		// Extract the source IP (SRC)
		matches := srcIPRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			srcIP := matches[1]
			if isExternalIP(srcIP) {
				uniqueIPs[srcIP] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Output unique external IPs
	for ip := range uniqueIPs {
		fmt.Println(ip)
	}
}
