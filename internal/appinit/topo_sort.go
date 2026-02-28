package appinit

import (
	"fmt"
	"strings"
)

// TopoSort returns modules in dependency order. Disabled modules are excluded.
// If module A depends on key K and module B provides K, B appears before A.
// Dependencies on keys not provided by any enabled module are silently ignored.
// Returns an error if a dependency cycle is detected.
func TopoSort(modules []Module) ([]Module, error) {
	// Filter to enabled modules only.
	enabled := make([]Module, 0, len(modules))
	for _, m := range modules {
		if m.Enabled() {
			enabled = append(enabled, m)
		}
	}

	if len(enabled) == 0 {
		return nil, nil
	}

	// Build a map from provides-key to the module that provides it.
	provider := make(map[Provides]Module, len(enabled))
	for _, m := range enabled {
		for _, p := range m.Provides() {
			provider[p] = m
		}
	}

	// Build adjacency list: module name -> set of module names it depends on.
	// An edge from A to B means "A depends on B" (B must come first).
	type nameSet = map[string]struct{}
	deps := make(map[string]nameSet, len(enabled))
	byName := make(map[string]Module, len(enabled))

	for _, m := range enabled {
		byName[m.Name()] = m
		deps[m.Name()] = make(nameSet)
	}

	for _, m := range enabled {
		for _, key := range m.DependsOn() {
			prov, ok := provider[key]
			if !ok {
				// No enabled module provides this key; skip.
				continue
			}
			if prov.Name() == m.Name() {
				// Self-dependency; skip.
				continue
			}
			deps[m.Name()][prov.Name()] = struct{}{}
		}
	}

	// Kahn's algorithm for topological sort.
	inDegree := make(map[string]int, len(enabled))
	for _, m := range enabled {
		inDegree[m.Name()] = len(deps[m.Name()])
	}

	queue := make([]string, 0, len(enabled))
	for _, m := range enabled {
		if inDegree[m.Name()] == 0 {
			queue = append(queue, m.Name())
		}
	}

	sorted := make([]Module, 0, len(enabled))
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		sorted = append(sorted, byName[name])

		// For each module that depends on the current one, decrement in-degree.
		for _, m := range enabled {
			if _, ok := deps[m.Name()][name]; ok {
				inDegree[m.Name()]--
				if inDegree[m.Name()] == 0 {
					queue = append(queue, m.Name())
				}
			}
		}
	}

	if len(sorted) != len(enabled) {
		// Cycle detected — report the modules involved.
		cycleMembers := make([]string, 0)
		for _, m := range enabled {
			if inDegree[m.Name()] > 0 {
				cycleMembers = append(cycleMembers, m.Name())
			}
		}
		return nil, fmt.Errorf("dependency cycle detected among modules: [%s]",
			strings.Join(cycleMembers, ", "))
	}

	return sorted, nil
}
