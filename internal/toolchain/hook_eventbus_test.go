package toolchain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
)

func TestEventBusHook_Post(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		toolName    string
		agentName   string
		sessionKey  string
		toolErr     error
		wantSuccess bool
		wantErrMsg  string
	}{
		{
			give:        "successful tool execution publishes success event",
			toolName:    "exec",
			agentName:   "executor",
			sessionKey:  "session-1",
			toolErr:     nil,
			wantSuccess: true,
		},
		{
			give:        "failed tool execution publishes failure event",
			toolName:    "exec",
			agentName:   "executor",
			sessionKey:  "session-2",
			toolErr:     errors.New("command failed"),
			wantSuccess: false,
			wantErrMsg:  "command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			bus := eventbus.New()

			var received *ToolExecutedEvent
			eventbus.SubscribeTyped(bus, func(e ToolExecutedEvent) {
				received = &e
			})

			hook := NewEventBusHook(bus)
			ctx := HookContext{
				ToolName:   tt.toolName,
				AgentName:  tt.agentName,
				SessionKey: tt.sessionKey,
				Ctx:        context.Background(),
			}

			// Call Pre to record start time, then Post to publish event.
			_, err := hook.Pre(ctx)
			require.NoError(t, err)

			err = hook.Post(ctx, "some-result", tt.toolErr)
			require.NoError(t, err)
			require.NotNil(t, received, "event was not published")
			assert.Equal(t, tt.toolName, received.ToolName)
			assert.Equal(t, tt.agentName, received.AgentName)
			assert.Equal(t, tt.sessionKey, received.SessionKey)
			assert.Equal(t, tt.wantSuccess, received.Success)
			assert.Equal(t, tt.wantErrMsg, received.Error)
			assert.Greater(t, received.Duration, int64(0))
		})
	}
}

func TestEventBusHook_PreContinues(t *testing.T) {
	t.Parallel()

	hook := NewEventBusHook(eventbus.New())
	result, err := hook.Pre(HookContext{
		ToolName:   "test",
		AgentName:  "agent",
		SessionKey: "sess",
		Ctx:        context.Background(),
	})
	require.NoError(t, err)
	assert.Equal(t, Continue, result.Action)
}

func TestEventBusHook_PostWithoutPre(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()

	var received *ToolExecutedEvent
	eventbus.SubscribeTyped(bus, func(e ToolExecutedEvent) {
		received = &e
	})

	hook := NewEventBusHook(bus)
	// Call Post without Pre — duration should be zero but no panic.
	err := hook.Post(HookContext{
		ToolName:   "test",
		AgentName:  "agent",
		SessionKey: "sess",
		Ctx:        context.Background(),
	}, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, received, "event was not published")
	assert.Zero(t, received.Duration)
}

func TestEventBusHook_Metadata(t *testing.T) {
	t.Parallel()

	hook := NewEventBusHook(eventbus.New())
	assert.Equal(t, "eventbus", hook.Name())
	assert.Equal(t, 50, hook.Priority())
}

func TestToolExecutedEvent_EventName(t *testing.T) {
	t.Parallel()

	e := ToolExecutedEvent{}
	assert.Equal(t, "tool.executed", e.EventName())
}
