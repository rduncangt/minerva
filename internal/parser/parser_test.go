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
		expectedSourcePort      int
		expectedDestinationPort int
		expectedProtocol        string
		expectedAction          string
		expectedReason          string
		expectedPacketLength    int
		expectedTTL             int
	}{
		{
			"2025-01-05T00:01:08.143626-05:00 SRC=192.0.2.1 DST=192.0.2.2 PROTO=TCP SPT=12345 DPT=80 action=DROP reason=PORTSCAN LEN=500 TTL=64",
			"2025-01-05T00:01:08.143626-05:00",
			"192.0.2.1",
			"192.0.2.2",
			12345,
			80,
			"TCP",
			"DROP",
			"PORTSCAN",
			500,
			64,
		},
	}

	for _, test := range tests {
		timestamp, srcIP, dstIP, spt, dpt, proto, action, reason, packetLength, ttl := ExtractFields(test.line)

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
			t.Errorf("Expected source port %d, got %d", test.expectedSourcePort, spt)
		}
		if dpt != test.expectedDestinationPort {
			t.Errorf("Expected destination port %d, got %d", test.expectedDestinationPort, dpt)
		}
		if proto != test.expectedProtocol {
			t.Errorf("Expected protocol %q, got %q", test.expectedProtocol, proto)
		}
		if action != test.expectedAction {
			t.Errorf("Expected action %q, got %q", test.expectedAction, action)
		}
		if reason != test.expectedReason {
			t.Errorf("Expected reason %q, got %q", test.expectedReason, reason)
		}
		if packetLength != test.expectedPacketLength {
			t.Errorf("Expected packet length %d, got %d", test.expectedPacketLength, packetLength)
		}
		if ttl != test.expectedTTL {
			t.Errorf("Expected TTL %d, got %d", test.expectedTTL, ttl)
		}
	}
}

func TestExtractFields_EdgeCases(t *testing.T) {
	tests := []struct {
		line                    string
		expectedSourceIP        string
		expectedDestinationIP   string
		expectedSourcePort      int
		expectedDestinationPort int
		expectedProtocol        string
		expectedAction          string
		expectedReason          string
		expectedPacketLength    int
		expectedTTL             int
	}{
		{
			line:                    "SRC=192.0.2.1 DST=192.0.2.2 PROTO=TCP SPT=12345 DPT=80 action=DROP reason=TEST LEN=512 TTL=128",
			expectedSourceIP:        "192.0.2.1",
			expectedDestinationIP:   "192.0.2.2",
			expectedSourcePort:      12345,
			expectedDestinationPort: 80,
			expectedProtocol:        "TCP",
			expectedAction:          "DROP",
			expectedReason:          "TEST",
			expectedPacketLength:    512,
			expectedTTL:             128,
		},
		{
			line:                    "PROTO=UDP SPT=12345 DPT=443",
			expectedSourceIP:        "unknown",
			expectedDestinationIP:   "unknown",
			expectedSourcePort:      12345,
			expectedDestinationPort: 443,
			expectedProtocol:        "UDP",
			expectedAction:          "unknown",
			expectedReason:          "unknown",
			expectedPacketLength:    0,
			expectedTTL:             0,
		},
		{
			line:                    "Invalid log entry",
			expectedSourceIP:        "unknown",
			expectedDestinationIP:   "unknown",
			expectedSourcePort:      0,
			expectedDestinationPort: 0,
			expectedProtocol:        "unknown",
			expectedAction:          "unknown",
			expectedReason:          "unknown",
			expectedPacketLength:    0,
			expectedTTL:             0,
		},
	}

	for _, test := range tests {
		_, srcIP, dstIP, spt, dpt, proto, action, reason, packetLength, ttl := ExtractFields(test.line)

		if srcIP != test.expectedSourceIP {
			t.Errorf("Expected SRC IP %q, got %q", test.expectedSourceIP, srcIP)
		}
		if dstIP != test.expectedDestinationIP {
			t.Errorf("Expected DST IP %q, got %q", test.expectedDestinationIP, dstIP)
		}
		if spt != test.expectedSourcePort {
			t.Errorf("Expected SRC port %d, got %d", test.expectedSourcePort, spt)
		}
		if dpt != test.expectedDestinationPort {
			t.Errorf("Expected DST port %d, got %d", test.expectedDestinationPort, dpt)
		}
		if proto != test.expectedProtocol {
			t.Errorf("Expected protocol %q, got %q", test.expectedProtocol, proto)
		}
		if action != test.expectedAction {
			t.Errorf("Expected action %q, got %q", test.expectedAction, action)
		}
		if reason != test.expectedReason {
			t.Errorf("Expected reason %q, got %q", test.expectedReason, reason)
		}
		if packetLength != test.expectedPacketLength {
			t.Errorf("Expected packet length %d, got %d", test.expectedPacketLength, packetLength)
		}
		if ttl != test.expectedTTL {
			t.Errorf("Expected TTL %d, got %d", test.expectedTTL, ttl)
		}
	}
}

