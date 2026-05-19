package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
	"github.com/langoai/lango/internal/sqlitedriver"
	"github.com/langoai/lango/internal/storagebroker"
)

func TestWave37EnsureDataDirDefaultsAndCreateFailure(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lango-home")
	state := &State{Options: Options{LangoDir: dir}}

	err := phaseEnsureDataDir().Run(context.Background(), state)
	require.NoError(t, err)

	assert.Equal(t, dir, state.LangoDir)
	assert.Equal(t, dir, state.Result.LangoDir)
	assert.Equal(t, filepath.Join(dir, "lango.db"), state.Options.DBPath)
	assert.Equal(t, filepath.Join(dir, "keyfile"), state.Options.KeyfilePath)

	info, err := os.Stat(filepath.Join(dir, "skills"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(dataDirPerm), info.Mode().Perm())

	filePath := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(filePath, []byte("blocked"), 0600))

	err = phaseEnsureDataDir().Run(context.Background(), &State{Options: Options{LangoDir: filePath}})
	require.Error(t, err)
	assert.ErrorContains(t, err, "create data directory")
}

func TestWave37LoadEnvelopeFileBranches(t *testing.T) {
	missingState := &State{LangoDir: t.TempDir()}
	err := phaseLoadEnvelopeFile().Run(context.Background(), missingState)
	require.NoError(t, err)
	assert.Nil(t, missingState.Envelope)

	corruptDir := t.TempDir()
	require.NoError(t, os.WriteFile(security.EnvelopeFilePath(corruptDir), []byte("{not-json"), 0600))
	err = phaseLoadEnvelopeFile().Run(context.Background(), &State{LangoDir: corruptDir})
	require.Error(t, err)
	assert.ErrorContains(t, err, "load envelope")
	assert.ErrorIs(t, err, security.ErrEnvelopeCorrupt)

	validDir := t.TempDir()
	env, mk, err := security.NewEnvelope("valid-passphrase")
	require.NoError(t, err)
	defer security.ZeroBytes(mk)
	require.NoError(t, security.StoreEnvelopeFile(validDir, env))

	validState := &State{LangoDir: validDir}
	err = phaseLoadEnvelopeFile().Run(context.Background(), validState)
	require.NoError(t, err)
	require.NotNil(t, validState.Envelope)
	assert.Equal(t, security.EnvelopeVersion, validState.Envelope.Version)
}

func TestWave37DetectEncryptionRejectsLegacyEncryptedHeaders(t *testing.T) {
	dir := t.TempDir()

	missingPath := filepath.Join(dir, "missing.db")
	state := &State{Options: Options{DBPath: missingPath}}
	err := phaseDetectEncryption().Run(context.Background(), state)
	require.NoError(t, err)
	assert.False(t, state.DBEncrypted)
	assert.False(t, state.NeedsDBKey)

	plainPath := filepath.Join(dir, "plain.db")
	require.NoError(t, os.WriteFile(plainPath, append([]byte("SQLite format 3\x00"), []byte("rest")...), 0600))
	state = &State{Options: Options{DBPath: plainPath}}
	err = phaseDetectEncryption().Run(context.Background(), state)
	require.NoError(t, err)
	assert.False(t, state.DBEncrypted)
	assert.False(t, state.NeedsDBKey)

	legacyPath := filepath.Join(dir, "legacy-encrypted.db")
	require.NoError(t, os.WriteFile(legacyPath, []byte("not sqlite header"), 0600))
	state = &State{Options: Options{DBPath: legacyPath}}
	err = phaseDetectEncryption().Run(context.Background(), state)
	require.Error(t, err)
	assert.ErrorIs(t, err, sqlitedriver.ErrLegacyEncryptedOrUnreadableDB)
	assert.True(t, state.DBEncrypted)
}

