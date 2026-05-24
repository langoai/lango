package app

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/extension"
	"github.com/langoai/lango/internal/prompt"
)

// wireExtensionRegistry loads installed extension packs at startup and
// merges their modes into the effective config. A disabled subsystem or
// missing/empty extensions dir is a valid no-op.
//
// The registry is attached to app.ExtensionRegistry so the CLI `extension
// list` subcommand can read it without re-walking the filesystem, and so
// future phases (pack-contributed tools, MCP, etc.) can surface the set
// of loaded packs through app state.
func wireExtensionRegistry(app *App) {
	if app == nil || app.Config == nil {
		return
	}
	cfg := app.Config.Extensions.ResolveExtensions()
	if !cfg.IsEnabled() {
		return
	}
	reg, err := extension.LoadRegistry(cfg.ResolvedDir(), cfg.EnforceIntegrity)
	if err != nil {
		slog.Warn("extension registry load failed", "error", err)
		return
	}
	app.ExtensionRegistry = reg

	// Extension-origin modes merge into the effective Config.Modes via a
	// mapstructure-compatible insert so all downstream consumers
	// (ResolveModes, /mode slash command, middleware allowlist) see them
	// without special-casing.
	extModes := reg.Modes()
	if len(extModes) > 0 {
		if app.Config.Modes == nil {
			app.Config.Modes = make(map[string]config.SessionMode)
		}
		for _, m := range extModes {
			if m.Name == "" {
				continue
			}
			// User-configured modes take precedence; do not overwrite.
			if _, exists := app.Config.Modes[m.Name]; exists {
				continue
			}
			app.Config.Modes[m.Name] = config.SessionMode{
				Name:       m.Name,
				Tools:      m.Tools,
				Skills:     m.Skills,
				SystemHint: m.SystemHint,
			}
		}
	}

	// Orphan sweep under skills dir: log any ext-<name>/ subtree whose
	// parent pack is missing.
	if skillsDir := app.Config.Skill.SkillsDir; skillsDir != "" {
		resolved := expandTildeAbs(skillsDir)
		if _, err := os.Stat(resolved); err == nil {
			extension.LogOrphanSubdirs(resolved, reg, slog.Default())
		}
	}
}

// extensionPromptSections reads prompt files declared by healthy packs and
// returns them as PromptSections ready for the prompt builder. Files that
// cannot be read are logged and skipped — a missing prompt file does not
// prevent startup.
func extensionPromptSections(reg *extension.Registry) []prompt.PromptSection {
	if reg == nil {
		return nil
	}
	sources := reg.PromptSources()
	if len(sources) == 0 {
		return nil
	}
	sections := make([]prompt.PromptSection, 0, len(sources))
	for _, ps := range sources {
		data, err := os.ReadFile(ps.AbsolutePath)
		if err != nil {
			slog.Warn("extension prompt file unreadable",
				"pack", ps.PackName,
				"path", ps.AbsolutePath,
				"error", err,
			)
			continue
		}
		content := string(data)
		if content == "" {
			continue
		}
		id := prompt.SectionID("extension_" + ps.PackName + "_" + ps.Section)
		sections = append(sections, prompt.NewStaticSection(id, 850, ps.Section, content))
	}
	return sections
}

// expandTildeAbs expands a leading "~/" and makes the result absolute.
// A lookup failure returns the input unchanged.
func expandTildeAbs(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
