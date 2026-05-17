package security

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/keyring"
	internalsecurity "github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/storagebroker"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubKeyCheckerProvider struct {
	hasKey bool
}

type stubKMSCryptoProvider struct{}

func (stubKMSCryptoProvider) Sign(_ context.Context, _ string, payload []byte) ([]byte, error) {
	return append([]byte("sig:"), payload...), nil
}

func (stubKMSCryptoProvider) Encrypt(_ context.Context, _ string, plaintext []byte) ([]byte, error) {
	return append([]byte("enc:"), plaintext...), nil
}

func (stubKMSCryptoProvider) Decrypt(_ context.Context, _ string, ciphertext []byte) ([]byte, error) {
	return bytes.TrimPrefix(ciphertext, []byte("enc:")), nil
}

func (s stubKeyCheckerProvider) Get(service, key string) (string, error) {
	return "", keyring.ErrNotFound
}
func (s stubKeyCheckerProvider) Set(service, key, value string) error { return nil }
func (s stubKeyCheckerProvider) Delete(service, key string) error     { return nil }
func (s stubKeyCheckerProvider) HasKey(service, key string) bool      { return s.hasKey }

func executeSecurityCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func executeSecurityCmdWithInput(t *testing.T, cmd *cobra.Command, input string, args ...string) (string, string, error) {
	t.Helper()
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetIn(bytes.NewBufferString(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func installSecurityPromptSeams(
	t *testing.T,
	passphraseFn func(io.Writer, string) (string, error),
	confirmFn func(io.Writer, string, string) (string, error),
) {
	t.Helper()
	origRequire := securityRequireInteractiveInput
	origPassphrase := securityPassphrase
	origConfirm := securityPassphraseConfirm
	t.Cleanup(func() {
		securityRequireInteractiveInput = origRequire
		securityPassphrase = origPassphrase
		securityPassphraseConfirm = origConfirm
	})

	securityRequireInteractiveInput = func(io.Reader, string) error { return nil }
	if passphraseFn != nil {
		securityPassphrase = passphraseFn
	}
	if confirmFn != nil {
		securityPassphraseConfirm = confirmFn
	}
}

func persistentSecurityBootLoader(t *testing.T, cfg *config.Config) func() (*bootstrap.Result, error) {
	t.Helper()
	return func() (*bootstrap.Result, error) {
		store, err := session.NewEntStore(cfg.Session.DatabasePath)
		if err != nil {
			return nil, err
		}
		crypto := internalsecurity.NewLocalCryptoProvider()
		salt := bytes.Repeat([]byte{1}, internalsecurity.SaltSize)
		if err := crypto.InitializeWithSalt("test-passphrase", salt); err != nil {
			_ = store.Close()
			return nil, err
		}
		if err := store.SetSalt(internalsecurity.SecurityConfigDefault, salt); err != nil {
			_ = store.Close()
			return nil, err
		}
		checksum := crypto.CalculateChecksum("test-passphrase", salt)
		if err := store.SetChecksum(internalsecurity.SecurityConfigDefault, checksum); err != nil {
			_ = store.Close()
			return nil, err
		}
		facade := storage.NewFacade(
			nil,
			nil,
			storage.WithEntClient(store.Client()),
			storage.WithSessionClient(store.Client()),
		)
		return &bootstrap.Result{
			Config:  cfg,
			Storage: facade,
			Crypto:  crypto,
		}, nil
	}
}

func keyRegistryBootLoader(t *testing.T, cfg *config.Config, seed func(reg *internalsecurity.KeyRegistry)) func() (*bootstrap.Result, error) {
	t.Helper()
	return func() (*bootstrap.Result, error) {
		store, err := session.NewEntStore(cfg.Session.DatabasePath)
		if err != nil {
			return nil, err
		}
		facade := storage.NewFacade(nil, nil, storage.WithEntClient(store.Client()))
		if seed != nil {
			seed(facade.KeyRegistry())
		}
		return &bootstrap.Result{
			Config:  cfg,
			Storage: facade,
		}, nil
	}
}

func dummyBootLoader() func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	}
}

func TestNewSecurityCmd_Structure(t *testing.T) {
	cmd := NewSecurityCmd(dummyBootLoader())
	require.NotNil(t, cmd)

	assert.Equal(t, "security", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	expected := []string{
		"change-passphrase", "recovery", "migrate-passphrase",
		"secrets", "status",
		"keyring", "db-migrate", "db-decrypt", "kms",
	}

	subCmds := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subCmds[sub.Use] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestNewSecurityCmd_SubcommandCount(t *testing.T) {
	cmd := NewSecurityCmd(dummyBootLoader())
	// 9 subcommands: change-passphrase, recovery, migrate-passphrase (deprecated),
	// secrets, status, keyring, db-migrate, db-decrypt, kms
	assert.Equal(t, 9, len(cmd.Commands()), "expected 9 security subcommands")
}

func TestSecurityInteractiveGuardsUseCommandInput(t *testing.T) {
	guardErr := errors.New("interactive guard stopped command")
	prevRequire := securityRequireInteractiveInput
	prevDetect := detectSecureProvider
	t.Cleanup(func() {
		securityRequireInteractiveInput = prevRequire
		detectSecureProvider = prevDetect
	})

	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "local"
	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	}

	tests := []struct {
		name  string
		build func() *cobra.Command
		setup func()
	}{
		{
			name: "change-passphrase",
			build: func() *cobra.Command {
				return newChangePassphraseCmd(bootLoader)
			},
		},
		{
			name: "migrate-passphrase",
			build: func() *cobra.Command {
				return newMigratePassphraseCmd(bootLoader)
			},
		},
		{
			name: "keyring-store",
			build: func() *cobra.Command {
				return newKeyringStoreCmd(bootLoader)
			},
			setup: func() {
				detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
					return stubKeyCheckerProvider{}, keyring.TierBiometric
				}
			},
		},
		{
			name: "recovery-setup",
			build: func() *cobra.Command {
				return newRecoverySetupCmd(bootLoader)
			},
		},
		{
			name: "recovery-restore",
			build: func() *cobra.Command {
				return newRecoveryRestoreCmd()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detectSecureProvider = prevDetect
			if tt.setup != nil {
				tt.setup()
			}
			input := bytes.NewBufferString("typed input\n")
			var gotInput io.Reader
			securityRequireInteractiveInput = func(in io.Reader, message string) error {
				gotInput = in
				if strings.TrimSpace(message) == "" {
					t.Fatal("guard message should be actionable")
				}
				return guardErr
			}

			cmd := tt.build()
			cmd.SetIn(input)
			err := cmd.Execute()
			require.ErrorIs(t, err, guardErr)
			assert.Same(t, input, gotInput)
		})
	}
}

func TestSecretsCmd_HasSubcommands(t *testing.T) {
	cmd := NewSecurityCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "secrets" {
			secretsSubs := make(map[string]bool)
			for _, ssub := range sub.Commands() {
				secretsSubs[ssub.Use] = true
			}
			assert.True(t, secretsSubs["list"], "secrets should have list subcommand")
			assert.True(t, secretsSubs["set <name>"], "secrets should have set subcommand")
			assert.True(t, secretsSubs["delete <name>"], "secrets should have delete subcommand")
			return
		}
	}
	t.Fatal("secrets subcommand not found")
}

