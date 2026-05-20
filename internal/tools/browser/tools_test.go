package browser

import (
	"context"
	"net/url"
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
		{
			name:    "unknown action is rejected",
			params:  map[string]interface{}{"action": "scroll"},
			wantErr: "unknown action: scroll",
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
		{
			name:    "action requires action",
			tool:    "browser_action",
			params:  map[string]interface{}{},
			wantErr: "missing action parameter",
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

func TestBuildTools_Metadata(t *testing.T) {
	t.Parallel()

	tools := BuildTools(newStubBrowserSessionManager())
	require.Len(t, tools, 6)

	testCases := []struct {
		name            string
		safety          agent.SafetyLevel
		activity        agent.ActivityKind
		readOnly        bool
		concurrencySafe bool
		required        []string
	}{
		{
			name:     "browser_navigate",
			safety:   agent.SafetyLevelDangerous,
			activity: agent.ActivityExecute,
			required: []string{"url"},
		},
		{
			name:     "browser_search",
			safety:   agent.SafetyLevelDangerous,
			activity: agent.ActivityQuery,
			required: []string{"query"},
		},
		{
			name:     "browser_observe",
			safety:   agent.SafetyLevelSafe,
			activity: agent.ActivityRead,
			readOnly: true,
		},
		{
			name:            "browser_extract",
			safety:          agent.SafetyLevelSafe,
			activity:        agent.ActivityRead,
			readOnly:        true,
			concurrencySafe: true,
		},
		{
			name:     "browser_action",
			safety:   agent.SafetyLevelDangerous,
			activity: agent.ActivityExecute,
			required: []string{"action"},
		},
		{
			name:     "browser_screenshot",
			safety:   agent.SafetyLevelSafe,
			activity: agent.ActivityRead,
			readOnly: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tool := testBrowserToolByName(t, tools, tc.name)
			assert.NotEmpty(t, tool.Description)
			assert.NotNil(t, tool.Handler)
			assert.Equal(t, tc.safety, tool.SafetyLevel)
			assert.Equal(t, "browser", tool.Capability.Category)
			assert.Equal(t, tc.activity, tool.Capability.Activity)
			assert.Equal(t, tc.readOnly, tool.Capability.ReadOnly)
			assert.Equal(t, tc.concurrencySafe, tool.Capability.ConcurrencySafe)
			assert.NotEmpty(t, tool.Capability.Aliases)
			assert.NotEmpty(t, tool.Capability.SearchHints)
			assertSchemaRequired(t, tool.Parameters, tc.required)
		})
	}

	actionTool := testBrowserToolByName(t, tools, "browser_action")
	actionParam := schemaProperty(t, actionTool.Parameters, "action")
	assert.ElementsMatch(
		t,
		[]string{actionClick, actionType, actionEval, actionGetText, actionGetInfo, actionWait},
		actionParam["enum"],
	)

	extractTool := testBrowserToolByName(t, tools, "browser_extract")
	modeParam := schemaProperty(t, extractTool.Parameters, "mode")
	assert.ElementsMatch(t, []string{"summary", "links", "article", "search_results"}, modeParam["enum"])
}

func TestBrowserExtract_RejectsUnknownModeWithoutBrowser(t *testing.T) {
	t.Parallel()

	tool := &Tool{}

	result, err := tool.Extract("session-test", "unknown", 5)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "unknown extract mode: unknown")
}

func TestBrowserSearch_RejectsEmptyQueryWithoutBrowser(t *testing.T) {
	t.Parallel()

	state := NewRequestState()
	tool := &Tool{}

	result, err := tool.Search(WithRequestState(context.Background(), state), "session-test", "", 5)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "query is required")
	assert.Zero(t, state.searchCount)
}

func TestBrowserSearch_StopsAtRequestStateLimitBeforeBrowserNavigation(t *testing.T) {
	t.Parallel()

	state := NewRequestState()
	state.RecordSearch("first", "https://example.com/first")
	state.RecordSearch("second", "https://example.com/second")
	tool := &Tool{}

	result, err := tool.Search(WithRequestState(context.Background(), state), "session-test", "third", 5)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrSearchLimitReached)
	assert.True(t, state.IsLimitReached())
	assert.Equal(t, []string{"first", "second", "third"}, state.queries)
	assert.Equal(t, "https://example.com/second", state.CurrentURL())
}

func TestHighLevelHelpers_ClampLimitAndBuildSearchURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 5, clampLimit(0, 5))
	assert.Equal(t, 5, clampLimit(-1, 5))
	assert.Equal(t, 7, clampLimit(7, 5))
	assert.Equal(t, maxExtractionLimit, clampLimit(maxExtractionLimit+1, 5))

	rawURL := buildSearchURL("golang browser tools")
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "duckduckgo.com", parsed.Host)
	assert.Equal(t, "/html/", parsed.Path)
	assert.Equal(t, "golang browser tools", parsed.Query().Get("q"))
}

func TestHighLevelScripts_IncludeConfiguredLimitsAndExpectedSelectors(t *testing.T) {
	t.Parallel()

	snapshot := snapshotScript(3, 4)
	assert.Contains(t, snapshot, `document.querySelectorAll("a[href]")`)
	assert.Contains(t, snapshot, `document.querySelectorAll("a[href],button,input,textarea,select,[role='button'],[role='link']")`)
	assert.Contains(t, snapshot, `.slice(0, 3);`)
	assert.Contains(t, snapshot, `.slice(0, 4);`)
	assert.Contains(t, snapshot, `pageType: pageType`)

	article := articleScript()
	assert.Contains(t, article, `document.querySelector("article")`)
	assert.Contains(t, article, `pageType: "article"`)
	assert.Contains(t, article, `slice(0, 5000)`)

	searchResults := searchResultsScript(6)
	assert.Contains(t, searchResults, `"article[data-testid='result']"`)
	assert.Contains(t, searchResults, `".result__snippet"`)
	assert.Contains(t, searchResults, `results.slice(0, 6)`)
	assert.Contains(t, searchResults, `pageType: "search_results"`)
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

func assertSchemaRequired(t *testing.T, schema map[string]interface{}, want []string) {
	t.Helper()

	if len(want) == 0 {
		assert.NotContains(t, schema, "required")
		return
	}

	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.ElementsMatch(t, want, required)
}

func schemaProperty(t *testing.T, schema map[string]interface{}, name string) map[string]interface{} {
	t.Helper()

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)
	rawProperty, ok := properties[name]
	require.True(t, ok)
	property, ok := rawProperty.(map[string]interface{})
	require.True(t, ok)
	return property
}
