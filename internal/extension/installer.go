package extension

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// extSkillPrefix is reserved for extension-owned skill subdirectories
// under the user's skillsDir. User-authored skills at the top level
// shadow ext-<pack>/<name>/ siblings with the same bare name.
const extSkillPrefix = "ext-"

// Installer executes the inspect/install/remove pipeline. It holds the
// destination paths and (for install-time collision detection) a snapshot
// of what's already installed.
type Installer struct {
	ExtensionsDir string
	SkillsDir     string
}

// InspectReport is the side-effect-free output of Inspect. It is both
// human-printable (via String) and machine-friendly (exported fields).
type InspectReport struct {
	Manifest       *Manifest
	ManifestSHA256 string
	FileHashes     map[string]string
	PlannedWrites  []string
	SkippedWrites  []string // informational: "tools/MCP/providers: not installed by v1"
	SourceRef      string
}

// Inspect reads the pack via the given Source, validates it, computes
// hashes, and produces a report. Inspect does not write to ExtensionsDir
// or SkillsDir. The caller must invoke WorkingCopy.Cleanup when done;
// Inspect does not own the WC lifecycle (that belongs to the CLI layer
// so it can reuse the WC for a subsequent Install without re-cloning).
func (i *Installer) Inspect(ctx context.Context, src Source) (*InspectReport, *WorkingCopy, error) {
	wc, err := src.Fetch(ctx)
	if err != nil {
		return nil, nil, err
	}
	report := &InspectReport{
		Manifest:       wc.Manifest,
		ManifestSHA256: wc.ManifestSHA256,
		FileHashes:     wc.FileHashes,
		SourceRef:      wc.SourceRef,
		PlannedWrites:  i.plannedWrites(wc.Manifest),
		SkippedWrites: []string{
			"tools:     not supported in v1 packs",
			"mcp:       not supported in v1 packs",
			"providers: not supported in v1 packs",
		},
	}
	return report, wc, nil
}

// InstallOptions controls install behavior.
type InstallOptions struct {
	// AllowOverwrite re-installs over an existing pack with the same name.
	// Phase 4 default: false — the caller must remove first.
	AllowOverwrite bool
}

// Install runs the full pipeline: load existing registry → collision check →
// stage → copy pack + skills → write .installed → atomic rename. On any
// error after writes begin, every file written during this install is
// rolled back before returning. The caller owns wc.Cleanup.
func (i *Installer) Install(_ context.Context, src Source, wc *WorkingCopy, opts InstallOptions) error {
	// Registry snapshot for collision detection. This runs every install;
	// a concurrent second install would race, but two CLI invocations
	// operating on the same extensions dir is out of scope for Phase 4.
	reg, err := LoadRegistry(i.ExtensionsDir, false)
	if err != nil {
		return fmt.Errorf("load existing registry: %w", err)
	}
	if _, exists := reg.Lookup(wc.Manifest.Name); exists && !opts.AllowOverwrite {
		return fmt.Errorf("pack %q is already installed; run `lango extension remove %s` first", wc.Manifest.Name, wc.Manifest.Name)
	}
	if err := i.detectCollisions(wc.Manifest, reg); err != nil {
		return err
	}

	// Stage directory for atomic install.
	if err := os.MkdirAll(i.ExtensionsDir, 0o755); err != nil {
		return fmt.Errorf("create extensions dir: %w", err)
	}
	stagingDir, err := os.MkdirTemp(i.ExtensionsDir, "."+wc.Manifest.Name+".staging-")
	if err != nil {
		return fmt.Errorf("create staging dir: %w", err)
	}
	extSkillDir := filepath.Join(i.SkillsDir, extSkillPrefix+wc.Manifest.Name)

	// Roll back any writes done by this install on failure.
	rollback := func() {
		_ = os.RemoveAll(stagingDir)
		_ = os.RemoveAll(extSkillDir)
	}

	// Copy pack files (manifest + skills + prompts) into staging.
	if err := copyPackFiles(wc, stagingDir); err != nil {
		rollback()
		return err
	}

	// Copy skills into <skillsDir>/ext-<name>/. Each skill lives at
	// <skillsDir>/ext-<name>/<skill-name>/SKILL.md with resource files
	// preserved.
	if err := i.copySkillsToStore(wc, extSkillDir); err != nil {
		rollback()
		return err
	}

	// Write .installed metadata.
	meta := &InstalledMeta{
		Name:           wc.Manifest.Name,
		Version:        wc.Manifest.Version,
		InstalledAt:    time.Now().UTC(),
		Source:         sourceDescription(src),
		SourceRef:      wc.SourceRef,
		ManifestSHA256: wc.ManifestSHA256,
		FileHashes:     wc.FileHashes,
	}
	if err := WriteInstalledMeta(stagingDir, meta); err != nil {
		rollback()
		return err
	}

	// Atomic rename into final location.
	finalDir := filepath.Join(i.ExtensionsDir, wc.Manifest.Name)
	if opts.AllowOverwrite {
		_ = os.RemoveAll(finalDir)
	}
	if err := os.Rename(stagingDir, finalDir); err != nil {
		rollback()
		return fmt.Errorf("move staging → final: %w", err)
	}

	slog.Info("extension.installed", "pack", wc.Manifest.Name, "version", wc.Manifest.Version)
	return nil
}

