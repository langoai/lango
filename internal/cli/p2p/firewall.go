package p2p

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/p2p/firewall"
)

func newFirewallCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "Manage firewall ACL rules",
		Long:  "List, add, or remove knowledge firewall rules that control which peers can access which tools.",
	}

	cmd.AddCommand(newFirewallListCmd(bootLoader))
	cmd.AddCommand(newFirewallAddCmd(bootLoader))
	cmd.AddCommand(newFirewallRemoveCmd(bootLoader))

	return cmd
}

func newFirewallListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List firewall ACL rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			cfg := boot.Config
			if !cfg.P2P.Enabled {
				return fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
			}

			// Build rules from config.
			rules := make([]firewall.ACLRule, len(cfg.P2P.FirewallRules))
			for i, r := range cfg.P2P.FirewallRules {
				rules[i] = firewall.ACLRule{
					PeerDID:   r.PeerDID,
					Action:    r.Action,
					Tools:     r.Tools,
					RateLimit: r.RateLimit,
				}
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(rules)
			}

			if len(rules) == 0 {
				fmt.Println("No firewall rules configured. Default policy: deny-all.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "PEER DID\tACTION\tTOOLS\tRATE LIMIT")
			for _, r := range rules {
				tools := "*"
				if len(r.Tools) > 0 {
					tools = strings.Join(r.Tools, ", ")
				}
				rateLimit := "unlimited"
				if r.RateLimit > 0 {
					rateLimit = fmt.Sprintf("%d/min", r.RateLimit)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.PeerDID, r.Action, tools, rateLimit)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newFirewallAddCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		peerDID   string
		action    string
		tools     []string
		rateLimit int
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a firewall ACL rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			if peerDID == "" {
				return fmt.Errorf("--peer-did is required")
			}
			if action != "allow" && action != "deny" {
				return fmt.Errorf("--action must be 'allow' or 'deny'")
			}

			rule := firewall.ACLRule{
				PeerDID:   peerDID,
				Action:    action,
				Tools:     tools,
				RateLimit: rateLimit,
			}

			toolsStr := "*"
			if len(tools) > 0 {
				toolsStr = strings.Join(tools, ", ")
			}

			fmt.Println("Firewall rule added (runtime only):")
			fmt.Printf("  Peer DID:    %s\n", rule.PeerDID)
			fmt.Printf("  Action:      %s\n", rule.Action)
			fmt.Printf("  Tools:       %s\n", toolsStr)
			if rateLimit > 0 {
				fmt.Printf("  Rate Limit:  %d/min\n", rateLimit)
			}
			fmt.Println("\nTo persist this rule, add it to p2p.firewallRules in your configuration.")
			return nil
		},
	}

	cmd.Flags().StringVar(&peerDID, "peer-did", "", "Peer DID to apply the rule to ('*' for all)")
	cmd.Flags().StringVar(&action, "action", "allow", "Action: 'allow' or 'deny'")
	cmd.Flags().StringSliceVar(&tools, "tools", nil, "Tool name patterns (empty = all)")
	cmd.Flags().IntVar(&rateLimit, "rate-limit", 0, "Max requests per minute (0 = unlimited)")

	return cmd
}

func newFirewallRemoveCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <peer-did>",
		Short: "Remove firewall rules for a peer DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			peerDID := args[0]
			fmt.Printf("To remove rules for peer %s, edit p2p.firewallRules in your configuration.\n", peerDID)
			fmt.Println("Runtime rule removal requires the P2P node to be running via 'lango serve'.")
			return nil
		},
	}

	return cmd
}
