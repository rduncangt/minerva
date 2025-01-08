package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
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

func main() {
	// Regex to match IP addresses
	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	uniqueIPs := make(map[string]bool)

	// Scan input line by line
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		ips := ipRegex.FindAllString(line, -1)
		for _, ip := range ips {
			// Only add external IPs to the map
			if isExternalIP(ip) {
				uniqueIPs[ip] = true
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
