package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/langowarny/lango/internal/cli/prompt"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/session"
	"github.com/spf13/cobra"
)

func newSecretsCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage encrypted secrets",
	}

	cmd.AddCommand(newSecretsListCmd(cfgLoader))
	cmd.AddCommand(newSecretsSetCmd(cfgLoader))
	cmd.AddCommand(newSecretsDeleteCmd(cfgLoader))

	return cmd
}

func newSecretsListCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored secrets (values are never shown)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, err := session.NewEntStore(cfg.Session.DatabasePath)
			if err != nil {
				return fmt.Errorf("open session store: %w", err)
			}
			defer store.Close()

			secretsStore, err := initLocalCrypto(cfg, store)
			if err != nil {
				return err
			}

			ctx := context.Background()
			secrets, err := secretsStore.List(ctx)
			if err != nil {
				return fmt.Errorf("list secrets: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(secrets)
			}

			if len(secrets) == 0 {
				fmt.Println("No secrets stored.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tKEY\tCREATED\tUPDATED\tACCESS_COUNT")
			for _, s := range secrets {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
					s.Name,
					s.KeyName,
					s.CreatedAt.Format("2006-01-02 15:04"),
					s.UpdatedAt.Format("2006-01-02 15:04"),
					s.AccessCount,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newSecretsSetCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "set <name>",
		Short: "Store an encrypted secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal")
			}

			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, err := session.NewEntStore(cfg.Session.DatabasePath)
			if err != nil {
				return fmt.Errorf("open session store: %w", err)
			}
			defer store.Close()

			secretsStore, err := initLocalCrypto(cfg, store)
			if err != nil {
				return err
			}

			value, err := prompt.Passphrase("Enter secret value: ")
			if err != nil {
				return fmt.Errorf("read secret value: %w", err)
			}

			ctx := context.Background()
			if err := secretsStore.Store(ctx, name, []byte(value)); err != nil {
				return fmt.Errorf("store secret: %w", err)
			}

			fmt.Printf("Secret '%s' stored successfully.\n", name)
			return nil
		},
	}
}

func newSecretsDeleteCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a stored secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, err := session.NewEntStore(cfg.Session.DatabasePath)
			if err != nil {
				return fmt.Errorf("open session store: %w", err)
			}
			defer store.Close()

			secretsStore, err := initLocalCrypto(cfg, store)
			if err != nil {
				return err
			}

			if !force {
				if !prompt.IsInteractive() {
					return fmt.Errorf("use --force for non-interactive deletion")
				}
				fmt.Printf("Delete secret '%s'? [y/N] ", name)
				var answer string
				fmt.Scanln(&answer)
				if answer != "y" && answer != "Y" && answer != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			ctx := context.Background()
			if err := secretsStore.Delete(ctx, name); err != nil {
				return fmt.Errorf("delete secret: %w", err)
			}

			fmt.Printf("Secret '%s' deleted.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}
