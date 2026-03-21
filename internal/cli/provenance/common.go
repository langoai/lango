package provenance

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/observability/token"
	"github.com/langoai/lango/internal/p2p/identity"
	provenancepkg "github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/wallet"
)

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
	bundle := provenancepkg.NewBundleService(checkpoints, treeStore, attrs, attribution)

	return &services{
		checkpoints: checkpoints,
		treeStore:   treeStore,
		tree:        tree,
		attrs:       attrs,
		attribution: attribution,
		bundle:      bundle,
	}
}

func loadSigner(ctx context.Context, boot *bootstrap.Result) (string, provenancepkg.BundleSignFunc, error) {
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

	return did.ID, func(ctx context.Context, payload []byte) ([]byte, error) {
		return wp.SignMessage(ctx, payload)
	}, nil
}