func TestWave37AcquireCredentialSetsFirstRunGuessAndWrapsErrors(t *testing.T) {
	origAcquire := acquirePassphrase
	t.Cleanup(func() { acquirePassphrase = origAcquire })

	var captured passphrase.Options
	acquirePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		captured = opts
		return "from-keyfile", passphrase.SourceKeyfile, nil
	}

	dir := t.TempDir()
	state := &State{
		Options: Options{
			DBPath:              filepath.Join(dir, "missing.db"),
			KeyfilePath:         filepath.Join(dir, "keyfile"),
			SkipSecureDetection: true,
		},
	}
	err := phaseAcquireCredential().Run(context.Background(), state)
	require.NoError(t, err)

	assert.True(t, state.FirstRunGuess)
	assert.True(t, captured.AllowCreation)
	assert.Equal(t, state.Options.KeyfilePath, captured.KeyfilePath)
	assert.Equal(t, "from-keyfile", state.Passphrase)
	assert.Equal(t, passphrase.SourceKeyfile, state.PassSource)

	require.NoError(t, os.WriteFile(state.Options.DBPath, []byte("SQLite format 3\x00"), 0600))
	acquirePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		captured = opts
		return "", passphrase.SourceKeyfile, errors.New("no credential")
	}

	err = phaseAcquireCredential().Run(context.Background(), state)
	require.Error(t, err)
	assert.False(t, state.FirstRunGuess)
	assert.False(t, captured.AllowCreation)
	assert.ErrorContains(t, err, "acquire passphrase: no credential")
}

func TestWave37OpenDatabaseDirectAndBrokerBranches(t *testing.T) {
	directDB := filepath.Join(t.TempDir(), "direct.db")
	directState := &State{Options: Options{DBPath: directDB}}

	err := phaseOpenDatabase().Run(context.Background(), directState)
	require.NoError(t, err)
	require.NotNil(t, directState.Client)
	require.NotNil(t, directState.RawDB)
	require.NoError(t, directState.RawDB.Ping())
	phaseOpenDatabase().Cleanup(directState)
	assert.Error(t, directState.RawDB.Ping())

	origStart := startStorageBroker
	t.Cleanup(func() { startStorageBroker = origStart })

	startStorageBroker = func(context.Context) (storagebroker.API, error) {
		return nil, errors.New("broker unavailable")
	}
	err = phaseOpenDatabase().Run(context.Background(), &State{
		Options: Options{DBPath: filepath.Join(t.TempDir(), "broker.db"), StartStorageBroker: true},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "start storage broker: broker unavailable")

	openErrBroker := &wave37Broker{openErr: errors.New("open failed")}
	startStorageBroker = func(context.Context) (storagebroker.API, error) { return openErrBroker, nil }
	state := &State{
		Options:   Options{DBPath: filepath.Join(t.TempDir(), "broker-open.db"), StartStorageBroker: true},
		MasterKey: []byte("12345678901234567890123456789012"),
	}
	err = phaseOpenDatabase().Run(context.Background(), state)
	require.Error(t, err)
	assert.ErrorContains(t, err, "storage broker open_db: open failed")
	assert.Equal(t, 1, openErrBroker.closeCalls)
	require.Len(t, openErrBroker.openRequests, 1)
	assert.Equal(t, state.Options.DBPath, openErrBroker.openRequests[0].DBPath)
	assert.NotEmpty(t, openErrBroker.openRequests[0].PayloadKey)
	assert.NotEqual(t, make([]byte, len(openErrBroker.openRequests[0].PayloadKey)), openErrBroker.openRequests[0].PayloadKey)
	assert.Equal(t, security.PayloadKeyVersionV1, openErrBroker.openRequests[0].PayloadVersion)

	successBroker := &wave37Broker{}
	startStorageBroker = func(context.Context) (storagebroker.API, error) { return successBroker, nil }
	err = phaseOpenDatabase().Run(context.Background(), state)
	require.NoError(t, err)
	assert.Same(t, successBroker, state.Broker)
	assert.Same(t, successBroker, state.Result.Broker)
	phaseOpenDatabase().Cleanup(state)
	assert.Nil(t, state.Broker)
	assert.Equal(t, 1, successBroker.closeCalls)
}

func TestWave37LoadSecurityStateDirectAndBrokerBranches(t *testing.T) {
	client, rawDB, err := openDatabase(filepath.Join(t.TempDir(), "security.db"), "", false, 0)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = client.Close()
		_ = rawDB.Close()
	})

	state := &State{RawDB: rawDB}
	err = phaseLoadSecurityState().Run(context.Background(), state)
	require.NoError(t, err)
	assert.True(t, state.FirstRun)
	assert.Nil(t, state.Salt)
	assert.Nil(t, state.Checksum)

	wantSalt := []byte("0123456789abcdef")
	wantChecksum := []byte("checksum")
	require.NoError(t, storeSalt(rawDB, wantSalt))
	require.NoError(t, storeChecksum(rawDB, wantChecksum))

	state = &State{RawDB: rawDB}
	err = phaseLoadSecurityState().Run(context.Background(), state)
	require.NoError(t, err)
	assert.False(t, state.FirstRun)
	assert.Equal(t, wantSalt, state.Salt)
	assert.Equal(t, wantChecksum, state.Checksum)

	broker := &wave37Broker{
		loadSecurityStateResult: storagebroker.LoadSecurityStateResult{
			Salt:     []byte("broker-salt"),
			Checksum: []byte("broker-checksum"),
			FirstRun: false,
		},
	}
	state = &State{Broker: broker}
	err = phaseLoadSecurityState().Run(context.Background(), state)
	require.NoError(t, err)
	assert.Equal(t, []byte("broker-salt"), state.Salt)
	assert.Equal(t, []byte("broker-checksum"), state.Checksum)
	assert.False(t, state.FirstRun)

	state = &State{
		Broker:   &wave37Broker{loadSecurityStateErr: errors.New("load failed")},
		Envelope: &security.MasterKeyEnvelope{PendingMigration: true},
	}
	err = phaseLoadSecurityState().Run(context.Background(), state)
	require.Error(t, err)
	assert.ErrorContains(t, err, "load security state for pending migration: load failed")
}

