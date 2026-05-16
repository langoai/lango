package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func TestBuildSmartAccountTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := buildSmartAccountTools(nil)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "session_key_create requires targets",
			tool:    "session_key_create",
			params:  map[string]interface{}{"duration": "1h"},
			wantErr: "missing targets parameter",
		},
		{
			name:    "session_key_create requires duration",
			tool:    "session_key_create",
			params:  map[string]interface{}{"targets": []interface{}{"0x123"}},
			wantErr: "missing duration parameter",
		},
		{
			name:    "session_key_revoke requires session_id",
			tool:    "session_key_revoke",
			params:  map[string]interface{}{},
			wantErr: "missing session_id parameter",
		},
		{
			name:    "session_execute requires session_id",
			tool:    "session_execute",
			params:  map[string]interface{}{"target": "0x123"},
			wantErr: "missing session_id parameter",
		},
		{
			name:    "session_execute requires target",
			tool:    "session_execute",
			params:  map[string]interface{}{"session_id": "sess-1"},
			wantErr: "missing target parameter",
		},
		{
			name:    "policy_check requires target",
			tool:    "policy_check",
			params:  map[string]interface{}{},
			wantErr: "missing target parameter",
		},
		{
			name:    "module_install requires module_type",
			tool:    "module_install",
			params:  map[string]interface{}{"address": "0x123"},
			wantErr: "missing module_type parameter",
		},
		{
			name:    "module_install requires address",
			tool:    "module_install",
			params:  map[string]interface{}{"module_type": float64(1)},
			wantErr: "missing address parameter",
		},
		{
			name:    "module_uninstall requires module_type",
			tool:    "module_uninstall",
			params:  map[string]interface{}{"address": "0x123"},
			wantErr: "missing module_type parameter",
		},
		{
			name:    "module_uninstall requires address",
			tool:    "module_uninstall",
			params:  map[string]interface{}{"module_type": float64(1)},
			wantErr: "missing address parameter",
		},
		{
			name:    "paymaster_approve requires token_address",
			tool:    "paymaster_approve",
			params:  map[string]interface{}{"paymaster_address": "0x123", "amount": "1.00"},
			wantErr: "missing token_address parameter",
		},
		{
			name:    "paymaster_approve requires paymaster_address",
			tool:    "paymaster_approve",
			params:  map[string]interface{}{"token_address": "0x123", "amount": "1.00"},
			wantErr: "missing paymaster_address parameter",
		},
		{
			name:    "paymaster_approve requires amount",
			tool:    "paymaster_approve",
			params:  map[string]interface{}{"token_address": "0x123", "paymaster_address": "0xabc"},
			wantErr: "missing amount parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findSmartAccountTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func findSmartAccountTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
