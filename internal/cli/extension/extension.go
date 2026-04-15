// Package extension provides the `lango extension` command tree: inspect,
// install, list, and remove. The CLI is a thin veneer over the
// internal/extension package — parsing flags, resolving config paths,
// rendering reports, and gating confirmations. Business logic (manifest
// validation, path safety, registry lookup) lives in internal/extension.
package extension

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/extension"
)

const (
	exitOK           = 0
	exitUserError    = 1
	exitInternal     = 2
	exitUserDeclined = 3
)

// configLoader returns the effective config (matches the existing CLI
// convention used by memory/learning/agent subcommands).
type configLoader func() (*config.Config, error)

// NewExtensionCmd wires the root `extension` command and all subcommands.
func NewExtensionCmd(loader configLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Install, inspect, and manage extension packs",
		Long: "Install, inspect, and manage Lango extension packs (skills + modes + prompts).\n\n" +
			"Trust model: every install is inspect + confirm. Inspect prints the\n" +
			"pack's identity, SHA-256 hashes, and planned filesystem writes before\n" +
			"anything is written. `--yes` skips the interactive confirmation but\n" +
			"still prints the inspect report to stdout.",
	}
	cmd.AddCommand(newInspectCmd(loader))
	cmd.AddCommand(newInstallCmd(loader))
	cmd.AddCommand(newListCmd(loader))
	cmd.AddCommand(newRemoveCmd(loader))
	return cmd
}

// resolveSource infers whether the argument is a local directory or a git
// URL. Local directories are detected by the filesystem; everything else
// is treated as a git URL (with optional `#<ref>` suffix).
func resolveSource(arg string) extension.Source {
	if info, err := os.Stat(arg); err == nil && info.IsDir() {
		return extension.NewLocalSource(arg)
	}
	return extension.NewGitSource(arg)
}

// installerFor produces an Installer rooted at the config-resolved paths.
func installerFor(cfg *config.Config) (*extension.Installer, error) {
	ext := cfg.Extensions.ResolveExtensions()
	if !ext.IsEnabled() {
		return nil, extension.ErrNotEnabled
	}
	skillsDir := cfg.Skill.SkillsDir
	if skillsDir == "" {
		return nil, fmt.Errorf("skills.skillsDir is not configured; extensions require a skills directory")
	}
	return &extension.Installer{
		ExtensionsDir: expandTildeAbs(ext.ResolvedDir()),
		SkillsDir:     expandTildeAbs(skillsDir),
	}, nil
}

// expandTildeAbs expands leading "~/" and returns an absolute path.
func expandTildeAbs(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}

func newInspectCmd(loader configLoader) *cobra.Command {
	var outputFmt string
	cmd := &cobra.Command{
		Use:     "inspect <source>",
		Short:   "Print a side-effect-free report about a pack",
		Args:    cobra.ExactArgs(1),
		Example: "  lango extension inspect ./python-dev\n  lango extension inspect https://example.com/pack.git#abc123",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutput(outputFmt); err != nil {
				return cliExit(cmd, exitInternal, err)
			}
			cfg, err := loader()
			if err != nil {
				return cliExit(cmd, exitInternal, fmt.Errorf("load config: %w", err))
			}
			inst, err := installerFor(cfg)
			if err != nil {
				return cliExit(cmd, exitUserError, err)
			}
			src := resolveSource(args[0])
			report, wc, err := inst.Inspect(cmd.Context(), src)
			if wc != nil {
				defer func() { _ = wc.Cleanup() }()
			}
			if err != nil {
				return cliExit(cmd, exitUserError, err)
			}
			return renderInspect(cmd.OutOrStdout(), report, resolveOutput(outputFmt, cmd.OutOrStdout()))
		},
	}
	cmd.Flags().StringVar(&outputFmt, "output", "", "Output format: table (default for TTY), json, plain")
	return cmd
}

