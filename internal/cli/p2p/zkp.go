package p2p

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

// availableCircuits lists the circuits that the ZKP system can compile.
var availableCircuits = []struct {
	ID          string
	Description string
}{
	{"identity", "Prove agent identity without revealing private key"},
	{"capability", "Prove possession of a capability without revealing all capabilities"},
	{"reputation", "Prove reputation score meets a threshold without revealing exact value"},
	{"attestation", "Prove attestation validity with timestamp range assertions"},
}

func newZKPCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zkp",
		Short: "Manage zero-knowledge proof settings",
		Long:  "Inspect ZKP configuration, available circuits, and proving scheme.",
	}

	cmd.AddCommand(newZKPStatusCmd(bootLoader))
	cmd.AddCommand(newZKPCircuitsCmd())

	return cmd
}

func newZKPStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show ZKP configuration",
		Long:          "Display the current ZKP proving scheme, SRS mode, and configuration.",
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

			cfg := boot.Config.P2P

			status := map[string]interface{}{
				"zkHandshake":      cfg.ZKHandshake,
				"zkAttestation":    cfg.ZKAttestation,
				"provingScheme":    cfg.ZKP.ProvingScheme,
				"srsMode":          cfg.ZKP.SRSMode,
				"srsPath":          cfg.ZKP.SRSPath,
				"proofCacheDir":    cfg.ZKP.ProofCacheDir,
				"maxCredentialAge": cfg.ZKP.MaxCredentialAge,
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), status)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "ZKP Configuration")
			fmt.Fprintf(cmd.OutOrStdout(), "  ZK Handshake:       %v\n", cfg.ZKHandshake)
			fmt.Fprintf(cmd.OutOrStdout(), "  ZK Attestation:     %v\n", cfg.ZKAttestation)
			fmt.Fprintf(cmd.OutOrStdout(), "  Proving Scheme:     %s\n", cfg.ZKP.ProvingScheme)
			fmt.Fprintf(cmd.OutOrStdout(), "  SRS Mode:           %s\n", cfg.ZKP.SRSMode)
			if cfg.ZKP.SRSPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  SRS Path:           %s\n", cfg.ZKP.SRSPath)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Proof Cache Dir:    %s\n", cfg.ZKP.ProofCacheDir)
			if cfg.ZKP.MaxCredentialAge != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  Max Credential Age: %s\n", cfg.ZKP.MaxCredentialAge)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func newZKPCircuitsCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "circuits",
		Short:         "List available ZKP circuits",
		Long:          "List all available zero-knowledge proof circuits and their descriptions.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}

			type circuitInfo struct {
				ID          string `json:"id"`
				Description string `json:"description"`
			}

			circuits := make([]circuitInfo, len(availableCircuits))
			for i, c := range availableCircuits {
				circuits[i] = circuitInfo{
					ID:          c.ID,
					Description: c.Description,
				}
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), circuits)
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "CIRCUIT\tDESCRIPTION")
			for _, c := range circuits {
				fmt.Fprintf(w, "%s\t%s\n", c.ID, c.Description)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
