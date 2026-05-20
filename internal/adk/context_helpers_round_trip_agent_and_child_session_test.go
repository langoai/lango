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

func TestContextHelpersRoundTripAgentAndChildSession(t *testing.T) {
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

func TestChildSessionServiceAdapterForwardsAndSummarizes(t *testing.T) {
	t.Parallel()

	store := &contextHelpersRoundTripAgentAndChildSessionChildStore{
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
	summarizer := &contextHelpersRoundTripAgentAndChildSessionSummarizer{summary: "summary: done"}
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

func TestChildSessionServiceAdapterMergePropagatesSummarizerError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("summarizer failed")
	adapter := NewChildSessionServiceAdapter(
		&contextHelpersRoundTripAgentAndChildSessionChildStore{child: &session.ChildSession{Key: "child-1"}},
		&contextHelpersRoundTripAgentAndChildSessionSummarizer{err: wantErr},
	)

	err := adapter.MergeWithSummary("child-1")
	require.ErrorIs(t, err, wantErr)
}

func TestPIIRedactingModelAdapterRedactsInputsAndResponses(t *testing.T) {
	t.Parallel()

	inner := &contextHelpersRoundTripAgentAndChildSessionModel{
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
	assert.Equal(t, "contextHelpersRoundTripAgentAndChildSession1-model", adapter.Name())
	assert.Equal(t, "email me at [REDACTED]", req.Contents[0].Parts[0].Text)
	assert.Equal(t, "tool saw [SECRET:api-key]", req.Contents[1].Parts[0].Text)
	assert.Equal(t, "do not redact user@example.com here", req.Contents[2].Parts[0].Text)
	require.Len(t, responses, 1)
	assert.Equal(t, "model leaked [SECRET:api-key]", responses[0].Content.Parts[0].Text)
	assert.True(t, inner.stream)
}

func TestPluginConstructorsExposeCallbacksAndMapKeys(t *testing.T) {
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

	fakeTool := contextHelpersRoundTripAgentAndChildSessionNamedTool{name: "contextHelpersRoundTripAgentAndChildSession_tool"}
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

func TestAdaptToolForAgentWithTimeoutInjectsNameAndReportsTimeout(t *testing.T) {
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
	got, err := runnable.Run(&contextHelpersRoundTripAgentAndChildSessionToolContext{Context: context.Background()}, map[string]any{"input": "x"})
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

type contextHelpersRoundTripAgentAndChildSessionChildStore struct {
	child *session.ChildSession

	forkParent string
	forkAgent  string
	forkConfig session.ChildSessionConfig

	gotChildKey     string
	mergeChildKey   string
	mergeSummary    string
	discardChildKey string
}

func (s *contextHelpersRoundTripAgentAndChildSessionChildStore) ForkChild(parentKey, agentName string, cfg session.ChildSessionConfig) (*session.ChildSession, error) {
	s.forkParent = parentKey
	s.forkAgent = agentName
	s.forkConfig = cfg
	return s.child, nil
}

func (s *contextHelpersRoundTripAgentAndChildSessionChildStore) MergeChild(childKey string, summary string) error {
	s.mergeChildKey = childKey
	s.mergeSummary = summary
	return nil
}

func (s *contextHelpersRoundTripAgentAndChildSessionChildStore) DiscardChild(childKey string) error {
	s.discardChildKey = childKey
	return nil
}

func (s *contextHelpersRoundTripAgentAndChildSessionChildStore) GetChild(childKey string) (*session.ChildSession, error) {
	s.gotChildKey = childKey
	return s.child, nil
}

func (s *contextHelpersRoundTripAgentAndChildSessionChildStore) ChildrenOf(string) ([]*session.ChildSession, error) {
	return []*session.ChildSession{s.child}, nil
}

type contextHelpersRoundTripAgentAndChildSessionSummarizer struct {
	summary     string
	err         error
	gotMessages []session.Message
}

func (s *contextHelpersRoundTripAgentAndChildSessionSummarizer) Summarize(messages []session.Message) (string, error) {
	s.gotMessages = messages
	if s.err != nil {
		return "", s.err
	}
	return s.summary, nil
}

type contextHelpersRoundTripAgentAndChildSessionModel struct {
	responseText string
	requests     []*adkmodel.LLMRequest
	stream       bool
}

func (m *contextHelpersRoundTripAgentAndChildSessionModel) Name() string {
	return "contextHelpersRoundTripAgentAndChildSession1-model"
}

func (m *contextHelpersRoundTripAgentAndChildSessionModel) GenerateContent(_ context.Context, req *adkmodel.LLMRequest, stream bool) iter.Seq2[*adkmodel.LLMResponse, error] {
	m.requests = append(m.requests, req)
	m.stream = stream
	return func(yield func(*adkmodel.LLMResponse, error) bool) {
		yield(&adkmodel.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{{Text: m.responseText}}},
		}, nil)
	}
}

type contextHelpersRoundTripAgentAndChildSessionNamedTool struct {
	name string
}

func (t contextHelpersRoundTripAgentAndChildSessionNamedTool) Name() string {
	return t.name
}

func (contextHelpersRoundTripAgentAndChildSessionNamedTool) Description() string {
	return "fake tool"
}

func (contextHelpersRoundTripAgentAndChildSessionNamedTool) IsLongRunning() bool {
	return false
}

type contextHelpersRoundTripAgentAndChildSessionToolContext struct {
	context.Context
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) FunctionCallID() string {
	return "call-contextHelpersRoundTripAgentAndChildSession1"
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) Actions() *adksession.EventActions {
	return &adksession.EventActions{}
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) SearchMemory(context.Context, string) (*memory.SearchResponse, error) {
	return nil, nil
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) ToolConfirmation() *toolconfirmation.ToolConfirmation {
	return nil
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) RequestConfirmation(string, any) error {
	return nil
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) UserContent() *genai.Content {
	return nil
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) InvocationID() string {
	return "inv-contextHelpersRoundTripAgentAndChildSession1"
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) AgentName() string {
	return "operator"
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) ReadonlyState() adksession.ReadonlyState {
	return nil
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) UserID() string {
	return "user-contextHelpersRoundTripAgentAndChildSession1"
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) AppName() string {
	return "lango"
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) SessionID() string {
	return "session-contextHelpersRoundTripAgentAndChildSession1"
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) Branch() string {
	return ""
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) Artifacts() adkagent.Artifacts {
	return nil
}

func (c *contextHelpersRoundTripAgentAndChildSessionToolContext) State() adksession.State {
	return nil
}

var (
	_ adkmodel.LLM    = (*contextHelpersRoundTripAgentAndChildSessionModel)(nil)
	_ adktool.Tool    = contextHelpersRoundTripAgentAndChildSessionNamedTool{}
	_ adktool.Context = (*contextHelpersRoundTripAgentAndChildSessionToolContext)(nil)
)
