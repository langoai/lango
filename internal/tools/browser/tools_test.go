package browser

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowserAction_RequiresConditionalInputs(t *testing.T) {
	t.Parallel()

	actionTool := testBrowserActionTool(t, newStubBrowserSessionManager())

	testCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "click requires selector",
			params:  map[string]interface{}{"action": "click"},
			wantErr: "selector required for click action",
		},
		{
			name:    "type requires selector",
			params:  map[string]interface{}{"action": "type", "text": "hello"},
			wantErr: "selector required for type action",
		},
		{
			name:    "type requires text",
			params:  map[string]interface{}{"action": "type", "selector": "#search"},
			wantErr: "text required for type action",
		},
		{
			name:    "eval requires javascript text",
			params:  map[string]interface{}{"action": "eval"},
			wantErr: "text (JavaScript) required for eval action",
		},
		{
			name:    "get text requires selector",
			params:  map[string]interface{}{"action": "get_text"},
			wantErr: "selector required for get_text action",
		},
		{
			name:    "get element info requires selector",
			params:  map[string]interface{}{"action": "get_element_info"},
			wantErr: "selector required for get_element_info action",
		},
		{
			name:    "wait requires selector",
			params:  map[string]interface{}{"action": "wait"},
			wantErr: "selector required for wait action",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := actionTool.Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, result)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestBrowserTools_RequireTopLevelInputs(t *testing.T) {
	t.Parallel()

	tools := BuildTools(newStubBrowserSessionManager())

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "navigate requires url",
			tool:    "browser_navigate",
			params:  map[string]interface{}{},
			wantErr: "missing url parameter",
		},
		{
			name:    "search requires query",
			tool:    "browser_search",
			params:  map[string]interface{}{},
			wantErr: "missing query parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := testBrowserToolByName(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, result)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestBrowserAction_BlocksEvalForP2PBeforeSessionCreation(t *testing.T) {
	t.Parallel()

	sm := &SessionManager{
		tool: &Tool{
			sessions: make(map[string]*Session),
		},
	}
	actionTool := testBrowserActionTool(t, sm)

	result, err := actionTool.Handler(
		ctxkeys.WithP2PRequest(context.Background()),
		map[string]interface{}{
			"action": "eval",
			"text":   "() => document.title",
		},
	)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrEvalBlockedP2P)
	assert.Empty(t, sm.sessionID)
	assert.Empty(t, sm.tool.sessions)
}

func testBrowserActionTool(t *testing.T, sm *SessionManager) *agent.Tool {
	t.Helper()

	return testBrowserToolByName(t, BuildTools(sm), "browser_action")
}

func testBrowserToolByName(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("%s tool not found", name)
	return nil
}

func newStubBrowserSessionManager() *SessionManager {
	return &SessionManager{
		tool: &Tool{
			sessions: map[string]*Session{
				"session-test": {ID: "session-test"},
			},
		},
		sessionID: "session-test",
	}
}