func TestKeyringCmd_HasSubcommands(t *testing.T) {
	cmd := NewSecurityCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "keyring" {
			keyringCmds := make(map[string]bool)
			for _, ksub := range sub.Commands() {
				keyringCmds[ksub.Use] = true
			}
			assert.True(t, keyringCmds["store"], "keyring should have store subcommand")
			assert.True(t, keyringCmds["clear"], "keyring should have clear subcommand")
			assert.True(t, keyringCmds["status"], "keyring should have status subcommand")
			return
		}
	}
	t.Fatal("keyring subcommand not found")
}

func TestKMSCmd_HasSubcommands(t *testing.T) {
	cmd := NewSecurityCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "kms" {
			kmsCmds := make(map[string]bool)
			for _, ksub := range sub.Commands() {
				kmsCmds[ksub.Use] = true
			}
			assert.True(t, kmsCmds["status"], "kms should have status subcommand")
			assert.True(t, kmsCmds["test"], "kms should have test subcommand")
			assert.True(t, kmsCmds["keys"], "kms should have keys subcommand")
			return
		}
	}
	t.Fatal("kms subcommand not found")
}

func TestKeyringStatus_WritesTextToCommandWriter(t *testing.T) {
	original := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return stubKeyCheckerProvider{hasKey: true}, keyring.TierBiometric
	}
	t.Cleanup(func() { detectSecureProvider = original })

	out, err := executeSecurityCmd(t, newKeyringStatusCmd())

	require.NoError(t, err)
	assert.Contains(t, out, "Hardware Keyring Status")
	assert.Contains(t, out, "Security Tier:   biometric")
	assert.Contains(t, out, "Has Passphrase:  true")
}

func TestKeyringStatus_WritesJSONToCommandWriter(t *testing.T) {
	original := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return stubKeyCheckerProvider{hasKey: false}, keyring.TierTPM
	}
	t.Cleanup(func() { detectSecureProvider = original })

	out, err := executeSecurityCmd(t, newKeyringStatusCmd(), "--output", "json")

	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, true, decoded["available"])
	assert.Equal(t, "tpm", decoded["security_tier"])
	assert.Equal(t, false, decoded["has_passphrase"])
}

func TestChangePassphrase_WritesSuccessToCommandWriter(t *testing.T) {
	original := executeChangePassphrase
	executeChangePassphrase = func(
		_ func() (*bootstrap.Result, error),
		_ io.Reader,
		_ io.Writer,
		_ io.Writer,
	) (string, error) {
		return "Passphrase changed. No data was re-encrypted.", nil
	}
	t.Cleanup(func() { executeChangePassphrase = original })

	out, err := executeSecurityCmd(t, newChangePassphraseCmd(dummyBootLoader()))

	require.NoError(t, err)
	assert.Contains(t, out, "Passphrase changed. No data was re-encrypted.")
}

func TestMigratePassphrase_WritesStatusToCommandWriter(t *testing.T) {
	original := executeMigratePassphrase
	executeMigratePassphrase = func(cmd *cobra.Command, _ func() (*bootstrap.Result, error)) error {
		fmt.Fprintln(cmd.OutOrStdout(), "This process will re-encrypt all your stored secrets with a new passphrase.")
		fmt.Fprintln(cmd.OutOrStdout(), "Migrating secrets...")
		fmt.Fprintln(cmd.OutOrStdout(), "Migration completed successfully!")
		return nil
	}
	t.Cleanup(func() { executeMigratePassphrase = original })

	out, err := executeSecurityCmd(t, newMigratePassphraseCmd(dummyBootLoader()))

	require.NoError(t, err)
	assert.Contains(t, out, "WARN: migrate-passphrase is deprecated.")
	assert.Contains(t, out, "This process will re-encrypt all your stored secrets with a new passphrase.")
	assert.Contains(t, out, "Migrating secrets...")
	assert.Contains(t, out, "Migration completed successfully!")
}

func TestRecoverySetup_WritesMnemonicAndSuccessToCommandWriter(t *testing.T) {
	original := executeRecoverySetup
	executeRecoverySetup = func(cmd *cobra.Command, _ func() (*bootstrap.Result, error)) error {
		fmt.Fprintln(cmd.OutOrStdout(), "RECOVERY MNEMONIC — write this down and store securely")
		fmt.Fprintln(cmd.OutOrStdout(), " 1. alpha")
		fmt.Fprintln(cmd.OutOrStdout(), "Recovery mnemonic slot added successfully.")
		return nil
	}
	t.Cleanup(func() { executeRecoverySetup = original })

	out, err := executeSecurityCmd(t, newRecoverySetupCmd(dummyBootLoader()))

	require.NoError(t, err)
	assert.Contains(t, out, "RECOVERY MNEMONIC — write this down and store securely")
	assert.Contains(t, out, "Recovery mnemonic slot added successfully.")
}

