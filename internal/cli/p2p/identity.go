package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	p2pidentity "github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/wallet"
)

func buildIdentityView(did string, peerID string, keyStorage string, listenAddrs []string) map[string]interface{} {
	return map[string]interface{}{
		"did":         did,
		"peerId":      peerID,
		"listenAddrs": listenAddrs,
		"keyStorage":  keyStorage,
	}
}

func resolveIdentityDID(boot *bootstrap.Result) string {
	if boot == nil || boot.LangoDir == "" {
		return resolveLegacyIdentityDID(boot)
	}

	bundle, err := p2pidentity.LoadBundleFile(boot.LangoDir)
	if err == nil && bundle != nil {
		did, err := p2pidentity.ComputeDIDv2(bundle)
		if err == nil {
			return did
		}
	}

	return resolveLegacyIdentityDID(boot)
}

func resolveLegacyIdentityDID(boot *bootstrap.Result) string {
	wp := loadReadOnlyWalletProvider(boot)
	if wp == nil {
		return ""
	}

	pub, err := wp.PublicKey(context.Background())
	if err != nil {
		return ""
	}

	did, err := p2pidentity.DIDFromPublicKey(pub)
	if err != nil {
		return ""
	}
	return did.ID
}

func loadReadOnlyWalletProvider(boot *bootstrap.Result) wallet.WalletProvider {
	if boot == nil || boot.Config == nil || boot.Storage == nil || boot.Crypto == nil {
		return nil
	}

	secrets := boot.Storage.SecretsStore(boot.Crypto)
	if secrets == nil {
		return nil
	}

	switch boot.Config.Payment.WalletProvider {
	case "", "local":
		return wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
	case "composite":
		local := wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
		return wallet.NewCompositeWallet(wallet.NewRPCWallet(), local, nil)
	default:
		return wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
	}
}

func newIdentityCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Show local DID and peer identity",
		Long:  "Show local DID and peer identity (creates an ephemeral node). For the running server's identity, use GET /api/p2p/identity.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			deps, err := initP2PDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			peerID := deps.node.PeerID().String()
			addrs := deps.node.Multiaddrs()

			listenAddrs := make([]string, len(addrs))
			for i, a := range addrs {
				listenAddrs[i] = a.String()
			}
			did := resolveIdentityDID(boot)
			view := buildIdentityView(did, peerID, deps.keyStorage, listenAddrs)

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(view)
			}

			fmt.Println("P2P Identity")
			if did != "" {
				fmt.Printf("  DID:          %s\n", did)
			}
			fmt.Printf("  Peer ID:      %s\n", peerID)
			fmt.Printf("  Key Storage:  %s\n", deps.keyStorage)
			fmt.Printf("  Listen Addrs:\n")
			for _, a := range listenAddrs {
				fmt.Printf("    %s\n", a)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
