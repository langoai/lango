// Package extension implements the Phase 4 extension-pack subsystem. A pack
// is a read-only bundle of skills, session modes, and prompt fragments
// declared by an extension.yaml manifest. Installation is gated by an
// inspect-then-confirm trust model (see design.md). v1 packs deliberately
// cannot ship tools, MCP servers, providers, or arbitrary code — those
// surfaces ship in later phases with their own trust review.
package extension

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SchemaV1 is the only manifest schema version accepted by the v1 parser.
const SchemaV1 = "lango.extension/v1"

// Manifest is the parsed form of extension.yaml.
type Manifest struct {
	Schema      string   `yaml:"schema"`
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author,omitempty"`
	License     string   `yaml:"license,omitempty"`
	Homepage    string   `yaml:"homepage,omitempty"`
	Contents    Contents `yaml:"contents"`
}

// Contents is the closed set of v1 pack content types. The YAML decoder is
// configured in strict mode so any unknown key under `contents` rejects the
// manifest rather than silently ignoring it — the core trust invariant.
type Contents struct {
	Skills  []SkillRef  `yaml:"skills,omitempty"`
	Modes   []ModeRef   `yaml:"modes,omitempty"`
	Prompts []PromptRef `yaml:"prompts,omitempty"`
}

// SkillRef is a pointer from the manifest to a SKILL.md-rooted directory
// inside the pack. The path is interpreted relative to the pack root.
type SkillRef struct {
	Name string `yaml:"name"`
	// Path is the relative path to the SKILL.md file OR to the directory
	// containing it. Both forms are accepted; the installer normalizes to
	// the directory root at copy time.
	Path string `yaml:"path"`
}

// ModeRef declares a session mode. Shape matches config.SessionMode, but we
// keep a local type to avoid a hard dep on the config package from the
// manifest parser.
type ModeRef struct {
	Name       string   `yaml:"name"`
	Tools      []string `yaml:"tools,omitempty"`
	Skills     []string `yaml:"skills,omitempty"`
	SystemHint string   `yaml:"systemHint,omitempty"`
}

// PromptRef is a file that appends to the effective system prompt at
// runtime. The Section label is optional; when non-empty the prompt
// composer groups prompts under that header.
type PromptRef struct {
	Path    string `yaml:"path"`
	Section string `yaml:"section,omitempty"`
}