func TestRecoveryRestore_WritesSuccessToCommandWriter(t *testing.T) {
	original := executeRecoveryRestore
	executeRecoveryRestore = func(cmd *cobra.Command, _ io.Writer) error {
		fmt.Fprintln(cmd.OutOrStdout(), "Recovery complete. The new passphrase is now active.")
		return nil
	}
	t.Cleanup(func() { executeRecoveryRestore = original })

	out, err := executeSecurityCmd(t, newRecoveryRestoreCmd())

	require.NoError(t, err)
	assert.Contains(t, out, "Recovery complete. The new passphrase is now active.")
}

func TestChangePassphrase_WritesWarningsToCommandErrorWriter(t *testing.T) {
	original := executeChangePassphrase
	executeChangePassphrase = func(
		_ func() (*bootstrap.Result, error),
		_ io.Reader,
		_ io.Writer,
		errWriter io.Writer,
	) (string, error) {
		fmt.Fprintln(errWriter, "warning: keyring update failed: boom")
		return "Passphrase changed. No data was re-encrypted.", nil
	}
	t.Cleanup(func() { executeChangePassphrase = original })

	out, errOut, err := executeSecurityCmdWithInput(t, newChangePassphraseCmd(dummyBootLoader()), "")
	require.NoError(t, err)
	assert.Contains(t, out, "Passphrase changed. No data was re-encrypted.")
	assert.Contains(t, errOut, "warning: keyring update failed: boom")
}

func TestRecoveryRestore_WritesWarningsToCommandErrorWriter(t *testing.T) {
	original := executeRecoveryRestore
	executeRecoveryRestore = func(cmd *cobra.Command, errWriter io.Writer) error {
		fmt.Fprintln(errWriter, "warning: update keyfile: boom")
		fmt.Fprintln(cmd.OutOrStdout(), "Recovery complete. The new passphrase is now active.")
		return nil
	}
	t.Cleanup(func() { executeRecoveryRestore = original })

	out, errOut, err := executeSecurityCmdWithInput(t, newRecoveryRestoreCmd(), "")
	require.NoError(t, err)
	assert.Contains(t, out, "Recovery complete. The new passphrase is now active.")
	assert.Contains(t, errOut, "warning: update keyfile: boom")
}

func TestChangePassphrase_PromptsUseCommandWriter(t *testing.T) {
	installSecurityPromptSeams(
		t,
		func(out io.Writer, promptText string) (string, error) {
			fmt.Fprint(out, promptText)
			return "old-passphrase-12345", nil
		},
		func(out io.Writer, promptText, confirmPrompt string) (string, error) {
			fmt.Fprint(out, promptText)
			fmt.Fprint(out, confirmPrompt)
			return "new-passphrase-12345", nil
		},
	)
	originalDetect := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return nil, keyring.TierNone
	}
	t.Cleanup(func() { detectSecureProvider = originalDetect })

	dir := t.TempDir()
	provider := internalsecurity.NewLocalCryptoProvider()
	env, err := provider.InitializeNewEnvelope("old-passphrase-12345")
	require.NoError(t, err)
	require.NoError(t, internalsecurity.StoreEnvelopeFile(dir, env))

	cmd := newChangePassphraseCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Crypto:   provider,
			LangoDir: dir,
		}, nil
	})

	out, err := executeSecurityCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Enter CURRENT passphrase: ")
	assert.Contains(t, out, "Enter NEW passphrase: ")
	assert.Contains(t, out, "Confirm NEW passphrase: ")
	assert.Contains(t, out, "Passphrase changed. No data was re-encrypted.")
}

func TestMigratePassphrase_PromptsUseCommandWriter(t *testing.T) {
	installSecurityPromptSeams(
		t,
		nil,
		func(out io.Writer, promptText, confirmPrompt string) (string, error) {
			fmt.Fprint(out, promptText)
			fmt.Fprint(out, confirmPrompt)
			return "new-passphrase-12345", nil
		},
	)

	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "local"
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "migrate-prompts.db")
	cmd := newMigratePassphraseCmd(persistentSecurityBootLoader(t, cfg))

	out, err := executeSecurityCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Enter NEW passphrase: ")
	assert.Contains(t, out, "Confirm NEW passphrase: ")
	assert.Contains(t, out, "Migration completed successfully!")
}

func TestKeyringStore_PassphrasePromptUsesCommandWriter(t *testing.T) {
	installSecurityPromptSeams(
		t,
		func(out io.Writer, promptText string) (string, error) {
			fmt.Fprint(out, promptText)
			return "stored-passphrase", nil
		},
		nil,
	)
	originalDetect := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return stubKeyCheckerProvider{hasKey: false}, keyring.TierBiometric
	}
	t.Cleanup(func() { detectSecureProvider = originalDetect })

	cmd := newKeyringStoreCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeSecurityCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Enter passphrase to store: ")
	assert.Contains(t, out, "Passphrase stored with biometric protection.")
}

func TestRecoverySetup_PassphrasePromptUsesCommandWriter(t *testing.T) {
	installSecurityPromptSeams(
		t,
		func(out io.Writer, promptText string) (string, error) {
			fmt.Fprint(out, promptText)
			return "old-passphrase-12345", nil
		},
		nil,
	)

	provider := internalsecurity.NewLocalCryptoProvider()
	_, err := provider.InitializeNewEnvelope("old-passphrase-12345")
	require.NoError(t, err)
	cmd := newRecoverySetupCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Crypto: provider}, nil
	})

	out, err := executeSecurityCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "setup aborted")
	assert.Contains(t, out, "Enter current passphrase to authorize setup: ")
}

