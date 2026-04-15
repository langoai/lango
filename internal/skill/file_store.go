package skill

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
	if err := os.MkdirAll(dir, 0o700); err != nil {
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

// extPrefix reserves the "ext-" directory-name prefix for extension-pack-
// owned skill subtrees under the skills directory. The ListActive walker
// descends into each ext-<pack>/ dir as if it were the top level, but
// records SourcePack on each SkillEntry it loads.
const extPrefix = "ext-"

// ListActive scans all skill directories and returns entries with status=active.
// Directories whose names begin with "ext-" are treated as extension-owned
// skill roots: their direct children are each walked for SKILL.md, and the
// parent's "ext-<pack>" name (minus the prefix) is recorded on each loaded
// entry as SourcePack.
//
// Name precedence: user-authored skills (found directly under s.dir) shadow
// extension-authored skills with the same bare name. When both exist, the
// user entry is returned and a debug log is emitted for the shadowed pack
// entry.
func (s *FileSkillStore) ListActive(_ context.Context) ([]SkillEntry, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list skills dir: %w", err)
	}

	// Two-pass walk: collect user-authored first so we can detect shadowing
	// when extension-owned skills are added in the second pass.
	userByName := map[string]SkillEntry{}
	extByName := map[string][]SkillEntry{} // name → entries from different packs (collision detection)

	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if strings.HasPrefix(e.Name(), extPrefix) {
			packName := strings.TrimPrefix(e.Name(), extPrefix)
			s.walkExtPack(filepath.Join(s.dir, e.Name()), packName, extByName)
			continue
		}
		entry, ok := s.loadSkillDir(filepath.Join(s.dir, e.Name()), "")
		if !ok {
			continue
		}
		userByName[entry.Name] = entry
	}

	// Cross-extension collision detection: same skill name in two different
	// ext-<pack>/ subtrees is a runtime error (installer normally prevents
	// this, but manual edits or old filesystems may produce it).
	for name, sources := range extByName {
		if len(sources) > 1 {
			packs := make([]string, 0, len(sources))
			for _, e := range sources {
				packs = append(packs, e.SourcePack)
			}
			return nil, fmt.Errorf("skill name %q provided by multiple extension packs %v — resolve before continuing", name, packs)
		}
	}

	result := make([]SkillEntry, 0, len(userByName)+len(extByName))
	for _, e := range userByName {
		result = append(result, e)
	}
	for name, sources := range extByName {
		if _, shadowed := userByName[name]; shadowed {
			s.logger.Debugw("skill.name.shadowed_by_user",
				"name", name,
				"pack", sources[0].SourcePack,
			)
			continue
		}
		result = append(result, sources[0])
	}
	return result, nil
}

// walkExtPack scans <s.dir>/ext-<pack>/ and collects every active skill
// directly under it into extByName keyed by skill name.
func (s *FileSkillStore) walkExtPack(packDir, packName string, extByName map[string][]SkillEntry) {
	entries, err := os.ReadDir(packDir)
	if err != nil {
		s.logger.Warnw("skip ext pack dir", "dir", packDir, "error", err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		entry, ok := s.loadSkillDir(filepath.Join(packDir, e.Name()), packName)
		if !ok {
			continue
		}
		extByName[entry.Name] = append(extByName[entry.Name], entry)
	}
}

// loadSkillDir reads <dir>/SKILL.md, parses it, and returns the active
// SkillEntry. The second return is false when the dir should be skipped
// (missing SKILL.md, parse error, or inactive status). sourcePack is the
// pack attribution string or "" for non-extension skills.
func (s *FileSkillStore) loadSkillDir(dir, sourcePack string) (SkillEntry, bool) {
	path := filepath.Join(dir, "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		s.logger.Debugw("skip skill dir (no SKILL.md)", "dir", dir)
		return SkillEntry{}, false
	}
	entry, err := ParseSkillMD(data)
	if err != nil {
		s.logger.Warnw("skip invalid skill", "dir", dir, "error", err)
		return SkillEntry{}, false
	}
	if entry.Status != "active" {
		return SkillEntry{}, false
	}
	entry.SourcePack = sourcePack
	return *entry, true
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

// SaveResource writes a resource file under a skill's directory.
func (s *FileSkillStore) SaveResource(_ context.Context, skillName, relPath string, data []byte) error {
	path := filepath.Join(s.dir, skillName, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create resource dir: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// DiscoverProjectSkills scans projectRoot/.lango/skills/ for SKILL.md files.
// Returns discovered skills without modifying the store.
func (s *FileSkillStore) DiscoverProjectSkills(_ context.Context, projectRoot string) ([]SkillEntry, error) {
	skillsDir := filepath.Join(projectRoot, ".lango", "skills")

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read project skills dir: %w", err)
	}

	var result []SkillEntry
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}

		path := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		data, err := os.ReadFile(path)
		if err != nil {
			s.logger.Debugw("skip project skill dir (no SKILL.md)", "dir", e.Name())
			continue
		}

		entry, err := ParseSkillMD(data)
		if err != nil {
			s.logger.Warnw("skip invalid project skill", "dir", e.Name(), "error", err)
			continue
		}

		result = append(result, *entry)
	}

	return result, nil
}

// EnsureDefaults deploys embedded default skills that don't already exist.
func (s *FileSkillStore) EnsureDefaults(defaultFS fs.FS) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
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
		if skillName == "." || strings.HasPrefix(skillName, ".") {
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

		if err := os.MkdirAll(targetDir, 0o700); err != nil {
			return fmt.Errorf("create default skill dir %q: %w", targetDir, err)
		}

		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			return fmt.Errorf("write default skill %q: %w", targetPath, err)
		}

		s.logger.Debugw("deployed default skill", "name", skillName)
		return nil
	})
}
