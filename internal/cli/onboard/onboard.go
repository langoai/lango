// Package onboard implements the lango onboard command.
package onboard

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/langowarny/lango/internal/bootstrap"
	"github.com/langowarny/lango/internal/cli/common"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/configstore"
)

// NewCommand creates the onboard command.
func NewCommand() *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Interactive setup wizard for Lango",
		Long: `The onboard command guides you through setting up Lango for the first time.

An interactive menu-based editor lets you configure each section independently:
  - Agent:      Provider, Model, Tokens, Fallback settings
  - Server:     Host, Port, HTTP/WebSocket toggles
  - Channels:   Telegram, Discord, Slack tokens
  - Tools:      Exec timeouts, Browser, Filesystem limits
  - Security:   Session DB, TTL, PII interceptor, Signer
  - Knowledge:  Learning limits, Skills, Context per layer
  - Providers:  Manage multiple provider configurations

Configuration is saved as an encrypted profile in ~/.lango/lango.db.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOnboard(profileName)
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "default", "Profile name to create or edit")

	return cmd
}

func runOnboard(profileName string) error {
	// 1. Bootstrap: DB + crypto + configstore initialization.
	//    This must happen before BubbleTea starts because passphrase
	//    acquisition requires terminal input.
	boot, err := bootstrap.Run(bootstrap.Options{})
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.DBClient.Close()

	ctx := context.Background()

	// 2. Load existing profile or start with defaults.
	initialCfg, isNew, err := loadOrDefault(ctx, boot.ConfigStore, profileName)
	if err != nil {
		return fmt.Errorf("load profile %q: %w", profileName, err)
	}

	// 3. Run TUI wizard with the initial config.
	p := tea.NewProgram(NewWizardWithConfig(initialCfg))
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

	// 4. Save the edited config as an encrypted profile.
	cfg := wizard.Config()
	if err := boot.ConfigStore.Save(ctx, profileName, cfg); err != nil {
		return fmt.Errorf("save profile %q: %w", profileName, err)
	}

	// 5. Activate the profile if it is new.
	if isNew {
		if err := boot.ConfigStore.SetActive(ctx, profileName); err != nil {
			return fmt.Errorf("activate profile %q: %w", profileName, err)
		}
	}

	// 6. Print next steps.
	printNextSteps(cfg, profileName)

	return nil
}

// loadOrDefault loads an existing profile or returns a default config.
// Returns (config, isNew, error).
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

// printNextSteps shows hints after successful onboarding.
func printNextSteps(cfg *config.Config, profileName string) {
	fmt.Printf("\n✓ Configuration saved to encrypted profile %q\n", profileName)
	fmt.Println("  Storage: ~/.lango/lango.db")

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Set your API key environment variable:")

	if meta, ok := common.GetProviderMetadata(cfg.Agent.Provider); ok && meta.EnvVar != "" {
		fmt.Printf("     export %s=your-key-here\n", meta.EnvVar)
	}

	// Check channels
	if cfg.Channels.Telegram.Enabled {
		fmt.Println("  2. Set your Telegram token:")
		fmt.Println("     export TELEGRAM_BOT_TOKEN=your-token")
	}
	if cfg.Channels.Discord.Enabled {
		fmt.Println("  2. Set your Discord token:")
		fmt.Println("     export DISCORD_BOT_TOKEN=your-token")
	}
	if cfg.Channels.Slack.Enabled {
		fmt.Println("  2. Set your Slack tokens:")
		fmt.Println("     export SLACK_BOT_TOKEN=your-bot-token")
		fmt.Println("     export SLACK_APP_TOKEN=your-app-token")
	}

	fmt.Println("\n  3. Start Lango:")
	fmt.Println("     lango serve")
	fmt.Println("\n  4. (Optional) Run doctor to verify setup:")
	fmt.Println("     lango doctor")
	fmt.Println("\n  Profile management:")
	fmt.Println("     lango config list    — list all profiles")
	fmt.Println("     lango config use     — switch active profile")

	// Write environment hints to a file for easy sourcing
	_ = os.WriteFile(".lango.env.example", []byte(generateEnvExample(cfg)), 0644)
	fmt.Println("\n  See .lango.env.example for a template of required environment variables.")
}

func generateEnvExample(cfg *config.Config) string {
	example := "# Lango Environment Variables\n"
	example += "# Copy this file to .lango.env and fill in your values\n\n"

	if meta, ok := common.GetProviderMetadata(cfg.Agent.Provider); ok && meta.EnvVar != "" {
		example += fmt.Sprintf("%s=your-%s-api-key\n", meta.EnvVar, cfg.Agent.Provider)
	}

	if cfg.Channels.Telegram.Enabled {
		example += "TELEGRAM_BOT_TOKEN=your-telegram-bot-token\n"
	}
	if cfg.Channels.Discord.Enabled {
		example += "DISCORD_BOT_TOKEN=your-discord-bot-token\n"
	}
	if cfg.Channels.Slack.Enabled {
		example += "SLACK_BOT_TOKEN=xoxb-your-bot-token\n"
		example += "SLACK_APP_TOKEN=xapp-your-app-token\n"
	}

	return example
}