func TestRecoveryRestore_PromptsUseCommandWriter(t *testing.T) {
	mnemonic, err := internalsecurity.GenerateRecoveryMnemonic()
	require.NoError(t, err)
	installSecurityPromptSeams(
		t,
		func(out io.Writer, promptText string) (string, error) {
			fmt.Fprint(out, promptText)
			return mnemonic, nil
		},
		func(out io.Writer, promptText, confirmPrompt string) (string, error) {
			fmt.Fprint(out, promptText)
			fmt.Fprint(out, confirmPrompt)
			return "new-passphrase-12345", nil
		},
	)
	originalDetect := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return nil, keyring.TierNone
	}
	t.Cleanup(func() { detectSecureProvider = originalDetect })

	home := t.TempDir()
	t.Setenv("HOME", home)
	langoDir := filepath.Join(home, ".lango")
	require.NoError(t, os.MkdirAll(langoDir, 0o700))
	env, mk, err := internalsecurity.NewEnvelope("old-passphrase-12345")
	require.NoError(t, err)
	defer internalsecurity.ZeroBytes(mk)
	require.NoError(t, env.AddSlot(
		internalsecurity.KEKSlotMnemonic,
		"recovery",
		mk,
		mnemonic,
		internalsecurity.NewDefaultKDFParams(),
	))
	require.NoError(t, internalsecurity.StoreEnvelopeFile(langoDir, env))

	out, err := executeSecurityCmd(t, newRecoveryRestoreCmd())

	require.NoError(t, err)
	assert.Contains(t, out, "Enter 24-word recovery mnemonic: ")
	assert.Contains(t, out, "Enter NEW passphrase: ")
	assert.Contains(t, out, "Confirm NEW passphrase: ")
	assert.Contains(t, out, "Recovery complete. The new passphrase is now active.")
}

func TestSecurityStatus_NonInteractiveWarningUsesInjectedErrorWriter(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
	})

	acquireNonInteractivePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		return "", 0, assert.AnError
	}
	var errBuf bytes.Buffer

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))
	env, _, err := internalsecurity.NewEnvelope("test-passphrase")
	require.NoError(t, err)
	section := readDBStatusNonInteractive(dir, dbPath, env, true, &errBuf)
	assert.False(t, section.available)
	assert.Contains(t, errBuf.String(), "warning: status non-interactive passphrase:")
}

func TestSecurityStatus_NonInteractiveWarningUsesCommandErrorWriter(t *testing.T) {
	origReadStatus := readStatusDBNonInteractive
	t.Cleanup(func() {
		readStatusDBNonInteractive = origReadStatus
	})

	readStatusDBNonInteractive = func(
		langoDir, dbPath string,
		envelope *internalsecurity.MasterKeyEnvelope,
		needsKey bool,
		warningWriter io.Writer,
	) dbStatusResult {
		fmt.Fprintln(warningWriter, "warning: status non-interactive passphrase: boom")
		return dbStatusResult{}
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	langoDir := filepath.Join(home, ".lango")
	require.NoError(t, os.MkdirAll(langoDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(langoDir, "lango.db"), []byte{}, 0o600))
	env, _, err := internalsecurity.NewEnvelope("test-passphrase")
	require.NoError(t, err)
	require.NoError(t, internalsecurity.StoreEnvelopeFile(langoDir, env))

	cmd := newStatusCmd(dummyBootLoader())
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err = cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Security Status")
	assert.NotContains(t, out.String(), "warning: status non-interactive passphrase:")
	assert.Contains(t, errOut.String(), "warning: status non-interactive passphrase:")
}

func TestReadDBStatusNonInteractive_PassesDetectedSecureProvider(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	origDetect := detectSecureProvider
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
		detectSecureProvider = origDetect
	})

	wantProvider := stubKeyCheckerProvider{hasKey: true}
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return wantProvider, keyring.TierBiometric
	}

	acquireNonInteractivePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		assert.Equal(t, wantProvider, opts.KeyringProvider)
		return "", 0, passphrase.ErrNoNonInteractiveSource
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))

	got := readDBStatusNonInteractive(dir, dbPath, nil, true, io.Discard)
	assert.False(t, got.available)
}

func TestReadDBStatusNonInteractive_UsesBrokerStarterWithoutProcess(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	origStartBroker := statusStartBroker
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
		statusStartBroker = origStartBroker
	})

	acquireNonInteractivePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		return "legacy-passphrase", passphrase.SourceKeyfile, nil
	}

	var gotReq storagebroker.DBStatusSummaryRequest
	statusStartBroker = func(context.Context) (statusBroker, error) {
		return &stubStatusBroker{
			summary: storagebroker.DBStatusSummaryResult{
				Available:      true,
				EncryptionKeys: 2,
				StoredSecrets:  3,
			},
			onSummary: func(req storagebroker.DBStatusSummaryRequest) {
				gotReq = req
			},
		}, nil
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))

	got := readDBStatusNonInteractive(dir, dbPath, nil, true, io.Discard)
	assert.True(t, got.available)
	assert.Equal(t, 2, got.encryptionKeys)
	assert.Equal(t, 3, got.storedSecrets)
	assert.Equal(t, dbPath, gotReq.DBPath)
	assert.Equal(t, "legacy-passphrase", gotReq.EncryptionKey)
	assert.False(t, gotReq.RawKey)
}

func TestReadDBStatusNonInteractive_BrokerStartFailureDegrades(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	origStartBroker := statusStartBroker
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
		statusStartBroker = origStartBroker
	})

	acquireNonInteractivePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		return "legacy-passphrase", passphrase.SourceKeyfile, nil
	}
	statusStartBroker = func(context.Context) (statusBroker, error) {
		return nil, errors.New("broker unavailable")
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))

	got := readDBStatusNonInteractive(dir, dbPath, nil, true, io.Discard)
	assert.False(t, got.available)
	assert.Equal(t, 0, got.encryptionKeys)
	assert.Equal(t, 0, got.storedSecrets)
}

func TestReadDBStatusNonInteractive_LoadsActiveConfigFromBroker(t *testing.T) {
	origAcquire := acquireNonInteractivePassphrase
	origStartBroker := statusStartBroker
	t.Cleanup(func() {
		acquireNonInteractivePassphrase = origAcquire
		statusStartBroker = origStartBroker
	})

	acquireNonInteractivePassphrase = func(opts passphrase.Options) (string, passphrase.Source, error) {
		return "legacy-passphrase", passphrase.SourceKeyfile, nil
	}

	cfg := config.DefaultConfig()
	cfg.Security.Exportability.Enabled = true
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	broker := &stubStatusBroker{
		summary: storagebroker.DBStatusSummaryResult{Available: true},
		config:  storagebroker.ConfigLoadActiveResult{Config: raw},
	}
	statusStartBroker = func(context.Context) (statusBroker, error) {
		return broker, nil
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "status.db")
	require.NoError(t, os.WriteFile(dbPath, []byte{}, 0o600))

	got := readDBStatusNonInteractive(dir, dbPath, nil, true, io.Discard)
	require.NotNil(t, got.config)
	assert.True(t, got.config.Security.Exportability.Enabled)
	assert.True(t, broker.closed)
}

