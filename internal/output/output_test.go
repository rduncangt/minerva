package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteJSONOutput(t *testing.T) {
	// Mock data for testing
	data := []map[string]string{
		{"key1": "value1"},
		{"key2": "value2"},
	}

	// Use a buffer to capture output
	var buf bytes.Buffer
	err := WriteJSONOutput(data, &buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Parse the output to validate its structure
	var result []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Compare the result with the original data
	if len(result) != len(data) {
		t.Fatalf("Expected %d items, got %d", len(data), len(result))
	}
	for i, item := range result {
		for key, value := range data[i] {
			if item[key] != value {
				t.Errorf("Expected %q: %q, got %q", key, value, item[key])
			}
		}
	}
}
