package economy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newEscrowCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "escrow",
		Short: "Manage escrow agreements",
	}

	cmd.AddCommand(
		newEscrowStatusCmd(cfgLoader),
		newEscrowListCmd(cfgLoader),
		newEscrowShowCmd(cfgLoader),
		newEscrowSentinelCmd(cfgLoader),
	)
	return cmd
}

func newEscrowStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show escrow configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled || !cfg.Economy.Escrow.Enabled {
				fmt.Println("Escrow is disabled.")
				return nil
			}

			fmt.Println("Escrow Configuration:")
			fmt.Printf("  Default Timeout: %s\n", cfg.Economy.Escrow.DefaultTimeout)
			fmt.Printf("  Max Milestones:  %d\n", cfg.Economy.Escrow.MaxMilestones)
			fmt.Printf("  Auto Release:    %v\n", cfg.Economy.Escrow.AutoRelease)
			fmt.Printf("  Dispute Window:  %s\n", cfg.Economy.Escrow.DisputeWindow)
			return nil
		},
	}
}

func newEscrowListCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List escrow configuration summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled {
				fmt.Println("Economy layer is disabled. Enable with economy.enabled=true")
				return nil
			}

			if !cfg.Economy.Escrow.Enabled {
				fmt.Println("Escrow is disabled. Enable with economy.escrow.enabled=true")
				return nil
			}

			oc := cfg.Economy.Escrow.OnChain
			fmt.Println("Escrow Summary:")
			fmt.Printf("  On-Chain Escrow:  %s\n", enabledStr(oc.Enabled))
			if oc.Enabled {
				fmt.Printf("  Mode:             %s\n", valueOrDefault(oc.Mode, "hub"))
				if oc.HubAddress != "" {
					fmt.Printf("  Hub Address:      %s\n", oc.HubAddress)
				}
				if oc.VaultFactoryAddress != "" {
					fmt.Printf("  Vault Factory:    %s\n", oc.VaultFactoryAddress)
				}
			}
			fmt.Printf("  Auto Release:     %v\n", cfg.Economy.Escrow.AutoRelease)
			fmt.Printf("  Default Timeout:  %s\n", cfg.Economy.Escrow.DefaultTimeout)

			fmt.Println("\nUse 'lango economy escrow show' for detailed on-chain configuration.")
			return nil
		},
	}
}

func newEscrowShowCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show detailed on-chain escrow configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled || !cfg.Economy.Escrow.Enabled {
				fmt.Println("Escrow is disabled.")
				return nil
			}

			if id != "" {
				fmt.Printf("Escrow ID %q: use 'lango serve' and escrow agent tools for live data\n", id)
				return nil
			}

			oc := cfg.Economy.Escrow.OnChain
			fmt.Println("On-Chain Escrow Configuration:")
			fmt.Printf("  Enabled:              %s\n", enabledStr(oc.Enabled))
			fmt.Printf("  Mode:                 %s\n", valueOrDefault(oc.Mode, "hub"))
			fmt.Printf("  Hub Address:          %s\n", valueOrDefault(oc.HubAddress, "(not set)"))
			fmt.Printf("  Vault Factory:        %s\n", valueOrDefault(oc.VaultFactoryAddress, "(not set)"))
			fmt.Printf("  Vault Implementation: %s\n", valueOrDefault(oc.VaultImplementation, "(not set)"))
			fmt.Printf("  Arbitrator:           %s\n", valueOrDefault(oc.ArbitratorAddress, "(not set)"))
			fmt.Printf("  Token Address:        %s\n", valueOrDefault(oc.TokenAddress, "(not set)"))
			fmt.Printf("  Poll Interval:        %s\n", oc.PollInterval)

			st := cfg.Economy.Escrow.Settlement
			fmt.Println("\nSettlement:")
			fmt.Printf("  Receipt Timeout:      %s\n", st.ReceiptTimeout)
			fmt.Printf("  Max Retries:          %d\n", st.MaxRetries)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Escrow ID to show (future use)")
	return cmd
}

func newEscrowSentinelCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sentinel",
		Short: "Security sentinel monitoring",
	}

	cmd.AddCommand(newEscrowSentinelStatusCmd(cfgLoader))
	return cmd
}

func newEscrowSentinelStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sentinel engine status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			if !cfg.Economy.Enabled || !cfg.Economy.Escrow.Enabled {
				fmt.Println("Escrow is disabled. Sentinel is not active.")
				return nil
			}

			if !cfg.Economy.Escrow.OnChain.Enabled {
				fmt.Println("On-chain escrow is disabled. Sentinel monitors on-chain events.")
				return nil
			}

			fmt.Println("Sentinel Engine:")
			fmt.Printf("  Status:  active (monitors on-chain escrow events)\n")
			fmt.Printf("  Mode:    %s\n", valueOrDefault(cfg.Economy.Escrow.OnChain.Mode, "hub"))
			fmt.Println("\nThe sentinel engine runs within the application server.")
			fmt.Println("Use 'lango serve' to start and 'lango economy escrow sentinel alerts'")
			fmt.Println("(via agent tools) to view detected alerts.")
			return nil
		},
	}
}

func enabledStr(v bool) string {
	if v {
		return "enabled"
	}
	return "disabled"
}

func valueOrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
