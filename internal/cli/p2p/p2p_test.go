package p2p

import (
	"strings"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummyBootLoader returns a boot loader that always errors.
// Used for testing command structure without actually bootstrapping.
func dummyBootLoader() func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	}
}

func TestNewP2PCmd_Structure(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	require.NotNil(t, cmd)

	assert.Equal(t, "p2p", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	// Verify all expected subcommands exist.
	expected := []string{
		"status", "peers", "connect", "disconnect",
		"firewall", "discover", "identity", "reputation",
		"pricing", "session", "sandbox",
	}

	subCmds := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subCmds[strings.Fields(sub.Use)[0]] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestNewP2PCmd_SubcommandCount(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	assert.Equal(t, 11, len(cmd.Commands()), "expected 11 P2P subcommands")
}

func TestStatusCmd_HasFlags(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "status" {
			jsonFlag := sub.Flags().Lookup("json")
			assert.NotNil(t, jsonFlag, "status command should have --json flag")
			return
		}
	}
	t.Fatal("status subcommand not found")
}

func TestFirewallCmd_HasSubcommands(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "firewall" {
			firewallSubs := make(map[string]bool)
			for _, fsub := range sub.Commands() {
				firewallSubs[fsub.Use] = true
			}
			assert.True(t, firewallSubs["list"], "firewall should have list subcommand")
			assert.True(t, firewallSubs["add"], "firewall should have add subcommand")
			assert.True(t, firewallSubs["remove <peer-did>"], "firewall should have remove subcommand")
			return
		}
	}
	t.Fatal("firewall subcommand not found")
}

func TestSessionCmd_HasSubcommands(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "session" {
			sessionSubs := make(map[string]bool)
			for _, ssub := range sub.Commands() {
				sessionSubs[ssub.Use] = true
			}
			assert.True(t, sessionSubs["list"], "session should have list subcommand")
			assert.True(t, sessionSubs["revoke"], "session should have revoke subcommand")
			assert.True(t, sessionSubs["revoke-all"], "session should have revoke-all subcommand")
			return
		}
	}
	t.Fatal("session subcommand not found")
}

func TestSandboxCmd_HasSubcommands(t *testing.T) {
	cmd := NewP2PCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "sandbox" {
			sandboxSubs := make(map[string]bool)
			for _, ssub := range sub.Commands() {
				sandboxSubs[ssub.Use] = true
			}
			assert.True(t, sandboxSubs["status"], "sandbox should have status subcommand")
			assert.True(t, sandboxSubs["test"], "sandbox should have test subcommand")
			assert.True(t, sandboxSubs["cleanup"], "sandbox should have cleanup subcommand")
			return
		}
	}
	t.Fatal("sandbox subcommand not found")
}
