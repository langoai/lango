package bootstrap

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/sqlitedriver"
)

func TestEnsureDataDirResolvesDefaultsAndPreservesExplicitPaths(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "custom-lango")
	explicitDB := filepath.Join(t.TempDir(), "explicit.db")
	explicitKeyfile := filepath.Join(t.TempDir(), "explicit-keyfile")
	state := &State{
		Options: Options{
			LangoDir:    dir,
			DBPath:      explicitDB,
			KeyfilePath: explicitKeyfile,
		},
	}

	err := phaseEnsureDataDir().Run(context.Background(), state)
	require.NoError(t, err)

	assert.NotEmpty(t, state.Home)
	assert.Equal(t, dir, state.LangoDir)
	assert.Equal(t, dir, state.Result.LangoDir)
	assert.Equal(t, explicitDB, state.Options.DBPath)
	assert.Equal(t, explicitKeyfile, state.Options.KeyfilePath)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(dataDirPerm), info.Mode().Perm())

	skillsInfo, err := os.Stat(filepath.Join(dir, "skills"))
	require.NoError(t, err)
	assert.True(t, skillsInfo.IsDir())
	assert.Equal(t, os.FileMode(dataDirPerm), skillsInfo.Mode().Perm())

	_, err = os.Stat(filepath.Join(dir, ".write-test"))
	assert.True(t, os.IsNotExist(err))
}

func TestDetectEncryptionHandlesMissingPlainAndLegacyHeaders(t *testing.T) {
	t.Parallel()

	missingState := &State{Options: Options{DBPath: filepath.Join(t.TempDir(), "missing.db")}}
	err := phaseDetectEncryption().Run(context.Background(), missingState)
	require.NoError(t, err)
	assert.False(t, missingState.DBEncrypted)
	assert.False(t, missingState.NeedsDBKey)

	plainPath := filepath.Join(t.TempDir(), "plain.db")
	require.NoError(t, os.WriteFile(plainPath, []byte("SQLite format 3\x00more"), 0600))
	plainState := &State{Options: Options{DBPath: plainPath}, NeedsDBKey: true}
	err = phaseDetectEncryption().Run(context.Background(), plainState)
	require.NoError(t, err)
	assert.False(t, plainState.DBEncrypted)
	assert.False(t, plainState.NeedsDBKey)

	legacyPath := filepath.Join(t.TempDir(), "legacy.db")
	require.NoError(t, os.WriteFile(legacyPath, []byte("not sqlite header"), 0600))
	legacyState := &State{Options: Options{DBPath: legacyPath}}
	err = phaseDetectEncryption().Run(context.Background(), legacyState)
	require.Error(t, err)
	assert.True(t, legacyState.DBEncrypted)
	assert.ErrorIs(t, err, sqlitedriver.ErrLegacyEncryptedOrUnreadableDB)
}

func TestUnwrapOrCreateMasterKeyBranches(t *testing.T) {
	t.Parallel()

	existingMK := []byte("12345678901234567890123456789012")
	existingState := &State{
		LangoDir:      t.TempDir(),
		Passphrase:    "existing-passphrase",
		FirstRunGuess: true,
		MasterKey:     append([]byte(nil), existingMK...),
	}
	err := phaseUnwrapOrCreateMK().Run(context.Background(), existingState)
	require.NoError(t, err)
	assert.Equal(t, existingMK, existingState.MasterKey)
	assert.Nil(t, existingState.Envelope)

	firstRunDir := t.TempDir()
	firstRunState := &State{
		LangoDir:      firstRunDir,
		Passphrase:    "first-run-passphrase",
		FirstRunGuess: true,
	}
	err = phaseUnwrapOrCreateMK().Run(context.Background(), firstRunState)
	require.NoError(t, err)
	require.NotNil(t, firstRunState.Envelope)
	require.Len(t, firstRunState.MasterKey, security.KeySize)
	assert.True(t, security.HasEnvelopeFile(firstRunDir))

	legacyState := &State{FirstRunGuess: false}
	err = phaseUnwrapOrCreateMK().Run(context.Background(), legacyState)
	require.NoError(t, err)
	assert.True(t, legacyState.LegacyMode)
	assert.Nil(t, legacyState.MasterKey)
}

func TestMigrateEnvelopeRejectsBrokerAndPendingRekeyBranches(t *testing.T) {
	t.Parallel()

	err := phaseMigrateEnvelope().Run(context.Background(), &State{
		LegacyMode: true,
		Broker:     &stubBrokerClient{},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "legacy envelope migration requires non-broker bootstrap path")

	err = phaseMigrateEnvelope().Run(context.Background(), &State{
		Envelope: &security.MasterKeyEnvelope{PendingMigration: true},
		Broker:   &stubBrokerClient{},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "pending envelope migration retry requires non-broker bootstrap path")

	err = phaseMigrateEnvelope().Run(context.Background(), &State{
		Envelope: &security.MasterKeyEnvelope{PendingMigration: true},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "pending migration retry requires unwrapped master key")

	err = phaseMigrateEnvelope().Run(context.Background(), &State{
		Envelope: &security.MasterKeyEnvelope{PendingRekey: true},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, sqlitedriver.ErrLegacyEncryptedOrUnreadableDB))
}

func TestDeriveKeyPhasesNoopAndPopulateResult(t *testing.T) {
	t.Parallel()

	emptyState := &State{}
	require.NoError(t, phaseDeriveIdentityKey().Run(context.Background(), emptyState))
	require.NoError(t, phaseDerivePQKey().Run(context.Background(), emptyState))
	assert.Nil(t, emptyState.IdentityKey)
	assert.Nil(t, emptyState.Result.IdentityKey)
	assert.Nil(t, emptyState.PQSigningKeySeed)
	assert.Nil(t, emptyState.Result.PQSigningKeySeed)

	keyedState := &State{MasterKey: []byte("12345678901234567890123456789012")}
	require.NoError(t, phaseDeriveIdentityKey().Run(context.Background(), keyedState))
	require.NoError(t, phaseDerivePQKey().Run(context.Background(), keyedState))

	assert.NotNil(t, keyedState.IdentityKey)
	assert.Equal(t, keyedState.IdentityKey, keyedState.Result.IdentityKey)
	require.Len(t, keyedState.PQSigningKeySeed, 32)
	assert.Equal(t, keyedState.PQSigningKeySeed, keyedState.Result.PQSigningKeySeed)
}
