package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// Precompiled regex patterns for performance.
var (
	timestampRegex = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?([+\-]\d{2}:\d{2}|Z)?\b`)
	ipRegex        = regexp.MustCompile(`SRC=(([0-9]{1,3}\.){3}[0-9]{1,3}|([a-fA-F0-9:]+))`)
	dstRegex       = regexp.MustCompile(`DST=(([0-9]{1,3}\.){3}[0-9]{1,3}|([a-fA-F0-9:]+))`)
	sptRegex       = regexp.MustCompile(`SPT=(\d+)`)
	dptRegex       = regexp.MustCompile(`DPT=(\d+)`)
	protoRegex     = regexp.MustCompile(`PROTO=(\w+)`)
	actionRegex    = regexp.MustCompile(`action=(\w+)`)
	reasonRegex    = regexp.MustCompile(`reason=([\w\-]+)`)
	lengthRegex    = regexp.MustCompile(`LEN=(\d+)`)
	ttlRegex       = regexp.MustCompile(`TTL=(\d+)`)
)

// IsValidLine checks if a log line is well-formed.
func IsValidLine(line string) bool {
	return timestampRegex.MatchString(line) &&
		ipRegex.MatchString(line) &&
		dstRegex.MatchString(line) &&
		sptRegex.MatchString(line) &&
		dptRegex.MatchString(line) &&
		protoRegex.MatchString(line) &&
		actionRegex.MatchString(line) &&
		reasonRegex.MatchString(line) &&
		lengthRegex.MatchString(line) &&
		ttlRegex.MatchString(line)
}

// IsFlaggedLog checks if a log line indicates a potential threat.
func IsFlaggedLog(line string) bool {
	FlaggedReasons := []string{
		"POLICY-INPUT-GEN-DISCARD",
		"PORTSCAN",
		"INTRUSION-DETECTED",
		"MALFORMED-PACKET",
	}

	if !strings.Contains(line, "action=DROP") {
		return false
	}

	for _, reason := range FlaggedReasons {
		if strings.Contains(line, reason) {
			return true
		}
	}
	return false
}

// ExtractFields extracts fields of interest from a log line.
func ExtractFields(line string) (string, string, string, int, int, string, string, string, int, int) {
	timestamp := timestampRegex.FindString(line)
	srcIP := getFirstGroup(ipRegex.FindStringSubmatch(line))
	dstIP := getFirstGroup(dstRegex.FindStringSubmatch(line))
	spt := parsePort(sptRegex.FindStringSubmatch(line))
	dpt := parsePort(dptRegex.FindStringSubmatch(line))
	proto := getFirstGroup(protoRegex.FindStringSubmatch(line))
	action := getFirstGroup(actionRegex.FindStringSubmatch(line))
	reason := getFirstGroup(reasonRegex.FindStringSubmatch(line))
	packetLength := parsePort(lengthRegex.FindStringSubmatch(line))
	ttl := parsePort(ttlRegex.FindStringSubmatch(line))

	return nonEmpty(timestamp, "unknown"),
		nonEmpty(srcIP, "unknown"),
		nonEmpty(dstIP, "unknown"),
		spt,
		dpt,
		nonEmpty(proto, "unknown"),
		nonEmpty(action, "unknown"),
		nonEmpty(reason, "unknown"),
		packetLength,
		ttl
}

// parsePort safely parses a port or numeric field. Returns 0 if missing or invalid.
func parsePort(match []string) int {
	if len(match) > 1 {
		return atoiSafe(match[1])
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

// nonEmpty returns the default value if the input is an empty string.
func nonEmpty(input, defaultValue string) string {
	if input == "" {
		return defaultValue
	}
	return input
}
