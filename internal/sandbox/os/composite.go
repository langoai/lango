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
