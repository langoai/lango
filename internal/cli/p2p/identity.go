package p2p

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	p2pidentity "github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/wallet"
)

func buildIdentityView(did any, peerID string, keyStorage string, listenAddrs []string) map[string]interface{} {
	return map[string]interface{}{
		"did":         did,
		"peerId":      peerID,
		"listenAddrs": listenAddrs,
		"keyStorage":  keyStorage,
	}
}

func didJSONValue(did string) any {
	if did == "" {
		return nil
	}
	return did
}

func resolveIdentityDID(boot *bootstrap.Result) string {
	if boot == nil || boot.LangoDir == "" {
		return resolveLegacyIdentityDID(boot)
	}

	bundle, err := p2pidentity.LoadBundleFile(boot.LangoDir)
	if err == nil && bundle != nil && (len(boot.IdentityKey) == 0 || bundleMatchesIdentityKey(bundle, boot.IdentityKey)) {
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
	case "rpc":
		return nil
	case "composite":
		local := wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
		return wallet.NewCompositeWallet(wallet.NewRPCWallet(), local, nil)
	default:
		return wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
	}
}

func bundleMatchesIdentityKey(bundle *p2pidentity.IdentityBundle, identityKey ed25519.PrivateKey) bool {
	if bundle == nil || len(identityKey) == 0 {
		return false
	}

	pub, ok := identityKey.Public().(ed25519.PublicKey)
	if !ok || len(pub) == 0 {
		return false
	}

	return bundle.SigningKey.Algorithm == "ed25519" && bytes.Equal(bundle.SigningKey.PublicKey, pub)
}

type identityCommandData struct {
	did         string
	peerID      string
	keyStorage  string
	listenAddrs []string
}

var loadIdentityCommandData = func(ctx context.Context, boot *bootstrap.Result) (identityCommandData, func(), error) {
	deps, err := initP2PDeps(ctx, boot)
	if err != nil {
		return identityCommandData{}, nil, err
	}

	peerID := deps.node.PeerID().String()
	addrs := deps.node.Multiaddrs()
	listenAddrs := make([]string, len(addrs))
	for i, a := range addrs {
		listenAddrs[i] = a.String()
	}

	return identityCommandData{
		did:         resolveIdentityDID(boot),
		peerID:      peerID,
		keyStorage:  deps.keyStorage,
		listenAddrs: listenAddrs,
	}, deps.cleanup, nil
}

func newIdentityCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "identity",
		Short:         "Show local DID and peer identity",
		Long:          "Show local DID and peer identity (creates an ephemeral node). For the running server's identity, use GET /api/p2p/identity.",
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

			data, cleanup, err := loadIdentityCommandData(cmd.Context(), boot)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			view := buildIdentityView(didJSONValue(data.did), data.peerID, data.keyStorage, data.listenAddrs)

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), view)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "P2P Identity")
			if data.did != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  DID:          %s\n", data.did)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Peer ID:      %s\n", data.peerID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Key Storage:  %s\n", data.keyStorage)
			fmt.Fprintln(cmd.OutOrStdout(), "  Listen Addrs:")
			for _, a := range data.listenAddrs {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", a)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
