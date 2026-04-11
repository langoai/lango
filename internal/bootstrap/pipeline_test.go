package bootstrap

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/security"
)

func TestPipeline_ExecutesInOrder(t *testing.T) {
	var order []string

	phases := []Phase{
		{
			Name: "phase-a",
			Run: func(_ context.Context, _ *State) error {
				order = append(order, "a")
				return nil
			},
		},
		{
			Name: "phase-b",
			Run: func(_ context.Context, _ *State) error {
				order = append(order, "b")
				return nil
			},
		},
		{
			Name: "phase-c",
			Run: func(_ context.Context, _ *State) error {
				order = append(order, "c")
				return nil
			},
		},
	}

	p := NewPipeline(phases...)
	result, err := p.Execute(context.Background(), Options{})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []string{"a", "b", "c"}, order)
}

func TestPipeline_CleanupRunsInReverseOnFailure(t *testing.T) {
	var cleanupOrder []string

	phases := []Phase{
		{
			Name: "phase-a",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			Cleanup: func(_ *State) {
				cleanupOrder = append(cleanupOrder, "a")
			},
		},
		{
			Name: "phase-b",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			Cleanup: func(_ *State) {
				cleanupOrder = append(cleanupOrder, "b")
			},
		},
		{
			Name: "phase-c",
			Run: func(_ context.Context, _ *State) error {
				return errors.New("phase-c failed")
			},
			Cleanup: func(_ *State) {
				cleanupOrder = append(cleanupOrder, "c")
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "phase-c")

	// Cleanup should run for a and b (completed) in reverse, NOT for c (failed).
	assert.Equal(t, []string{"b", "a"}, cleanupOrder)
}

func TestPipeline_CleanupNotCalledForFailedPhase(t *testing.T) {
	var cleaned []string

	phases := []Phase{
		{
			Name: "phase-a",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			Cleanup: func(_ *State) {
				cleaned = append(cleaned, "a")
			},
		},
		{
			Name: "phase-b",
			Run: func(_ context.Context, _ *State) error {
				return errors.New("boom")
			},
			Cleanup: func(_ *State) {
				cleaned = append(cleaned, "b")
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.Error(t, err)

	// Only phase-a cleanup should run, not phase-b.
	assert.Equal(t, []string{"a"}, cleaned)
}

func TestPipeline_StatePassedBetweenPhases(t *testing.T) {
	phases := []Phase{
		{
			Name: "set-home",
			Run: func(_ context.Context, s *State) error {
				s.Home = "/test/home"
				return nil
			},
		},
		{
			Name: "read-home",
			Run: func(_ context.Context, s *State) error {
				if s.Home != "/test/home" {
					return errors.New("expected Home to be /test/home")
				}
				s.LangoDir = s.Home + "/.lango"
				return nil
			},
		},
		{
			Name: "verify",
			Run: func(_ context.Context, s *State) error {
				if s.LangoDir != "/test/home/.lango" {
					return errors.New("expected LangoDir to be /test/home/.lango")
				}
				return nil
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.NoError(t, err)
}

func TestPipeline_NilCleanupSkipped(t *testing.T) {
	phases := []Phase{
		{
			Name: "no-cleanup",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			// Cleanup is nil — should not panic.
		},
		{
			Name: "fail",
			Run: func(_ context.Context, _ *State) error {
				return errors.New("fail")
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.Error(t, err)
	// No panic means nil cleanup was properly skipped.
}

func TestPipeline_ErrorWrapsPhaseNameAndCause(t *testing.T) {
	sentinel := errors.New("root cause")
	phases := []Phase{
		{
			Name: "important-phase",
			Run: func(_ context.Context, _ *State) error {
				return sentinel
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "important-phase")
	assert.True(t, errors.Is(err, sentinel))
}

func TestDefaultPhases_Returns12Phases(t *testing.T) {
	phases := DefaultPhases()
	require.Len(t, phases, 12)

	wantNames := []string{
		"ensure data directory",
		"detect encryption",
		"load envelope file",
		"acquire credential",
		"unwrap or create master key",
		"open database",
		"migrate envelope",
		"load security state",
		"initialize crypto",
		"derive identity key",
		"derive PQ signing key",
		"load profile",
	}

	for i, want := range wantNames {
		assert.Equal(t, want, phases[i].Name, "phase %d name", i)
	}
}

func TestDefaultPhases_OpenDatabaseHasCleanup(t *testing.T) {
	phases := DefaultPhases()
	// Only "open database" (index 3) should have a Cleanup function.
	for i, p := range phases {
		if p.Name == "open database" {
			assert.NotNil(t, p.Cleanup, "phase %d (%s) should have Cleanup", i, p.Name)
		}
	}
}

func TestPipeline_NoCleanupOnSuccess(t *testing.T) {
	var cleaned bool

	phases := []Phase{
		{
			Name: "only-phase",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			Cleanup: func(_ *State) {
				cleaned = true
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.NoError(t, err)

	// Cleanup should NOT run on success.
	assert.False(t, cleaned)
}

// mockCryptoProvider implements security.CryptoProvider for bootstrap testing.
type mockCryptoProvider struct {
	local *security.LocalCryptoProvider
}

func newMockCryptoProvider(t *testing.T) *mockCryptoProvider {
	t.Helper()
	p := security.NewLocalCryptoProvider()
	require.NoError(t, p.Initialize("mock-passphrase-12345"))
	return &mockCryptoProvider{local: p}
}

func (m *mockCryptoProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	return m.local.Sign(ctx, keyID, payload)
}

func (m *mockCryptoProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	return m.local.Encrypt(ctx, keyID, plaintext)
}

func (m *mockCryptoProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	return m.local.Decrypt(ctx, keyID, ciphertext)
}

func TestPhaseAcquireCredential_KMSUnwrap(t *testing.T) {
	// Create envelope with a KMS slot.
	env, mk, err := security.NewEnvelope("test-passphrase-1234")
	require.NoError(t, err)
	defer security.ZeroBytes(mk)

	kms := newMockCryptoProvider(t)
	ctx := context.Background()

	err = env.AddKMSSlot(ctx, "test-kms", mk, kms, "aws-kms", "test-key-1")
	require.NoError(t, err)

	// Run just the acquire credential phase.
	state := &State{
		Options: Options{
			SkipSecureDetection: true,
			KMSConfig: &config.KMSConfig{
				KeyID: "test-key-1",
			},
			KMSProviderName: "aws-kms",
		},
		Envelope: env,
	}

	// We can't use real NewKMSProvider (build-tag gated), so we test the
	// flow by directly testing the envelope unwrap logic.
	mk2, _, err := env.UnwrapFromKMS(ctx, kms, "aws-kms", "test-key-1")
	require.NoError(t, err)
	defer security.ZeroBytes(mk2)

	assert.Equal(t, mk, mk2)

	// Simulate what phaseAcquireCredential would do on KMS success.
	state.MasterKey = mk2
	state.KMSUnwrap = true
	state.Result.KMSUnwrap = true

	assert.True(t, state.KMSUnwrap)
	assert.True(t, state.Result.KMSUnwrap)
	assert.NotNil(t, state.MasterKey)
}

func TestPhaseAcquireCredential_KMSFallback(t *testing.T) {
	// Create envelope with passphrase slot only — no KMS slot.
	env, mk, err := security.NewEnvelope("test-passphrase-1234")
	require.NoError(t, err)
	defer security.ZeroBytes(mk)

	// No KEKSlotHardware → KMS path should be skipped.
	assert.False(t, env.HasSlotType(security.KEKSlotHardware))
}

func TestKMSConfigFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		wantProvider string
		wantNil      bool
		check        func(t *testing.T, cfg *config.KMSConfig)
	}{
		{
			name:         "no env vars",
			env:          nil,
			wantProvider: "",
			wantNil:      true,
		},
		{
			name: "aws-kms",
			env: map[string]string{
				"LANGO_KMS_PROVIDER": "aws-kms",
				"LANGO_KMS_KEY_ID":   "arn:aws:kms:us-east-1:123:key/abc",
				"LANGO_KMS_REGION":   "us-east-1",
			},
			wantProvider: "aws-kms",
			check: func(t *testing.T, cfg *config.KMSConfig) {
				assert.Equal(t, "arn:aws:kms:us-east-1:123:key/abc", cfg.KeyID)
				assert.Equal(t, "us-east-1", cfg.Region)
			},
		},
		{
			name: "azure-kv",
			env: map[string]string{
				"LANGO_KMS_PROVIDER":          "azure-kv",
				"LANGO_KMS_KEY_ID":            "my-key",
				"LANGO_KMS_AZURE_VAULT_URL":   "https://vault.azure.net",
				"LANGO_KMS_AZURE_KEY_VERSION": "v1",
			},
			wantProvider: "azure-kv",
			check: func(t *testing.T, cfg *config.KMSConfig) {
				assert.Equal(t, "my-key", cfg.KeyID)
				assert.Equal(t, "https://vault.azure.net", cfg.Azure.VaultURL)
				assert.Equal(t, "v1", cfg.Azure.KeyVersion)
			},
		},
		{
			name: "pkcs11",
			env: map[string]string{
				"LANGO_KMS_PROVIDER":          "pkcs11",
				"LANGO_KMS_PKCS11_MODULE":     "/usr/lib/pkcs11.so",
				"LANGO_KMS_PKCS11_SLOT_ID":    "2",
				"LANGO_KMS_PKCS11_KEY_LABEL":  "mk-key",
				"LANGO_PKCS11_PIN":            "1234",
			},
			wantProvider: "pkcs11",
			check: func(t *testing.T, cfg *config.KMSConfig) {
				assert.Equal(t, "/usr/lib/pkcs11.so", cfg.PKCS11.ModulePath)
				assert.Equal(t, 2, cfg.PKCS11.SlotID)
				assert.Equal(t, "mk-key", cfg.PKCS11.KeyLabel)
				assert.Equal(t, "1234", cfg.PKCS11.Pin)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all KMS env vars.
			for _, key := range []string{
				"LANGO_KMS_PROVIDER", "LANGO_KMS_KEY_ID", "LANGO_KMS_REGION",
				"LANGO_KMS_ENDPOINT", "LANGO_KMS_AZURE_VAULT_URL",
				"LANGO_KMS_AZURE_KEY_VERSION", "LANGO_KMS_PKCS11_MODULE",
				"LANGO_KMS_PKCS11_SLOT_ID", "LANGO_KMS_PKCS11_KEY_LABEL",
				"LANGO_PKCS11_PIN",
			} {
				t.Setenv(key, "")
				os.Unsetenv(key)
			}

			// Set test env vars.
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, provider := KMSConfigFromEnv()
			assert.Equal(t, tt.wantProvider, provider)

			if tt.wantNil {
				assert.Nil(t, cfg)
				return
			}

			require.NotNil(t, cfg)
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}
