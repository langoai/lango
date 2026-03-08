package module

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sa "github.com/langoai/lango/internal/smartaccount"
)

func TestRegistry_Register(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	desc := &ModuleDescriptor{
		Name:    "TestValidator",
		Address: common.HexToAddress("0x1111"),
		Type:    sa.ModuleTypeValidator,
		Version: "1.0.0",
	}

	err := r.Register(desc)
	require.NoError(t, err)

	got, err := r.Get(desc.Address)
	require.NoError(t, err)
	assert.Equal(t, "TestValidator", got.Name)
	assert.Equal(t, sa.ModuleTypeValidator, got.Type)
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	desc := &ModuleDescriptor{
		Name:    "TestValidator",
		Address: common.HexToAddress("0x1111"),
		Type:    sa.ModuleTypeValidator,
		Version: "1.0.0",
	}

	require.NoError(t, r.Register(desc))

	err := r.Register(desc)
	assert.ErrorIs(t, err, sa.ErrModuleAlreadyInstalled)
}

func TestRegistry_Register_IsolatesCopy(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	desc := &ModuleDescriptor{
		Name:     "TestValidator",
		Address:  common.HexToAddress("0x1111"),
		Type:     sa.ModuleTypeValidator,
		Version:  "1.0.0",
		InitData: []byte{0x01, 0x02},
	}

	require.NoError(t, r.Register(desc))

	// Mutate original should not affect registry.
	desc.Name = "Mutated"
	desc.InitData[0] = 0xFF

	got, err := r.Get(desc.Address)
	require.NoError(t, err)
	assert.Equal(t, "TestValidator", got.Name)
	assert.Equal(t, byte(0x01), got.InitData[0])
}

func TestRegistry_Get_NotFound(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	_, err := r.Get(common.HexToAddress("0x9999"))
	assert.ErrorIs(t, err, sa.ErrModuleNotInstalled)
}

func TestRegistry_Get_ReturnsCopy(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	desc := &ModuleDescriptor{
		Name:     "TestValidator",
		Address:  common.HexToAddress("0x1111"),
		Type:     sa.ModuleTypeValidator,
		Version:  "1.0.0",
		InitData: []byte{0x01},
	}
	require.NoError(t, r.Register(desc))

	got, err := r.Get(desc.Address)
	require.NoError(t, err)

	// Mutating returned copy should not affect registry.
	got.Name = "Mutated"
	got.InitData[0] = 0xFF

	got2, err := r.Get(desc.Address)
	require.NoError(t, err)
	assert.Equal(t, "TestValidator", got2.Name)
	assert.Equal(t, byte(0x01), got2.InitData[0])
}

func TestRegistry_List(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	descs := []*ModuleDescriptor{
		{
			Name:    "Charlie",
			Address: common.HexToAddress("0x3333"),
			Type:    sa.ModuleTypeExecutor,
			Version: "1.0.0",
		},
		{
			Name:    "Alpha",
			Address: common.HexToAddress("0x1111"),
			Type:    sa.ModuleTypeValidator,
			Version: "1.0.0",
		},
		{
			Name:    "Bravo",
			Address: common.HexToAddress("0x2222"),
			Type:    sa.ModuleTypeHook,
			Version: "1.0.0",
		},
	}

	for _, d := range descs {
		require.NoError(t, r.Register(d))
	}

	list := r.List()
	require.Len(t, list, 3)
	// Should be sorted by name.
	assert.Equal(t, "Alpha", list[0].Name)
	assert.Equal(t, "Bravo", list[1].Name)
	assert.Equal(t, "Charlie", list[2].Name)
}

func TestRegistry_List_Empty(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	list := r.List()
	assert.Empty(t, list)
}

func TestRegistry_ListByType(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	descs := []*ModuleDescriptor{
		{
			Name:    "Val1",
			Address: common.HexToAddress("0x1111"),
			Type:    sa.ModuleTypeValidator,
			Version: "1.0.0",
		},
		{
			Name:    "Exec1",
			Address: common.HexToAddress("0x2222"),
			Type:    sa.ModuleTypeExecutor,
			Version: "1.0.0",
		},
		{
			Name:    "Val2",
			Address: common.HexToAddress("0x3333"),
			Type:    sa.ModuleTypeValidator,
			Version: "2.0.0",
		},
	}

	for _, d := range descs {
		require.NoError(t, r.Register(d))
	}

	validators := r.ListByType(sa.ModuleTypeValidator)
	require.Len(t, validators, 2)
	assert.Equal(t, "Val1", validators[0].Name)
	assert.Equal(t, "Val2", validators[1].Name)

	executors := r.ListByType(sa.ModuleTypeExecutor)
	require.Len(t, executors, 1)
	assert.Equal(t, "Exec1", executors[0].Name)

	hooks := r.ListByType(sa.ModuleTypeHook)
	assert.Empty(t, hooks)
}