func newInstallCmd(loader configLoader) *cobra.Command {
	var yes bool
	var outputFmt string
	cmd := &cobra.Command{
		Use:     "install <source>",
		Short:   "Install a pack with inspect + confirm",
		Args:    cobra.ExactArgs(1),
		Example: "  lango extension install ./python-dev\n  lango extension install --yes ./python-dev\n  lango extension install https://example.com/pack.git",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutput(outputFmt); err != nil {
				return cliExit(cmd, exitInternal, err)
			}
			cfg, err := loader()
			if err != nil {
				return cliExit(cmd, exitInternal, fmt.Errorf("load config: %w", err))
			}
			inst, err := installerFor(cfg)
			if err != nil {
				return cliExit(cmd, exitUserError, err)
			}
			src := resolveSource(args[0])
			report, wc, err := inst.Inspect(cmd.Context(), src)
			if wc != nil {
				defer func() { _ = wc.Cleanup() }()
			}
			if err != nil {
				return cliExit(cmd, exitUserError, err)
			}
			if err := renderInspect(cmd.OutOrStdout(), report, resolveOutput(outputFmt, cmd.OutOrStdout())); err != nil {
				return cliExit(cmd, exitInternal, err)
			}
			if !yes {
				ok, promptErr := promptConfirm(cmd.InOrStdin(), cmd.OutOrStdout(), "Install this pack?")
				if promptErr != nil {
					return cliExit(cmd, exitUserDeclined, promptErr)
				}
				if !ok {
					fmt.Fprintln(cmd.OutOrStdout(), "install cancelled by user")
					return cliExit(cmd, exitUserDeclined, nil)
				}
			}
			if err := inst.Install(cmd.Context(), src, wc, extension.InstallOptions{}); err != nil {
				return cliExit(cmd, exitUserError, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "installed %s@%s\n", report.Manifest.Name, report.Manifest.Version)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the interactive confirmation; inspect output is still printed")
	cmd.Flags().StringVar(&outputFmt, "output", "", "Output format for the inspect report: table (default for TTY), json, plain")
	return cmd
}

func newListCmd(loader configLoader) *cobra.Command {
	var outputFmt string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed extension packs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutput(outputFmt); err != nil {
				return cliExit(cmd, exitInternal, err)
			}
			cfg, err := loader()
			if err != nil {
				return cliExit(cmd, exitInternal, fmt.Errorf("load config: %w", err))
			}
			ext := cfg.Extensions.ResolveExtensions()
			reg, err := extension.LoadRegistry(expandTildeAbs(ext.ResolvedDir()), ext.EnforceIntegrity)
			if err != nil {
				return cliExit(cmd, exitInternal, err)
			}
			return renderList(cmd.OutOrStdout(), reg.List(), resolveOutput(outputFmt, cmd.OutOrStdout()))
		},
	}
	cmd.Flags().StringVar(&outputFmt, "output", "", "Output format: table (default for TTY), json, plain")
	return cmd
}

