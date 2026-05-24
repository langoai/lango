package crypto

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil, nil, nil, nil)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "encrypt requires data",
			tool:    "crypto_encrypt",
			params:  map[string]interface{}{},
			wantErr: "data is required",
		},
		{
			name:    "decrypt requires ciphertext",
			tool:    "crypto_decrypt",
			params:  map[string]interface{}{},
			wantErr: "ciphertext is required",
		},
		{
			name:    "sign requires data",
			tool:    "crypto_sign",
			params:  map[string]interface{}{},
			wantErr: "data is required",
		},
		{
			name:    "hash requires data",
			tool:    "crypto_hash",
			params:  map[string]interface{}{},
			wantErr: "data is required",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := findCryptoTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, result)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func findCryptoTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
