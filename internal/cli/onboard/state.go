package onboard

import (
	"github.com/langowarny/lango/internal/config"
)

// ConfigState holds the current configuration state and tracking for dirty fields.
type ConfigState struct {
	Current *config.Config
	Dirty   map[string]bool
}

// NewConfigState creates a new state with default config.
func NewConfigState() *ConfigState {
	return NewConfigStateWith(config.DefaultConfig())
}

// NewConfigStateWith creates a new state with the given config.
func NewConfigStateWith(cfg *config.Config) *ConfigState {
	return &ConfigState{
		Current: cfg,
		Dirty:   make(map[string]bool),
	}
}

// MarkDirty marks a field as modified.
func (s *ConfigState) MarkDirty(field string) {
	s.Dirty[field] = true
}

// IsDirty checks if a field is modified.
func (s *ConfigState) IsDirty(field string) bool {
	return s.Dirty[field]
}

// UpdateField updates a field using reflection (simplified for now).
// In a real implementation, we might want type-safe updates per form.
func (s *ConfigState) UpdateField(path string, value interface{}) {
	// TODO: Implement reflection-based update or explicit setters
	s.MarkDirty(path)
}
