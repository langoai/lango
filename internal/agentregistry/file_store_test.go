package agentregistry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_Load(t *testing.T) {
	tests := []struct {
		give     string
		setup    func(t *testing.T, dir string)
		wantLen  int
		wantName string
		wantErr  bool
	}{
		{
			give: "valid AGENT.md in subdirectory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				agentDir := filepath.Join(dir, "test-agent")
				require.NoError(t, os.MkdirAll(agentDir, 0o755))
				content := []byte("---\nname: test-agent\ndescription: A test agent\n---\n\nTest instructions.")
				require.NoError(t, os.WriteFile(filepath.Join(agentDir, "AGENT.md"), content, 0o644))
			},
			wantLen:  1,
			wantName: "test-agent",
		},
		{
			give: "skip directories without AGENT.md",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Directory with AGENT.md
				agentDir := filepath.Join(dir, "valid-agent")
				require.NoError(t, os.MkdirAll(agentDir, 0o755))
				content := []byte("---\nname: valid-agent\n---\n\nInstructions.")
				require.NoError(t, os.WriteFile(filepath.Join(agentDir, "AGENT.md"), content, 0o644))

				// Directory without AGENT.md
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "no-agent"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "no-agent", "README.md"), []byte("not an agent"), 0o644))
			},
			wantLen:  1,
			wantName: "valid-agent",
		},
		{
			give:    "empty directory",
			setup:   func(t *testing.T, dir string) { t.Helper() },
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			store := NewFileStore(dir)
			defs, err := store.Load()

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, defs, tt.wantLen)

			if tt.wantName != "" && len(defs) > 0 {
				assert.Equal(t, tt.wantName, defs[0].Name)
				assert.Equal(t, SourceUser, defs[0].Source)
			}
		})
	}
}

func TestFileStore_Load_NonexistentDir(t *testing.T) {
	store := NewFileStore("/nonexistent/path/to/agents")
	defs, err := store.Load()
	require.NoError(t, err)
	assert.Nil(t, defs)
}