func TestWave37InitCryptoEnvelopeFirstRunAndMismatchBranches(t *testing.T) {
	env, mk, err := security.NewEnvelope("envelope-passphrase")
	require.NoError(t, err)
	defer security.ZeroBytes(mk)

	envelopeState := &State{Envelope: env, MasterKey: mk}
	err = phaseInitCrypto().Run(context.Background(), envelopeState)
	require.NoError(t, err)
	require.NotNil(t, envelopeState.Crypto)
	assert.Same(t, envelopeState.Crypto, envelopeState.Result.Crypto)
	ciphertext, err := envelopeState.Crypto.Encrypt(context.Background(), "local", []byte("payload"))
	require.NoError(t, err)
	plaintext, err := envelopeState.Crypto.Decrypt(context.Background(), "local", ciphertext)
	require.NoError(t, err)
	assert.Equal(t, []byte("payload"), plaintext)

	broker := &wave37Broker{}
	firstRunState := &State{
		Broker:     broker,
		FirstRun:   true,
		Passphrase: "first-run-passphrase",
	}
	err = phaseInitCrypto().Run(context.Background(), firstRunState)
	require.NoError(t, err)
	require.Len(t, broker.storedSalt, security.SaltSize)
	assert.NotEmpty(t, broker.storedChecksum)

	provider := security.NewLocalCryptoProvider()
	require.NoError(t, provider.Initialize("correct-passphrase"))
	mismatchState := &State{
		Passphrase: "wrong-passphrase",
		Salt:       provider.Salt(),
		Checksum:   provider.CalculateChecksum("correct-passphrase", provider.Salt()),
	}
	err = phaseInitCrypto().Run(context.Background(), mismatchState)
	require.Error(t, err)
	assert.ErrorContains(t, err, "passphrase checksum mismatch")
}

