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

func TestReadLines_Empty(t *testing.T) {
	reader := strings.NewReader("")
	lines, err := ReadLines(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("Expected 0 lines, got %d", len(lines))
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

func TestReverseLines_Empty(t *testing.T) {
	var lines []string
	reversed := ReverseLines(lines)
	if len(reversed) != 0 {
		t.Errorf("Expected empty slice, got %v", reversed)
	}
}

func TestReverseLines_SingleElement(t *testing.T) {
	lines := []string{"only"}
	reversed := ReverseLines(lines)
	expected := []string{"only"}
	if len(reversed) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(reversed))
	}
	if reversed[0] != expected[0] {
		t.Errorf("Expected element %q, got %q", expected[0], reversed[0])
	}
}
