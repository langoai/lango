package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/logging"
	p2pidentity "github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/wallet"
	"go.uber.org/zap"
)

func buildIdentityView(did string, peerID string, keyStorage string, listenAddrs []string) map[string]interface{} {
	return map[string]interface{}{
		"did":         did,
		"peerId":      peerID,
		"listenAddrs": listenAddrs,
		"keyStorage":  keyStorage,
	}
}

type identityDIDProvider interface {
	DID(ctx context.Context) (*p2pidentity.DID, error)
}

type staticIdentityProvider struct {
	did string
}

func (p *staticIdentityProvider) DID(context.Context) (*p2pidentity.DID, error) {
	if p == nil || p.did == "" {
		return nil, nil
	}
	return &p2pidentity.DID{ID: p.did, Version: 2}, nil
}

func loadP2PIdentityProvider(boot *bootstrap.Result) identityDIDProvider {
	if boot == nil || boot.Config == nil {
		return nil
	}

	logger := logging.Sugar()
	if logger == nil {
		l, _ := zap.NewProduction()
		logger = l.Sugar()
	}

	var secrets *security.SecretsStore
	if boot.Crypto != nil && boot.Storage != nil {
		secrets = boot.Storage.SecretsStore(boot.Crypto)
	}

	var wp wallet.WalletProvider
	switch boot.Config.Payment.WalletProvider {
	case "", "local":
		if secrets != nil {
			wp = wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
		}
	case "rpc":
		wp = wallet.NewRPCWallet()
	case "composite":
		if secrets != nil {
			local := wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
			wp = wallet.NewCompositeWallet(wallet.NewRPCWallet(), local, nil)
		}
	}

	if wp != nil && boot.IdentityKey != nil && boot.LangoDir != "" {
		walletPub, err := wp.PublicKey(context.Background())
		if err == nil {
			legacyProv := p2pidentity.NewProvider(wp, logger)
			if bp, err := p2pidentity.NewBundleProvider(p2pidentity.BundleProviderConfig{
				SigningKey:       boot.IdentityKey,
				SettlementPub:    walletPub,
				PQSigningKeySeed: boot.PQSigningKeySeed,
				LangoDir:         boot.LangoDir,
				Legacy:           legacyProv,
				Logger:           logger,
			}); err == nil {
				return bp
			}
		}
	}

	if boot.LangoDir != "" {
		bundle, err := p2pidentity.LoadBundleFile(boot.LangoDir)
		if err == nil && bundle != nil {
			did, err := p2pidentity.ComputeDIDv2(bundle)
			if err == nil {
				return &staticIdentityProvider{did: did}
			}
		}
	}

	if wp != nil {
		return p2pidentity.NewProvider(wp, logger)
	}

	return nil
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
			identityProv := loadP2PIdentityProvider(boot)

			listenAddrs := make([]string, len(addrs))
			for i, a := range addrs {
				listenAddrs[i] = a.String()
			}
			did := ""
			if identityProv != nil {
				if got, err := identityProv.DID(cmd.Context()); err == nil && got != nil {
					did = got.ID
				}
			}
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
