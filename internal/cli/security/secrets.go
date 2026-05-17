package security

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
)

var (
	secretsRequireInteractiveTerminal = prompt.RequireInteractiveTerminal
	secretsPassphrase                 = prompt.PassphraseIO
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
	var output string

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List stored secrets (values are never shown)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			secretsStore, err := secretsStoreFromBoot(boot)
			if err != nil {
				return err
			}

			ctx := context.Background()
			secrets, err := secretsStore.List(ctx)
			if err != nil {
				return fmt.Errorf("list secrets: %w", err)
			}

			type secretMeta struct {
				Name        string `json:"name"`
				KeyName     string `json:"key_name"`
				CreatedAt   string `json:"created_at"`
				UpdatedAt   string `json:"updated_at"`
				AccessCount int    `json:"access_count"`
			}
			out := make([]secretMeta, 0, len(secrets))
			for _, s := range secrets {
				out = append(out, secretMeta{
					Name:        s.Name,
					KeyName:     s.KeyName,
					CreatedAt:   s.CreatedAt.Format(time.RFC3339),
					UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
					AccessCount: s.AccessCount,
				})
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), out)
			}

			if len(secrets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No secrets stored.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
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

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func newSecretsSetCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var valueHex string

	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Store an encrypted secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			secretsStore, err := secretsStoreFromBoot(boot)
			if err != nil {
				return err
			}

			var raw []byte
			if valueHex != "" {
				// Non-interactive: decode hex value (with optional 0x prefix).
				decoded, err := hex.DecodeString(strings.TrimPrefix(valueHex, "0x"))
				if err != nil {
					return fmt.Errorf("decode hex value: %w", err)
				}
				raw = decoded
			} else {
				if err := secretsRequireInteractiveTerminal(
					"this command requires an interactive terminal (use --value-hex for non-interactive)",
				); err != nil {
					return err
				}
				value, err := secretsPassphrase(cmd.OutOrStdout(), "Enter secret value: ")
				if err != nil {
					return fmt.Errorf("read secret value: %w", err)
				}
				raw = []byte(value)
			}

			ctx := context.Background()
			if err := secretsStore.Store(ctx, name, raw); err != nil {
				return fmt.Errorf("store secret: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Secret '%s' stored successfully.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&valueHex, "value-hex", "", "Hex-encoded value to store (non-interactive, optional 0x prefix)")
	return cmd
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
			defer boot.Close()

			secretsStore, err := secretsStoreFromBoot(boot)
			if err != nil {
				return err
			}

			if !force {
				if err := prompt.RequireTTYInput(cmd.InOrStdin(), "use --force for non-interactive deletion"); err != nil {
					return err
				}
				ok, err := prompt.ConfirmDenyOnEOFIO(
					cmd.InOrStdin(),
					cmd.OutOrStdout(),
					fmt.Sprintf("Delete secret '%s'?", name),
				)
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			ctx := context.Background()
			if err := secretsStore.Delete(ctx, name); err != nil {
				return fmt.Errorf("delete secret: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Secret '%s' deleted.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}
