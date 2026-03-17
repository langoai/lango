package configcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
)

// NewConfigCmd creates the "config" parent command with all profile subcommands.
// bootLoader is called to obtain the bootstrap result (DB + crypto + config).
// The caller wires get/set/keys subcommands separately if needed.
func NewConfigCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration profile management",
		Long: `Configuration profile management.

Manage multiple configuration profiles for different environments or setups.

See Also:
  lango settings  - Interactive settings editor (TUI)
  lango onboard   - Guided setup wizard
  lango doctor    - Diagnose configuration issues`,
	}

	cmd.AddCommand(newListCmd(bootLoader))
	cmd.AddCommand(newCreateCmd(bootLoader))
	cmd.AddCommand(newUseCmd(bootLoader))
	cmd.AddCommand(newDeleteCmd(bootLoader))
	cmd.AddCommand(newImportCmd(bootLoader))
	cmd.AddCommand(newExportCmd(bootLoader))
	cmd.AddCommand(newValidateCmd(bootLoader))

	return cmd
}

func newListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			profiles, err := boot.ConfigStore.List(context.Background())
			if err != nil {
				return fmt.Errorf("list profiles: %w", err)
			}

			if len(profiles) == 0 {
				fmt.Println("No profiles found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tACTIVE\tVERSION\tCREATED\tUPDATED")
			for _, p := range profiles {
				active := ""
				if p.Active {
					active = "*"
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
					p.Name,
					active,
					p.Version,
					p.CreatedAt.Format(time.DateTime),
					p.UpdatedAt.Format(time.DateTime),
				)
			}
			return w.Flush()
		},
	}
}

func newCreateCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var preset string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new profile (optionally from a preset)",
		Long: `Create a new configuration profile.

Use --preset to start from a purpose-built template:
  minimal       Basic AI agent (quick start)
  researcher    Knowledge, RAG, Graph (research/analysis)
  collaborator  P2P team, payment, workspace (collaboration)
  full          All features enabled (power user)

Examples:
  lango config create my-profile
  lango config create research-bot --preset researcher`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if preset != "" && !config.IsValidPreset(preset) {
				return fmt.Errorf("unknown preset %q (valid: minimal, researcher, collaborator, full)", preset)
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			ctx := context.Background()

			exists, err := boot.ConfigStore.Exists(ctx, name)
			if err != nil {
				return fmt.Errorf("check profile: %w", err)
			}
			if exists {
				return fmt.Errorf("profile %q already exists", name)
			}

			var cfg *config.Config
			if preset != "" {
				cfg = config.PresetConfig(preset)
			} else {
				cfg = config.DefaultConfig()
			}

			if err := boot.ConfigStore.Save(ctx, name, cfg); err != nil {
				return fmt.Errorf("create profile: %w", err)
			}

			if preset != "" {
				fmt.Printf("Profile %q created from preset %q.\n", name, preset)
			} else {
				fmt.Printf("Profile %q created with default configuration.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&preset, "preset", "", "Preset template (minimal, researcher, collaborator, full)")
	return cmd
}

func newUseCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Switch to a different configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			if err := boot.ConfigStore.SetActive(context.Background(), name); err != nil {
				return fmt.Errorf("switch profile: %w", err)
			}

			fmt.Printf("Switched to profile %q.\n", name)
			return nil
		},
	}
}

func newDeleteCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !force {
				fmt.Printf("Delete profile %q? This cannot be undone. [y/N]: ", name)
				var answer string
				_, _ = fmt.Scanln(&answer)
				if answer != "y" && answer != "Y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			if err := boot.ConfigStore.Delete(context.Background(), name); err != nil {
				return fmt.Errorf("delete profile: %w", err)
			}

			fmt.Printf("Profile %q deleted.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	return cmd
}

func newImportCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import and encrypt a JSON config (source file is deleted after import)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			ctx := context.Background()
			if err := configstore.MigrateFromJSON(ctx, boot.ConfigStore, filePath, profileName); err != nil {
				return fmt.Errorf("import config: %w", err)
			}

			fmt.Printf("Imported %q as profile %q (now active).\n", filePath, profileName)
			fmt.Println("Source file deleted for security.")
			return nil
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "default", "name for the imported profile")
	return cmd
}

func newExportCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "export <name>",
		Short: "Export a profile as plaintext JSON (requires passphrase verification)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			cfg, err := boot.ConfigStore.Load(context.Background(), name)
			if err != nil {
				return fmt.Errorf("load profile: %w", err)
			}

			fmt.Fprintln(os.Stderr, "WARNING: exported configuration contains sensitive values in plaintext.")

			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}
}

func newValidateCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the active configuration profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			if err := config.Validate(boot.Config); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			fmt.Printf("Profile %q configuration is valid.\n", boot.ProfileName)
			return nil
		},
	}
}
