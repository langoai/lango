package prompt

import (
	"sort"
	"strings"
)

// Builder assembles prompt sections into a single system prompt string.
type Builder struct {
	sections []PromptSection
}

// NewBuilder creates an empty Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Add appends a section. If a section with the same ID already exists,
// it is replaced (last-writer-wins).
func (b *Builder) Add(s PromptSection) *Builder {
	for i, existing := range b.sections {
		if existing.ID() == s.ID() {
			b.sections[i] = s
			return b
		}
	}
	b.sections = append(b.sections, s)
	return b
}

// Remove removes a section by ID.
func (b *Builder) Remove(id SectionID) *Builder {
	for i, s := range b.sections {
		if s.ID() == id {
			b.sections = append(b.sections[:i], b.sections[i+1:]...)
			return b
		}
	}
	return b
}

// Has returns true if a section with the given ID is registered.
func (b *Builder) Has(id SectionID) bool {
	for _, s := range b.sections {
		if s.ID() == id {
			return true
		}
	}
	return false
}

// Clone returns a deep copy of the builder so the caller can diverge
// independently (e.g. per sub-agent customization).
func (b *Builder) Clone() *Builder {
	clone := &Builder{sections: make([]PromptSection, len(b.sections))}
	copy(clone.sections, b.sections)
	return clone
}

// Build sorts sections by priority and joins their rendered content.
func (b *Builder) Build() string {
	sorted := make([]PromptSection, len(b.sections))
	copy(sorted, b.sections)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority() < sorted[j].Priority()
	})

	var parts []string
	for _, s := range sorted {
		rendered := s.Render()
		if rendered != "" {
			parts = append(parts, rendered)
		}
	}
	return strings.Join(parts, "\n\n")
}
