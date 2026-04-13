package provenance

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/observability/token"
	"github.com/langoai/lango/internal/p2p/identity"
	provenancepkg "github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/wallet"
)

const dateTimeFormat = "2006-01-02 15:04:05"

// isProvenanceDisabled prints a notice and returns true if provenance is not enabled.
func isProvenanceDisabled(boot *bootstrap.Result, cmd *cobra.Command) bool {
	if boot.Config.Provenance.Enabled {
		return false
	}
	cmd.Println("Provenance is disabled. Enable with: lango config set provenance.enabled true")
	return true
}

type services struct {
	checkpoints provenancepkg.CheckpointStore
	treeStore   provenancepkg.SessionTreeStore
	tree        *provenancepkg.SessionTree
	attrs       provenancepkg.AttributionStore
	attribution *provenancepkg.AttributionService
	bundle      *provenancepkg.BundleService
}

func loadServices(boot *bootstrap.Result) *services {
	checkpoints := provenancepkg.CheckpointStore(provenancepkg.NewMemoryStore())
	treeStore := provenancepkg.SessionTreeStore(provenancepkg.NewMemoryTreeStore())
	attrs := provenancepkg.AttributionStore(provenancepkg.NewMemoryAttributionStore())
	var tokenStore provenancepkg.TokenUsageReader

	if boot != nil && boot.DBClient != nil {
		checkpoints = provenancepkg.NewEntCheckpointStore(boot.DBClient)
		treeStore = provenancepkg.NewEntSessionTreeStore(boot.DBClient)
		attrs = provenancepkg.NewEntAttributionStore(boot.DBClient)
		tokenStore = token.NewEntTokenStore(boot.DBClient)
	}

	tree := provenancepkg.NewSessionTree(treeStore)
	attribution := provenancepkg.NewAttributionService(attrs, checkpoints, tokenStore)
	ed25519Verifier := func(didStr string, payload, signature []byte) error {
		pubkey, err := identity.ParseDIDPublicKey(didStr)
		if err != nil {
			return err
		}
		return security.VerifyEd25519(pubkey, payload, signature)
	}
	verifiers := map[string]provenancepkg.SignatureVerifyFunc{
		security.AlgorithmSecp256k1Keccak256: identity.VerifyMessageSignature,
		security.AlgorithmEd25519:            ed25519Verifier,
	}
	bundle := provenancepkg.NewBundleService(checkpoints, treeStore, attrs, attribution, verifiers)

	return &services{
		checkpoints: checkpoints,
		treeStore:   treeStore,
		tree:        tree,
		attrs:       attrs,
		attribution: attribution,
		bundle:      bundle,
	}
}

// cliBundleSigner wraps a WalletProvider to satisfy provenance.BundleSigner.
type cliBundleSigner struct {
	wp wallet.WalletProvider
}

func (s *cliBundleSigner) Sign(ctx context.Context, payload []byte) ([]byte, error) {
	return s.wp.SignMessage(ctx, payload)
}

func (s *cliBundleSigner) Algorithm() string {
	return provenancepkg.AlgorithmSecp256k1Keccak256
}

func loadSigner(ctx context.Context, boot *bootstrap.Result) (string, provenancepkg.BundleSigner, error) {
	if boot == nil || boot.DBClient == nil || boot.Crypto == nil {
		return "", nil, fmt.Errorf("signed provenance export requires initialized bootstrap crypto")
	}
	if !boot.Config.Payment.Enabled {
		return "", nil, fmt.Errorf("signed provenance export requires payment.enabled=true")
	}

	keys := security.NewKeyRegistry(boot.DBClient)
	secrets := security.NewSecretsStore(boot.DBClient, keys, boot.Crypto)

	var wp wallet.WalletProvider
	switch boot.Config.Payment.WalletProvider {
	case "", "local":
		wp = wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
	case "composite":
		local := wallet.NewLocalWallet(secrets, boot.Config.Payment.Network.RPCURL, boot.Config.Payment.Network.ChainID)
		wp = wallet.NewCompositeWallet(wallet.NewRPCWallet(), local, nil)
	default:
		return "", nil, fmt.Errorf("wallet provider %q cannot sign provenance bundles in CLI mode", boot.Config.Payment.WalletProvider)
	}

	pub, err := wp.PublicKey(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("load wallet public key: %w", err)
	}
	did, err := identity.DIDFromPublicKey(pub)
	if err != nil {
		return "", nil, fmt.Errorf("derive signer DID: %w", err)
	}

	return did.ID, &cliBundleSigner{wp: wp}, nil
}
