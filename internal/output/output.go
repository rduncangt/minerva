package output

import (
	"encoding/json"
	"log"
	"os"
)

// WriteJSONOutput writes results as a JSON array to stdout.
func WriteJSONOutput(results interface{}) {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate JSON output: %v", err)
	}
	os.Stdout.Write(jsonData)
}
