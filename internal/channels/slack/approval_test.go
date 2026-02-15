package slack

import (
	"context"
	"testing"
	"time"

	slackapi "github.com/slack-go/slack"
	"github.com/langowarny/lango/internal/approval"
)

// MockApprovalClient extends MockClient with UpdateMessage.
type MockApprovalClient struct {
	MockClient
	UpdateMessageFunc func(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error)
}

func (m *MockApprovalClient) UpdateMessage(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error) {
	if m.UpdateMessageFunc != nil {
		return m.UpdateMessageFunc(channelID, timestamp, options...)
	}
	return channelID, timestamp, "", nil
}

func TestSlackApprovalProvider_CanHandle(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{give: "slack:ch:usr", want: true},
		{give: "slack:C123:U456", want: true},
		{give: "telegram:123:456", want: false},
		{give: "discord:ch:usr", want: false},
		{give: "", want: false},
	}

	p := NewApprovalProvider(&MockApprovalClient{}, 30*time.Second)
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if got := p.CanHandle(tt.give); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.give, got, tt.want)
			}
		})
	}
}

func TestSlackApprovalProvider_Approve(t *testing.T) {
	client := &MockApprovalClient{
		MockClient: MockClient{
			PostMessageFunc: func(channelID string, options ...slackapi.MsgOption) (string, string, error) {
				return "ts-123", channelID, nil
			},
		},
	}
	p := NewApprovalProvider(client, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-1",
		ToolName:   "exec",
		SessionKey: "slack:C123:U456",
		CreatedAt:  time.Now(),
	}

	done := make(chan struct{})
	var approved bool
	var err error

	go func() {
		approved, err = p.RequestApproval(context.Background(), req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// Simulate button click
	p.HandleInteractive("approve:test-req-1")

	select {
	case <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !approved {
			t.Error("expected approved=true")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestSlackApprovalProvider_Deny(t *testing.T) {
	client := &MockApprovalClient{
		MockClient: MockClient{
			PostMessageFunc: func(channelID string, options ...slackapi.MsgOption) (string, string, error) {
				return "ts-456", channelID, nil
			},
		},
	}
	p := NewApprovalProvider(client, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-2",
		ToolName:   "fs_delete",
		SessionKey: "slack:C123:U456",
		CreatedAt:  time.Now(),
	}

	done := make(chan struct{})
	var approved bool
	var err error

	go func() {
		approved, err = p.RequestApproval(context.Background(), req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	p.HandleInteractive("deny:test-req-2")

	select {
	case <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if approved {
			t.Error("expected approved=false")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestSlackApprovalProvider_Timeout(t *testing.T) {
	client := &MockApprovalClient{
		MockClient: MockClient{
			PostMessageFunc: func(channelID string, options ...slackapi.MsgOption) (string, string, error) {
				return "ts-789", channelID, nil
			},
		},
	}
	p := NewApprovalProvider(client, 100*time.Millisecond)

	req := approval.ApprovalRequest{
		ID:         "test-req-3",
		ToolName:   "exec",
		SessionKey: "slack:C123:U456",
		CreatedAt:  time.Now(),
	}

	approved, err := p.RequestApproval(context.Background(), req)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if approved {
		t.Error("expected approved=false on timeout")
	}
}

func TestSlackApprovalProvider_UnknownAction(t *testing.T) {
	p := NewApprovalProvider(&MockApprovalClient{}, 5*time.Second)

	// Should not panic on unknown action
	p.HandleInteractive("unknown:action")
}