func TestConfirmWord_UsesCommandStreams(t *testing.T) {
	var out bytes.Buffer
	err := confirmWord(bytes.NewBufferString("alpha\n"), &out, []string{"alpha", "beta"}, 1)

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Enter word 1 to confirm: ")
}

func TestConfirmWord_AllowsMatchingPartialLineOnEOF(t *testing.T) {
	var out bytes.Buffer
	err := confirmWord(bytes.NewBufferString("alpha"), &out, []string{"alpha", "beta"}, 1)

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Enter word 1 to confirm: ")
}

func TestRecoverySetup_UsesCommandStreamsForWrittenDownConfirmation(t *testing.T) {
	var out bytes.Buffer
	ok, err := prompt.ConfirmDenyOnEOFIO(bytes.NewBufferString("y\n"), &out, "Have you written down all 24 words?")

	require.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, out.String(), "Have you written down all 24 words? [y/N]: ")
}

func TestKeyringClear_AbortUsesCommandStreams(t *testing.T) {
	original := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return nil, keyring.TierNone
	}
	t.Cleanup(func() { detectSecureProvider = original })
	t.Setenv("HOME", t.TempDir())

	out, errOut, err := executeSecurityCmdWithInput(t, newKeyringClearCmd(), "n\n")

	require.NoError(t, err)
	assert.Empty(t, errOut)
	assert.Contains(t, out, "Remove passphrase from all keyring backends? [y/N]: ")
	assert.Contains(t, out, "Aborted.")
}

func TestKeyringClear_ConfirmUsesCommandStreams(t *testing.T) {
	original := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return nil, keyring.TierNone
	}
	t.Cleanup(func() { detectSecureProvider = original })
	t.Setenv("HOME", t.TempDir())

	out, errOut, err := executeSecurityCmdWithInput(t, newKeyringClearCmd(), "y\n")

	require.NoError(t, err)
	assert.Empty(t, errOut)
	assert.Contains(t, out, "Remove passphrase from all keyring backends? [y/N]: ")
	assert.Contains(t, out, "No stored passphrase found in any backend.")
}

func TestKeyringClear_ForceWritesToCommandWriter(t *testing.T) {
	original := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return stubKeyCheckerProvider{hasKey: true}, keyring.TierBiometric
	}
	t.Cleanup(func() { detectSecureProvider = original })
	t.Setenv("HOME", t.TempDir())

	out, errOut, err := executeSecurityCmdWithInput(t, newKeyringClearCmd(), "", "--force")

	require.NoError(t, err)
	assert.Empty(t, errOut)
	assert.Contains(t, out, "Removed passphrase from secure provider.")
}

func TestKeyringClear_EOFAbortsWithoutClearing(t *testing.T) {
	original := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return nil, keyring.TierNone
	}
	t.Cleanup(func() { detectSecureProvider = original })
	t.Setenv("HOME", t.TempDir())

	out, errOut, err := executeSecurityCmdWithInput(t, newKeyringClearCmd(), "")

	require.NoError(t, err)
	assert.Empty(t, errOut)
	assert.Contains(t, out, "Remove passphrase from all keyring backends? [y/N]: ")
	assert.Contains(t, out, "Aborted.")
}

func TestKeyringClear_NonInteractiveRequiresForce(t *testing.T) {
	original := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return nil, keyring.TierNone
	}
	t.Cleanup(func() { detectSecureProvider = original })

	cmd := newKeyringClearCmd()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = reader.Close() })
	t.Cleanup(func() { _ = writer.Close() })
	cmd.SetIn(reader)

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "use --force for non-interactive deletion")
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestKeyringStore_AlreadyStoredWritesToCommandWriter(t *testing.T) {
	original := detectSecureProvider
	detectSecureProvider = func() (keyring.Provider, keyring.SecurityTier) {
		return stubKeyCheckerProvider{hasKey: true}, keyring.TierBiometric
	}
	t.Cleanup(func() { detectSecureProvider = original })

	cmd := newKeyringStoreCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	out, err := executeSecurityCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Passphrase is already stored in the secure keyring.")
	assert.Contains(t, out, "Next launch will load it automatically.")
}

func TestKMSStatus_WritesTextToCommandWriter(t *testing.T) {
	cmd := newKMSStatusCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.Security.Signer.Provider = "aws-kms"
		cfg.Security.KMS.KeyID = "arn:aws:kms:us-east-1:123456789012:key/example"
		cfg.Security.KMS.Region = "us-east-1"
		cfg.Security.KMS.FallbackToLocal = true
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeSecurityCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "KMS Status")
	assert.Contains(t, out, "Provider:      aws-kms")
	assert.Contains(t, out, "Key ID:        arn:aws:kms:us-east-1:123456789012:key/example")
}

func TestKMSStatus_WritesJSONToCommandWriter(t *testing.T) {
	cmd := newKMSStatusCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.Security.Signer.Provider = "aws-kms"
		cfg.Security.KMS.KeyID = "arn:aws:kms:us-east-1:123456789012:key/example"
		cfg.Security.KMS.Region = "us-east-1"
		cfg.Security.KMS.FallbackToLocal = false
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeSecurityCmd(t, cmd, "--output", "json")
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "aws-kms", decoded["provider"])
	assert.Equal(t, "arn:aws:kms:us-east-1:123456789012:key/example", decoded["key_id"])
	assert.Equal(t, "disabled", decoded["fallback"])
}

func TestKMSKeys_WritesEmptyStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "kms-keys.db")
	cmd := newKMSKeysCmd(keyRegistryBootLoader(t, cfg, nil))

	out, err := executeSecurityCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No keys registered.")
}

