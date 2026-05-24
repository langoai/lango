package smartaccount

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeSmartAccountCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestSmartAccountInfo_WritesTextToCommandWriter(t *testing.T) {
	original := loadInfoAccountResult
	loadInfoAccountResult = func(_ BootLoader) (infoAccountResult, func(), error) {
		return infoAccountResult{
			Address:    "0x1234abcd5678ef901234abcdef567890abcdef12",
			IsDeployed: true,
			Owner:      "0x5678abcd1234ef567890abcdef1234567890abcd",
			ChainID:    84532,
			EntryPoint: "0x0000000071727De22E5E9d8BAf0edAc6f37da032",
			Modules: []infoModuleEntry{
				{Name: "LangoSessionValidator", Type: "validator", Address: "0xaaaa"},
			},
			Paymaster: true,
		}, func() {}, nil
	}
	t.Cleanup(func() { loadInfoAccountResult = original })

	cmd := infoCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Smart Account Info")
	assert.Contains(t, out, "Address:")
	assert.Contains(t, out, "LangoSessionValidator")
}

func TestSmartAccountInfo_WritesJSONToCommandWriter(t *testing.T) {
	original := loadInfoAccountResult
	loadInfoAccountResult = func(_ BootLoader) (infoAccountResult, func(), error) {
		return infoAccountResult{
			Address:    "0x1234abcd5678ef901234abcdef567890abcdef12",
			IsDeployed: false,
			Owner:      "0x5678abcd1234ef567890abcdef1234567890abcd",
			ChainID:    84532,
			EntryPoint: "0x0000000071727De22E5E9d8BAf0edAc6f37da032",
			Modules:    nil,
			Paymaster:  false,
		}, func() {}, nil
	}
	t.Cleanup(func() { loadInfoAccountResult = original })

	cmd := infoCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "0x1234abcd5678ef901234abcdef567890abcdef12", decoded["address"])
	assert.Equal(t, false, decoded["isDeployed"])
	assert.Equal(t, float64(84532), decoded["chainId"])
}
