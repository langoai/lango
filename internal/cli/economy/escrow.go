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
				fmt.Fprintln(cmd.OutOrStdout(), "Escrow is disabled.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Escrow Configuration:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Default Timeout: %s\n", cfg.Economy.Escrow.DefaultTimeout)
			fmt.Fprintf(cmd.OutOrStdout(), "  Max Milestones:  %d\n", cfg.Economy.Escrow.MaxMilestones)
			fmt.Fprintf(cmd.OutOrStdout(), "  Auto Release:    %v\n", cfg.Economy.Escrow.AutoRelease)
			fmt.Fprintf(cmd.OutOrStdout(), "  Dispute Window:  %s\n", cfg.Economy.Escrow.DisputeWindow)
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
				fmt.Fprintln(cmd.OutOrStdout(), "Economy layer is disabled. Enable with economy.enabled=true")
				return nil
			}

			if !cfg.Economy.Escrow.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "Escrow is disabled. Enable with economy.escrow.enabled=true")
				return nil
			}

			oc := cfg.Economy.Escrow.OnChain
			fmt.Fprintln(cmd.OutOrStdout(), "Escrow Summary:")
			fmt.Fprintf(cmd.OutOrStdout(), "  On-Chain Escrow:  %s\n", enabledStr(oc.Enabled))
			if oc.Enabled {
				fmt.Fprintf(cmd.OutOrStdout(), "  Mode:             %s\n", valueOrDefault(oc.Mode, "hub"))
				if oc.HubAddress != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  Hub Address:      %s\n", oc.HubAddress)
				}
				if oc.VaultFactoryAddress != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  Vault Factory:    %s\n", oc.VaultFactoryAddress)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Auto Release:     %v\n", cfg.Economy.Escrow.AutoRelease)
			fmt.Fprintf(cmd.OutOrStdout(), "  Default Timeout:  %s\n", cfg.Economy.Escrow.DefaultTimeout)

			fmt.Fprintln(cmd.OutOrStdout(), "\nUse 'lango economy escrow show' for detailed on-chain configuration.")
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
				fmt.Fprintln(cmd.OutOrStdout(), "Escrow is disabled.")
				return nil
			}

			if id != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Escrow ID %q: use 'lango serve' and the escrow_status agent tool for live data\n", id)
				return nil
			}

			oc := cfg.Economy.Escrow.OnChain
			fmt.Fprintln(cmd.OutOrStdout(), "On-Chain Escrow Configuration:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Enabled:              %s\n", enabledStr(oc.Enabled))
			fmt.Fprintf(cmd.OutOrStdout(), "  Mode:                 %s\n", valueOrDefault(oc.Mode, "hub"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Hub Address:          %s\n", valueOrDefault(oc.HubAddress, "(not set)"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Vault Factory:        %s\n", valueOrDefault(oc.VaultFactoryAddress, "(not set)"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Vault Implementation: %s\n", valueOrDefault(oc.VaultImplementation, "(not set)"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Arbitrator:           %s\n", valueOrDefault(oc.ArbitratorAddress, "(not set)"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Token Address:        %s\n", valueOrDefault(oc.TokenAddress, "(not set)"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Poll Interval:        %s\n", oc.PollInterval)
			fmt.Fprintf(cmd.OutOrStdout(), "  Confirmation Depth:   %d\n", oc.ConfirmationDepth)

			st := cfg.Economy.Escrow.Settlement
			fmt.Fprintln(cmd.OutOrStdout(), "\nSettlement:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Receipt Timeout:      %s\n", st.ReceiptTimeout)
			fmt.Fprintf(cmd.OutOrStdout(), "  Max Retries:          %d\n", st.MaxRetries)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Optional escrow ID; prints live-status runtime guidance for that escrow")
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
				fmt.Fprintln(cmd.OutOrStdout(), "Escrow is disabled. Sentinel is not active.")
				return nil
			}

			if !cfg.Economy.Escrow.OnChain.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "On-chain escrow is disabled. Sentinel monitors on-chain events.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Sentinel Engine:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Status:  active (monitors on-chain escrow events)\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Mode:    %s\n", valueOrDefault(cfg.Economy.Escrow.OnChain.Mode, "hub"))
			fmt.Fprintln(cmd.OutOrStdout(), "\nThe sentinel engine runs within the application server.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use 'lango serve' to start the application server, then inspect detected alerts via the sentinel_alerts agent tool.")
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
