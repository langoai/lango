package toolchain

import (
	"context"
	"errors"
	"testing"

	"github.com/langoai/lango/internal/eventbus"
)

func TestEventBusHook_Post(t *testing.T) {
	tests := []struct {
		give          string
		toolName      string
		agentName     string
		sessionKey    string
		toolErr       error
		wantSuccess   bool
		wantErrMsg    string
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
			bus := eventbus.New()

			var received *ToolExecutedEvent
			eventbus.SubscribeTyped(bus, func(e ToolExecutedEvent) {
				received = &e
			})

			hook := NewEventBusHook(bus)
			err := hook.Post(HookContext{
				ToolName:   tt.toolName,
				AgentName:  tt.agentName,
				SessionKey: tt.sessionKey,
				Ctx:        context.Background(),
			}, "some-result", tt.toolErr)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if received == nil {
				t.Fatal("event was not published")
			}
			if received.ToolName != tt.toolName {
				t.Errorf("ToolName = %q, want %q", received.ToolName, tt.toolName)
			}
			if received.AgentName != tt.agentName {
				t.Errorf("AgentName = %q, want %q", received.AgentName, tt.agentName)
			}
			if received.SessionKey != tt.sessionKey {
				t.Errorf("SessionKey = %q, want %q", received.SessionKey, tt.sessionKey)
			}
			if received.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", received.Success, tt.wantSuccess)
			}
			if received.Error != tt.wantErrMsg {
				t.Errorf("Error = %q, want %q", received.Error, tt.wantErrMsg)
			}
		})
	}
}

func TestEventBusHook_Metadata(t *testing.T) {
	hook := NewEventBusHook(eventbus.New())
	if hook.Name() != "eventbus" {
		t.Errorf("Name() = %q, want %q", hook.Name(), "eventbus")
	}
	if hook.Priority() != 50 {
		t.Errorf("Priority() = %d, want 50", hook.Priority())
	}
}

func TestToolExecutedEvent_EventName(t *testing.T) {
	e := ToolExecutedEvent{}
	if e.EventName() != "tool.executed" {
		t.Errorf("EventName() = %q, want %q", e.EventName(), "tool.executed")
	}
}
