package security

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	p2pidentity "github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
	"github.com/langoai/lango/internal/storagebroker"
)

// TestReadEnvelopeStatus_Present verifies the passphrase-free envelope reader
// extracts the expected fields from an on-disk envelope file.
func TestReadEnvelopeStatus_Present(t *testing.T) {
	dir := t.TempDir()
	env, mk, err := security.NewEnvelope("smoke-test-pass")
	require.NoError(t, err)
	defer security.ZeroBytes(mk)

	// Add a mnemonic slot to exercise the RecoverySetup branch.
	mn := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	require.NoError(t, env.AddSlot(security.KEKSlotMnemonic, "recovery", mk, mn, security.NewDefaultKDFParams()))
	require.NoError(t, security.StoreEnvelopeFile(dir, env))

	got := readEnvelopeStatus(dir)
	assert.True(t, got.Present)
	assert.Equal(t, security.EnvelopeVersion, got.Version)
	assert.Equal(t, 2, got.SlotCount)
	assert.True(t, got.RecoverySetup, "mnemonic slot should flip RecoverySetup to true")
	assert.Contains(t, got.SlotTypes, string(security.KEKSlotPassphrase))
	assert.Contains(t, got.SlotTypes, string(security.KEKSlotMnemonic))
	assert.False(t, got.PendingMigration)
	assert.False(t, got.PendingRekey)
}

// TestReadEnvelopeStatus_Missing verifies the reader degrades to an empty
// struct when no envelope file exists.
func TestReadEnvelopeStatus_Missing(t *testing.T) {
	dir := t.TempDir()
	got := readEnvelopeStatus(dir)
	assert.False(t, got.Present)
	assert.Equal(t, 0, got.SlotCount)
	assert.False(t, got.RecoverySetup)
}

// TestReadEnvelopeStatus_CorruptReturnsEmpty verifies the reader never panics
// or propagates errors when the envelope file is malformed.
func TestReadEnvelopeStatus_CorruptReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := security.EnvelopeFilePath(dir)
	require.NoError(t, os.WriteFile(path, []byte("{not-json"), 0o600))

	got := readEnvelopeStatus(dir)
	assert.False(t, got.Present, "corrupt envelope must degrade to absent")
}

func TestReadEnvelopeStatus_KMSAndPendingFlags(t *testing.T) {
	dir := t.TempDir()
	env, mk, err := security.NewEnvelope("smoke-test-pass")
	require.NoError(t, err)
	defer security.ZeroBytes(mk)

	env.PendingMigration = true
	env.PendingRekey = true
	require.NoError(t, env.AddKMSSlot(
		context.Background(),
		"primary-kms",
		mk,
		statusFakeKMSProvider{},
		"aws-kms",
		"arn:aws:kms:us-east-1:123456789012:key/example",
	))
	require.NoError(t, security.StoreEnvelopeFile(dir, env))

	got := readEnvelopeStatus(dir)
	assert.True(t, got.Present)
	assert.Equal(t, 2, got.SlotCount)
	assert.Contains(t, got.SlotTypes, string(security.KEKSlotHardware))
	assert.True(t, got.KMSProtected)
	assert.Equal(t, "aws-kms", got.KMSProvider)
	assert.True(t, got.PendingMigration)
	assert.True(t, got.PendingRekey)
}

func TestReadIdentityBundleStatus_PresentWithPQKey(t *testing.T) {
	dir := t.TempDir()
	bundle := &p2pidentity.IdentityBundle{
		Version: 1,
		SigningKey: p2pidentity.PublicKeyEntry{
			Algorithm: "ed25519",
			PublicKey: bytes.Repeat([]byte{0x01}, 32),
		},
		SettlementKey: p2pidentity.PublicKeyEntry{
			Algorithm: "secp256k1-keccak256",
			PublicKey: bytes.Repeat([]byte{0x02}, 33),
		},
		LegacyDID: "did:lango:legacy",
		PQSigningKey: &p2pidentity.PublicKeyEntry{
			Algorithm: "mldsa65",
			PublicKey: bytes.Repeat([]byte{0x03}, 64),
		},
		CreatedAt: time.Unix(1_700_000_000, 0).UTC(),
	}
	require.NoError(t, p2pidentity.StoreBundleFile(dir, bundle))
	wantDID, err := p2pidentity.ComputeDIDv2(bundle)
	require.NoError(t, err)

	got := readIdentityBundleStatus(dir)
	assert.True(t, got.Present)
	assert.Equal(t, wantDID, got.DIDv2)
	assert.Equal(t, "ed25519", got.SigningAlgorithm)
	assert.True(t, got.HasSettlement)
	assert.Equal(t, "did:lango:legacy", got.LegacyDID)
	assert.True(t, got.PQSigningKeyAvailable)
	assert.Equal(t, "mldsa65", got.PQSigningAlgorithm)
}

