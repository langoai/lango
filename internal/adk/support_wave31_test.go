package adk

import (
	"context"
	"errors"
	"iter"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/memory"
	adkmodel "google.golang.org/adk/model"
	adksession "google.golang.org/adk/session"
	adktool "google.golang.org/adk/tool"
	"google.golang.org/adk/tool/toolconfirmation"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

func TestWave31ContextHelpersRoundTripAgentAndChildSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assert.Empty(t, AgentNameFromContext(ctx))
	_, ok := ChildSessionFromContext(ctx)
	assert.False(t, ok)

	ctx = WithAgentName(ctx, "planner")
	assert.Equal(t, "planner", AgentNameFromContext(ctx))

	info := ChildSessionInfo{
		ChildKey:  "child-1",
		ParentKey: "parent-1",
		AgentName: "planner",
	}
	ctx = WithChildSession(ctx, info)

	got, ok := ChildSessionFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, info, got)
	assert.Equal(t, "planner", AgentNameFromContext(ctx))
}

func TestWave31ChildSessionServiceAdapterForwardsAndSummarizes(t *testing.T) {
	t.Parallel()

	store := &wave31ChildStore{
		child: &session.ChildSession{
			Key:       "child-1",
			ParentKey: "parent-1",
			AgentName: "researcher",
			History: []session.Message{
				{Role: types.RoleUser, Content: "inspect"},
				{Role: types.RoleAssistant, Content: "done"},
			},
		},
	}
	summarizer := &wave31Summarizer{summary: "summary: done"}
	adapter := NewChildSessionServiceAdapter(store, summarizer)

	cfg := session.ChildSessionConfig{MaxMessages: 4, InheritHistory: 2, SummarizeOnMerge: true}
	child, err := adapter.Fork("parent-1", "researcher", cfg)
	require.NoError(t, err)
	assert.Equal(t, "parent-1", store.forkParent)
	assert.Equal(t, "researcher", store.forkAgent)
	assert.Equal(t, cfg, store.forkConfig)
	assert.Same(t, store.child, child)

	require.NoError(t, adapter.MergeWithSummary("child-1"))
	assert.Equal(t, "child-1", store.gotChildKey)
	assert.Equal(t, store.child.History, summarizer.gotMessages)
	assert.Equal(t, "child-1", store.mergeChildKey)
	assert.Equal(t, "summary: done", store.mergeSummary)

	require.NoError(t, adapter.Discard("child-1"))
	assert.Equal(t, "child-1", store.discardChildKey)
}

func TestWave31ChildSessionServiceAdapterMergePropagatesSummarizerError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("summarizer failed")
	adapter := NewChildSessionServiceAdapter(
		&wave31ChildStore{child: &session.ChildSession{Key: "child-1"}},
		&wave31Summarizer{err: wantErr},
	)

	err := adapter.MergeWithSummary("child-1")
	require.ErrorIs(t, err, wantErr)
}

func TestWave31PIIRedactingModelAdapterRedactsInputsAndResponses(t *testing.T) {
	t.Parallel()

	inner := &wave31Model{
		responseText: "model leaked sk-live-secret",
	}
	scanner := agent.NewSecretScanner()
	scanner.Register("api-key", []byte("sk-live-secret"))
	redactor := agent.NewPIIRedactor(agent.PIIConfig{RedactEmail: true})
	adapter := NewPIIRedactingModelAdapter(inner, redactor, scanner)

	req := &adkmodel.LLMRequest{
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "email me at user@example.com"}}},
			{Role: "tool", Parts: []*genai.Part{{Text: "tool saw sk-live-secret"}}},
			{Role: "model", Parts: []*genai.Part{{Text: "do not redact user@example.com here"}}},
		},
	}

	var responses []*adkmodel.LLMResponse
	for resp, err := range adapter.GenerateContent(context.Background(), req, true) {
		require.NoError(t, err)
		responses = append(responses, resp)
	}

	require.Len(t, inner.requests, 1)
	assert.Equal(t, "wave31-model", adapter.Name())
	assert.Equal(t, "email me at [REDACTED]", req.Contents[0].Parts[0].Text)
	assert.Equal(t, "tool saw [SECRET:api-key]", req.Contents[1].Parts[0].Text)
	assert.Equal(t, "do not redact user@example.com here", req.Contents[2].Parts[0].Text)
	require.Len(t, responses, 1)
	assert.Equal(t, "model leaked [SECRET:api-key]", responses[0].Content.Parts[0].Text)
	assert.True(t, inner.stream)
}

func TestWave31PluginConstructorsExposeCallbacksAndMapKeys(t *testing.T) {
	t.Parallel()

	eventPlugin, err := NewEventLoggingPlugin()
	require.NoError(t, err)
	assert.Equal(t, "lango-event-logger", eventPlugin.Name())
	require.NotNil(t, eventPlugin.OnEventCallback())

	evt := &adksession.Event{
		Author: "planner",
		Actions: adksession.EventActions{
			TransferToAgent: "researcher",
		},
		LLMResponse: adkmodel.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{
				{Text: "hello"},
				{FunctionCall: &genai.FunctionCall{Name: "search"}},
			}},
		},
	}
	got, err := eventPlugin.OnEventCallback()(nil, evt)
	require.NoError(t, err)
	assert.Same(t, evt, got)

	toolPlugin, err := NewBeforeToolLoggingPlugin()
	require.NoError(t, err)
	assert.Equal(t, "lango-tool-logger", toolPlugin.Name())
	require.NotNil(t, toolPlugin.BeforeToolCallback())
	require.NotNil(t, toolPlugin.AfterToolCallback())
	require.NotNil(t, toolPlugin.OnToolErrorCallback())

	fakeTool := wave31NamedTool{name: "wave31_tool"}
	block, err := toolPlugin.BeforeToolCallback()(nil, fakeTool, map[string]any{"x": 1})
	require.NoError(t, err)
	assert.Nil(t, block)
	replacement, err := toolPlugin.AfterToolCallback()(nil, fakeTool, nil, map[string]any{"b": 2, "a": 1}, nil)
	require.NoError(t, err)
	assert.Nil(t, replacement)
	replacement, err = toolPlugin.OnToolErrorCallback()(nil, fakeTool, nil, errors.New("boom"))
	require.NoError(t, err)
	assert.Nil(t, replacement)

	assert.Nil(t, mapKeys(nil))
	keys := mapKeys(map[string]any{"b": 2, "a": 1})
	sort.Strings(keys)
	assert.Equal(t, []string{"a", "b"}, keys)
}