// Remove runs the ordered deletion: .installed → ext-<name>/ → pack dir.
// Each step is best-effort after the first — missing files are OK but
// surprise errors surface via the returned error.
func (i *Installer) Remove(_ context.Context, name string) error {
	packDir := filepath.Join(i.ExtensionsDir, name)
	if _, err := os.Stat(packDir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrPackNotFound, name)
	}

	// Delete .installed first so even on a mid-way failure, the pack is no
	// longer considered installed on next startup.
	metaPath := filepath.Join(packDir, installedFileName)
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", metaPath, err)
	}

	extSkillDir := filepath.Join(i.SkillsDir, extSkillPrefix+name)
	if err := os.RemoveAll(extSkillDir); err != nil {
		return fmt.Errorf("remove %s: %w", extSkillDir, err)
	}
	if err := os.RemoveAll(packDir); err != nil {
		return fmt.Errorf("remove %s: %w", packDir, err)
	}
	slog.Info("extension.removed", "pack", name)
	return nil
}

// plannedWrites returns the paths Install would write for a given manifest.
// Used by Inspect for user-visible preview.
func (i *Installer) plannedWrites(m *Manifest) []string {
	packDir := filepath.Join(i.ExtensionsDir, m.Name)
	extSkillDir := filepath.Join(i.SkillsDir, extSkillPrefix+m.Name)
	out := []string{
		filepath.Join(packDir, manifestFileName),
		filepath.Join(packDir, installedFileName),
	}
	for _, s := range m.Contents.Skills {
		// Pack-side mirror.
		out = append(out, filepath.Join(packDir, filepath.FromSlash(s.Path)))
		// SkillsDir-side copy.
		out = append(out, filepath.Join(extSkillDir, s.Name, "SKILL.md"))
	}
	for _, p := range m.Contents.Prompts {
		out = append(out, filepath.Join(packDir, filepath.FromSlash(p.Path)))
	}
	return out
}

