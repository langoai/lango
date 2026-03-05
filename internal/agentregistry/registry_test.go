package agentregistry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Register_OverridePriority(t *testing.T) {
	r := New()

	builtin := &AgentDefinition{
		Name:        "agent-a",
		Description: "builtin version",
		Status:      StatusActive,
		Source:      SourceBuiltin,
	}
	user := &AgentDefinition{
		Name:        "agent-a",
		Description: "user version",
		Status:      StatusActive,
		Source:      SourceUser,
	}

	r.Register(builtin)
	r.Register(user)

	got, ok := r.Get("agent-a")
	require.True(t, ok)
	assert.Equal(t, "user version", got.Description)
	assert.Equal(t, SourceUser, got.Source)

	// Insertion order should not duplicate.
	all := r.All()
	assert.Len(t, all, 1)
}

func TestRegistry_Active(t *testing.T) {
	r := New()

	r.Register(&AgentDefinition{Name: "charlie", Status: StatusActive})
	r.Register(&AgentDefinition{Name: "alice", Status: StatusActive})
	r.Register(&AgentDefinition{Name: "bob", Status: StatusDisabled})
	r.Register(&AgentDefinition{Name: "dave", Status: StatusDraft})

	active := r.Active()
	require.Len(t, active, 2)
	assert.Equal(t, "alice", active[0].Name)
	assert.Equal(t, "charlie", active[1].Name)
}

func TestRegistry_Get(t *testing.T) {
	r := New()
	r.Register(&AgentDefinition{Name: "test-agent", Description: "test", Status: StatusActive})

	got, ok := r.Get("test-agent")
	require.True(t, ok)
	assert.Equal(t, "test", got.Description)

	_, ok = r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_Specs(t *testing.T) {
	r := New()
	r.Register(&AgentDefinition{
		Name:          "operator",
		Description:   "System ops",
		Instruction:   "Handle system operations.",
		Status:        StatusActive,
		Prefixes:      []string{"exec", "fs_"},
		Keywords:      []string{"run", "execute"},
		Accepts:       "A command",
		Returns:       "Command output",
		CannotDo:      []string{"web browsing"},
		AlwaysInclude: false,
	})
	r.Register(&AgentDefinition{
		Name:   "disabled-agent",
		Status: StatusDisabled,
	})

	specs := r.Specs()
	require.Len(t, specs, 1)

	spec := specs[0]
	assert.Equal(t, "operator", spec.Name)
	assert.Equal(t, "System ops", spec.Description)
	assert.Equal(t, "Handle system operations.", spec.Instruction)
	assert.Equal(t, []string{"exec", "fs_"}, spec.Prefixes)
	assert.Equal(t, []string{"run", "execute"}, spec.Keywords)
	assert.Equal(t, "A command", spec.Accepts)
	assert.Equal(t, "Command output", spec.Returns)
	assert.Equal(t, []string{"web browsing"}, spec.CannotDo)
	assert.False(t, spec.AlwaysInclude)
}

// mockStore implements Store for testing.
type mockStore struct {
	defs []*AgentDefinition
	err  error
}

func (m *mockStore) Load() ([]*AgentDefinition, error) {
	return m.defs, m.err
}

func TestRegistry_LoadFromStore(t *testing.T) {
	tests := []struct {
		give      string
		giveStore Store
		wantLen   int
		wantErr   bool
	}{
		{
			give: "loads all definitions",
			giveStore: &mockStore{
				defs: []*AgentDefinition{
					{Name: "a", Status: StatusActive},
					{Name: "b", Status: StatusActive},
				},
			},
			wantLen: 2,
		},
		{
			give:      "store error propagates",
			giveStore: &mockStore{err: assert.AnError},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			r := New()
			err := r.LoadFromStore(tt.giveStore)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, r.All(), tt.wantLen)
		})
	}
}
