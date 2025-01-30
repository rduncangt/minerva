package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// IsSuspiciousLog checks if a log line indicates a potential threat.
func IsSuspiciousLog(line string) bool {
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

// ExtractFields extracts fields of interest from a log line.
func ExtractFields(line string) (string, string, string, int, int, string) {
	timestampRegex := regexp.MustCompile(`^\S+`) // First word as timestamp
	ipRegex := regexp.MustCompile(`SRC=(([0-9]{1,3}\.){3}[0-9]{1,3}|([a-fA-F0-9:]+))`)
	dstRegex := regexp.MustCompile(`DST=(([0-9]{1,3}\.){3}[0-9]{1,3}|([a-fA-F0-9:]+))`)
	sptRegex := regexp.MustCompile(`SPT=(\d+)`)
	dptRegex := regexp.MustCompile(`DPT=(\d+)`)
	protoRegex := regexp.MustCompile(`PROTO=(\w+)`)

	timestamp := timestampRegex.FindString(line)
	srcIP := getFirstGroup(ipRegex.FindStringSubmatch(line))
	dstIP := getFirstGroup(dstRegex.FindStringSubmatch(line))
	spt := parsePort(sptRegex.FindStringSubmatch(line))
	dpt := parsePort(dptRegex.FindStringSubmatch(line))
	proto := getFirstGroup(protoRegex.FindStringSubmatch(line))

	// Ensure empty strings are returned for missing fields
	if timestamp == "" {
		timestamp = "unknown"
	}
	if srcIP == "" {
		srcIP = "unknown"
	}
	if dstIP == "" {
		dstIP = "unknown"
	}
	if proto == "" {
		proto = "unknown"
	}

	return timestamp, srcIP, dstIP, spt, dpt, proto
}

// parsePort safely parses a port field. Returns 0 if missing or invalid.
func parsePort(match []string) int {
	if len(match) > 1 {
		port := match[1]
		return atoiSafe(port)
	}
	return 0
}

// atoiSafe converts a string to an integer, returning 0 if conversion fails.
func atoiSafe(s string) int {
	port, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return port
}

// getFirstGroup retrieves the first capturing group from a regex match.
func getFirstGroup(match []string) string {
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
