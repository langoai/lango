package provenance

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	provenancepkg "github.com/langoai/lango/internal/provenance"
)

func newBundleCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Export or import signed provenance bundles",
	}
	cmd.AddCommand(newBundleExportCmd(bootLoader))
	cmd.AddCommand(newBundleImportCmd(bootLoader))
	return cmd
}

func newBundleExportCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		out       string
		redaction string
	)

	cmd := &cobra.Command{
		Use:   "export <session-key>",
		Short: "Export a signed provenance bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.DBClient.Close()

			svcs := loadServices(boot)
			did, signFn, err := loadSigner(context.Background(), boot)
			if err != nil {
				return err
			}
			bundle, data, err := svcs.bundle.Export(cmd.Context(), args[0], provenancepkg.RedactionLevel(redaction), did, signFn)
			if err != nil {
				return fmt.Errorf("export provenance bundle: %w", err)
			}
			if out != "" {
				if err := os.WriteFile(out, data, 0o600); err != nil {
					return fmt.Errorf("write bundle file: %w", err)
				}
				cmd.Printf("Exported bundle for %s to %s (redaction=%s)\n", args[0], out, bundle.RedactionLevel)
				return nil
			}
			cmd.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&out, "out", "", "Optional file path to write bundle JSON")
	cmd.Flags().StringVar(&redaction, "redaction", string(provenancepkg.RedactionContent), "Redaction level (none, content, full)")
	return cmd
}

func newBundleImportCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "import <file>",
		Short: "Import a signed provenance bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.DBClient.Close()

			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read bundle file: %w", err)
			}
			svcs := loadServices(boot)
			bundle, err := svcs.bundle.Import(cmd.Context(), data)
			if err != nil {
				return fmt.Errorf("import provenance bundle: %w", err)
			}
			cmd.Printf("Imported bundle signed by %s (redaction=%s)\n", bundle.SignerDID, bundle.RedactionLevel)
			return nil
		},
	}
}
