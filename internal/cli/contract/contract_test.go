package contract

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeContractCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func executeContractCmdSplit(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	t.Helper()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestABILoad_WritesTextToCommandWriter(t *testing.T) {
	abiPath := filepath.Join(t.TempDir(), "erc20.json")
	require.NoError(t, os.WriteFile(abiPath, []byte(`[{"type":"function","name":"balanceOf","stateMutability":"view","inputs":[{"name":"owner","type":"address"}],"outputs":[{"name":"","type":"uint256"}]},{"type":"event","name":"Transfer","inputs":[{"name":"from","type":"address","indexed":true},{"name":"to","type":"address","indexed":true},{"name":"value","type":"uint256","indexed":false}],"anonymous":false}]`), 0o644))

	cfg := config.DefaultConfig()
	cfg.Payment.Network.ChainID = 84532
	cmd := newABILoadCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeContractCmd(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--file", abiPath,
	)
	require.NoError(t, err)
	assert.Contains(t, out, "ABI Loaded")
	assert.Contains(t, out, "Chain ID: 84532")
	assert.Contains(t, out, "Methods:  1")
	assert.Contains(t, out, "Events:   1")
}

func TestABILoad_WritesJSONToCommandWriter(t *testing.T) {
	abiPath := filepath.Join(t.TempDir(), "erc20.json")
	require.NoError(t, os.WriteFile(abiPath, []byte(`[{"type":"function","name":"balanceOf","stateMutability":"view","inputs":[{"name":"owner","type":"address"}],"outputs":[{"name":"","type":"uint256"}]}]`), 0o644))

	cfg := config.DefaultConfig()
	cfg.Payment.Network.ChainID = 84532
	cmd := newABILoadCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeContractCmd(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--file", abiPath,
		"--output", "json",
	)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "0x036CbD53842c5426634e7929541eC2318f3dCF7e", decoded["address"])
	assert.Equal(t, float64(84532), decoded["chainId"])
	assert.Equal(t, float64(1), decoded["methods"])
	assert.Equal(t, float64(0), decoded["events"])
	assert.Equal(t, "loaded", decoded["status"])
}

func TestContractRead_WritesTextToCommandWriter(t *testing.T) {
	abiPath := filepath.Join(t.TempDir(), "erc20.json")
	require.NoError(t, os.WriteFile(abiPath, []byte(`[{"type":"function","name":"balanceOf","stateMutability":"view","inputs":[{"name":"owner","type":"address"}],"outputs":[{"name":"","type":"uint256"}]}]`), 0o644))

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.Network.ChainID = 84532
	cmd := newReadCmd(func() (*config.Config, error) { return cfg, nil })

	stdout, stderr, err := executeContractCmdSplit(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--abi", abiPath,
		"--method", "balanceOf",
		"--args", "0x1234abcd5678ef901234abcdef567890abcdef12",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "Note: contract read requires a running RPC connection.")
	assert.Contains(t, stdout, "Contract Read (validated)")
	assert.Contains(t, stdout, "Method:   balanceOf")
	assert.Contains(t, stdout, "Chain ID: 84532")
}

func TestContractRead_WritesJSONToCommandWriter(t *testing.T) {
	abiPath := filepath.Join(t.TempDir(), "erc20.json")
	require.NoError(t, os.WriteFile(abiPath, []byte(`[{"type":"function","name":"balanceOf","stateMutability":"view","inputs":[{"name":"owner","type":"address"}],"outputs":[{"name":"","type":"uint256"}]}]`), 0o644))

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.Network.ChainID = 84532
	cmd := newReadCmd(func() (*config.Config, error) { return cfg, nil })

	stdout, stderr, err := executeContractCmdSplit(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--abi", abiPath,
		"--method", "balanceOf",
		"--args", "0x1234abcd5678ef901234abcdef567890abcdef12",
		"--output", "json",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "Use 'lango serve' and the contract_read agent tool for live queries.")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &decoded))
	assert.Equal(t, "0x036CbD53842c5426634e7929541eC2318f3dCF7e", decoded["address"])
	assert.Equal(t, "balanceOf", decoded["method"])
	assert.Equal(t, float64(84532), decoded["chainId"])
}

func TestContractCall_WritesTextToCommandWriter(t *testing.T) {
	abiPath := filepath.Join(t.TempDir(), "erc20.json")
	require.NoError(t, os.WriteFile(abiPath, []byte(`[{"type":"function","name":"transfer","stateMutability":"nonpayable","inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[]}]`), 0o644))

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.Network.ChainID = 84532
	cmd := newCallCmd(func() (*config.Config, error) { return cfg, nil })

	stdout, stderr, err := executeContractCmdSplit(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--abi", abiPath,
		"--method", "transfer",
		"--args", "0x5678abcd1234ef567890abcdef1234567890abcd,1000000",
		"--value", "0.01",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "Note: contract call requires a running RPC connection and wallet.")
	assert.Contains(t, stdout, "Contract Call (validated)")
	assert.Contains(t, stdout, "Method:   transfer")
	assert.Contains(t, stdout, "Value:    0.01 ETH")
}

func TestContractCall_WritesJSONToCommandWriter(t *testing.T) {
	abiPath := filepath.Join(t.TempDir(), "erc20.json")
	require.NoError(t, os.WriteFile(abiPath, []byte(`[{"type":"function","name":"transfer","stateMutability":"nonpayable","inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[]}]`), 0o644))

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.Network.ChainID = 84532
	cmd := newCallCmd(func() (*config.Config, error) { return cfg, nil })

	stdout, stderr, err := executeContractCmdSplit(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--abi", abiPath,
		"--method", "transfer",
		"--args", "0x5678abcd1234ef567890abcdef1234567890abcd,1000000",
		"--value", "0.01",
		"--output", "json",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "Use 'lango serve' and the contract_call agent tool for live transactions.")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &decoded))
	assert.Equal(t, "0x036CbD53842c5426634e7929541eC2318f3dCF7e", decoded["address"])
	assert.Equal(t, "transfer", decoded["method"])
	assert.Equal(t, "0.01", decoded["value"])
	assert.Equal(t, float64(84532), decoded["chainId"])
}

func TestContractRead_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := newReadCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, _, err := executeContractCmdSplit(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--abi", "./erc20.json",
		"--method", "balanceOf",
		"--output", "yaml",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestContractCall_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := newCallCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, _, err := executeContractCmdSplit(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--abi", "./erc20.json",
		"--method", "transfer",
		"--output", "yaml",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestABILoad_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := newABILoadCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeContractCmd(t, cmd,
		"--address", "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"--file", "./erc20.json",
		"--output", "yaml",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}
