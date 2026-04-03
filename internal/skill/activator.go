package skill

import (
	"context"
	"fmt"
	"path/filepath"
)

// Activator checks which skills should activate based on file paths being edited.
type Activator struct {
	registry *Registry
}

// NewActivator creates an Activator that queries the given registry for path-based skill activation.
func NewActivator(registry *Registry) *Activator {
	return &Activator{registry: registry}
}

// CheckPaths returns skills whose Paths globs match any of the given edited paths.
// This is a pure query — does not change skill status.
func (a *Activator) CheckPaths(ctx context.Context, editedPaths []string) ([]*SkillEntry, error) {
	if len(editedPaths) == 0 {
		return nil, nil
	}

	skills, err := a.registry.ListActiveSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active skills: %w", err)
	}

	var matched []*SkillEntry
	for i := range skills {
		sk := &skills[i]
		if len(sk.Paths) == 0 {
			continue
		}
		if matchesAny(sk.Paths, editedPaths) {
			matched = append(matched, sk)
		}
	}

	return matched, nil
}

// matchesAny reports whether any glob pattern matches any edited path.
func matchesAny(globs []string, paths []string) bool {
	for _, g := range globs {
		for _, p := range paths {
			ok, err := filepath.Match(g, p)
			if err != nil {
				// Malformed glob pattern — skip it.
				continue
			}
			if ok {
				return true
			}
		}
	}
	return false
}
