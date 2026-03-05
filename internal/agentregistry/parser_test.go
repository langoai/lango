package agentregistry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAgentMD(t *testing.T) {
	tests := []struct {
		give    string
		wantDef *AgentDefinition
		wantErr string
	}{
		{
			give: "valid full AGENT.md",
			wantDef: &AgentDefinition{
				Name:             "operator",
				Description:      "System operations agent",
				Status:           StatusActive,
				Prefixes:         []string{"exec", "fs_"},
				Keywords:         []string{"run", "execute"},
				Capabilities:     []string{"shell", "file-io"},
				Accepts:          "A command to execute",
				Returns:          "Command output",
				CannotDo:         []string{"web browsing"},
				AlwaysInclude:    false,
				SessionIsolation: true,
				Instruction:      "You are the operator agent.\n\nHandle system operations.",
			},
		},
		{
			give: "minimal name only",
			wantDef: &AgentDefinition{
				Name:   "minimal",
				Status: StatusActive,
			},
		},
		{
			give: "all fields populated",
			wantDef: &AgentDefinition{
				Name:             "full-agent",
				Description:      "A fully specified agent",
				Status:           StatusDisabled,
				Prefixes:         []string{"a_", "b_", "c_"},
				Keywords:         []string{"alpha", "beta", "gamma"},
				Capabilities:     []string{"cap1", "cap2"},
				Accepts:          "Structured input",
				Returns:          "Structured output",
				CannotDo:         []string{"x", "y", "z"},
				AlwaysInclude:    true,
				SessionIsolation: true,
				Instruction:      "Full instruction body.",
			},
		},
		{
			give:    "missing name",
			wantErr: "agent name is required",
		},
		{
			give:    "missing frontmatter",
			wantErr: "missing frontmatter delimiter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			content := buildTestContent(tt.give)
			def, err := ParseAgentMD(content)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantDef.Name, def.Name)
			assert.Equal(t, tt.wantDef.Description, def.Description)
			assert.Equal(t, tt.wantDef.Status, def.Status)
			assert.Equal(t, tt.wantDef.Prefixes, def.Prefixes)
			assert.Equal(t, tt.wantDef.Keywords, def.Keywords)
			assert.Equal(t, tt.wantDef.Capabilities, def.Capabilities)
			assert.Equal(t, tt.wantDef.Accepts, def.Accepts)
			assert.Equal(t, tt.wantDef.Returns, def.Returns)
			assert.Equal(t, tt.wantDef.CannotDo, def.CannotDo)
			assert.Equal(t, tt.wantDef.AlwaysInclude, def.AlwaysInclude)
			assert.Equal(t, tt.wantDef.SessionIsolation, def.SessionIsolation)
			assert.Equal(t, tt.wantDef.Instruction, def.Instruction)
		})
	}
}

func TestRoundtrip(t *testing.T) {
	original := &AgentDefinition{
		Name:             "roundtrip-agent",
		Description:      "Test roundtrip parsing",
		Status:           StatusActive,
		Prefixes:         []string{"rt_"},
		Keywords:         []string{"test", "roundtrip"},
		Capabilities:     []string{"testing"},
		Accepts:          "Test input",
		Returns:          "Test output",
		CannotDo:         []string{"production work"},
		AlwaysInclude:    true,
		SessionIsolation: false,
		Instruction:      "You are a roundtrip test agent.\n\nHandle test operations.",
	}

	rendered, err := RenderAgentMD(original)
	require.NoError(t, err)

	parsed, err := ParseAgentMD(rendered)
	require.NoError(t, err)

	assert.Equal(t, original.Name, parsed.Name)
	assert.Equal(t, original.Description, parsed.Description)
	assert.Equal(t, original.Status, parsed.Status)
	assert.Equal(t, original.Prefixes, parsed.Prefixes)
	assert.Equal(t, original.Keywords, parsed.Keywords)
	assert.Equal(t, original.Capabilities, parsed.Capabilities)
	assert.Equal(t, original.Accepts, parsed.Accepts)
	assert.Equal(t, original.Returns, parsed.Returns)
	assert.Equal(t, original.CannotDo, parsed.CannotDo)
	assert.Equal(t, original.AlwaysInclude, parsed.AlwaysInclude)
	assert.Equal(t, original.SessionIsolation, parsed.SessionIsolation)
	assert.Equal(t, original.Instruction, parsed.Instruction)
}

// buildTestContent returns AGENT.md content for a named test case.
func buildTestContent(name string) []byte {
	switch name {
	case "valid full AGENT.md":
		return []byte(`---
name: operator
description: System operations agent
status: active
prefixes:
  - exec
  - fs_
keywords:
  - run
  - execute
capabilities:
  - shell
  - file-io
accepts: A command to execute
returns: Command output
cannot_do:
  - web browsing
session_isolation: true
---

You are the operator agent.

Handle system operations.`)

	case "minimal name only":
		return []byte(`---
name: minimal
---
`)

	case "all fields populated":
		return []byte(`---
name: full-agent
description: A fully specified agent
status: disabled
prefixes:
  - a_
  - b_
  - c_
keywords:
  - alpha
  - beta
  - gamma
capabilities:
  - cap1
  - cap2
accepts: Structured input
returns: Structured output
cannot_do:
  - x
  - "y"
  - z
always_include: true
session_isolation: true
---

Full instruction body.`)

	case "missing name":
		return []byte(`---
description: An agent without a name
---

Some instructions.`)

	case "missing frontmatter":
		return []byte(`No frontmatter here, just plain text.`)

	default:
		return nil
	}
}
