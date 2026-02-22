// Package onboard implements the lango onboard command — a guided 5-step wizard.
package onboard

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
)

// NewCommand creates the onboard command.
func NewCommand() *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Guided 5-step setup wizard for Lango",
		Long: `The onboard command walks you through configuring Lango in five guided steps:

  1. Provider Setup   — Choose an AI provider and enter API credentials
  2. Agent Config     — Select model, max tokens, and temperature
  3. Channel Setup    — Configure Telegram, Discord, or Slack
  4. Security & Auth  — Enable privacy interceptor and PII protection
  5. Test Config      — Validate your configuration

For the full configuration editor with all options, use "lango settings".

All settings including API keys are saved in an encrypted profile (~/.lango/lango.db).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOnboard(profileName)
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "default", "Profile name to create or edit")

	return cmd
}

func runOnboard(profileName string) error {
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
	fmt.Println("\n  3. Fine-tune all settings:")
	fmt.Println("     lango settings")
	fmt.Println("\n  Profile management:")
	fmt.Println("     lango config list    \u2014 list all profiles")
	fmt.Println("     lango config use     \u2014 switch active profile")
}
