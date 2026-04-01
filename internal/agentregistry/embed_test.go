package agentregistry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedStore_Load(t *testing.T) {
	store := NewEmbeddedStore()
	defs, err := store.Load()
	require.NoError(t, err)
	require.Len(t, defs, 8)

	// Verify all expected agents are present.
	wantNames := map[string]bool{
		"operator":   false,
		"navigator":  false,
		"vault":      false,
		"librarian":  false,
		"automator":  false,
		"planner":    false,
		"chronicler": false,
		"ontologist": false,
	}

	for _, def := range defs {
		_, ok := wantNames[def.Name]
		require.True(t, ok, "unexpected agent: %s", def.Name)
		wantNames[def.Name] = true

		assert.Equal(t, SourceEmbedded, def.Source)
		assert.Equal(t, StatusActive, def.Status)
		assert.NotEmpty(t, def.Description)
		assert.NotEmpty(t, def.Instruction)
	}

	for name, found := range wantNames {
		assert.True(t, found, "missing agent: %s", name)
	}
}

func TestEmbeddedStore_LoadAndRegister(t *testing.T) {
	r := New()
	err := r.LoadFromStore(NewEmbeddedStore())
	require.NoError(t, err)

	// All 8 agents are active.
	active := r.Active()
	assert.Len(t, active, 8)

	// Specs conversion works for all.
	specs := r.Specs()
	assert.Len(t, specs, 8)

	// Planner should have AlwaysInclude set.
	planner, ok := r.Get("planner")
	require.True(t, ok)
	assert.True(t, planner.AlwaysInclude)
	assert.Empty(t, planner.Prefixes)
}
