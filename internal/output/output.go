package output

import (
	"encoding/json"
	"io"
)

// WriteJSONOutput writes the given data as JSON to the provided writer.
func WriteJSONOutput(data interface{}, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	return encoder.Encode(data)
}
