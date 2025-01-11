package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

// WriteJSONOutput writes the given data as JSON to the provided writer.
func WriteJSONOutput(data interface{}, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	return encoder.Encode(data)
}

// WriteIPSummaryTable writes an IP summary table to the provided writer.
func WriteIPSummaryTable(summary []map[string]interface{}, w io.Writer) error {
	// Initialize tab writer for aligned columns
	writer := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)

	// Write the headers
	fmt.Fprintln(writer, "Date\tSource IP\tFrequency\tPort(s) Targeted\tLog Level\tAction Taken\tGeolocation\tNotes")

	// Write the rows
	for _, row := range summary {
		fmt.Fprintf(writer, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			row["date"],           // Date
			row["source_ip"],      // Source IP
			row["frequency"],      // Frequency
			row["ports_targeted"], // Port(s) Targeted
			row["log_level"],      // Log Level
			row["action_taken"],   // Action Taken
			row["geolocation"],    // Geolocation
			row["notes"],          // Notes
		)
	}

	// Flush the tab writer
	return writer.Flush()
}
