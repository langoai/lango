package p2p

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
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
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func executeIdentityCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

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

func TestDidJSONValue_EmptyReturnsNil(t *testing.T) {
	if got := didJSONValue(""); got != nil {
		t.Fatalf("didJSONValue(\"\") = %v, want nil", got)
	}

	if got := didJSONValue("did:lango:02abcdef"); got != "did:lango:02abcdef" {
		t.Fatalf("didJSONValue(non-empty) = %v, want string", got)
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

func TestResolveIdentityDID_MismatchedPersistedBundleFallsBackToLegacy(t *testing.T) {
	dir := t.TempDir()
	_, currentPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	persistedPub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	bundle := &p2pidentity.IdentityBundle{
		Version: 1,
		SigningKey: p2pidentity.PublicKeyEntry{
			Algorithm: "ed25519",
			PublicKey: persistedPub,
		},
		SettlementKey: p2pidentity.PublicKeyEntry{
			Algorithm: "secp256k1-keccak256",
			PublicKey: []byte("settlement-key"),
		},
		LegacyDID: "did:lango:02abcdef",
	}
	if err := p2pidentity.StoreBundleFile(dir, bundle); err != nil {
		t.Fatalf("StoreBundleFile() error = %v", err)
	}

	boot, expected := newLegacyIdentityBoot(t, "local")
	boot.LangoDir = dir
	boot.IdentityKey = currentPriv

	got := resolveIdentityDID(boot)
	if got != expected {
		t.Fatalf("resolveIdentityDID() = %q, want %q", got, expected)
	}
}

func TestResolveIdentityDID_LegacyWalletFallback(t *testing.T) {
	boot, expected := newLegacyIdentityBoot(t, "local")

	got := resolveIdentityDID(boot)
	if got != expected {
		t.Fatalf("resolveIdentityDID() = %q, want %q", got, expected)
	}
}

func TestResolveIdentityDID_UnknownWalletProviderFallsBackToLocal(t *testing.T) {
	boot, expected := newLegacyIdentityBoot(t, "mystery-provider")

	got := resolveIdentityDID(boot)
	if got != expected {
		t.Fatalf("resolveIdentityDID() = %q, want %q", got, expected)
	}
}

func TestResolveIdentityDID_RpcWalletProviderDoesNotFallbackToLegacy(t *testing.T) {
	boot, _ := newLegacyIdentityBoot(t, "rpc")

	got := resolveIdentityDID(boot)
	if got != "" {
		t.Fatalf("resolveIdentityDID() = %q, want empty string", got)
	}
}

func TestIdentityCmd_WritesTextToCommandWriter(t *testing.T) {
	original := loadIdentityCommandData
	loadIdentityCommandData = func(_ *bootstrap.Result) (identityCommandData, func(), error) {
		return identityCommandData{
			did:        "did:lango:v2:test-identity",
			peerID:     "12D3KooWTestPeer",
			keyStorage: "secrets-store",
			listenAddrs: []string{
				"/ip4/127.0.0.1/tcp/9000",
				"/ip6/::/tcp/9000",
			},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadIdentityCommandData = original })

	cmd := newIdentityCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeIdentityCmd(t, cmd)
	require.NoError(t, err)
	require.Contains(t, out, "P2P Identity")
	require.Contains(t, out, "DID:          did:lango:v2:test-identity")
	require.Contains(t, out, "Peer ID:      12D3KooWTestPeer")
	require.Contains(t, out, "Key Storage:  secrets-store")
	require.Contains(t, out, "/ip4/127.0.0.1/tcp/9000")
}

func TestIdentityCmd_WritesJSONToCommandWriter(t *testing.T) {
	original := loadIdentityCommandData
	loadIdentityCommandData = func(_ *bootstrap.Result) (identityCommandData, func(), error) {
		return identityCommandData{
			did:         "",
			peerID:      "12D3KooWJsonPeer",
			keyStorage:  "file",
			listenAddrs: []string{"/ip4/127.0.0.1/tcp/9000"},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadIdentityCommandData = original })

	cmd := newIdentityCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeIdentityCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Equal(t, nil, decoded["did"])
	require.Equal(t, "12D3KooWJsonPeer", decoded["peerId"])
	require.Equal(t, "file", decoded["keyStorage"])
	require.Len(t, decoded["listenAddrs"], 1)
}

func newLegacyIdentityBoot(t *testing.T, walletProvider string) (*bootstrap.Result, string) {
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
				WalletProvider: walletProvider,
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
