package parser

import "testing"

func TestIsSuspiciousLog(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"action=DROP reason=PORTSCAN SRC=192.0.2.1 DST=192.0.2.2", true},
		{"action=DROP reason=INTRUSION-DETECTED SRC=192.0.2.3 DST=192.0.2.4", true},
		{"action=ALLOW reason=WHITELIST SRC=192.0.2.5 DST=192.0.2.6", false},
		{"no-action-log SRC=192.0.2.7 DST=192.0.2.8", false},
	}

	for _, test := range tests {
		result := IsSuspiciousLog(test.line)
		if result != test.expected {
			t.Errorf("Expected IsSuspiciousLog(%q) to be %v, got %v", test.line, test.expected, result)
		}
	}
}
