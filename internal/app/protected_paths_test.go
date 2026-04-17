package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

func TestResolvedProtectedPaths(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{
		DataRoot: root,
		Session: config.SessionConfig{
			DatabasePath: filepath.Join(root, "lango.db"),
		},
		Tools: config.ToolsConfig{
			Exec: config.ExecToolConfig{
				AdditionalProtectedPaths: []string{filepath.Join(root, "custom")},
			},
		},
	}

	got := resolvedProtectedPaths(cfg, &bootstrap.Result{LangoDir: root})

	assert.Contains(t, got, filepath.Join(root, "lango.db"))
	assert.Contains(t, got, filepath.Join(root, "graph.db"))
	assert.Contains(t, got, filepath.Join(root, "envelope.json"))
	assert.Contains(t, got, filepath.Join(root, "keyfile"))
	assert.Contains(t, got, filepath.Join(root, "custom"))
}

func TestResolvedProtectedPaths_GraphOverride(t *testing.T) {
	root := t.TempDir()
	graphPath := filepath.Join(t.TempDir(), "custom-graph.db")
	cfg := &config.Config{
		DataRoot: root,
		Session: config.SessionConfig{
			DatabasePath: filepath.Join(root, "lango.db"),
		},
		Graph: config.GraphConfig{
			DatabasePath: graphPath,
		},
	}

	got := resolvedProtectedPaths(cfg, nil)
	assert.Contains(t, got, graphPath)
	assert.NotContains(t, got, filepath.Join(root, "graph.db"))
}