func TestWave37LoadProfileCreatesDefaultAndLoadsForcedProfile(t *testing.T) {
	client, rawDB, crypto := wave37ProfileDependencies(t)
	state := &State{Client: client, RawDB: rawDB, Crypto: crypto}

	err := phaseLoadProfile().Run(context.Background(), state)
	require.NoError(t, err)
	require.NotNil(t, state.Result.Config)
	assert.Equal(t, "default", state.Result.ProfileName)
	require.NotNil(t, state.Result.ConfigStore)
	require.NotNil(t, state.Result.Storage)
	exists, err := state.Result.Storage.ConfigProfiles().Exists(context.Background(), "default")
	require.NoError(t, err)
	assert.True(t, exists)

	client, rawDB, crypto = wave37ProfileDependencies(t)
	store := configstore.NewStore(client, crypto)
	custom := config.DefaultConfig()
	custom.ContextProfile = config.ContextProfileFull
	custom.Knowledge.Enabled = false
	explicit := map[string]bool{"knowledge.enabled": true}
	require.NoError(t, store.Save(context.Background(), "custom", custom, explicit))

	state = &State{
		Options: Options{ForceProfile: "custom"},
		Client:  client,
		RawDB:   rawDB,
		Crypto:  crypto,
	}
	err = phaseLoadProfile().Run(context.Background(), state)
	require.NoError(t, err)
	require.NotNil(t, state.Result.Config)
	assert.Equal(t, "custom", state.Result.ProfileName)
	assert.Equal(t, explicit, state.Result.ExplicitKeys)
	assert.False(t, state.Result.Config.Knowledge.Enabled)
	assert.True(t, state.Result.Config.Graph.Enabled)

	state = &State{
		Options: Options{ForceProfile: "missing"},
		Client:  client,
		RawDB:   rawDB,
		Crypto:  crypto,
	}
	err = phaseLoadProfile().Run(context.Background(), state)
	require.Error(t, err)
	assert.ErrorContains(t, err, `load profile "missing"`)
}

func wave37ProfileDependencies(t *testing.T) (*ent.Client, *sql.DB, security.CryptoProvider) {
	t.Helper()
	client, rawDB, err := openDatabase(filepath.Join(t.TempDir(), "profiles.db"), "", false, 0)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = client.Close()
		_ = rawDB.Close()
	})
	crypto := security.NewLocalCryptoProvider()
	require.NoError(t, crypto.Initialize("profile-passphrase"))
	return client, rawDB, crypto
}

type wave37Broker struct {
	stubBrokerClient

	openErr      error
	openRequests []storagebroker.OpenDBRequest
	closeCalls   int

	loadSecurityStateResult storagebroker.LoadSecurityStateResult
	loadSecurityStateErr    error
	storedSalt              []byte
	storedChecksum          []byte
}

func (b *wave37Broker) OpenDB(_ context.Context, req storagebroker.OpenDBRequest) (storagebroker.OpenDBResult, error) {
	req.MasterKey = append([]byte(nil), req.MasterKey...)
	req.PayloadKey = append([]byte(nil), req.PayloadKey...)
	b.openRequests = append(b.openRequests, req)
	if b.openErr != nil {
		return storagebroker.OpenDBResult{}, b.openErr
	}
	b.opened = true
	return storagebroker.OpenDBResult{Opened: true}, nil
}

func (b *wave37Broker) LoadSecurityState(context.Context) (storagebroker.LoadSecurityStateResult, error) {
	if b.loadSecurityStateErr != nil {
		return storagebroker.LoadSecurityStateResult{}, b.loadSecurityStateErr
	}
	return b.loadSecurityStateResult, nil
}

func (b *wave37Broker) StoreSalt(_ context.Context, salt []byte) error {
	b.storedSalt = append([]byte(nil), salt...)
	return nil
}

func (b *wave37Broker) StoreChecksum(_ context.Context, checksum []byte) error {
	b.storedChecksum = append([]byte(nil), checksum...)
	return nil
}

func (b *wave37Broker) Close(context.Context) error {
	b.closeCalls++
	return nil
}