func TestKMSKeys_WritesTableOutputToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "kms-keys.db")
	cmd := newKMSKeysCmd(keyRegistryBootLoader(t, cfg, func(reg *internalsecurity.KeyRegistry) {
		_, err := reg.RegisterKey(context.Background(), "primary-signing", "arn:aws:kms:us-east-1:123456789012:key/example", internalsecurity.KeyTypeSigning)
		require.NoError(t, err)
	}))

	out, err := executeSecurityCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "REMOTE KEY ID")
	assert.Contains(t, out, "primary-signing")
	assert.Contains(t, out, "signing")
}

func TestKMSKeys_WritesJSONToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "kms-keys.db")
	cmd := newKMSKeysCmd(keyRegistryBootLoader(t, cfg, func(reg *internalsecurity.KeyRegistry) {
		_, err := reg.RegisterKey(context.Background(), "default-encryption", "arn:aws:kms:us-east-1:123456789012:key/example", internalsecurity.KeyTypeEncryption)
		require.NoError(t, err)
	}))

	out, err := executeSecurityCmd(t, cmd, "--output", "json")
	require.NoError(t, err)
	var decoded []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "default-encryption", decoded[0]["Name"])
	assert.Equal(t, "encryption", decoded[0]["Type"])
}

func TestKMSTest_WritesToCommandWriter(t *testing.T) {
	original := newKMSProvider
	newKMSProvider = func(_ internalsecurity.KMSProviderName, _ config.KMSConfig) (internalsecurity.CryptoProvider, error) {
		return stubKMSCryptoProvider{}, nil
	}
	t.Cleanup(func() { newKMSProvider = original })

	cmd := newKMSTestCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.Security.Signer.Provider = "aws-kms"
		cfg.Security.KMS.KeyID = "arn:aws:kms:us-east-1:123456789012:key/example"
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeSecurityCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, `Testing KMS roundtrip with key "arn:aws:kms:us-east-1:123456789012:key/example"...`)
	assert.Contains(t, out, "  Encrypt: OK")
	assert.Contains(t, out, "  Decrypt: OK")
	assert.Contains(t, out, "  Roundtrip: PASS")
}

func TestKMSWrap_WritesToCommandWriter(t *testing.T) {
	original := newKMSProvider
	newKMSProvider = func(_ internalsecurity.KMSProviderName, _ config.KMSConfig) (internalsecurity.CryptoProvider, error) {
		return stubKMSCryptoProvider{}, nil
	}
	t.Cleanup(func() { newKMSProvider = original })

	dir := t.TempDir()
	provider := internalsecurity.NewLocalCryptoProvider()
	env, err := provider.InitializeNewEnvelope("test-passphrase-12345")
	require.NoError(t, err)
	require.NoError(t, internalsecurity.StoreEnvelopeFile(dir, env))

	cmd := newKMSWrapCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.Security.Signer.Provider = "aws-kms"
		return &bootstrap.Result{Config: cfg, Crypto: provider, LangoDir: dir}, nil
	})

	out, err := executeSecurityCmd(t, cmd, "--provider", "aws-kms", "--key-id", "arn:aws:kms:us-east-1:123456789012:key/example")
	require.NoError(t, err)
	assert.Contains(t, out, "KMS slot added (provider=aws-kms, keyID=arn:aws:kms:us-east-1:123456789012:key/example)")
	assert.Contains(t, out, "Next bootstrap can use KMS for passphraseless unlock.")
}

func TestKMSDetach_WritesToCommandWriter(t *testing.T) {
	dir := t.TempDir()
	provider := internalsecurity.NewLocalCryptoProvider()
	env, err := provider.InitializeNewEnvelope("test-passphrase-12345")
	require.NoError(t, err)

	mk := provider.MasterKey()
	require.NotNil(t, mk)
	defer internalsecurity.ZeroBytes(mk)
	require.NoError(t, env.AddKMSSlot(context.Background(), "kms", mk, stubKMSCryptoProvider{}, "aws-kms", "arn:aws:kms:us-east-1:123456789012:key/example"))
	require.NoError(t, internalsecurity.StoreEnvelopeFile(dir, env))

	cmd := newKMSDetachCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		return &bootstrap.Result{Config: cfg, Crypto: provider, LangoDir: dir}, nil
	})

	out, err := executeSecurityCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "KMS slot ")
	assert.Contains(t, out, " removed.")
}

func TestKMSDetach_MultipleSlotsGuidanceUsesCommandWriter(t *testing.T) {
	dir := t.TempDir()
	provider := internalsecurity.NewLocalCryptoProvider()
	env, err := provider.InitializeNewEnvelope("test-passphrase-12345")
	require.NoError(t, err)

	mk := provider.MasterKey()
	require.NotNil(t, mk)
	defer internalsecurity.ZeroBytes(mk)
	require.NoError(t, env.AddKMSSlot(context.Background(), "kms-a", mk, stubKMSCryptoProvider{}, "aws-kms", "key-a"))
	require.NoError(t, env.AddKMSSlot(context.Background(), "kms-b", mk, stubKMSCryptoProvider{}, "aws-kms", "key-b"))
	require.NoError(t, internalsecurity.StoreEnvelopeFile(dir, env))

	cmd := newKMSDetachCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		return &bootstrap.Result{Config: cfg, Crypto: provider, LangoDir: dir}, nil
	})

	out, err := executeSecurityCmd(t, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--slot-id required when multiple KMS slots exist")
	assert.Contains(t, out, "Multiple KMS slots found. Specify --slot-id:")
	assert.Contains(t, out, "provider=aws-kms")
}

func TestSecretsList_WritesEmptyStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-secrets.db")
	cmd := newSecretsListCmd(persistentSecurityBootLoader(t, cfg))

	out, err := executeSecurityCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No secrets stored.")
}

func TestSecretsSetAndList_UseCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-secrets.db")
	bootLoader := persistentSecurityBootLoader(t, cfg)

	setCmd := newSecretsSetCmd(bootLoader)
	out, err := executeSecurityCmd(t, setCmd, "api-key", "--value-hex", "746f6b656e")
	require.NoError(t, err)
	assert.Contains(t, out, "Secret 'api-key' stored successfully.")

	listCmd := newSecretsListCmd(bootLoader)
	listOut, err := executeSecurityCmd(t, listCmd)
	require.NoError(t, err)
	assert.Contains(t, listOut, "NAME")
	assert.Contains(t, listOut, "api-key")

	jsonOut, err := executeSecurityCmd(t, listCmd, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, jsonOut, `"name": "api-key"`)
}

