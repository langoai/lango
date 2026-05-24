package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func TestBuildContractTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := buildContractTools(nil)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "contract_read requires address",
			tool:    "contract_read",
			params:  map[string]interface{}{"abi": "[]", "method": "balanceOf"},
			wantErr: "missing address parameter",
		},
		{
			name:    "contract_read requires abi",
			tool:    "contract_read",
			params:  map[string]interface{}{"address": "0x123", "method": "balanceOf"},
			wantErr: "missing abi parameter",
		},
		{
			name:    "contract_read requires method",
			tool:    "contract_read",
			params:  map[string]interface{}{"address": "0x123", "abi": "[]"},
			wantErr: "missing method parameter",
		},
		{
			name:    "contract_call requires address",
			tool:    "contract_call",
			params:  map[string]interface{}{"abi": "[]", "method": "transfer"},
			wantErr: "missing address parameter",
		},
		{
			name:    "contract_call requires abi",
			tool:    "contract_call",
			params:  map[string]interface{}{"address": "0x123", "method": "transfer"},
			wantErr: "missing abi parameter",
		},
		{
			name:    "contract_call requires method",
			tool:    "contract_call",
			params:  map[string]interface{}{"address": "0x123", "abi": "[]"},
			wantErr: "missing method parameter",
		},
		{
			name:    "contract_abi_load requires address",
			tool:    "contract_abi_load",
			params:  map[string]interface{}{"abi": "[]"},
			wantErr: "missing address parameter",
		},
		{
			name:    "contract_abi_load requires abi",
			tool:    "contract_abi_load",
			params:  map[string]interface{}{"address": "0x123"},
			wantErr: "missing abi parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findContractTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func findContractTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
