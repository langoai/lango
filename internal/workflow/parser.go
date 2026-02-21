package workflow

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// validAgents is the set of recognized agent names.
var validAgents = map[string]bool{
	"executor":       true,
	"researcher":     true,
	"planner":        true,
	"memory-manager": true,
}

// Parse parses YAML data into a Workflow.
func Parse(data []byte) (*Workflow, error) {
	var w Workflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("parse workflow YAML: %w", err)
	}
	if err := Validate(&w); err != nil {
		return nil, fmt.Errorf("validate workflow: %w", err)
	}
	return &w, nil
}

// ParseFile reads a YAML file and parses it into a Workflow.
func ParseFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file %q: %w", path, err)
	}
	return Parse(data)
}

// Validate checks that a Workflow is well-formed.
func Validate(w *Workflow) error {
	if w.Name == "" {
		return ErrWorkflowNameEmpty
	}
	if len(w.Steps) == 0 {
		return ErrNoWorkflowSteps
	}

	// Check step ID uniqueness and build lookup.
	seen := make(map[string]bool, len(w.Steps))
	for _, s := range w.Steps {
		if s.ID == "" {
			return ErrStepIDEmpty
		}
		if seen[s.ID] {
			return fmt.Errorf("duplicate step id %q", s.ID)
		}
		seen[s.ID] = true
	}

	// Validate depends_on references and agent names.
	for _, s := range w.Steps {
		for _, dep := range s.DependsOn {
			if !seen[dep] {
				return fmt.Errorf("step %q depends on unknown step %q", s.ID, dep)
			}
		}
		if s.Agent != "" && !validAgents[s.Agent] {
			return fmt.Errorf("step %q has unknown agent %q", s.ID, s.Agent)
		}
	}

	// Cycle detection using DFS.
	if err := detectCycles(w.Steps); err != nil {
		return err
	}

	return nil
}

// detectCycles performs DFS-based cycle detection on the step dependency graph.
func detectCycles(steps []Step) error {
	const (
		white = 0 // unvisited
		gray  = 1 // in current DFS path
		black = 2 // fully processed
	)

	// Build adjacency: step -> steps that depend on it (children in dependency graph).
	// For cycle detection we traverse the depends_on edges.
	adj := make(map[string][]string, len(steps))
	for _, s := range steps {
		for _, dep := range s.DependsOn {
			adj[s.ID] = append(adj[s.ID], dep)
		}
	}

	color := make(map[string]int, len(steps))
	for _, s := range steps {
		color[s.ID] = white
	}

	var visit func(id string) error
	visit = func(id string) error {
		color[id] = gray
		for _, dep := range adj[id] {
			switch color[dep] {
			case gray:
				return fmt.Errorf("circular dependency detected involving step %q", dep)
			case white:
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		color[id] = black
		return nil
	}

	for _, s := range steps {
		if color[s.ID] == white {
			if err := visit(s.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
