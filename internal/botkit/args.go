package botkit

import (
	"encoding/json"
)

// ParseJSON parses a JSON-encoded string into a specified generic type.
// It takes a JSON string as input, attempts to unmarshal it into an instance of the specified type T,
// and returns the parsed value or an error if the parsing fails.
//
// Type Parameters:
//   - T: The type into which the JSON string should be unmarshaled.
//
// Parameters:
//   - src: The JSON-encoded string to be parsed.
//
// Returns:
//   - T: The parsed value of type T.
//   - error: An error if JSON unmarshaling fails or if the input is invalid.
func ParseJSON[T any](src string) (T, error) {
	var args T

	if err := json.Unmarshal([]byte(src), &args); err != nil {
		return *(new(T)), err
	}

	return args, nil
}
