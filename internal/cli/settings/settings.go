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
gives you free navigation across every configuration section. Categories are organized
into groups:

  Core:           Providers, Agent, Server, Session
  Communication:  Channels, Tools, Multi-Agent, A2A Protocol
  AI & Knowledge: Knowledge, Skill, Observational Memory, Embedding & RAG,
                  Graph Store, Librarian
  Infrastructure: Payment, Cron Scheduler, Background Tasks, Workflow Engine
  P2P Network:    P2P Network, P2P ZKP, P2P Pricing, P2P Owner Protection,
                  P2P Sandbox
  Security:       Security, Auth, Security DB Encryption,
                  Security KMS

Press "/" to search across all categories by keyword.

All settings including API keys are saved in an encrypted profile (~/.lango/lango.db).

See Also:
  lango config   - View/manage configuration profiles
  lango onboard  - Guided setup wizard
  lango doctor   - Diagnose configuration issues`,
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
	if err := boot.ConfigStore.Save(ctx, profileName, cfg); err != nil {
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
	cfg, err := store.Load(ctx, name)
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
