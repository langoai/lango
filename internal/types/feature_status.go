package types

// FeatureStatus describes the initialization state of a subsystem feature.
// Used by doctor checks, status command, and TUI for structured diagnostics.
type FeatureStatus struct {
	// Name is the human-readable feature name (e.g., "Embedding & RAG").
	Name string `json:"name"`
	// Enabled indicates whether the feature is active.
	Enabled bool `json:"enabled"`
	// Healthy indicates whether the feature initialized without errors.
	Healthy bool `json:"healthy"`
	// AutoEnabled indicates the feature was auto-enabled (not explicitly configured).
	AutoEnabled bool `json:"autoEnabled,omitempty"`
	// Reason explains why the feature is disabled or unhealthy (empty when OK).
	Reason string `json:"reason,omitempty"`
	// Suggestion provides an actionable next step for the user (empty when OK).
	Suggestion string `json:"suggestion,omitempty"`
}