func TestExtractFields_MissingFields(t *testing.T) {
	tests := []struct {
		line                    string
		expectedSourceIP        string
		expectedDestinationIP   string
		expectedSourcePort      int
		expectedDestinationPort int
		expectedProtocol        string
		expectedAction          string
		expectedReason          string
		expectedPacketLength    int
		expectedTTL             int
	}{
		{
			line:                    "DST=192.0.2.2 PROTO=TCP SPT=12345 DPT=80",
			expectedSourceIP:        "unknown",
			expectedDestinationIP:   "192.0.2.2",
			expectedSourcePort:      12345,
			expectedDestinationPort: 80,
			expectedProtocol:        "TCP",
			expectedAction:          "unknown",
			expectedReason:          "unknown",
			expectedPacketLength:    0,
			expectedTTL:             0,
		},
		{
			line:                    "SRC=192.0.2.1 PROTO=UDP DPT=443",
			expectedSourceIP:        "192.0.2.1",
			expectedDestinationIP:   "unknown",
			expectedSourcePort:      0,
			expectedDestinationPort: 443,
			expectedProtocol:        "UDP",
			expectedAction:          "unknown",
			expectedReason:          "unknown",
			expectedPacketLength:    0,
			expectedTTL:             0,
		},
		{
			line:                    "SRC=192.0.2.1 DST=192.0.2.2 SPT=54321",
			expectedSourceIP:        "192.0.2.1",
			expectedDestinationIP:   "192.0.2.2",
			expectedSourcePort:      54321,
			expectedDestinationPort: 0,
			expectedProtocol:        "unknown",
			expectedAction:          "unknown",
			expectedReason:          "unknown",
			expectedPacketLength:    0,
			expectedTTL:             0,
		},
	}

	for _, test := range tests {
		_, srcIP, dstIP, spt, dpt, proto, action, reason, packetLength, ttl := ExtractFields(test.line)

		if srcIP != test.expectedSourceIP {
			t.Errorf("Expected SRC IP %q, got %q", test.expectedSourceIP, srcIP)
		}
		if dstIP != test.expectedDestinationIP {
			t.Errorf("Expected DST IP %q, got %q", test.expectedDestinationIP, dstIP)
		}
		if spt != test.expectedSourcePort {
			t.Errorf("Expected SRC port %d, got %d", test.expectedSourcePort, spt)
		}
		if dpt != test.expectedDestinationPort {
			t.Errorf("Expected DST port %d, got %d", test.expectedDestinationPort, dpt)
		}
		if proto != test.expectedProtocol {
			t.Errorf("Expected protocol %q, got %q", test.expectedProtocol, proto)
		}
		if action != test.expectedAction {
			t.Errorf("Expected action %q, got %q", test.expectedAction, action)
		}
		if reason != test.expectedReason {
			t.Errorf("Expected reason %q, got %q", test.expectedReason, reason)
		}
		if packetLength != test.expectedPacketLength {
			t.Errorf("Expected packet length %d, got %d", test.expectedPacketLength, packetLength)
		}
		if ttl != test.expectedTTL {
			t.Errorf("Expected TTL %d, got %d", test.expectedTTL, ttl)
		}
	}
}

func TestExtractFields_IPv6(t *testing.T) {
	tests := []struct {
		line                  string
		expectedSourceIP      string
		expectedDestinationIP string
		expectedProtocol      string
		expectedAction        string
		expectedReason        string
		expectedPacketLength  int
		expectedTTL           int
	}{
		{
			line:                  "SRC=2001:0db8::1 DST=2001:0db8::2 PROTO=TCP action=DROP reason=PORTSCAN LEN=400 TTL=64",
			expectedSourceIP:      "2001:0db8::1",
			expectedDestinationIP: "2001:0db8::2",
			expectedProtocol:      "TCP",
			expectedAction:        "DROP",
			expectedReason:        "PORTSCAN",
			expectedPacketLength:  400,
			expectedTTL:           64,
		},
		{
			line:                  "SRC=INVALID_IP DST=2001:0db8::1 PROTO=UDP",
			expectedSourceIP:      "unknown",
			expectedDestinationIP: "2001:0db8::1",
			expectedProtocol:      "UDP",
			expectedAction:        "unknown",
			expectedReason:        "unknown",
			expectedPacketLength:  0,
			expectedTTL:           0,
		},
	}

	for _, test := range tests {
		_, srcIP, dstIP, _, _, proto, action, reason, packetLength, ttl := ExtractFields(test.line)

		if srcIP != test.expectedSourceIP {
			t.Errorf("Expected SRC IP %q, got %q", test.expectedSourceIP, srcIP)
		}
		if dstIP != test.expectedDestinationIP {
			t.Errorf("Expected DST IP %q, got %q", test.expectedDestinationIP, dstIP)
		}
		if proto != test.expectedProtocol {
			t.Errorf("Expected protocol %q, got %q", test.expectedProtocol, proto)
		}
		if action != test.expectedAction {
			t.Errorf("Expected action %q, got %q", test.expectedAction, action)
		}
		if reason != test.expectedReason {
			t.Errorf("Expected reason %q, got %q", test.expectedReason, reason)
		}
		if packetLength != test.expectedPacketLength {
			t.Errorf("Expected packet length %d, got %d", test.expectedPacketLength, packetLength)
		}
		if ttl != test.expectedTTL {
			t.Errorf("Expected TTL %d, got %d", test.expectedTTL, ttl)
		}
	}
}
