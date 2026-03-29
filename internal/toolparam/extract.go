package toolparam

import "fmt"

// ErrMissingParam indicates a required parameter was empty or absent.
type ErrMissingParam struct {
	Name string
}

func (e *ErrMissingParam) Error() string {
	return fmt.Sprintf("missing %s parameter", e.Name)
}

// RequireString extracts a required string parameter.
// Returns ErrMissingParam when the key is absent or the value is empty.
func RequireString(params map[string]interface{}, key string) (string, error) {
	v, _ := params[key].(string)
	if v == "" {
		return "", &ErrMissingParam{Name: key}
	}
	return v, nil
}

// OptionalString extracts an optional string parameter with a fallback.
func OptionalString(params map[string]interface{}, key, fallback string) string {
	if v, ok := params[key].(string); ok && v != "" {
		return v
	}
	return fallback
}

// OptionalInt extracts an optional integer parameter with a fallback.
// JSON numbers arrive as float64, so this handles the conversion.
func OptionalInt(params map[string]interface{}, key string, fallback int) int {
	if v, ok := params[key].(float64); ok {
		return int(v)
	}
	return fallback
}

// OptionalBool extracts an optional boolean parameter with a fallback.
func OptionalBool(params map[string]interface{}, key string, fallback bool) bool {
	if v, ok := params[key].(bool); ok {
		return v
	}
	return fallback
}

// RequireFloat64 extracts a required float64 parameter.
// Returns ErrMissingParam when the key is absent.
func RequireFloat64(params map[string]interface{}, key string) (float64, error) {
	v, ok := params[key].(float64)
	if !ok {
		return 0, &ErrMissingParam{Name: key}
	}
	return v, nil
}

// OptionalFloat64 extracts an optional float64 parameter with a fallback.
func OptionalFloat64(params map[string]interface{}, key string, fallback float64) float64 {
	if v, ok := params[key].(float64); ok {
		return v
	}
	return fallback
}

// StringSlice extracts a string slice from a parameter that arrives as []interface{}.
func StringSlice(params map[string]interface{}, key string) []string {
	raw, ok := params[key].([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
