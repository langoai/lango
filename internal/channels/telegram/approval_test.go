package telegram

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/langowarny/lango/internal/approval"
)

// MockApprovalBotAPI extends MockBotAPI with Request support.
type MockApprovalBotAPI struct {
	MockBotAPI
	RequestFunc func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

func (m *MockApprovalBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	if m.RequestFunc != nil {
		return m.RequestFunc(c)
	}
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func TestApprovalProvider_CanHandle(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{give: "telegram:123:456", want: true},
		{give: "telegram:0:0", want: true},
		{give: "discord:ch:usr", want: false},
		{give: "slack:ch:usr", want: false},
		{give: "", want: false},
	}

	p := NewApprovalProvider(&MockApprovalBotAPI{}, 30*time.Second)
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if got := p.CanHandle(tt.give); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.give, got, tt.want)
			}
		})
	}
}

func TestApprovalProvider_Approve(t *testing.T) {
	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-1",
		ToolName:   "exec",
		SessionKey: "telegram:123:456",
		CreatedAt:  time.Now(),
	}

	done := make(chan struct{})
	var resp approval.ApprovalResponse
	var err error

	go func() {
		resp, err = p.RequestApproval(context.Background(), req)
		close(done)
	}()

	// Wait for the message to be sent
	time.Sleep(50 * time.Millisecond)

	// Simulate approve callback
	p.HandleCallback(&tgbotapi.CallbackQuery{
		ID:   "cb-1",
		Data: "approve:test-req-1",
		Message: &tgbotapi.Message{
			MessageID: 100,
			Chat:      &tgbotapi.Chat{ID: 123},
			Text:      "Tool 'exec' requires approval",
		},
	})

	select {
	case <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Approved {
			t.Error("expected approved=true")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for approval")
	}

	// Verify keyboard was removed: edit message should have been sent
	hasEdit := false
	for _, msg := range bot.SentMessages {
		if _, ok := msg.(tgbotapi.EditMessageTextConfig); ok {
			hasEdit = true
			break
		}
	}
	if !hasEdit {
		t.Error("expected edit message to remove keyboard")
	}
}

func TestApprovalProvider_Deny(t *testing.T) {
	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-2",
		ToolName:   "fs_delete",
		SessionKey: "telegram:123:456",
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

	p.HandleCallback(&tgbotapi.CallbackQuery{
		ID:   "cb-2",
		Data: "deny:test-req-2",
		Message: &tgbotapi.Message{
			MessageID: 101,
			Chat:      &tgbotapi.Chat{ID: 123},
			Text:      "Tool 'fs_delete' requires approval",
		},
	})

	select {
	case <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Approved {
			t.Error("expected approved=false")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for denial")
	}
}

func TestApprovalProvider_Timeout(t *testing.T) {
	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 100*time.Millisecond) // short timeout

	req := approval.ApprovalRequest{
		ID:         "test-req-3",
		ToolName:   "exec",
		SessionKey: "telegram:123:456",
		CreatedAt:  time.Now(),
	}

	resp, err := p.RequestApproval(context.Background(), req)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if resp.Approved {
		t.Error("expected approved=false on timeout")
	}

	// Verify expired message was edited
	hasEdit := false
	for _, msg := range bot.SentMessages {
		if edit, ok := msg.(tgbotapi.EditMessageTextConfig); ok {
			if edit.Text == "🔐 Tool approval — ⏱ Expired" {
				hasEdit = true
			}
		}
	}
	if !hasEdit {
		t.Error("expected expired message edit on timeout")
	}
}

func TestApprovalProvider_ContextCancellation(t *testing.T) {
	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 30*time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	req := approval.ApprovalRequest{
		ID:         "test-req-4",
		ToolName:   "exec",
		SessionKey: "telegram:123:456",
		CreatedAt:  time.Now(),
	}

	done := make(chan struct{})
	var err error

	go func() {
		_, err = p.RequestApproval(ctx, req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		if err == nil {
			t.Fatal("expected context cancelled error")
		}
		// Verify expired message was edited
		hasEdit := false
		for _, msg := range bot.SentMessages {
			if edit, ok := msg.(tgbotapi.EditMessageTextConfig); ok {
				if edit.Text == "🔐 Tool approval — ⏱ Expired" {
					hasEdit = true
				}
			}
		}
		if !hasEdit {
			t.Error("expected expired message edit on context cancellation")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for cancellation")
	}
}

func TestApprovalProvider_AlwaysAllow(t *testing.T) {
	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-always",
		ToolName:   "exec",
		SessionKey: "telegram:123:456",
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

	// Simulate always-allow callback
	p.HandleCallback(&tgbotapi.CallbackQuery{
		ID:   "cb-always",
		Data: "always:test-req-always",
	})

	select {
	case <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Approved {
			t.Error("expected approved=true")
		}
		if !resp.AlwaysAllow {
			t.Error("expected alwaysAllow=true")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for always-allow")
	}
}

func TestApprovalProvider_InvalidSessionKey(t *testing.T) {
	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-5",
		ToolName:   "exec",
		SessionKey: "telegram",
		CreatedAt:  time.Now(),
	}

	_, err := p.RequestApproval(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid session key")
	}
}

func TestApprovalProvider_UnknownCallback(t *testing.T) {
	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 5*time.Second)

	// Should not panic on unknown callback data
	p.HandleCallback(&tgbotapi.CallbackQuery{
		ID:   "cb-unknown",
		Data: "unknown:action",
	})

	// Should not panic on nil
	p.HandleCallback(nil)
}

func TestApprovalProvider_DuplicateCallback(t *testing.T) {
	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-dup",
		ToolName:   "exec",
		SessionKey: "telegram:123:456",
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

	// First callback — should succeed
	p.HandleCallback(&tgbotapi.CallbackQuery{
		ID:   "cb-dup-1",
		Data: "approve:test-req-dup",
	})

	// Second callback — should be silently ignored (LoadAndDelete already removed it)
	p.HandleCallback(&tgbotapi.CallbackQuery{
		ID:   "cb-dup-2",
		Data: "deny:test-req-dup",
	})

	select {
	case <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Approved {
			t.Error("expected approved=true from first callback")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}