func TestWave31AdaptToolForAgentWithTimeoutInjectsNameAndReportsTimeout(t *testing.T) {
	t.Parallel()

	started := make(chan string, 1)
	internalTool := &agent.Tool{
		Name:        "slow_tool",
		Description: "Sleeps until its context expires",
		Parameters:  map[string]interface{}{"input": agent.ParameterDef{Type: "string"}},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			started <- AgentNameFromContext(ctx)
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	adapted, err := AdaptToolForAgentWithTimeout(internalTool, "operator", time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, "slow_tool", adapted.Name())
	assert.Equal(t, "Sleeps until its context expires", adapted.Description())

	runnable := adapted.(interface {
		Run(adktool.Context, any) (map[string]any, error)
	})
	got, err := runnable.Run(&wave31ToolContext{Context: context.Background()}, map[string]any{"input": "x"})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, `tool "slow_tool" timed out after 1ms`)
	select {
	case agentName := <-started:
		assert.Equal(t, "operator", agentName)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for tool handler to observe agent name")
	}
}

type wave31ChildStore struct {
	child *session.ChildSession

	forkParent string
	forkAgent  string
	forkConfig session.ChildSessionConfig

	gotChildKey     string
	mergeChildKey   string
	mergeSummary    string
	discardChildKey string
}

func (s *wave31ChildStore) ForkChild(parentKey, agentName string, cfg session.ChildSessionConfig) (*session.ChildSession, error) {
	s.forkParent = parentKey
	s.forkAgent = agentName
	s.forkConfig = cfg
	return s.child, nil
}

func (s *wave31ChildStore) MergeChild(childKey string, summary string) error {
	s.mergeChildKey = childKey
	s.mergeSummary = summary
	return nil
}

func (s *wave31ChildStore) DiscardChild(childKey string) error {
	s.discardChildKey = childKey
	return nil
}

func (s *wave31ChildStore) GetChild(childKey string) (*session.ChildSession, error) {
	s.gotChildKey = childKey
	return s.child, nil
}

func (s *wave31ChildStore) ChildrenOf(string) ([]*session.ChildSession, error) {
	return []*session.ChildSession{s.child}, nil
}

type wave31Summarizer struct {
	summary     string
	err         error
	gotMessages []session.Message
}

func (s *wave31Summarizer) Summarize(messages []session.Message) (string, error) {
	s.gotMessages = messages
	if s.err != nil {
		return "", s.err
	}
	return s.summary, nil
}

type wave31Model struct {
	responseText string
	requests     []*adkmodel.LLMRequest
	stream       bool
}

func (m *wave31Model) Name() string {
	return "wave31-model"
}

func (m *wave31Model) GenerateContent(_ context.Context, req *adkmodel.LLMRequest, stream bool) iter.Seq2[*adkmodel.LLMResponse, error] {
	m.requests = append(m.requests, req)
	m.stream = stream
	return func(yield func(*adkmodel.LLMResponse, error) bool) {
		yield(&adkmodel.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{{Text: m.responseText}}},
		}, nil)
	}
}

type wave31NamedTool struct {
	name string
}

func (t wave31NamedTool) Name() string {
	return t.name
}

func (wave31NamedTool) Description() string {
	return "fake tool"
}

func (wave31NamedTool) IsLongRunning() bool {
	return false
}

type wave31ToolContext struct {
	context.Context
}

func (c *wave31ToolContext) FunctionCallID() string {
	return "call-wave31"
}

func (c *wave31ToolContext) Actions() *adksession.EventActions {
	return &adksession.EventActions{}
}

func (c *wave31ToolContext) SearchMemory(context.Context, string) (*memory.SearchResponse, error) {
	return nil, nil
}

func (c *wave31ToolContext) ToolConfirmation() *toolconfirmation.ToolConfirmation {
	return nil
}

func (c *wave31ToolContext) RequestConfirmation(string, any) error {
	return nil
}

func (c *wave31ToolContext) UserContent() *genai.Content {
	return nil
}

func (c *wave31ToolContext) InvocationID() string {
	return "inv-wave31"
}

func (c *wave31ToolContext) AgentName() string {
	return "operator"
}

func (c *wave31ToolContext) ReadonlyState() adksession.ReadonlyState {
	return nil
}

func (c *wave31ToolContext) UserID() string {
	return "user-wave31"
}

func (c *wave31ToolContext) AppName() string {
	return "lango"
}

func (c *wave31ToolContext) SessionID() string {
	return "session-wave31"
}

func (c *wave31ToolContext) Branch() string {
	return ""
}

func (c *wave31ToolContext) Artifacts() adkagent.Artifacts {
	return nil
}

func (c *wave31ToolContext) State() adksession.State {
	return nil
}

var (
	_ adkmodel.LLM    = (*wave31Model)(nil)
	_ adktool.Tool    = wave31NamedTool{}
	_ adktool.Context = (*wave31ToolContext)(nil)
)