func TestReadIdentityBundleStatus_MissingAndCorruptReturnEmpty(t *testing.T) {
	missing := readIdentityBundleStatus(t.TempDir())
	assert.False(t, missing.Present)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(p2pidentity.BundleFilePath(dir), []byte("{not-json"), 0o600))

	corrupt := readIdentityBundleStatus(dir)
	assert.False(t, corrupt.Present)
}

func TestRenderStatus_IncludesEnvelopeIdentityAndPQSections(t *testing.T) {
	out := statusOutput{
		SignerProvider:       "local",
		EncryptionKeys:       4,
		StoredSecrets:        9,
		Interceptor:          "enabled",
		PIIRedaction:         "disabled",
		ApprovalPolicy:       "dangerous",
		ExportabilityEnabled: true,
		DBEncryption:         "disabled (plaintext)",
		Envelope: envelopeSection{
			Present:          true,
			Version:          1,
			SlotCount:        2,
			SlotTypes:        []string{"passphrase", "hardware"},
			RecoverySetup:    true,
			PendingMigration: true,
			PendingRekey:     true,
			KMSProtected:     true,
			KMSProvider:      "aws-kms",
		},
		IdentityBundle: identityBundleSection{
			Present:               true,
			DIDv2:                 "did:lango:v2:abc123",
			SigningAlgorithm:      "ed25519",
			HasSettlement:         true,
			LegacyDID:             "did:lango:legacy",
			PQSigningKeyAvailable: true,
			PQSigningAlgorithm:    "mldsa65",
		},
		DBAvailable:        true,
		PQHandshakeEnabled: true,
		PQHandshakeAlgo:    "X25519-MLKEM768",
	}

	var stdout bytes.Buffer
	err := renderStatus(&stdout, out, "table")
	require.NoError(t, err)
	rendered := stdout.String()
	assert.Contains(t, rendered, "KEK Slots:        2 (passphrase, hardware)")
	assert.Contains(t, rendered, "Recovery Setup:   enabled")
	assert.Contains(t, rendered, "KMS Protection:   enabled (aws-kms)")
	assert.Contains(t, rendered, "PendingMigration: TRUE (migration incomplete)")
	assert.Contains(t, rendered, "PendingRekey:     TRUE (PRAGMA rekey incomplete)")
	assert.Contains(t, rendered, "DID v2:           did:lango:v2:abc123")
	assert.Contains(t, rendered, "Settlement Key:   enabled")
	assert.Contains(t, rendered, "PQ Signing Key:   available (mldsa65)")
	assert.Contains(t, rendered, "PQ Handshake:       enabled (X25519-MLKEM768)")
}

func TestStatusHelpers_PQAlgorithmLabel(t *testing.T) {
	assert.Equal(t, "X25519-MLKEM768", pqAlgorithmLabel(true))
	assert.Equal(t, "", pqAlgorithmLabel(false))
}

// TestReadDBStatusNonInteractive_NoKeyfileNoKeyring verifies the non-interactive
// mini-bootstrap degrades gracefully when no credential is available.
func TestReadDBStatusNonInteractive_NoKeyfileNoKeyring(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nonexistent.db")

	got := readDBStatusNonInteractive(dir, dbPath, nil, false, io.Discard)
	assert.False(t, got.available)
	assert.Equal(t, 0, got.encryptionKeys)
	assert.Equal(t, 0, got.storedSecrets)
}

func TestReadDBStatusNonInteractive_SkipsPassphraseWhenKeyNotNeeded(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	origStartBroker := statusStartBroker
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
		statusStartBroker = origStartBroker
	})

	acquireNonInteractivePassphrase = func(passphrase.Options) (string, passphrase.Source, error) {
		t.Fatal("passphrase acquisition should not run when needsKey is false")
		return "", 0, nil
	}

	var gotReq storagebroker.DBStatusSummaryRequest
	statusStartBroker = func(context.Context) (statusBroker, error) {
		return &stubStatusBroker{
			summary: storagebroker.DBStatusSummaryResult{
				Available:      true,
				EncryptionKeys: 5,
				StoredSecrets:  8,
			},
			onSummary: func(req storagebroker.DBStatusSummaryRequest) {
				gotReq = req
			},
		}, nil
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))

	got := readDBStatusNonInteractive(dir, dbPath, nil, false, io.Discard)
	assert.True(t, got.available)
	assert.Equal(t, 5, got.encryptionKeys)
	assert.Equal(t, 8, got.storedSecrets)
	assert.Equal(t, dbPath, gotReq.DBPath)
	assert.Empty(t, gotReq.EncryptionKey)
	assert.False(t, gotReq.RawKey)
}

