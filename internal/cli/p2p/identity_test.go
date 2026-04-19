package p2p

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	p2pidentity "github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/wallet"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestBuildIdentityView_PreservesDidAndListenAddrs(t *testing.T) {
	listenAddrs := []string{
		"/ip4/127.0.0.1/tcp/9000",
		"/ip6/::/tcp/9000",
	}

	view := buildIdentityView("did:lango:v2:abcdef1234567890", "peer-123", "secure-store", listenAddrs)

	if got, want := view["did"], "did:lango:v2:abcdef1234567890"; got != want {
		t.Fatalf("did = %v, want %v", got, want)
	}
	if got, want := view["peerId"], "peer-123"; got != want {
		t.Fatalf("peerId = %v, want %v", got, want)
	}
	if got, want := view["keyStorage"], "secure-store"; got != want {
		t.Fatalf("keyStorage = %v, want %v", got, want)
	}

	gotAddrs, ok := view["listenAddrs"].([]string)
	if !ok {
		t.Fatalf("listenAddrs type = %T, want []string", view["listenAddrs"])
	}
	if !reflect.DeepEqual(gotAddrs, listenAddrs) {
		t.Fatalf("listenAddrs = %v, want %v", gotAddrs, listenAddrs)
	}
}

func TestResolveIdentityDID_ReadOnlyBundleLookup(t *testing.T) {
	dir := t.TempDir()
	bundle := &p2pidentity.IdentityBundle{
		Version: 1,
		SigningKey: p2pidentity.PublicKeyEntry{
			Algorithm: "ed25519",
			PublicKey: []byte("signing-key"),
		},
		SettlementKey: p2pidentity.PublicKeyEntry{
			Algorithm: "secp256k1-keccak256",
			PublicKey: []byte("settlement-key"),
		},
		LegacyDID: "did:lango:02abcdef",
	}
	expected, err := p2pidentity.ComputeDIDv2(bundle)
	if err != nil {
		t.Fatalf("ComputeDIDv2() error = %v", err)
	}
	if err := p2pidentity.StoreBundleFile(dir, bundle); err != nil {
		t.Fatalf("StoreBundleFile() error = %v", err)
	}

	before, err := os.ReadFile(p2pidentity.BundleFilePath(dir))
	if err != nil {
		t.Fatalf("ReadFile() before = %v", err)
	}

	got := resolveIdentityDID(&bootstrap.Result{LangoDir: dir})
	if got != expected {
		t.Fatalf("resolveIdentityDID() = %q, want %q", got, expected)
	}

	after, err := os.ReadFile(p2pidentity.BundleFilePath(dir))
	if err != nil {
		t.Fatalf("ReadFile() after = %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("bundle file changed during read-only DID lookup")
	}
}

func TestResolveIdentityDID_LegacyWalletFallback(t *testing.T) {
	boot, expected := newLegacyIdentityBoot(t)

	got := resolveIdentityDID(boot)
	if got != expected {
		t.Fatalf("resolveIdentityDID() = %q, want %q", got, expected)
	}
}

func newLegacyIdentityBoot(t *testing.T) (*bootstrap.Result, string) {
	t.Helper()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	cryptoProvider := security.NewLocalCryptoProvider()
	require.NoError(t, cryptoProvider.Initialize("test-passphrase-12345"))

	registry := security.NewKeyRegistry(client)
	require.NotNil(t, registry)
	_, err := registry.RegisterKey(context.Background(), "default", "local", security.KeyTypeEncryption)
	require.NoError(t, err)

	secrets := security.NewSecretsStore(client, registry, cryptoProvider)
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	keyBytes := crypto.FromECDSA(privateKey)
	require.NoError(t, secrets.Store(context.Background(), wallet.WalletKeyName, keyBytes))

	expectedPub := crypto.CompressPubkey(&privateKey.PublicKey)
	expectedDID, err := p2pidentity.DIDFromPublicKey(expectedPub)
	require.NoError(t, err)

	boot := &bootstrap.Result{
		Config: &config.Config{
			Payment: config.PaymentConfig{
				WalletProvider: "local",
				Network: config.PaymentNetworkConfig{
					RPCURL:  "http://localhost:8545",
					ChainID: 1,
				},
			},
		},
		Crypto: cryptoProvider,
		Storage: storage.NewFacade(nil, nil, storage.WithSecretsStoreFactory(func(security.CryptoProvider) *security.SecretsStore {
			return secrets
		})),
	}

	return boot, expectedDID.ID
}
