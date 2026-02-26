package security

import (
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		"migrate-passphrase", "secrets", "status",
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
	assert.Equal(t, 7, len(cmd.Commands()), "expected 7 security subcommands")
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

func TestBoolToStatus(t *testing.T) {
	assert.Equal(t, "enabled", boolToStatus(true))
	assert.Equal(t, "disabled", boolToStatus(false))
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
