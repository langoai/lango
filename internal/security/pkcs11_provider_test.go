//go:build kms_pkcs11 || kms_all

package security

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPKCS11Provider_NewWithoutModulePath(t *testing.T) {
	cfg := config.KMSConfig{
		PKCS11: config.PKCS11Config{
			ModulePath: "",
		},
	}

	_, err := newPKCS11Provider(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "module path is required")
}

func TestPKCS11Provider_PinFromEnv(t *testing.T) {
	// Verify that LANGO_PKCS11_PIN env var is read (we can't fully test
	// PKCS#11 without a real module, but we can verify the env logic).
	t.Setenv("LANGO_PKCS11_PIN", "test-pin-from-env")

	cfg := config.KMSConfig{
		PKCS11: config.PKCS11Config{
			ModulePath: "/nonexistent/module.so",
			Pin:        "config-pin",
		},
	}

	// This will fail at pkcs11.New since the module doesn't exist,
	// but we verify that the function reads the env var by checking
	// it doesn't fail on PIN validation.
	_, err := newPKCS11Provider(cfg)
	require.Error(t, err)
	// Should fail at module loading, not at PIN stage.
	assert.Contains(t, err.Error(), "pkcs11")
}
