package extension

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// installedFileName is the on-disk metadata record the installer writes
// next to the manifest. It fixes the SHA-256 of the manifest and every
// declared content file at install time so startup can detect tampering.
const installedFileName = ".installed"

// InstalledMeta is the on-disk shape of .installed.
type InstalledMeta struct {
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	InstalledAt    time.Time         `json:"installed_at"`
	Source         string            `json:"source"`
	SourceRef      string            `json:"source_ref,omitempty"`
	ManifestSHA256 string            `json:"manifest_sha256"`
	FileHashes     map[string]string `json:"file_hashes"`
}

// Status summarizes the runtime health of a discovered pack.
type Status string

const (
	StatusOK       Status = "ok"
	StatusTampered Status = "tampered"
	StatusBroken   Status = "broken" // manifest unparseable, permissions error, etc.
)

// InstalledPack is the in-memory record of a pack that startup discovered.
// It is safe to copy.
type InstalledPack struct {
	Manifest *Manifest
	Meta     *InstalledMeta
	RootDir  string
	Status   Status
	Warnings []string
}

// Registry is the set of loaded packs discovered from an extensions dir.
type Registry struct {
	packs []InstalledPack
}

// LoadRegistry walks extensionsDir, parses each */extension.yaml,
// recomputes hashes, and flags tampered or broken packs. A non-existent
// extensionsDir is a valid no-op. enforceIntegrity skips tampered packs
// from the returned pack list (they still appear in List() for CLI
// visibility, just not in Modes/PromptSources output).
func LoadRegistry(extensionsDir string, enforceIntegrity bool) (*Registry, error) {
	r := &Registry{}
	entries, err := os.ReadDir(extensionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return r, nil
		}
		return nil, fmt.Errorf("read extensions dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		pack := loadPack(filepath.Join(extensionsDir, e.Name()), enforceIntegrity)
		r.packs = append(r.packs, pack)
	}
	return r, nil
}

// List returns every discovered pack, including tampered and broken ones.
// Callers that want only the healthy subset should use OKPacks.
func (r *Registry) List() []InstalledPack {
	out := make([]InstalledPack, len(r.packs))
	copy(out, r.packs)
	return out
}

// OKPacks returns only packs whose on-disk state matches the recorded
// manifest. Tampered-and-enforced packs and broken packs are excluded.
func (r *Registry) OKPacks() []InstalledPack {
	out := make([]InstalledPack, 0, len(r.packs))
	for _, p := range r.packs {
		if p.Status == StatusOK {
			out = append(out, p)
		}
	}
	return out
}

// Modes returns the flat list of extension-origin modes from healthy packs.
// Duplicate mode names across packs are kept in the order discovered; the
// caller (config.ResolveModes or equivalent) resolves precedence.
func (r *Registry) Modes() []ModeRef {
	var out []ModeRef
	for _, p := range r.OKPacks() {
		out = append(out, p.Manifest.Contents.Modes...)
	}
	return out
}

// PromptSources returns every prompt file from healthy packs as
// (absolute path, section name) pairs. Callers read the file lazily.
type PromptSource struct {
	AbsolutePath string
	Section      string
	PackName     string
}

// PromptSources returns every prompt file declared by healthy packs.
func (r *Registry) PromptSources() []PromptSource {
	var out []PromptSource
	for _, p := range r.OKPacks() {
		for _, pr := range p.Manifest.Contents.Prompts {
			out = append(out, PromptSource{
				AbsolutePath: filepath.Join(p.RootDir, filepath.FromSlash(pr.Path)),
				Section:      pr.Section,
				PackName:     p.Manifest.Name,
			})
		}
	}
	return out
}

// Lookup returns the pack with the given name, or (zero, false).
func (r *Registry) Lookup(name string) (InstalledPack, bool) {
	for _, p := range r.packs {
		if p.Manifest != nil && p.Manifest.Name == name {
			return p, true
		}
	}
	return InstalledPack{}, false
}

// loadPack reads one pack directory. The returned InstalledPack always
// has a Status; Manifest/Meta may be nil when the pack is broken beyond
// partial recovery.
func loadPack(dir string, enforceIntegrity bool) InstalledPack {
	pack := InstalledPack{RootDir: dir}

	manifestBytes, err := os.ReadFile(filepath.Join(dir, manifestFileName))
	if err != nil {
		pack.Status = StatusBroken
		pack.Warnings = append(pack.Warnings, fmt.Sprintf("read manifest: %v", err))
		return pack
	}
	m, err := ParseManifest(strings.NewReader(string(manifestBytes)))
	if err != nil {
		pack.Status = StatusBroken
		pack.Warnings = append(pack.Warnings, fmt.Sprintf("parse manifest: %v", err))
		return pack
	}
	pack.Manifest = m

	metaBytes, err := os.ReadFile(filepath.Join(dir, installedFileName))
	if err != nil {
		pack.Status = StatusBroken
		pack.Warnings = append(pack.Warnings, fmt.Sprintf("read .installed: %v", err))
		return pack
	}
	var meta InstalledMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		pack.Status = StatusBroken
		pack.Warnings = append(pack.Warnings, fmt.Sprintf("parse .installed: %v", err))
		return pack
	}
	pack.Meta = &meta

	currentManifestSum := sha256.Sum256(manifestBytes)
	if hex.EncodeToString(currentManifestSum[:]) != meta.ManifestSHA256 {
		pack.Status = StatusTampered
		pack.Warnings = append(pack.Warnings, "manifest SHA-256 differs from recorded value")
	}
	for rel, recorded := range meta.FileHashes {
		abs, err := ResolvePath(dir, rel)
		if err != nil {
			pack.Status = StatusTampered
			pack.Warnings = append(pack.Warnings, fmt.Sprintf("path-safety failed for %q: %v", rel, err))
			continue
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			pack.Status = StatusTampered
			pack.Warnings = append(pack.Warnings, fmt.Sprintf("missing or unreadable %q", rel))
			continue
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != recorded {
			pack.Status = StatusTampered
			pack.Warnings = append(pack.Warnings, fmt.Sprintf("SHA-256 mismatch for %q", rel))
		}
	}

	if pack.Status == "" {
		pack.Status = StatusOK
	} else if pack.Status == StatusTampered {
		slog.Warn("extension.tamper.detected",
			"pack", m.Name,
			"warnings", pack.Warnings,
		)
		if enforceIntegrity {
			// When enforcing, strip Manifest so OKPacks/Modes/PromptSources
			// exclude the pack entirely. Keep the record for CLI visibility.
			pack.Manifest = nil
		}
	}

	return pack
}

// LogOrphanSubdirs scans a skills directory for ext-<name>/ subdirs that
// have no matching pack in the provided registry. Each orphan emits a
// structured warning. Phase 4 does not auto-delete orphans.
func LogOrphanSubdirs(skillsDir string, reg *Registry, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "ext-") {
			continue
		}
		packName := strings.TrimPrefix(e.Name(), "ext-")
		if _, ok := reg.Lookup(packName); ok {
			continue
		}
		log.Warn("extension.orphan.detected",
			"subdir", filepath.Join(skillsDir, e.Name()),
			"expected_pack", packName,
		)
	}
}

// WriteInstalledMeta marshals and writes the .installed record to dir.
func WriteInstalledMeta(dir string, meta *InstalledMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal installed meta: %w", err)
	}
	path := filepath.Join(dir, installedFileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// ErrPackNotFound signals that no pack with the requested name is installed.
var ErrPackNotFound = errors.New("pack not found")

// Context is unused in the Registry surface today; keep the import so
// future iterations can thread context through without a breaking change.
var _ = context.Background
