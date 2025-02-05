package main

import (
	"io"
	"os"
	"testing"
)

func TestEndToEndPipeline(t *testing.T) {
	// Prepare mock input log data.
	mockInput := `
action=DROP reason=PORTSCAN SRC=192.0.2.1 DST=192.0.2.2 PROTO=TCP SPT=12345 DPT=80
action=DROP reason=INTRUSION SRC=203.0.113.5 DST=198.51.100.1 PROTO=UDP SPT=54321 DPT=443
`
	// Create a temporary file to simulate standard input.
	tempInput, err := os.CreateTemp("", "mock_stdin")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempInput.Name())

	_, err = io.WriteString(tempInput, mockInput)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempInput.Seek(0, io.SeekStart)

	// Replace os.Stdin with the temporary file.
	oldStdin := os.Stdin
	os.Stdin = tempInput
	defer func() { os.Stdin = oldStdin }()

	// Create a temporary file to capture standard output.
	tempOutput, err := os.CreateTemp("", "mock_stdout")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempOutput.Name())

	oldStdout := os.Stdout
	os.Stdout = tempOutput
	defer func() { os.Stdout = oldStdout }()

	// Run the main function.
	main()

	// Validate that output is produced.
	tempOutput.Seek(0, io.SeekStart)
	output, err := io.ReadAll(tempOutput)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if len(output) == 0 {
		t.Fatalf("Expected JSON output, got empty output")
	}
}