// Compile-time name and version regexes. Kebab-case name: 2–64 chars,
// lowercase letters/digits, hyphen-separated. Version: simple semver
// (major.minor.patch with optional pre-release/build).
var (
	nameRegex    = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}[a-z0-9]$`)
	versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
)

// ErrSchemaMismatch indicates the manifest declared a schema version this
// parser doesn't understand.
var ErrSchemaMismatch = errors.New("manifest schema mismatch")

// ParseManifest decodes and validates a v1 extension manifest. An unknown
// top-level field under `contents` or a schema version other than
// SchemaV1 causes an error; see the manifest-schema requirement in
// openspec specs for the full contract.
func ParseManifest(r io.Reader) (*Manifest, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)

	var m Manifest
	if err := dec.Decode(&m); err != nil {
		if strings.Contains(err.Error(), "field ") && strings.Contains(err.Error(), "not found in type") {
			return nil, fmt.Errorf("unknown field in manifest: %w", err)
		}
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.Schema == "" {
		return nil, fmt.Errorf("manifest missing schema field; expected %q", SchemaV1)
	}
	if m.Schema != SchemaV1 {
		return nil, fmt.Errorf("%w: got %q, this lango supports %q — upgrade lango to install this pack", ErrSchemaMismatch, m.Schema, SchemaV1)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate enforces identity, semver, and path-safety rules. This runs at
// parse time; the installer re-validates paths at copy time as a defense-in-
// depth check against symlinks that escape the pack root.
func (m *Manifest) Validate() error {
	if !nameRegex.MatchString(m.Name) {
		return fmt.Errorf("invalid pack name %q: must match kebab-case (lowercase, digits, hyphens; 3–64 chars, no leading/trailing hyphen)", m.Name)
	}
	if !versionRegex.MatchString(m.Version) {
		return fmt.Errorf("invalid pack version %q: must be semver (e.g. 0.1.0)", m.Version)
	}
	if strings.TrimSpace(m.Description) == "" {
		return fmt.Errorf("manifest description must not be empty")
	}
	if m.Homepage != "" {
		if _, err := url.Parse(m.Homepage); err != nil {
			return fmt.Errorf("invalid homepage URL %q: %w", m.Homepage, err)
		}
	}
	seenSkills := map[string]struct{}{}
	for i, s := range m.Contents.Skills {
		if s.Name == "" {
			return fmt.Errorf("contents.skills[%d]: name is required", i)
		}
		if !nameRegex.MatchString(s.Name) {
			return fmt.Errorf("contents.skills[%d]: invalid name %q", i, s.Name)
		}
		if _, dup := seenSkills[s.Name]; dup {
			return fmt.Errorf("contents.skills: duplicate skill name %q within manifest", s.Name)
		}
		seenSkills[s.Name] = struct{}{}
		if err := validateContentPath(s.Path); err != nil {
			return fmt.Errorf("contents.skills[%d] (%s): %w", i, s.Name, err)
		}
	}
	seenModes := map[string]struct{}{}
	for i, mr := range m.Contents.Modes {
		if mr.Name == "" {
			return fmt.Errorf("contents.modes[%d]: name is required", i)
		}
		if !nameRegex.MatchString(mr.Name) {
			return fmt.Errorf("contents.modes[%d]: invalid name %q", i, mr.Name)
		}
		if _, dup := seenModes[mr.Name]; dup {
			return fmt.Errorf("contents.modes: duplicate mode name %q within manifest", mr.Name)
		}
		seenModes[mr.Name] = struct{}{}
	}
	for i, p := range m.Contents.Prompts {
		if err := validateContentPath(p.Path); err != nil {
			return fmt.Errorf("contents.prompts[%d]: %w", i, err)
		}
	}
	return nil
}

// validateContentPath enforces the pack-root path-safety rule at the
// string level. Absolute paths are rejected. Any ".." segment is rejected.
// Empty paths are rejected. Cross-platform separator check: both forward
// and backward slashes are considered.
func validateContentPath(p string) error {
	if p == "" {
		return fmt.Errorf("path must not be empty")
	}
	if filepath.IsAbs(p) || strings.HasPrefix(p, "/") || strings.HasPrefix(p, "\\") {
		return fmt.Errorf("path %q: absolute paths are not allowed", p)
	}
	cleaned := filepath.ToSlash(filepath.Clean(p))
	for _, seg := range strings.Split(cleaned, "/") {
		if seg == ".." {
			return fmt.Errorf("path %q: parent-directory (..) segments are not allowed", p)
		}
	}
	return nil
}

// ResolvePath joins a validated relative path to the given pack root and
// verifies that the resolved absolute target lies within the root after
// symlink evaluation. It returns the final absolute path or an error.
// Callers must invoke this at copy time as a belt-and-suspenders check
// against symlinks that escape the root.
func ResolvePath(root, relPath string) (string, error) {
	if err := validateContentPath(relPath); err != nil {
		return "", err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve pack root: %w", err)
	}
	// Evaluate symlinks on the root too so macOS /var→/private/var and
	// similar platform redirections compare correctly in the containment
	// check below.
	if resolvedRoot, err := filepath.EvalSymlinks(absRoot); err == nil {
		absRoot = resolvedRoot
	}
	joined := filepath.Join(absRoot, relPath)
	resolved, err := filepath.EvalSymlinks(joined)
	if err != nil {
		// A non-existent target fails EvalSymlinks. We still return the
		// joined path so copy-time consumers can distinguish a missing
		// file error from a symlink-escape attack. Containment is checked
		// below using the joined path as a best-effort fallback.
		if !strings.Contains(err.Error(), "no such file") {
			return "", fmt.Errorf("evaluate symlinks for %q: %w", relPath, err)
		}
		resolved = joined
	}
	rel, err := filepath.Rel(absRoot, resolved)
	if err != nil {
		return "", fmt.Errorf("cannot relate %q to pack root: %w", resolved, err)
	}
	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %q: resolved target escapes pack root", relPath)
	}
	return resolved, nil
}