func TestSecretsSetValueHexDoesNotPrompt(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-secrets-hex.db")
	bootLoader := persistentSecurityBootLoader(t, cfg)

	prevRequire := secretsRequireInteractiveInput
	prevPassphrase := secretsPassphrase
	t.Cleanup(func() {
		secretsRequireInteractiveInput = prevRequire
		secretsPassphrase = prevPassphrase
	})

	secretsRequireInteractiveInput = func(io.Reader, string) error {
		t.Fatal("interactive terminal check should not run with --value-hex")
		return nil
	}
	secretsPassphrase = func(out io.Writer, promptText string) (string, error) {
		t.Fatal("passphrase prompt should not run with --value-hex")
		return "", nil
	}

	setCmd := newSecretsSetCmd(bootLoader)
	out, err := executeSecurityCmd(t, setCmd, "api-key", "--value-hex", "746f6b656e")

	require.NoError(t, err)
	assert.NotContains(t, out, "Enter secret value")
	assert.Contains(t, out, "Secret 'api-key' stored successfully.")
}

func TestSecretsSetInteractivePromptUsesCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-secrets-interactive.db")
	bootLoader := persistentSecurityBootLoader(t, cfg)

	prevRequire := secretsRequireInteractiveInput
	prevPassphrase := secretsPassphrase
	t.Cleanup(func() {
		secretsRequireInteractiveInput = prevRequire
		secretsPassphrase = prevPassphrase
	})

	secretsRequireInteractiveInput = func(io.Reader, string) error {
		return nil
	}
	secretsPassphrase = func(out io.Writer, promptText string) (string, error) {
		fmt.Fprint(out, promptText)
		fmt.Fprintln(out)
		return "token", nil
	}

	setCmd := newSecretsSetCmd(bootLoader)
	out, err := executeSecurityCmd(t, setCmd, "api-key")

	require.NoError(t, err)
	assert.Contains(t, out, "Enter secret value: \n")
	assert.Contains(t, out, "Secret 'api-key' stored successfully.")
}

func TestSecretsSetInteractiveGuardUsesCommandInput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-secrets-guard.db")
	bootLoader := persistentSecurityBootLoader(t, cfg)

	guardErr := errors.New("interactive guard stopped command")
	prevRequire := secretsRequireInteractiveInput
	t.Cleanup(func() {
		secretsRequireInteractiveInput = prevRequire
	})

	input := bytes.NewBufferString("typed input\n")
	var gotInput io.Reader
	secretsRequireInteractiveInput = func(in io.Reader, message string) error {
		gotInput = in
		if !strings.Contains(message, "--value-hex") {
			t.Fatalf("guard message should mention --value-hex, got %q", message)
		}
		return guardErr
	}

	cmd := newSecretsSetCmd(bootLoader)
	cmd.SetIn(input)
	cmd.SetArgs([]string{"api-key"})
	err := cmd.Execute()
	require.ErrorIs(t, err, guardErr)
	assert.Same(t, input, gotInput)
}

func TestSecurityInspectionCommands_InvalidOutputFailFast(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-invalid-output.db")
	bootLoader := persistentSecurityBootLoader(t, cfg)

	commands := []*cobra.Command{
		newStatusCmd(func() (*bootstrap.Result, error) {
			t.Fatal("boot loader should not be called for invalid security status output")
			return nil, nil
		}),
		newKMSStatusCmd(func() (*bootstrap.Result, error) {
			t.Fatal("boot loader should not be called for invalid kms status output")
			return nil, nil
		}),
		newKMSKeysCmd(func() (*bootstrap.Result, error) {
			t.Fatal("boot loader should not be called for invalid kms keys output")
			return nil, nil
		}),
		newSecretsListCmd(func() (*bootstrap.Result, error) {
			t.Fatal("boot loader should not be called for invalid secrets list output")
			return nil, nil
		}),
		newKeyringStatusCmd(),
	}

	for _, cmd := range commands {
		_, err := executeSecurityCmd(t, cmd, "--output", "yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unknown output format "yaml"`)
	}

	_ = bootLoader
}

func TestSecretsDelete_UsesCommandStreams(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-secrets.db")
	bootLoader := persistentSecurityBootLoader(t, cfg)

	setCmd := newSecretsSetCmd(bootLoader)
	_, err := executeSecurityCmd(t, setCmd, "api-key", "--value-hex", "746f6b656e")
	require.NoError(t, err)

	deleteCmd := newSecretsDeleteCmd(bootLoader)
	abortOut, abortErrOut, err := executeSecurityCmdWithInput(t, deleteCmd, "n\n", "api-key")
	require.NoError(t, err)
	assert.Empty(t, abortErrOut)
	assert.Contains(t, abortOut, "Delete secret 'api-key'? [y/N]: ")
	assert.Contains(t, abortOut, "Aborted.")

	confirmCmd := newSecretsDeleteCmd(bootLoader)
	confirmOut, confirmErrOut, err := executeSecurityCmdWithInput(t, confirmCmd, "y\n", "api-key")
	require.NoError(t, err)
	assert.Empty(t, confirmErrOut)
	assert.Contains(t, confirmOut, "Delete secret 'api-key'? [y/N]: ")
	assert.Contains(t, confirmOut, "Secret 'api-key' deleted.")
}

func TestSecretsDelete_EOFAbortsWithoutDeleting(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-secrets-eof.db")
	bootLoader := persistentSecurityBootLoader(t, cfg)

	setCmd := newSecretsSetCmd(bootLoader)
	_, err := executeSecurityCmd(t, setCmd, "api-key", "--value-hex", "746f6b656e")
	require.NoError(t, err)

	deleteCmd := newSecretsDeleteCmd(bootLoader)
	abortOut, abortErrOut, err := executeSecurityCmdWithInput(t, deleteCmd, "", "api-key")
	require.NoError(t, err)
	assert.Empty(t, abortErrOut)
	assert.Contains(t, abortOut, "Delete secret 'api-key'? [y/N]: ")
	assert.Contains(t, abortOut, "Aborted.")
}

func TestSecretsDelete_NonInteractiveRequiresForce(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "security-secrets-noninteractive.db")
	bootLoader := persistentSecurityBootLoader(t, cfg)

	setCmd := newSecretsSetCmd(bootLoader)
	_, err := executeSecurityCmd(t, setCmd, "api-key", "--value-hex", "746f6b656e")
	require.NoError(t, err)

	cmd := newSecretsDeleteCmd(bootLoader)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = reader.Close() })
	t.Cleanup(func() { _ = writer.Close() })
	cmd.SetIn(reader)
	cmd.SetArgs([]string{"api-key"})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "use --force for non-interactive deletion")
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestBoolToStatus(t *testing.T) {
	assert.Equal(t, "enabled", boolToStatus(true))
	assert.Equal(t, "disabled", boolToStatus(false))
}

