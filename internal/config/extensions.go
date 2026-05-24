package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultExtensionsDir is the default on-disk root for installed extension
// packs. Tilde expansion to the user's home is performed by ResolvedDir at
// consumption time — the raw value stays portable across hosts.
const DefaultExtensionsDir = "~/.lango/extensions"

// ResolveExtensions returns a copy with defaults applied. A nil Enabled
// pointer becomes a pointer to true; an empty Dir becomes DefaultExtensionsDir.
// The receiver is not mutated.
func (c ExtensionsConfig) ResolveExtensions() ExtensionsConfig {
	out := c
	if out.Enabled == nil {
		t := true
		out.Enabled = &t
	}
	if out.Dir == "" {
		out.Dir = DefaultExtensionsDir
	}
	return out
}

// ResolvedDir expands a leading "~/" in Dir to the current user's home
// directory. Consumers that read files should use this rather than Dir
// directly. On lookup failure, the raw Dir is returned.
func (c ExtensionsConfig) ResolvedDir() string {
	dir := c.Dir
	if dir == "" {
		dir = DefaultExtensionsDir
	}
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return dir
		}
		return filepath.Join(home, dir[2:])
	}
	return dir
}

// IsEnabled returns true when the subsystem is enabled. A nil Enabled
// pointer means "default true."
func (c ExtensionsConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}
