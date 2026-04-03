package secrets

import (
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/security"
)

// BuildTools creates secrets agent tools.
func BuildTools(secretsStore *security.SecretsStore, refs *security.RefStore, scanner *agent.SecretScanner) []*agent.Tool {
	st := New(secretsStore, refs, scanner)
	return []*agent.Tool{
		{
			Name:        "secrets_store",
			Description: "Encrypt and store a secret value",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "secrets",
				Activity:    agent.ActivityWrite,
				SearchHints: []string{"store", "save", "password"},
			},
			Parameters: agent.Schema().
				Str("name", "Unique name for the secret").
				Str("value", "The secret value to store").
				Required("name", "value").
				Build(),
			Handler: st.Store,
		},
		{
			Name:        "secrets_get",
			Description: "Retrieve a stored secret as a reference token. Returns an opaque {{secret:name}} token that is resolved at execution time by exec tools. The actual secret value never enters the agent context.",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "secrets",
				Activity:    agent.ActivityRead,
				ReadOnly:    true,
				SearchHints: []string{"retrieve", "password"},
			},
			Parameters: agent.Schema().
				Str("name", "Name of the secret to retrieve").
				Required("name").
				Build(),
			Handler: st.Get,
		},
		{
			Name:        "secrets_list",
			Description: "List all stored secrets (metadata only, no values)",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "secrets",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: agent.Schema().Build(),
			Handler:     st.List,
		},
		{
			Name:        "secrets_delete",
			Description: "Delete a stored secret",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "secrets",
				Activity:    agent.ActivityManage,
				SearchHints: []string{"remove", "delete"},
			},
			Parameters: agent.Schema().
				Str("name", "Name of the secret to delete").
				Required("name").
				Build(),
			Handler: st.Delete,
		},
	}
}
