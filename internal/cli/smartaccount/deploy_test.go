package smartaccount

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartAccountDeploy_WritesTextToCommandWriter(t *testing.T) {
	original := loadDeployAccountResult
	loadDeployAccountResult = func(_ BootLoader) (deployAccountResult, func(), error) {
		return deployAccountResult{
			Address:    "0x1234abcd5678ef901234abcdef567890abcdef12",
			IsDeployed: true,
			Owner:      "0x5678abcd1234ef567890abcdef1234567890abcd",
			ChainID:    84532,
			EntryPoint: "0x0000000071727De22E5E9d8BAf0edAc6f37da032",
			Modules:    3,
		}, func() {}, nil
	}
	t.Cleanup(func() { loadDeployAccountResult = original })

	cmd := deployCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Smart Account Deployed")
	assert.Contains(t, out, "Address:")
	assert.Contains(t, out, "84532")
}

func TestSmartAccountDeploy_WritesJSONToCommandWriter(t *testing.T) {
	original := loadDeployAccountResult
	loadDeployAccountResult = func(_ BootLoader) (deployAccountResult, func(), error) {
		return deployAccountResult{
			Address:    "0x1234abcd5678ef901234abcdef567890abcdef12",
			IsDeployed: false,
			Owner:      "0x5678abcd1234ef567890abcdef1234567890abcd",
			ChainID:    84532,
			EntryPoint: "0x0000000071727De22E5E9d8BAf0edAc6f37da032",
			Modules:    1,
		}, func() {}, nil
	}
	t.Cleanup(func() { loadDeployAccountResult = original })

	cmd := deployCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "0x1234abcd5678ef901234abcdef567890abcdef12", decoded["address"])
	assert.Equal(t, false, decoded["isDeployed"])
	assert.Equal(t, float64(84532), decoded["chainId"])
	assert.Equal(t, float64(1), decoded["moduleCount"])
}

func TestSmartAccountDeploy_InvalidOutputRejectsBeforeLoad(t *testing.T) {
	original := loadDeployAccountResult
	called := false
	loadDeployAccountResult = func(_ BootLoader) (deployAccountResult, func(), error) {
		called = true
		return deployAccountResult{}, nil, assert.AnError
	}
	t.Cleanup(func() { loadDeployAccountResult = original })

	cmd := deployCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
	assert.Empty(t, out)
	assert.False(t, called)
}
