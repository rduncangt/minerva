package parser

import (
	"regexp"
	"strings"
)

// isSuspiciousLog checks if a log line indicates a potential threat.
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
func ExtractFields(line string) (string, string, string, string, string, string) {
	timestampRegex := regexp.MustCompile(`^\S+`) // First word as timestamp
	srcIPRegex := regexp.MustCompile(`SRC=(\d{1,3}(?:\.\d{1,3}){3})`)
	dstIPRegex := regexp.MustCompile(`DST=(\d{1,3}(?:\.\d{1,3}){3})`)
	sptRegex := regexp.MustCompile(`SPT=(\d+)`)
	dptRegex := regexp.MustCompile(`DPT=(\d+)`)
	protoRegex := regexp.MustCompile(`PROTO=(\w+)`)

	timestamp := timestampRegex.FindString(line)
	srcIP := getFirstGroup(srcIPRegex.FindStringSubmatch(line))
	dstIP := getFirstGroup(dstIPRegex.FindStringSubmatch(line))
	spt := getFirstGroup(sptRegex.FindStringSubmatch(line))
	dpt := getFirstGroup(dptRegex.FindStringSubmatch(line))
	proto := getFirstGroup(protoRegex.FindStringSubmatch(line))

	return timestamp, srcIP, dstIP, spt, dpt, proto
}

func getFirstGroup(match []string) string {
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