// detectCollisions walks existing OK packs and checks whether any
// skill or mode name in the incoming manifest already belongs to a
// different pack. Duplicates within the same pack are already blocked
// by Manifest.Validate.
func (i *Installer) detectCollisions(m *Manifest, reg *Registry) error {
	takenSkills := map[string]string{} // skill name → pack name
	takenModes := map[string]string{}
	for _, p := range reg.OKPacks() {
		if p.Manifest == nil || p.Manifest.Name == m.Name {
			continue
		}
		for _, s := range p.Manifest.Contents.Skills {
			takenSkills[s.Name] = p.Manifest.Name
		}
		for _, md := range p.Manifest.Contents.Modes {
			takenModes[md.Name] = p.Manifest.Name
		}
	}
	for _, s := range m.Contents.Skills {
		if owner, ok := takenSkills[s.Name]; ok {
			return fmt.Errorf("skill name %q is already owned by installed pack %q", s.Name, owner)
		}
	}
	for _, md := range m.Contents.Modes {
		if owner, ok := takenModes[md.Name]; ok {
			return fmt.Errorf("mode name %q is already owned by installed pack %q", md.Name, owner)
		}
	}
	return nil
}

// copyPackFiles copies the manifest and every declared content file from
// the working copy into the staging directory, preserving paths. Directory
// structure is recreated with mode 0o755; files with the source mode.
func copyPackFiles(wc *WorkingCopy, stagingDir string) error {
	manifestSrc := filepath.Join(wc.RootDir, manifestFileName)
	manifestDst := filepath.Join(stagingDir, manifestFileName)
	if err := copyFile(manifestSrc, manifestDst); err != nil {
		return fmt.Errorf("copy manifest: %w", err)
	}

	copyOne := func(rel string) error {
		src, err := ResolvePath(wc.RootDir, rel)
		if err != nil {
			return fmt.Errorf("path %q: %w", rel, err)
		}
		dst := filepath.Join(stagingDir, filepath.FromSlash(filepath.Clean(rel)))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return copyFile(src, dst)
	}
	for _, s := range wc.Manifest.Contents.Skills {
		if err := copyOne(s.Path); err != nil {
			return err
		}
	}
	for _, p := range wc.Manifest.Contents.Prompts {
		if err := copyOne(p.Path); err != nil {
			return err
		}
	}
	return nil
}

// copySkillsToStore mirrors manifest.Contents.Skills into
// <skillsDir>/ext-<pack>/<skill-name>/ so the existing SkillsDir walker
// discovers them. The manifest's path may point at SKILL.md or at its
// parent directory; either form copies the whole directory subtree.
func (i *Installer) copySkillsToStore(wc *WorkingCopy, extSkillDir string) error {
	if err := os.MkdirAll(extSkillDir, 0o755); err != nil {
		return fmt.Errorf("create ext-skills dir: %w", err)
	}
	for _, s := range wc.Manifest.Contents.Skills {
		srcRel := s.Path
		// If manifest points at SKILL.md, copy its containing directory.
		if strings.HasSuffix(filepath.Base(srcRel), "SKILL.md") {
			srcRel = filepath.Dir(srcRel)
		}
		srcAbs, err := ResolvePath(wc.RootDir, srcRel)
		if err != nil {
			return fmt.Errorf("skill %q: %w", s.Name, err)
		}
		dstAbs := filepath.Join(extSkillDir, s.Name)
		if err := copyTree(srcAbs, dstAbs); err != nil {
			return fmt.Errorf("copy skill %q: %w", s.Name, err)
		}
	}
	return nil
}

// copyFile is a small cross-platform file copy respecting source mode.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// copyTree duplicates a directory tree. Symlinks are not followed; the
// copy writes the symlink target's contents as a regular file after
// passing the same containment check used for manifest paths. Callers
// have already validated the source root, so this function focuses on
// the mechanical work.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

// sourceDescription produces the value to record in .installed.Source.
// Local directories record the absolute path; git sources record the URL.
func sourceDescription(src Source) string {
	switch s := src.(type) {
	case *LocalSource:
		abs, err := filepath.Abs(s.Dir)
		if err == nil {
			return "local:" + abs
		}
		return "local:" + s.Dir
	case *GitSource:
		return "git:" + s.URL
	default:
		return ""
	}
}

// ErrNotEnabled signals that a CLI install/remove was invoked with the
// subsystem disabled in config.
var ErrNotEnabled = errors.New("extensions subsystem is disabled; set extensions.enabled=true to continue")
