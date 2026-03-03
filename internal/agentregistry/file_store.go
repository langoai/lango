package agentregistry

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileStore loads agent definitions from a directory of AGENT.md files.
// Expected structure: <dir>/<name>/AGENT.md
type FileStore struct {
	dir string
}

// NewFileStore creates a FileStore that reads from the given directory.
func NewFileStore(dir string) *FileStore {
	return &FileStore{dir: dir}
}

// Load reads all AGENT.md files from subdirectories and returns parsed definitions.
func (s *FileStore) Load() ([]*AgentDefinition, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read agents dir %q: %w", s.dir, err)
	}

	var defs []*AgentDefinition
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		mdPath := filepath.Join(s.dir, entry.Name(), "AGENT.md")
		data, err := os.ReadFile(mdPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %q: %w", mdPath, err)
		}

		def, err := ParseAgentMD(data)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", mdPath, err)
		}
		def.Source = SourceUser
		defs = append(defs, def)
	}

	return defs, nil
}
