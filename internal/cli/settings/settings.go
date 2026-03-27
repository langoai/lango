// Package settings implements the lango settings command.
package settings

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
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

  Core:             Providers, Agent, Channels, Tools, Logging, Gatekeeper, Output Manager
  AI & Knowledge:   Knowledge, Skill, Observational Memory, Embedding & RAG
  Automation:       Cron, Background, Workflow, RunLedger
  Payment & Account: Payment, Smart Account
  P2P & Economy:    P2P Network, Economy, Escrow
  Integrations:     MCP, Observability
  Security:         Security, Auth, DB Encryption, KMS

All settings including API keys are saved in an encrypted profile (~/.lango/lango.db).

See Also:
  lango config get  - Read a config value by dot-path
  lango config set  - Set a config value (passphrase required)
  lango config keys - List available config keys
  lango onboard     - Guided setup wizard
  lango doctor      - Diagnose configuration issues`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSettings(profileName)
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "default", "Profile name to create or edit")

	return cmd
}

func runSettings(profileName string) error {
	boot, err := bootstrap.Run(bootstrap.Options{})
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.DBClient.Close()

	ctx := context.Background()

	initialCfg, isNew, err := loadOrDefault(ctx, boot.ConfigStore, profileName)
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
		fmt.Println("\nSettings cancelled.")
		return nil
	}

	if !editor.Completed {
		return nil
	}

	cfg := editor.Config()
	// TODO(step8): Track modified fields from editor as explicitKeys and merge.
	if err := boot.ConfigStore.Save(ctx, profileName, cfg, nil); err != nil {
		return fmt.Errorf("save profile %q: %w", profileName, err)
	}

	if isNew {
		if err := boot.ConfigStore.SetActive(ctx, profileName); err != nil {
			return fmt.Errorf("activate profile %q: %w", profileName, err)
		}
	}

	printNextSteps(profileName)

	return nil
}

func loadOrDefault(ctx context.Context, store *configstore.Store, name string) (*config.Config, bool, error) {
	cfg, _, err := store.Load(ctx, name)
	if err == nil {
		return cfg, false, nil
	}
	if errors.Is(err, configstore.ErrProfileNotFound) {
		return config.DefaultConfig(), true, nil
	}
	return nil, false, err
}

func printNextSteps(profileName string) {
	fmt.Printf("\n%s Configuration saved to encrypted profile %q\n", "\u2713", profileName)
	fmt.Println("  Storage: ~/.lango/lango.db")

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Start Lango:")
	fmt.Println("     lango serve")
	fmt.Println("\n  2. (Optional) Run doctor to verify setup:")
	fmt.Println("     lango doctor")
	fmt.Println("\n  Profile management:")
	fmt.Println("     lango config list    \u2014 list all profiles")
	fmt.Println("     lango config use     \u2014 switch active profile")
}
