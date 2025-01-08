package input

import (
	"strings"
	"testing"
)

func TestReadLines(t *testing.T) {
	data := "line1\nline2\nline3"
	reader := strings.NewReader(data)

	lines, err := ReadLines(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"line1", "line2", "line3"}
	if len(lines) != len(expected) {
		t.Fatalf("Expected %d lines, got %d", len(expected), len(lines))
	}

	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("Expected line %d to be %q, got %q", i, expected[i], line)
		}
	}
}

func TestReverseLines(t *testing.T) {
	lines := []string{"line1", "line2", "line3"}
	reversed := ReverseLines(lines)

	expected := []string{"line3", "line2", "line1"}
	for i, line := range reversed {
		if line != expected[i] {
			t.Errorf("Expected line %d to be %q, got %q", i, expected[i], line)
		}
	}
}