func TestRenderStatus_IncludesExportability(t *testing.T) {
	out := statusOutput{
		SignerProvider:       "local",
		ApprovalPolicy:       "dangerous",
		DBEncryption:         "disabled (plaintext)",
		ExportabilityEnabled: true,
	}

	var stdout bytes.Buffer
	err := renderStatus(&stdout, out, "table")
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Exportability:")
	assert.Contains(t, stdout.String(), "enabled")

	var jsonOut bytes.Buffer
	err = renderStatus(&jsonOut, out, "json")
	require.NoError(t, err)
	var decoded struct {
		ExportabilityEnabled bool `json:"exportability_enabled"`
	}
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &decoded))
	assert.True(t, decoded.ExportabilityEnabled)
}

func TestRenderStatus_KMSProviderAndFallbackSurface(t *testing.T) {
	out := statusOutput{
		SignerProvider:       "aws-kms",
		ApprovalPolicy:       "dangerous",
		DBEncryption:         "deprecated config (ignored)",
		ExportabilityEnabled: false,
		KMSProvider:          "aws-kms",
		KMSKeyID:             "arn:aws:kms:us-east-1:123456789012:key/example",
		KMSFallback:          "enabled",
		DBAvailable:          true,
	}

	var stdout bytes.Buffer
	err := renderStatus(&stdout, out, "table")
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Signer Provider:    aws-kms")
	assert.Contains(t, stdout.String(), "DB Encryption:      deprecated config (ignored)")
	assert.Contains(t, stdout.String(), "KMS Provider:       aws-kms")
	assert.Contains(t, stdout.String(), "KMS Fallback:       enabled")
}

func TestRenderStatus_JSONPreservesCurrentFieldSemantics(t *testing.T) {
	out := statusOutput{
		SignerProvider:       "unavailable",
		ApprovalPolicy:       "dangerous",
		DBEncryption:         "legacy encrypted or unreadable DB (unsupported)",
		ExportabilityEnabled: true,
		KMSProvider:          "gcp-kms",
		KMSKeyID:             "projects/demo/locations/global/keyRings/ring/cryptoKeys/key",
		KMSFallback:          "disabled",
		DBAvailable:          false,
	}

	var jsonOut bytes.Buffer
	err := renderStatus(&jsonOut, out, "json")
	require.NoError(t, err)

	var decoded struct {
		SignerProvider       string `json:"signer_provider"`
		DBEncryption         string `json:"db_encryption"`
		KMSProvider          string `json:"kms_provider"`
		KMSFallback          string `json:"kms_fallback"`
		ExportabilityEnabled bool   `json:"exportability_enabled"`
		DBAvailable          bool   `json:"db_available"`
	}
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &decoded))
	assert.Equal(t, "unavailable", decoded.SignerProvider)
	assert.Equal(t, "legacy encrypted or unreadable DB (unsupported)", decoded.DBEncryption)
	assert.Equal(t, "gcp-kms", decoded.KMSProvider)
	assert.Equal(t, "disabled", decoded.KMSFallback)
	assert.True(t, decoded.ExportabilityEnabled)
	assert.False(t, decoded.DBAvailable)
}

func TestLoadActiveStatusConfig_UsesExportabilitySetting(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Exportability.Enabled = true
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	loaded, ok := loadActiveStatusConfig(&stubActiveConfigLoader{
		result: storagebroker.ConfigLoadActiveResult{Config: raw},
	})
	require.True(t, ok)
	require.NotNil(t, loaded)
	assert.True(t, loaded.Security.Exportability.Enabled)
}

type stubActiveConfigLoader struct {
	result storagebroker.ConfigLoadActiveResult
	err    error
}

func (s *stubActiveConfigLoader) ConfigLoadActive(context.Context) (storagebroker.ConfigLoadActiveResult, error) {
	return s.result, s.err
}

type stubStatusBroker struct {
	summary   storagebroker.DBStatusSummaryResult
	config    storagebroker.ConfigLoadActiveResult
	err       error
	configErr error
	onSummary func(storagebroker.DBStatusSummaryRequest)
	closed    bool
}

func (s *stubStatusBroker) DBStatusSummary(
	_ context.Context,
	req storagebroker.DBStatusSummaryRequest,
) (storagebroker.DBStatusSummaryResult, error) {
	if s.onSummary != nil {
		s.onSummary(req)
	}
	return s.summary, s.err
}

func (s *stubStatusBroker) ConfigLoadActive(context.Context) (storagebroker.ConfigLoadActiveResult, error) {
	return s.config, s.configErr
}

func (s *stubStatusBroker) Close(context.Context) error {
	s.closed = true
	return nil
}

func TestIsKMSProvider(t *testing.T) {
	assert.True(t, isKMSProvider("aws-kms"))
	assert.True(t, isKMSProvider("gcp-kms"))
	assert.True(t, isKMSProvider("azure-kv"))
	assert.True(t, isKMSProvider("pkcs11"))
	assert.False(t, isKMSProvider("local"))
	assert.False(t, isKMSProvider("rpc"))
	assert.False(t, isKMSProvider(""))
}
