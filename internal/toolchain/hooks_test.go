package toolchain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test helpers ---

type stubPreHook struct {
	name     string
	priority int
	result   PreHookResult
	err      error
	called   bool
}

func (h *stubPreHook) Name() string  { return h.name }
func (h *stubPreHook) Priority() int { return h.priority }
func (h *stubPreHook) Pre(_ HookContext) (PreHookResult, error) {
	h.called = true
	return h.result, h.err
}

type stubPostHook struct {
	name      string
	priority  int
	err       error
	called    bool
	gotResult interface{}
	gotErr    error
}

func (h *stubPostHook) Name() string  { return h.name }
func (h *stubPostHook) Priority() int { return h.priority }
func (h *stubPostHook) Post(_ HookContext, result interface{}, toolErr error) error {
	h.called = true
	h.gotResult = result
	h.gotErr = toolErr
	return h.err
}

// --- HookRegistry tests ---

func TestHookRegistry_RunPre(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		preHooks   []*stubPreHook
		wantAction PreHookAction
		wantReason string
		wantErr    bool
	}{
		{
			give:       "no hooks returns Continue",
			preHooks:   nil,
			wantAction: Continue,
		},
		{
			give: "single hook returns Continue",
			preHooks: []*stubPreHook{
				{name: "noop", priority: 1, result: PreHookResult{Action: Continue}},
			},
			wantAction: Continue,
		},
		{
			give: "single hook returns Block",
			preHooks: []*stubPreHook{
				{name: "blocker", priority: 1, result: PreHookResult{Action: Block, BlockReason: "forbidden"}},
			},
			wantAction: Block,
			wantReason: "forbidden",
		},
		{
			give: "single hook returns Modify",
			preHooks: []*stubPreHook{
				{name: "modifier", priority: 1, result: PreHookResult{
					Action:         Modify,
					ModifiedParams: map[string]interface{}{"key": "new"},
				}},
			},
			wantAction: Modify,
		},
		{
			give: "hook error propagates",
			preHooks: []*stubPreHook{
				{name: "err", priority: 1, err: errors.New("hook failure")},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			reg := NewHookRegistry()
			for _, h := range tt.preHooks {
				reg.RegisterPre(h)
			}

			result, err := reg.RunPre(HookContext{
				ToolName: "test_tool",
				Params:   map[string]interface{}{"key": "val"},
				Ctx:      context.Background(),
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAction, result.Action)
			if tt.wantReason != "" {
				assert.Equal(t, tt.wantReason, result.BlockReason)
			}
		})
	}
}

func TestHookRegistry_RunPre_PriorityOrdering(t *testing.T) {
	t.Parallel()

	var order []string

	makeHook := func(name string, priority int) *orderPreHook {
		return &orderPreHook{name: name, priority: priority, order: &order}
	}

	reg := NewHookRegistry()
	// Register in reverse priority order to verify sorting.
	reg.RegisterPre(makeHook("third", 30))
	reg.RegisterPre(makeHook("first", 10))
	reg.RegisterPre(makeHook("second", 20))

	_, err := reg.RunPre(HookContext{Ctx: context.Background()})
	require.NoError(t, err)

	assert.Equal(t, []string{"first", "second", "third"}, order)
}

func TestHookRegistry_RunPre_BlockStopsEarly(t *testing.T) {
	t.Parallel()

	blocker := &stubPreHook{
		name:     "blocker",
		priority: 1,
		result:   PreHookResult{Action: Block, BlockReason: "stop"},
	}
	after := &stubPreHook{
		name:     "after",
		priority: 2,
		result:   PreHookResult{Action: Continue},
	}

	reg := NewHookRegistry()
	reg.RegisterPre(blocker)
	reg.RegisterPre(after)

	result, err := reg.RunPre(HookContext{Ctx: context.Background()})
	require.NoError(t, err)
	assert.Equal(t, Block, result.Action)
	assert.False(t, after.called, "hook after blocker should not have been called")
}

func TestHookRegistry_RunPre_ModifyPassesParams(t *testing.T) {
	t.Parallel()

	modifiedParams := map[string]interface{}{"key": "modified"}
	modifier := &stubPreHook{
		name:     "modifier",
		priority: 1,
		result:   PreHookResult{Action: Modify, ModifiedParams: modifiedParams},
	}

	// This hook captures the params it receives to verify modification propagation.
	capturer := &capturePreHook{name: "capturer", priority: 2}

	reg := NewHookRegistry()
	reg.RegisterPre(modifier)
	reg.RegisterPre(capturer)

	_, err := reg.RunPre(HookContext{
		Params: map[string]interface{}{"key": "original"},
		Ctx:    context.Background(),
	})
	require.NoError(t, err)

	assert.Equal(t, "modified", capturer.receivedParams["key"])
}

