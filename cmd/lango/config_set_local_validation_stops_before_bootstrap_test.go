package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigSetLocalValidationStopsBeforeBootstrap(t *testing.T) {
	for _, tt := range []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing value without from-env",
			args:    []string{"set", "agent.provider"},
			wantErr: "accepts 2 arg(s), received 1",
		},
		{
			name:    "from-env requires exactly one path arg",
			args:    []string{"set", "--from-env", "LANGO_CONFIG_VALUE"},
			wantErr: "accepts 1 arg with --from-env: <dot.path>",
		},
		{
			name:    "from-env fails before bootstrap when env is missing",
			args:    []string{"set", "agent.provider", "--from-env", "LANGO_CONFIG_MISSING"},
			wantErr: `environment variable "LANGO_CONFIG_MISSING" is not set`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LANGO_CONFIG_MISSING", "")
			require.NoError(t, os.Unsetenv("LANGO_CONFIG_MISSING"))

			cmd := configCmd()
			cmd.SetArgs(tt.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConfigKeysCommandWritesStaticCatalogWithoutBootstrap(t *testing.T) {
	cmd := configCmd()
	var out bytes.Buffer
	cmd.SetArgs([]string{"keys"})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "agent.provider")
	assert.Contains(t, out.String(), "economy.budget.defaultMax")
	assert.Contains(t, out.String(), "providers.<name>.apiKey")
}
