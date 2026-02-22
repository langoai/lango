package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
)

func newSecretsCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage encrypted secrets",
	}

	cmd.AddCommand(newSecretsListCmd(bootLoader))
	cmd.AddCommand(newSecretsSetCmd(bootLoader))
	cmd.AddCommand(newSecretsDeleteCmd(bootLoader))

	return cmd
}

func newSecretsListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored secrets (values are never shown)",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			secretsStore, err := secretsStoreFromBoot(boot)
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

func newSecretsSetCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "set <name>",
		Short: "Store an encrypted secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal")
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			secretsStore, err := secretsStoreFromBoot(boot)
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

func newSecretsDeleteCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a stored secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			secretsStore, err := secretsStoreFromBoot(boot)
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
