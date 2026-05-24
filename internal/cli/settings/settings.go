// Package settings implements the lango settings command.
package settings

import (
	"context"
	"errors"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/cliboot"
	cliprompt "github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/storage"
)

const nonInteractiveSettingsError = "settings requires an interactive terminal; " +
	"use 'lango config import' or 'lango config set' for scripted configuration"

var (
	requireInteractiveTerminal = cliprompt.RequireInteractiveInput
	runSettingsFn              = runSettings
	settingsBootResult         = cliboot.BootResult
)

// NewCommand creates the settings command.
func NewCommand() *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Full configuration editor for Lango",
		Long: `The settings command opens an interactive menu-based editor for all Lango configuration.

Unlike "lango onboard" (which is a guided wizard for first-time setup), this editor
gives you free navigation across every configuration section.

All categories are visible by default. Advanced items are marked with an ADV badge.
Press Tab to toggle between showing all categories and basic-only view.
Press "/" to search, or use smart filters: @basic, @advanced, @enabled, @modified.

  Core:             Providers, Agent, Channels, Tools, Server, Session, Logging, Gatekeeper, Output Manager
  AI & Knowledge:   Context Profile, Knowledge, Skill, Observational Memory, Embedding & RAG,
                    Graph, Librarian, Retrieval, Auto-Adjust, Context Budget, Agent Memory,
                    Multi-Agent, A2A Protocol, Hooks, Ontology
  Automation:       Cron, Background, Workflow, RunLedger, Provenance
  Payment & Account: Payment, Smart Account
  P2P & Economy:    P2P Network, P2P Workspace, P2P ZKP, P2P Pricing, P2P Owner, P2P Sandbox,
                    Economy, Risk, Negotiation, Escrow, On-Chain Escrow, Pricing
  Integrations:     MCP, Observability, Alerting
  Security:         Security, Auth, Legacy DB Encryption, KMS, OS Sandbox

All settings including API keys are saved in an encrypted profile (~/.lango/lango.db).

See Also:
  lango config get  - Read a config value by dot-path
  lango config set  - Set a config value (passphrase required)
  lango config keys - List available config keys
  lango onboard     - Guided setup wizard
  lango doctor      - Diagnose configuration issues`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireInteractiveTerminal(cmd.InOrStdin(), nonInteractiveSettingsError); err != nil {
				return err
			}
			return runSettingsFn(cmd.OutOrStdout(), profileName)
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "default", "Profile name to create or edit")

	return cmd
}

func runSettings(out io.Writer, profileName string) error {
	boot, err := settingsBootResult()
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.Close()

	ctx := context.Background()

	initialCfg, isNew, err := loadOrDefault(ctx, boot.Storage.ConfigProfiles(), profileName)
	if err != nil {
		return fmt.Errorf("load profile %q: %w", profileName, err)
	}

	tui.SetProfile(profileName)

	p := tea.NewProgram(NewEditorWithConfig(initialCfg))
	model, err := p.Run()
	if err != nil {
		return fmt.Errorf("settings editor: %w", err)
	}

	editor, ok := model.(*Editor)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if editor.Cancelled {
		fmt.Fprintln(out, "\nSettings cancelled.")
		return nil
	}

	if !editor.Completed {
		return nil
	}

	cfg := editor.Config()
	// Mark all context-related keys as explicitly set — the user has seen and
	// accepted these values in the TUI, so auto-enable must not override them.
	explicitKeys := make(map[string]bool, len(config.ContextRelatedKeys()))
	for _, k := range config.ContextRelatedKeys() {
		explicitKeys[k] = true
	}
	if err := boot.Storage.ConfigProfiles().Save(ctx, profileName, cfg, explicitKeys); err != nil {
		return fmt.Errorf("save profile %q: %w", profileName, err)
	}

	if isNew {
		if err := boot.Storage.ConfigProfiles().SetActive(ctx, profileName); err != nil {
			return fmt.Errorf("activate profile %q: %w", profileName, err)
		}
	}

	printNextSteps(out, profileName)

	return nil
}

func loadOrDefault(ctx context.Context, store storage.ConfigProfileStore, name string) (*config.Config, bool, error) {
	cfg, _, err := store.Load(ctx, name)
	if err == nil {
		return cfg, false, nil
	}
	if errors.Is(err, configstore.ErrProfileNotFound) {
		return config.DefaultConfig(), true, nil
	}
	return nil, false, err
}

func printNextSteps(out io.Writer, profileName string) {
	fmt.Fprintf(out, "\n%s Configuration saved to encrypted profile %q\n", "\u2713", profileName)
	fmt.Fprintln(out, "  Storage: ~/.lango/lango.db")

	fmt.Fprintln(out, "\nNext steps:")
	fmt.Fprintln(out, "  1. Start Lango:")
	fmt.Fprintln(out, "     lango serve")
	fmt.Fprintln(out, "\n  2. (Optional) Run doctor to verify setup:")
	fmt.Fprintln(out, "     lango doctor")
	fmt.Fprintln(out, "\n  Profile management:")
	fmt.Fprintln(out, "     lango config list    \u2014 list all profiles")
	fmt.Fprintln(out, "     lango config use     \u2014 switch active profile")
}
