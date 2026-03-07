package telegram

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/langoai/lango/internal/approval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Parallel()

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
			t.Parallel()

			assert.Equal(t, tt.want, p.CanHandle(tt.give))
		})
	}
}

func TestApprovalProvider_Approve(t *testing.T) {
	t.Parallel()

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
		require.NoError(t, err)
		assert.True(t, resp.Approved)
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
	assert.True(t, hasEdit, "expected edit message to remove keyboard")
}

func TestApprovalProvider_Deny(t *testing.T) {
	t.Parallel()

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
		require.NoError(t, err)
		assert.False(t, resp.Approved)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for denial")
	}
}

func TestApprovalProvider_Timeout(t *testing.T) {
	t.Parallel()

	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 100*time.Millisecond) // short timeout

	req := approval.ApprovalRequest{
		ID:         "test-req-3",
		ToolName:   "exec",
		SessionKey: "telegram:123:456",
		CreatedAt:  time.Now(),
	}

	resp, err := p.RequestApproval(context.Background(), req)
	require.Error(t, err)
	assert.False(t, resp.Approved)

	// Verify expired message was edited
	hasEdit := false
	for _, msg := range bot.SentMessages {
		if edit, ok := msg.(tgbotapi.EditMessageTextConfig); ok {
			if edit.Text == "🔐 Tool approval — ⏱ Expired" {
				hasEdit = true
			}
		}
	}
	assert.True(t, hasEdit, "expected expired message edit on timeout")
}

func TestApprovalProvider_ContextCancellation(t *testing.T) {
	t.Parallel()

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
		require.Error(t, err)
		// Verify expired message was edited
		hasEdit := false
		for _, msg := range bot.SentMessages {
			if edit, ok := msg.(tgbotapi.EditMessageTextConfig); ok {
				if edit.Text == "🔐 Tool approval — ⏱ Expired" {
					hasEdit = true
				}
			}
		}
		assert.True(t, hasEdit, "expected expired message edit on context cancellation")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for cancellation")
	}
}

func TestApprovalProvider_AlwaysAllow(t *testing.T) {
	t.Parallel()

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
		require.NoError(t, err)
		assert.True(t, resp.Approved)
		assert.True(t, resp.AlwaysAllow)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for always-allow")
	}
}

func TestApprovalProvider_InvalidSessionKey(t *testing.T) {
	t.Parallel()

	bot := &MockApprovalBotAPI{}
	p := NewApprovalProvider(bot, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-5",
		ToolName:   "exec",
		SessionKey: "telegram",
		CreatedAt:  time.Now(),
	}

	_, err := p.RequestApproval(context.Background(), req)
	require.Error(t, err)
}

func TestApprovalProvider_UnknownCallback(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
		require.NoError(t, err)
		assert.True(t, resp.Approved, "expected approved=true from first callback")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}