func TestReadDBStatusNonInteractive_SummaryErrorDegradesAndClosesBroker(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	origStartBroker := statusStartBroker
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
		statusStartBroker = origStartBroker
	})

	acquireNonInteractivePassphrase = func(passphrase.Options) (string, passphrase.Source, error) {
		return "legacy-passphrase", passphrase.SourceKeyfile, nil
	}

	broker := &stubStatusBroker{err: errors.New("summary failed")}
	statusStartBroker = func(context.Context) (statusBroker, error) {
		return broker, nil
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))

	got := readDBStatusNonInteractive(dir, dbPath, nil, true, io.Discard)
	assert.False(t, got.available)
	assert.Equal(t, 0, got.encryptionKeys)
	assert.Equal(t, 0, got.storedSecrets)
	assert.True(t, broker.closed)
}

func TestReadDBStatusNonInteractive_LegacyKeyringFailureRetriesKeyfile(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	origStartBroker := statusStartBroker
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
		statusStartBroker = origStartBroker
	})

	acquireCalls := 0
	acquireNonInteractivePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		acquireCalls++
		if acquireCalls == 1 {
			assert.NotEmpty(t, opts.KeyfilePath)
			return "stale-keyring-passphrase", passphrase.SourceKeyring, nil
		}
		assert.Nil(t, opts.KeyringProvider)
		return "keyfile-passphrase", passphrase.SourceKeyfile, nil
	}

	var gotKeys []string
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))

	broker := &statusRetryBroker{
		summary: storagebroker.DBStatusSummaryResult{
			Available:      true,
			EncryptionKeys: 7,
			StoredSecrets:  11,
		},
		failForKey: "stale-keyring-passphrase",
		gotKeys:    &gotKeys,
	}
	statusStartBroker = func(context.Context) (statusBroker, error) {
		return broker, nil
	}

	got := readDBStatusNonInteractive(dir, dbPath, nil, true, io.Discard)
	assert.True(t, got.available)
	assert.Equal(t, 7, got.encryptionKeys)
	assert.Equal(t, 11, got.storedSecrets)
	assert.Equal(t, []string{"stale-keyring-passphrase", "keyfile-passphrase"}, gotKeys)
	assert.Equal(t, 2, acquireCalls)
	assert.True(t, broker.closed)
}

func TestReadDBStatusNonInteractive_EnvelopeKeyringUnwrapRetriesKeyfile(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	origStartBroker := statusStartBroker
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
		statusStartBroker = origStartBroker
	})

	env, mk, err := security.NewEnvelope("keyfile-passphrase")
	require.NoError(t, err)
	wantDBKey := security.DeriveDBKeyHex(mk)
	defer security.ZeroBytes(mk)

	acquireCalls := 0
	acquireNonInteractivePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		acquireCalls++
		if acquireCalls == 1 {
			assert.NotEmpty(t, opts.KeyfilePath)
			return "wrong-keyring-passphrase", passphrase.SourceKeyring, nil
		}
		assert.Nil(t, opts.KeyringProvider)
		return "keyfile-passphrase", passphrase.SourceKeyfile, nil
	}

	var gotReq storagebroker.DBStatusSummaryRequest
	statusStartBroker = func(context.Context) (statusBroker, error) {
		return &stubStatusBroker{
			summary: storagebroker.DBStatusSummaryResult{Available: true},
			onSummary: func(req storagebroker.DBStatusSummaryRequest) {
				gotReq = req
			},
		}, nil
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))

	got := readDBStatusNonInteractive(dir, dbPath, env, true, io.Discard)
	assert.True(t, got.available)
	assert.Equal(t, wantDBKey, gotReq.EncryptionKey)
	assert.True(t, gotReq.RawKey)
	assert.Equal(t, 2, acquireCalls)
}

