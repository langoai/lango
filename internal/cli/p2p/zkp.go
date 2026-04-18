package p2p

import (
	"encoding/json"
	"fmt"
	"os"
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
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show ZKP configuration",
		Long:  "Display the current ZKP proving scheme, SRS mode, and configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(status)
			}

			fmt.Println("ZKP Configuration")
			fmt.Printf("  ZK Handshake:       %v\n", cfg.ZKHandshake)
			fmt.Printf("  ZK Attestation:     %v\n", cfg.ZKAttestation)
			fmt.Printf("  Proving Scheme:     %s\n", cfg.ZKP.ProvingScheme)
			fmt.Printf("  SRS Mode:           %s\n", cfg.ZKP.SRSMode)
			if cfg.ZKP.SRSPath != "" {
				fmt.Printf("  SRS Path:           %s\n", cfg.ZKP.SRSPath)
			}
			fmt.Printf("  Proof Cache Dir:    %s\n", cfg.ZKP.ProofCacheDir)
			if cfg.ZKP.MaxCredentialAge != "" {
				fmt.Printf("  Max Credential Age: %s\n", cfg.ZKP.MaxCredentialAge)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newZKPCircuitsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "circuits",
		Short: "List available ZKP circuits",
		Long:  "List all available zero-knowledge proof circuits and their descriptions.",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(circuits)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "CIRCUIT\tDESCRIPTION")
			for _, c := range circuits {
				fmt.Fprintf(w, "%s\t%s\n", c.ID, c.Description)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
