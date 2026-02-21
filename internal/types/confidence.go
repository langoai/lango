package types

// Confidence represents a confidence level for analysis results.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// Valid reports whether c is a known confidence level.
func (c Confidence) Valid() bool {
	switch c {
	case ConfidenceHigh, ConfidenceMedium, ConfidenceLow:
		return true
	}
	return false
}

// Values returns all known confidence levels.
func (c Confidence) Values() []Confidence {
	return []Confidence{ConfidenceHigh, ConfidenceMedium, ConfidenceLow}
}
