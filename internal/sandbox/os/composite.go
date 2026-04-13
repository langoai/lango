//go:build linux

package os

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// compositeIsolator applies multiple isolators in sequence (e.g., Landlock + seccomp on Linux).
type compositeIsolator struct {
	isolators []OSIsolator
}

// Compile-time interface compliance check.
var _ OSIsolator = (*compositeIsolator)(nil)

// NewCompositeIsolator returns an isolator that applies multiple isolators in sequence.
func NewCompositeIsolator(isolators ...OSIsolator) OSIsolator {
	return &compositeIsolator{isolators: isolators}
}

func (c *compositeIsolator) Apply(ctx context.Context, cmd *exec.Cmd, policy Policy) error {
	for _, iso := range c.isolators {
		if err := iso.Apply(ctx, cmd, policy); err != nil {
			return fmt.Errorf("apply %s: %w", iso.Name(), err)
		}
	}
	return nil
}

func (c *compositeIsolator) Available() bool {
	for _, iso := range c.isolators {
		if iso.Available() {
			return true
		}
	}
	return false
}

func (c *compositeIsolator) Name() string {
	names := make([]string, 0, len(c.isolators))
	for _, iso := range c.isolators {
		names = append(names, iso.Name())
	}
	return strings.Join(names, "+")
}

func (c *compositeIsolator) Reason() string {
	var reasons []string
	for _, iso := range c.isolators {
		if r := iso.Reason(); r != "" {
			reasons = append(reasons, iso.Name()+": "+r)
		}
	}
	return strings.Join(reasons, "; ")
}
