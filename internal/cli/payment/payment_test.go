package payment

import (
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummyBootLoader returns a boot loader that always errors.
func dummyBootLoader() func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	}
}

func TestNewPaymentCmd_Structure(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	require.NotNil(t, cmd)

	assert.Equal(t, "payment", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestNewPaymentCmd_Subcommands(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())

	expected := []string{"balance", "history", "limits", "info", "send"}

	subCmds := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subCmds[sub.Use] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestNewPaymentCmd_SubcommandCount(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	assert.Equal(t, 5, len(cmd.Commands()), "expected 5 payment subcommands")
}

func TestBalanceCmd_HasJSONFlag(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "balance" {
			jsonFlag := sub.Flags().Lookup("json")
			assert.NotNil(t, jsonFlag, "balance command should have --json flag")
			return
		}
	}
	t.Fatal("balance subcommand not found")
}

func TestSendCmd_HasRequiredFlags(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "send" {
			assert.NotNil(t, sub.Flags().Lookup("to"), "send should have --to flag")
			assert.NotNil(t, sub.Flags().Lookup("amount"), "send should have --amount flag")
			assert.NotNil(t, sub.Flags().Lookup("purpose"), "send should have --purpose flag")
			assert.NotNil(t, sub.Flags().Lookup("force"), "send should have --force flag")
			return
		}
	}
	t.Fatal("send subcommand not found")
}

func TestHistoryCmd_HasLimitFlag(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "history" {
			limitFlag := sub.Flags().Lookup("limit")
			assert.NotNil(t, limitFlag, "history command should have --limit flag")
			return
		}
	}
	t.Fatal("history subcommand not found")
}

func TestSubcommands_HaveShortDescription(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		assert.NotEmpty(t, sub.Short, "subcommand %q should have a Short description", sub.Use)
	}
}