func newRemoveCmd(loader configLoader) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Short:   "Remove an installed pack",
		Args:    cobra.ExactArgs(1),
		Example: "  lango extension remove python-dev\n  lango extension remove --yes python-dev",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loader()
			if err != nil {
				return cliExit(cmd, exitInternal, fmt.Errorf("load config: %w", err))
			}
			inst, err := installerFor(cfg)
			if err != nil {
				return cliExit(cmd, exitUserError, err)
			}
			name := args[0]
			packDir := filepath.Join(inst.ExtensionsDir, name)
			extSkillDir := filepath.Join(inst.SkillsDir, "ext-"+name)
			fmt.Fprintf(cmd.OutOrStdout(), "Will delete:\n  %s\n  %s\n", packDir, extSkillDir)
			if !yes {
				ok, promptErr := promptConfirm(cmd.InOrStdin(), cmd.OutOrStdout(), "Remove pack?")
				if promptErr != nil {
					return cliExit(cmd, exitUserDeclined, promptErr)
				}
				if !ok {
					fmt.Fprintln(cmd.OutOrStdout(), "remove cancelled by user")
					return cliExit(cmd, exitUserDeclined, nil)
				}
			}
			if err := inst.Remove(cmd.Context(), name); err != nil {
				if errors.Is(err, extension.ErrPackNotFound) {
					return cliExit(cmd, exitUserError, err)
				}
				return cliExit(cmd, exitInternal, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the interactive confirmation")
	return cmd
}

// promptConfirm reads a single line from stdin. Non-TTY stdin without
// --yes returns (false, error directing the user to pass --yes). "y"/"yes"
// (case-insensitive) is an accept; anything else is a deny.
func promptConfirm(in io.Reader, out io.Writer, prompt string) (bool, error) {
	if f, ok := in.(*os.File); ok && !term.IsTerminal(int(f.Fd())) {
		return false, fmt.Errorf("stdin is not a TTY; pass --yes for scripted runs")
	}
	fmt.Fprintf(out, "%s [y/N]: ", prompt)
	var resp string
	_, err := fmt.Fscanln(in, &resp)
	if err != nil && err != io.EOF {
		// Fscanln returns "unexpected newline" on an empty response — treat as deny, not error.
		if !strings.Contains(err.Error(), "unexpected newline") {
			return false, err
		}
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	return resp == "y" || resp == "yes", nil
}

// outputFormat identifies the rendering mode a subcommand uses.
type outputFormat int

const (
	outputTable outputFormat = iota
	outputJSON
	outputPlain
)

func validateOutput(flag string) error {
	switch flag {
	case "", "table", "json", "plain":
		return nil
	default:
		return fmt.Errorf("unknown output format %q (expected: table, json, plain)", flag)
	}
}

// resolveOutput maps the flag value to an outputFormat, defaulting to
// table on TTY and plain otherwise.
func resolveOutput(flag string, out io.Writer) outputFormat {
	switch flag {
	case "json":
		return outputJSON
	case "plain":
		return outputPlain
	case "table":
		return outputTable
	}
	// Auto-detect: table when stdout is a TTY, plain otherwise.
	if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		return outputTable
	}
	return outputPlain
}

// renderInspect writes the report in the requested format.
func renderInspect(w io.Writer, r *extension.InspectReport, format outputFormat) error {
	switch format {
	case outputJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(struct {
			Name           string            `json:"name"`
			Version        string            `json:"version"`
			Author         string            `json:"author,omitempty"`
			License        string            `json:"license,omitempty"`
			Homepage       string            `json:"homepage,omitempty"`
			ManifestSHA256 string            `json:"manifest_sha256"`
			FileHashes     map[string]string `json:"file_hashes"`
			PlannedWrites  []string          `json:"planned_writes"`
			SkippedWrites  []string          `json:"skipped_writes"`
			SourceRef      string            `json:"source_ref,omitempty"`
		}{
			Name:           r.Manifest.Name,
			Version:        r.Manifest.Version,
			Author:         r.Manifest.Author,
			License:        r.Manifest.License,
			Homepage:       r.Manifest.Homepage,
			ManifestSHA256: r.ManifestSHA256,
			FileHashes:     r.FileHashes,
			PlannedWrites:  r.PlannedWrites,
			SkippedWrites:  r.SkippedWrites,
			SourceRef:      r.SourceRef,
		})
	case outputPlain, outputTable:
		m := r.Manifest
		fmt.Fprintf(w, "Pack: %s@%s\n", m.Name, m.Version)
		if m.Author != "" {
			fmt.Fprintf(w, "Author:   %s\n", m.Author)
		}
		if m.License != "" {
			fmt.Fprintf(w, "License:  %s\n", m.License)
		}
		if m.Homepage != "" {
			fmt.Fprintf(w, "Homepage: %s\n", m.Homepage)
		}
		fmt.Fprintf(w, "Description: %s\n\n", m.Description)
		fmt.Fprintf(w, "Manifest SHA-256: %s\n", r.ManifestSHA256)
		if r.SourceRef != "" {
			fmt.Fprintf(w, "Source ref:       %s\n", r.SourceRef)
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Skills (%d):\n", len(m.Contents.Skills))
		for _, s := range m.Contents.Skills {
			fmt.Fprintf(w, "  - %s  (path: %s)\n", s.Name, s.Path)
		}
		fmt.Fprintf(w, "Modes (%d):\n", len(m.Contents.Modes))
		for _, md := range m.Contents.Modes {
			fmt.Fprintf(w, "  - %s\n", md.Name)
			if md.SystemHint != "" {
				fmt.Fprintf(w, "      hint: %s\n", truncate(md.SystemHint, 120))
			}
		}
		fmt.Fprintf(w, "Prompts (%d):\n", len(m.Contents.Prompts))
		for _, p := range m.Contents.Prompts {
			if p.Section != "" {
				fmt.Fprintf(w, "  - %s  [section=%s]\n", p.Path, p.Section)
			} else {
				fmt.Fprintf(w, "  - %s\n", p.Path)
			}
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Planned writes:")
		for _, p := range r.PlannedWrites {
			fmt.Fprintf(w, "  %s\n", p)
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "v1 packs do NOT install:")
		for _, s := range r.SkippedWrites {
			fmt.Fprintf(w, "  - %s\n", s)
		}
		fmt.Fprintln(w)
		return nil
	}
	return fmt.Errorf("unhandled output format")
}

// renderList renders the registry list.
func renderList(w io.Writer, packs []extension.InstalledPack, format outputFormat) error {
	switch format {
	case outputJSON:
		type row struct {
			Name           string    `json:"name"`
			Version        string    `json:"version"`
			Author         string    `json:"author,omitempty"`
			InstalledAt    time.Time `json:"installed_at,omitempty"`
			Source         string    `json:"source,omitempty"`
			Status         string    `json:"status"`
			ManifestSHA256 string    `json:"manifest_sha256,omitempty"`
		}
		rows := make([]row, 0, len(packs))
		for _, p := range packs {
			r := row{Status: string(p.Status)}
			if p.Manifest != nil {
				r.Name = p.Manifest.Name
				r.Version = p.Manifest.Version
				r.Author = p.Manifest.Author
			}
			if p.Meta != nil {
				r.InstalledAt = p.Meta.InstalledAt
				r.Source = p.Meta.Source
				r.ManifestSHA256 = p.Meta.ManifestSHA256
			}
			rows = append(rows, r)
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	case outputPlain:
		for _, p := range packs {
			name, version := "", ""
			if p.Manifest != nil {
				name, version = p.Manifest.Name, p.Manifest.Version
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", name, version, p.Status)
		}
		return nil
	case outputTable:
		fmt.Fprintf(w, "%-24s %-10s %-20s %-22s %s\n", "NAME", "VERSION", "AUTHOR", "INSTALLED", "STATUS")
		for _, p := range packs {
			name, version, author := "", "", ""
			var installed string
			if p.Manifest != nil {
				name = p.Manifest.Name
				version = p.Manifest.Version
				author = p.Manifest.Author
			}
			if p.Meta != nil {
				installed = p.Meta.InstalledAt.Format(time.RFC3339)
			}
			fmt.Fprintf(w, "%-24s %-10s %-20s %-22s %s\n", name, version, author, installed, p.Status)
		}
		return nil
	}
	return fmt.Errorf("unhandled output format")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// cliExit wraps a cobra RunE failure with a specific exit code using
// cobra.CheckErr-compatible behavior: the command's parent chain sets
// SilenceUsage via side effect and we return a wrapped error carrying the
// exit code. The main binary's errorHandler should translate cliError to
// os.Exit — for now we use os.Exit directly to match the Phase 4 CLI spec
// exit-code requirements without a cross-cutting refactor.
func cliExit(cmd *cobra.Command, code int, err error) error {
	if code == exitOK {
		return nil
	}
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	os.Exit(code)
	return nil
}

// Compile-time checks keep us honest about unused context surface areas.
var _ = context.Background
