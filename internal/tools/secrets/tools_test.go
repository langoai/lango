package secrets

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil, nil, nil)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "store requires name",
			tool:    "secrets_store",
			params:  map[string]interface{}{"value": "secret"},
			wantErr: "name is required",
		},
		{
			name:    "store requires value",
			tool:    "secrets_store",
			params:  map[string]interface{}{"name": "api-key"},
			wantErr: "value is required",
		},
		{
			name:    "get requires name",
			tool:    "secrets_get",
			params:  map[string]interface{}{},
			wantErr: "name is required",
		},
		{
			name:    "delete requires name",
			tool:    "secrets_delete",
			params:  map[string]interface{}{},
			wantErr: "name is required",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := findSecretsTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, result)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func findSecretsTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
