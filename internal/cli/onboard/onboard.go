// Package onboard implements the lango onboard command — a guided 5-step wizard.
package onboard

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cliboot"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/storage"
)

// NewCommand creates the onboard command.
func NewCommand() *cobra.Command {
	var (
		profileName string
		preset      string
	)

	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Guided 5-step setup wizard for Lango",
		Long: `The onboard command walks you through configuring Lango in five guided steps:

  1. Provider Setup   — Choose a provider (Anthropic, OpenAI, Gemini, Ollama, GitHub)
  2. Agent Config     — Select model (auto-fetched from provider), tokens, temperature
  3. Channel Setup    — Configure Telegram, Discord, or Slack
  4. Security & Auth  — Privacy interceptor, PII redaction, approval policy
  5. Test Config      — Validate your configuration

Use --preset to start from a purpose-built template:
  minimal       Basic AI agent (quick start)
  researcher    Knowledge, RAG, Graph (research/analysis)
  collaborator  P2P team, payment, workspace (collaboration)
  full          All features enabled (power user)

For the full configuration editor with all options, use "lango settings".

All settings including API keys are saved in an encrypted profile (~/.lango/lango.db).

See Also:
  lango settings - Interactive settings editor (TUI)
  lango config   - View/manage configuration profiles
  lango doctor   - Diagnose configuration issues`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOnboard(profileName, preset)
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "default", "Profile name to create or edit")
	cmd.Flags().StringVar(&preset, "preset", "", "Start from a preset (minimal, researcher, collaborator, full)")

	return cmd
}

func runOnboard(profileName, preset string) error {
	if preset != "" && !config.IsValidPreset(preset) {
		return fmt.Errorf("unknown preset %q (valid: minimal, researcher, collaborator, full)", preset)
	}

	boot, err := bootstrap.Run(bootstrap.Options{
		Version:            cliboot.Version,
		StartStorageBroker: true,
	})
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.Close()

	ctx := context.Background()

	initialCfg, isNew, err := loadOrDefault(ctx, boot.Storage.ConfigProfiles(), profileName, preset)
	if err != nil {
		return fmt.Errorf("load profile %q: %w", profileName, err)
	}

	tui.SetProfile(profileName)

	if preset != "" {
		fmt.Printf("\n  Using preset: %s\n\n", preset)
	}

	p := tea.NewProgram(NewWizard(initialCfg))
	model, err := p.Run()
	if err != nil {
		return fmt.Errorf("onboard wizard: %w", err)
	}

	wizard, ok := model.(*Wizard)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if wizard.Cancelled {
		fmt.Println("\nOnboard cancelled.")
		return nil
	}

	if !wizard.Completed {
		return nil
	}

	cfg := wizard.Config()
	// Onboard wizard: all fields set during onboard are considered explicit.
	if err := boot.Storage.ConfigProfiles().Save(ctx, profileName, cfg, nil); err != nil {
		return fmt.Errorf("save profile %q: %w", profileName, err)
	}

	if isNew {
		if err := boot.Storage.ConfigProfiles().SetActive(ctx, profileName); err != nil {
			return fmt.Errorf("activate profile %q: %w", profileName, err)
		}
	}

	printNextSteps(profileName, cfg)

	return nil
}

func loadOrDefault(ctx context.Context, store storage.ConfigProfileStore, name, preset string) (*config.Config, bool, error) {
	cfg, _, err := store.Load(ctx, name)
	if err == nil {
		return cfg, false, nil
	}
	if errors.Is(err, configstore.ErrProfileNotFound) {
		if preset != "" {
			return config.PresetConfig(preset), true, nil
		}
		return config.DefaultConfig(), true, nil
	}
	return nil, false, err
}

func printNextSteps(profileName string, cfg *config.Config) {
	fmt.Printf("\n%s Configuration saved to encrypted profile %q\n", tui.CheckPass, profileName)
	fmt.Println("  Storage: ~/.lango/lango.db")

	fmt.Println("\n  Next: lango serve")

	// Recommend features that are currently disabled.
	type rec struct {
		name     string
		desc     string
		enabled  bool
		category string
	}
	recommendations := []rec{
		{"Knowledge & RAG", "AI learns from conversations", cfg.Knowledge.Enabled, "Knowledge"},
		{"Observational Memory", "Auto-recognize patterns", cfg.ObservationalMemory.Enabled, "Observational Memory"},
		{"Cron Scheduler", "Automate scheduled tasks", cfg.Cron.Enabled, "Cron Scheduler"},
		{"MCP Integrations", "Connect external tool servers", cfg.MCP.Enabled, "MCP Settings"},
	}

	var disabled []rec
	for _, r := range recommendations {
		if !r.enabled {
			disabled = append(disabled, r)
		}
	}

	if len(disabled) > 0 {
		fmt.Println("\n  Recommended features (enable in 'lango settings'):")
		for _, r := range disabled {
			fmt.Printf("    %s %-22s %s\n",
				tui.MutedStyle.Render("\u2022"),
				r.name,
				tui.MutedStyle.Render("\u2014 "+r.desc),
			)
		}
	}

	// Advanced features hint
	fmt.Println("\n  Advanced features (when needed):")
	fmt.Printf("    %s %-22s %s\n",
		tui.MutedStyle.Render("\u2022"),
		"P2P Network",
		tui.MutedStyle.Render("\u2014 collaborate with other agents"),
	)
	fmt.Printf("    %s %-22s %s\n",
		tui.MutedStyle.Render("\u2022"),
		"Payment & Economy",
		tui.MutedStyle.Render("\u2014 on-chain payments and budget management"),
	)

	fmt.Println("\n  Quick presets:")
	fmt.Println("    lango config create <name> --preset researcher")
	fmt.Println("    lango config create <name> --preset collaborator")
	fmt.Println("    lango config create <name> --preset full")
}
