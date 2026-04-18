package bootstrap

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/security/passphrase"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storagebroker"
)

func TestRun_ShredsKeyfileAfterCryptoInit(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	keyfilePath := filepath.Join(dir, "keyfile")
	pass := "test-passphrase-for-shred"

	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, pass))

	result, err := Run(Options{
		LangoDir:            dir,
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		SkipSecureDetection: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = result.Close()
	})

	_, statErr := os.Stat(keyfilePath)
	assert.True(t, os.IsNotExist(statErr), "keyfile should be deleted after bootstrap")
}

func TestRun_KeepsKeyfileWhenOptedOut(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	keyfilePath := filepath.Join(dir, "keyfile")
	pass := "test-passphrase-for-keep"

	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, pass))

	result, err := Run(Options{
		LangoDir:            dir,
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		KeepKeyfile:         true,
		SkipSecureDetection: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = result.Close()
	})

	_, statErr := os.Stat(keyfilePath)
	assert.NoError(t, statErr, "keyfile should still exist when KeepKeyfile is true")
}

func TestRun_StartStorageBroker(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	keyfilePath := filepath.Join(dir, "keyfile")
	pass := "test-passphrase-for-broker"

	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, pass))

	origStart := startStorageBroker
	t.Cleanup(func() { startStorageBroker = origStart })
	startStorageBroker = func(ctx context.Context) (storagebroker.API, error) {
		return &stubBrokerClient{}, nil
	}

	result, err := Run(Options{
		LangoDir:            dir,
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		KeepKeyfile:         true,
		SkipSecureDetection: true,
		StartStorageBroker:  true,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Broker)

	health, err := result.Broker.Health(context.Background())
	require.NoError(t, err)
	assert.True(t, health.Opened)
	require.NotNil(t, result.Storage)

	sessionStore, err := result.Storage.OpenSessionStore()
	require.NoError(t, err)
	require.NotNil(t, sessionStore)

	summary, err := result.Storage.SecuritySummary(context.Background())
	require.NoError(t, err)
	assert.Zero(t, summary.EncryptionKeys)
	assert.Zero(t, summary.StoredSecrets)

	require.NoError(t, result.Close())
}

type stubBrokerClient struct {
	opened bool
}

func (s *stubBrokerClient) Health(context.Context) (storagebroker.HealthResult, error) {
	return storagebroker.HealthResult{Opened: s.opened}, nil
}

func (s *stubBrokerClient) OpenDB(context.Context, storagebroker.OpenDBRequest) (storagebroker.OpenDBResult, error) {
	s.opened = true
	return storagebroker.OpenDBResult{Opened: true}, nil
}

func (s *stubBrokerClient) DBStatusSummary(context.Context, storagebroker.DBStatusSummaryRequest) (storagebroker.DBStatusSummaryResult, error) {
	return storagebroker.DBStatusSummaryResult{}, nil
}

func (s *stubBrokerClient) EncryptPayload(context.Context, []byte) (storagebroker.EncryptPayloadResult, error) {
	return storagebroker.EncryptPayloadResult{
		Ciphertext: []byte("cipher"),
		Nonce:      make([]byte, 12),
		KeyVersion: 1,
	}, nil
}

func (s *stubBrokerClient) DecryptPayload(context.Context, []byte, []byte, int) (storagebroker.DecryptPayloadResult, error) {
	return storagebroker.DecryptPayloadResult{Plaintext: []byte("plain")}, nil
}

func (s *stubBrokerClient) LoadSecurityState(context.Context) (storagebroker.LoadSecurityStateResult, error) {
	return storagebroker.LoadSecurityStateResult{}, nil
}

func (s *stubBrokerClient) StoreSalt(context.Context, []byte) error { return nil }
func (s *stubBrokerClient) StoreChecksum(context.Context, []byte) error {
	return nil
}
func (s *stubBrokerClient) ConfigLoad(context.Context, string) (storagebroker.ConfigLoadResult, error) {
	return storagebroker.ConfigLoadResult{}, nil
}
func (s *stubBrokerClient) ConfigLoadActive(context.Context) (storagebroker.ConfigLoadActiveResult, error) {
	raw, _ := json.Marshal(config.DefaultConfig())
	return storagebroker.ConfigLoadActiveResult{Name: "default", Config: raw}, nil
}
func (s *stubBrokerClient) ConfigSave(context.Context, string, any, map[string]bool) error {
	return nil
}
func (s *stubBrokerClient) ConfigSetActive(context.Context, string) error {
	return nil
}
func (s *stubBrokerClient) ConfigList(context.Context) (storagebroker.ConfigListResult, error) {
	return storagebroker.ConfigListResult{}, nil
}
func (s *stubBrokerClient) ConfigDelete(context.Context, string) error { return nil }
func (s *stubBrokerClient) ConfigExists(context.Context, string) (storagebroker.ConfigExistsResult, error) {
	return storagebroker.ConfigExistsResult{Exists: true}, nil
}
func (s *stubBrokerClient) SessionCreate(context.Context, *session.Session) error { return nil }
func (s *stubBrokerClient) SessionGet(context.Context, string) (*session.Session, error) {
	return &session.Session{Key: "stub"}, nil
}
func (s *stubBrokerClient) SessionUpdate(context.Context, *session.Session) error { return nil }
func (s *stubBrokerClient) SessionDelete(context.Context, string) error           { return nil }
func (s *stubBrokerClient) SessionAppendMessage(context.Context, string, session.Message) error {
	return nil
}
func (s *stubBrokerClient) SessionEnd(context.Context, string) error { return nil }
func (s *stubBrokerClient) SessionList(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}
func (s *stubBrokerClient) SessionGetSalt(context.Context, string) ([]byte, error) {
	return nil, nil
}
func (s *stubBrokerClient) SessionSetSalt(context.Context, string, []byte) error { return nil }
func (s *stubBrokerClient) RecallIndexSession(context.Context, string) error     { return nil }
func (s *stubBrokerClient) RecallProcessPending(context.Context) error           { return nil }
func (s *stubBrokerClient) RecallSearch(context.Context, string, int) ([]search.SearchResult, error) {
	return nil, nil
}
func (s *stubBrokerClient) RecallGetSummary(context.Context, string) (string, error) { return "", nil }
func (s *stubBrokerClient) Close(context.Context) error                              { return nil }
