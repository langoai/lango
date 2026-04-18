package app

import (
	"path/filepath"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

func resolvedProtectedPaths(cfg *config.Config, boot *bootstrap.Result) []string {
	if cfg == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var paths []string
	add := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		paths = append(paths, clean)
	}

	add(cfg.DataRoot)
	add(cfg.Session.DatabasePath)

	graphPath := cfg.Graph.DatabasePath
	if graphPath == "" {
		if cfg.Session.DatabasePath != "" {
			graphPath = filepath.Join(filepath.Dir(cfg.Session.DatabasePath), "graph.db")
		} else if cfg.DataRoot != "" {
			graphPath = filepath.Join(cfg.DataRoot, "graph.db")
		}
	}
	add(graphPath)

	langoDir := cfg.DataRoot
	if boot != nil && boot.LangoDir != "" {
		langoDir = boot.LangoDir
	}
	if langoDir != "" {
		add(filepath.Join(langoDir, "envelope.json"))
		add(filepath.Join(langoDir, "keyfile"))
	}

	for _, path := range cfg.Tools.Exec.AdditionalProtectedPaths {
		add(path)
	}

	return paths
}
