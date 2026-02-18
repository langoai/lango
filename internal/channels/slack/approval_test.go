package slack

import (
	"context"
	"sync"
	"testing"
	"time"

	slackapi "github.com/slack-go/slack"
	"github.com/langowarny/lango/internal/approval"
)

// MockApprovalClient extends MockClient with UpdateMessage tracking.
type MockApprovalClient struct {
	MockClient
	UpdateMessageFunc func(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error)
	UpdateMessages    []struct {
		ChannelID string
		Timestamp string
		Options   []slackapi.MsgOption
	}
	mu sync.Mutex
}

func (m *MockApprovalClient) UpdateMessage(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error) {
	m.mu.Lock()
	m.UpdateMessages = append(m.UpdateMessages, struct {
		ChannelID string
		Timestamp string
		Options   []slackapi.MsgOption
	}{ChannelID: channelID, Timestamp: timestamp, Options: options})
	m.mu.Unlock()

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
	var resp approval.ApprovalResponse
	var err error

	go func() {
		resp, err = p.RequestApproval(context.Background(), req)
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
		if !resp.Approved {
			t.Error("expected approved=true")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	// Verify UpdateMessage was called to remove buttons
	client.mu.Lock()
	updateCount := len(client.UpdateMessages)
	client.mu.Unlock()
	if updateCount == 0 {
		t.Error("expected UpdateMessage to be called to remove buttons")
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
	var resp approval.ApprovalResponse
	var err error

	go func() {
		resp, err = p.RequestApproval(context.Background(), req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	p.HandleInteractive("deny:test-req-2")

	select {
	case <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Approved {
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

	resp, err := p.RequestApproval(context.Background(), req)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if resp.Approved {
		t.Error("expected approved=false on timeout")
	}

	// Verify expired message was sent via UpdateMessage
	client.mu.Lock()
	updateCount := len(client.UpdateMessages)
	client.mu.Unlock()
	if updateCount == 0 {
		t.Error("expected UpdateMessage to be called on timeout for expired message")
	}
}

func TestSlackApprovalProvider_UnknownAction(t *testing.T) {
	p := NewApprovalProvider(&MockApprovalClient{}, 5*time.Second)

	// Should not panic on unknown action
	p.HandleInteractive("unknown:action")
}

func TestSlackApprovalProvider_DuplicateAction(t *testing.T) {
	client := &MockApprovalClient{
		MockClient: MockClient{
			PostMessageFunc: func(channelID string, options ...slackapi.MsgOption) (string, string, error) {
				return "ts-dup", channelID, nil
			},
		},
	}
	p := NewApprovalProvider(client, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-dup",
		ToolName:   "exec",
		SessionKey: "slack:C123:U456",
		CreatedAt:  time.Now(),
	}

	done := make(chan struct{})
	var resp approval.ApprovalResponse
	var err error

	go func() {
		resp, err = p.RequestApproval(context.Background(), req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// First action — should succeed
	p.HandleInteractive("approve:test-req-dup")

	// Second action — should be silently ignored (LoadAndDelete already removed it)
	p.HandleInteractive("deny:test-req-dup")

	select {
	case <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Approved {
			t.Error("expected approved=true from first action")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	// Only one UpdateMessage call (from the first action)
	client.mu.Lock()
	updateCount := len(client.UpdateMessages)
	client.mu.Unlock()
	if updateCount != 1 {
		t.Errorf("expected 1 UpdateMessage call, got %d", updateCount)
	}
}
