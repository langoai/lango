package storeutil

import (
	"encoding/json"
	"fmt"
)

// MarshalField marshals v to json.RawMessage for store persistence.
// Returns an error so callers can abort the write rather than silently
// storing malformed JSON.
func MarshalField(v interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal field: %w", err)
	}
	return data, nil
}

// UnmarshalField unmarshals raw JSON data into the target pointer.
// Returns a wrapped error with context suitable for store methods.
func UnmarshalField(data []byte, target interface{}, context string) error {
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("unmarshal %s: %w", context, err)
	}
	return nil
}