func TestRunStatusNonInteractive_UsesDBConfigWhenAvailable(t *testing.T) {
	origReadStatus := readStatusDBNonInteractive
	t.Cleanup(func() {
		readStatusDBNonInteractive = origReadStatus
	})

	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "rpc"
	cfg.Security.Interceptor.Enabled = false
	cfg.Security.Interceptor.RedactPII = true
	cfg.Security.Interceptor.ApprovalPolicy = config.ApprovalPolicyNone
	cfg.Security.Exportability.Enabled = false
	cfg.P2P.EnablePQHandshake = true

	readStatusDBNonInteractive = func(
		langoDir, dbPath string,
		envelope *security.MasterKeyEnvelope,
		needsKey bool,
		warningWriter io.Writer,
	) dbStatusResult {
		assert.False(t, needsKey)
		assert.Equal(t, filepath.Join(langoDir, "lango.db"), dbPath)
		return dbStatusResult{
			available:      true,
			encryptionKeys: 3,
			storedSecrets:  6,
			config:         cfg,
		}
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".lango"), 0o700))

	var stdout bytes.Buffer
	err := runStatusNonInteractive(&stdout, io.Discard, "json")
	require.NoError(t, err)

	var decoded statusOutput
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &decoded))
	assert.True(t, decoded.DBAvailable)
	assert.Equal(t, "rpc", decoded.SignerProvider)
	assert.Equal(t, 3, decoded.EncryptionKeys)
	assert.Equal(t, 6, decoded.StoredSecrets)
	assert.Equal(t, "disabled", decoded.Interceptor)
	assert.Equal(t, "enabled", decoded.PIIRedaction)
	assert.Equal(t, "none", decoded.ApprovalPolicy)
	assert.False(t, decoded.ExportabilityEnabled)
	assert.True(t, decoded.PQHandshakeEnabled)
	assert.Equal(t, "X25519-MLKEM768", decoded.PQHandshakeAlgo)
}

func TestRunStatusNonInteractive_FallsBackToDefaultsWhenDBUnavailable(t *testing.T) {
	origReadStatus := readStatusDBNonInteractive
	t.Cleanup(func() {
		readStatusDBNonInteractive = origReadStatus
	})

	readStatusDBNonInteractive = func(
		langoDir, dbPath string,
		envelope *security.MasterKeyEnvelope,
		needsKey bool,
		warningWriter io.Writer,
	) dbStatusResult {
		return dbStatusResult{}
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".lango"), 0o700))

	var stdout bytes.Buffer
	err := runStatusNonInteractive(&stdout, io.Discard, "json")
	require.NoError(t, err)

	var decoded statusOutput
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &decoded))
	assert.False(t, decoded.DBAvailable)
	assert.Equal(t, "unavailable", decoded.SignerProvider)
	assert.Equal(t, "enabled", decoded.Interceptor)
	assert.Equal(t, "dangerous", decoded.ApprovalPolicy)
	assert.True(t, decoded.ExportabilityEnabled)
	assert.False(t, decoded.PQHandshakeEnabled)
	assert.Equal(t, "", decoded.PQHandshakeAlgo)
}

type statusFakeKMSProvider struct{}

func (statusFakeKMSProvider) Sign(context.Context, string, []byte) ([]byte, error) {
	return nil, errors.New("sign not implemented")
}

func (statusFakeKMSProvider) Encrypt(_ context.Context, _ string, plaintext []byte) ([]byte, error) {
	return append([]byte("wrapped:"), plaintext...), nil
}

func (statusFakeKMSProvider) Decrypt(context.Context, string, []byte) ([]byte, error) {
	return nil, errors.New("decrypt not implemented")
}

type statusRetryBroker struct {
	summary    storagebroker.DBStatusSummaryResult
	failForKey string
	gotKeys    *[]string
	closed     bool
}

func (s *statusRetryBroker) DBStatusSummary(
	_ context.Context,
	req storagebroker.DBStatusSummaryRequest,
) (storagebroker.DBStatusSummaryResult, error) {
	*s.gotKeys = append(*s.gotKeys, req.EncryptionKey)
	if req.EncryptionKey == s.failForKey {
		return storagebroker.DBStatusSummaryResult{}, errors.New("stale key")
	}
	return s.summary, nil
}

func (s *statusRetryBroker) ConfigLoadActive(context.Context) (storagebroker.ConfigLoadActiveResult, error) {
	return storagebroker.ConfigLoadActiveResult{}, errors.New("no config")
}

func (s *statusRetryBroker) Close(context.Context) error {
	s.closed = true
	return nil
}
