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

func TestExtractFields(t *testing.T) {
	tests := []struct {
		line                    string
		expectedTimestamp       string
		expectedSourceIP        string
		expectedDestinationIP   string
		expectedSourcePort      string
		expectedDestinationPort string
		expectedProtocol        string
	}{
		{
			"2025-01-05T00:01:08.143626-05:00 SRC=192.0.2.1 DST=192.0.2.2 PROTO=TCP SPT=12345 DPT=80",
			"2025-01-05T00:01:08.143626-05:00",
			"192.0.2.1",
			"192.0.2.2",
			"12345",
			"80",
			"TCP",
		},
		{
			"2025-01-05T00:01:09.000000-05:00 SRC=203.0.113.5 DST=198.51.100.1 PROTO=UDP SPT=54321 DPT=443",
			"2025-01-05T00:01:09.000000-05:00",
			"203.0.113.5",
			"198.51.100.1",
			"54321",
			"443",
			"UDP",
		},
	}

	for _, test := range tests {
		timestamp, srcIP, dstIP, spt, dpt, proto := ExtractFields(test.line)

		if timestamp != test.expectedTimestamp {
			t.Errorf("Expected timestamp %q, got %q", test.expectedTimestamp, timestamp)
		}
		if srcIP != test.expectedSourceIP {
			t.Errorf("Expected source IP %q, got %q", test.expectedSourceIP, srcIP)
		}
		if dstIP != test.expectedDestinationIP {
			t.Errorf("Expected destination IP %q, got %q", test.expectedDestinationIP, dstIP)
		}
		if spt != test.expectedSourcePort {
			t.Errorf("Expected source port %q, got %q", test.expectedSourcePort, spt)
		}
		if dpt != test.expectedDestinationPort {
			t.Errorf("Expected destination port %q, got %q", test.expectedDestinationPort, dpt)
		}
		if proto != test.expectedProtocol {
			t.Errorf("Expected protocol %q, got %q", test.expectedProtocol, proto)
		}
	}
}
