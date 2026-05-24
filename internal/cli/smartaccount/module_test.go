package smartaccount

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sa "github.com/langoai/lango/internal/smartaccount"
)

func executeSmartAccountModuleCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestSmartAccountModuleList_WritesTextToCommandWriter(t *testing.T) {
	original := loadModuleListEntries
	loadModuleListEntries = func(_ BootLoader) ([]listedModuleEntry, func(), error) {
		return []listedModuleEntry{
			{Name: "LangoSessionValidator", Type: "validator", Address: "0xaaaa", Version: "1.0.0"},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadModuleListEntries = original })

	cmd := moduleListCmd(nil)
	out, err := executeSmartAccountModuleCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "LangoSessionValidator")
}

func TestSmartAccountModuleList_WritesJSONToCommandWriter(t *testing.T) {
	original := loadModuleListEntries
	loadModuleListEntries = func(_ BootLoader) ([]listedModuleEntry, func(), error) {
		return []listedModuleEntry{
			{Name: "LangoSessionValidator", Type: "validator", Address: "0xaaaa", Version: "1.0.0"},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadModuleListEntries = original })

	cmd := moduleListCmd(nil)
	out, err := executeSmartAccountModuleCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "LangoSessionValidator", decoded[0]["name"])
}

func TestSmartAccountModuleList_WritesEmptyStateToCommandWriter(t *testing.T) {
	original := loadModuleListEntries
	loadModuleListEntries = func(_ BootLoader) ([]listedModuleEntry, func(), error) {
		return nil, func() {}, nil
	}
	t.Cleanup(func() { loadModuleListEntries = original })

	cmd := moduleListCmd(nil)
	out, err := executeSmartAccountModuleCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No modules registered.")
}

func TestSmartAccountModuleInstall_WritesSuccessToCommandWriter(t *testing.T) {
	original := installSmartAccountModule
	installSmartAccountModule = func(_ BootLoader, _ sa.ModuleType, _ common.Address) (string, func(), error) {
		return "0xtxhash", func() {}, nil
	}
	t.Cleanup(func() { installSmartAccountModule = original })

	cmd := moduleInstallCmd(nil)
	out, err := executeSmartAccountModuleCmd(t, cmd, "0xdddd1234567890abcdef1234567890abcdef1234", "--type", "executor")
	require.NoError(t, err)
	assert.Contains(t, out, "Module installed successfully.")
	assert.Contains(t, out, common.HexToAddress("0xdddd1234567890abcdef1234567890abcdef1234").Hex())
	assert.Contains(t, out, "executor")
	assert.Contains(t, out, "0xtxhash")
}