func TestHookRegistry_RunPost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		postHooks []*stubPostHook
		wantErr   bool
	}{
		{
			give:      "no hooks returns nil",
			postHooks: nil,
			wantErr:   false,
		},
		{
			give: "single hook success",
			postHooks: []*stubPostHook{
				{name: "logger", priority: 1},
			},
			wantErr: false,
		},
		{
			give: "hook error propagates",
			postHooks: []*stubPostHook{
				{name: "failing", priority: 1, err: errors.New("post failure")},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			reg := NewHookRegistry()
			for _, h := range tt.postHooks {
				reg.RegisterPost(h)
			}

			err := reg.RunPost(HookContext{Ctx: context.Background()}, "result", nil)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHookRegistry_RunPost_PriorityOrdering(t *testing.T) {
	t.Parallel()

	var order []string

	makeHook := func(name string, priority int) *orderPostHook {
		return &orderPostHook{name: name, priority: priority, order: &order}
	}

	reg := NewHookRegistry()
	reg.RegisterPost(makeHook("third", 30))
	reg.RegisterPost(makeHook("first", 10))
	reg.RegisterPost(makeHook("second", 20))

	err := reg.RunPost(HookContext{Ctx: context.Background()}, "result", nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"first", "second", "third"}, order)
}

func TestHookRegistry_RunPost_ErrorStopsEarly(t *testing.T) {
	t.Parallel()

	failing := &stubPostHook{name: "failing", priority: 1, err: errors.New("fail")}
	after := &stubPostHook{name: "after", priority: 2}

	reg := NewHookRegistry()
	reg.RegisterPost(failing)
	reg.RegisterPost(after)

	err := reg.RunPost(HookContext{Ctx: context.Background()}, "result", nil)
	require.Error(t, err)
	assert.False(t, after.called, "hook after failing should not have been called")
}

func TestHookRegistry_RunPost_ReceivesToolResult(t *testing.T) {
	t.Parallel()

	hook := &stubPostHook{name: "observer", priority: 1}

	reg := NewHookRegistry()
	reg.RegisterPost(hook)

	wantResult := "tool-output"
	wantErr := errors.New("tool error")

	err := reg.RunPost(HookContext{Ctx: context.Background()}, wantResult, wantErr)
	require.NoError(t, err)
	assert.Equal(t, wantResult, hook.gotResult)
	assert.Equal(t, wantErr, hook.gotErr)
}

// --- AgentName context helpers ---

func TestAgentNameContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		setName  string
		wantName string
	}{
		{
			give:     "empty context returns empty",
			wantName: "",
		},
		{
			give:     "set name is retrievable",
			setName:  "researcher",
			wantName: "researcher",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tt.setName != "" {
				ctx = WithAgentName(ctx, tt.setName)
			}
			got := AgentNameFromContext(ctx)
			assert.Equal(t, tt.wantName, got)
		})
	}
}

// --- ordering test helpers ---

type orderPreHook struct {
	name     string
	priority int
	order    *[]string
}

func (h *orderPreHook) Name() string  { return h.name }
func (h *orderPreHook) Priority() int { return h.priority }
func (h *orderPreHook) Pre(_ HookContext) (PreHookResult, error) {
	*h.order = append(*h.order, h.name)
	return PreHookResult{Action: Continue}, nil
}

type capturePreHook struct {
	name           string
	priority       int
	receivedParams map[string]interface{}
}

func (h *capturePreHook) Name() string  { return h.name }
func (h *capturePreHook) Priority() int { return h.priority }
func (h *capturePreHook) Pre(ctx HookContext) (PreHookResult, error) {
	h.receivedParams = ctx.Params
	return PreHookResult{Action: Continue}, nil
}

type orderPostHook struct {
	name     string
	priority int
	order    *[]string
}

func (h *orderPostHook) Name() string  { return h.name }
func (h *orderPostHook) Priority() int { return h.priority }
func (h *orderPostHook) Post(_ HookContext, _ interface{}, _ error) error {
	*h.order = append(*h.order, h.name)
	return nil
}
