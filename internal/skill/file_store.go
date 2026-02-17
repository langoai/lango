package skill

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

var _ SkillStore = (*FileSkillStore)(nil)

// FileSkillStore implements SkillStore using .lango/skills/<name>/SKILL.md files.
type FileSkillStore struct {
	dir    string
	logger *zap.SugaredLogger
}

// NewFileSkillStore creates a new file-based skill store rooted at dir.
func NewFileSkillStore(dir string, logger *zap.SugaredLogger) *FileSkillStore {
	return &FileSkillStore{dir: dir, logger: logger}
}

// Save creates or overwrites a skill's SKILL.md file.
func (s *FileSkillStore) Save(_ context.Context, entry SkillEntry) error {
	if entry.Name == "" {
		return fmt.Errorf("skill name is required")
	}

	if entry.Status == "" {
		entry.Status = "draft"
	}

	data, err := RenderSkillMD(&entry)
	if err != nil {
		return fmt.Errorf("render skill %q: %w", entry.Name, err)
	}

	dir := filepath.Join(s.dir, entry.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create skill dir %q: %w", dir, err)
	}

	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write skill file %q: %w", path, err)
	}

	s.logger.Debugw("skill saved", "name", entry.Name, "path", path)
	return nil
}

// Get reads and parses a skill's SKILL.md file.
func (s *FileSkillStore) Get(_ context.Context, name string) (*SkillEntry, error) {
	path := filepath.Join(s.dir, name, "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("skill not found: %s", name)
		}
		return nil, fmt.Errorf("read skill %q: %w", name, err)
	}

	entry, err := ParseSkillMD(data)
	if err != nil {
		return nil, fmt.Errorf("parse skill %q: %w", name, err)
	}

	return entry, nil
}

// ListActive scans all skill directories and returns entries with status=active.
func (s *FileSkillStore) ListActive(_ context.Context) ([]SkillEntry, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list skills dir: %w", err)
	}

	var result []SkillEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		path := filepath.Join(s.dir, e.Name(), "SKILL.md")
		data, err := os.ReadFile(path)
		if err != nil {
			s.logger.Debugw("skip skill dir (no SKILL.md)", "dir", e.Name())
			continue
		}

		entry, err := ParseSkillMD(data)
		if err != nil {
			s.logger.Warnw("skip invalid skill", "dir", e.Name(), "error", err)
			continue
		}

		if entry.Status == "active" {
			result = append(result, *entry)
		}
	}

	return result, nil
}

// Activate sets a skill's status to active by rewriting its SKILL.md.
func (s *FileSkillStore) Activate(ctx context.Context, name string) error {
	entry, err := s.Get(ctx, name)
	if err != nil {
		return err
	}

	entry.Status = "active"
	return s.Save(ctx, *entry)
}

// Delete removes a skill's directory entirely.
func (s *FileSkillStore) Delete(_ context.Context, name string) error {
	dir := filepath.Join(s.dir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("skill not found: %s", name)
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete skill %q: %w", name, err)
	}

	s.logger.Debugw("skill deleted", "name", name)
	return nil
}

// EnsureDefaults deploys embedded default skills that don't already exist.
func (s *FileSkillStore) EnsureDefaults(defaultFS fs.FS) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("ensure skills dir: %w", err)
	}

	return fs.WalkDir(defaultFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Base(path) != "SKILL.md" {
			return nil
		}

		// path is like "serve/SKILL.md" — extract skill name from parent dir.
		skillName := filepath.Dir(path)
		if skillName == "." {
			return nil
		}

		targetDir := filepath.Join(s.dir, skillName)
		targetPath := filepath.Join(targetDir, "SKILL.md")

		// Skip if already exists (user may have customized).
		if _, err := os.Stat(targetPath); err == nil {
			return nil
		}

		data, err := fs.ReadFile(defaultFS, path)
		if err != nil {
			s.logger.Warnw("read embedded skill", "path", path, "error", err)
			return nil
		}

		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return fmt.Errorf("create default skill dir %q: %w", targetDir, err)
		}

		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			return fmt.Errorf("write default skill %q: %w", targetPath, err)
		}

		s.logger.Debugw("deployed default skill", "name", skillName)
		return nil
	})
}
