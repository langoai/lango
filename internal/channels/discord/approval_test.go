package discord

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/langowarny/lango/internal/approval"
)

// MockApprovalSession extends MockSession with InteractionRespond and ChannelMessageEditComplex.
type MockApprovalSession struct {
	MockSession
	InteractionRespondFunc          func(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	ChannelMessageEditComplexFunc   func(edit *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
	SentComplexMessages             []*discordgo.MessageSend
}

func (m *MockApprovalSession) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	if m.InteractionRespondFunc != nil {
		return m.InteractionRespondFunc(interaction, resp, options...)
	}
	return nil
}

func (m *MockApprovalSession) ChannelMessageEditComplex(edit *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if m.ChannelMessageEditComplexFunc != nil {
		return m.ChannelMessageEditComplexFunc(edit, options...)
	}
	return &discordgo.Message{}, nil
}

func (m *MockApprovalSession) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.SentComplexMessages = append(m.SentComplexMessages, data)
	m.SentMessages = append(m.SentMessages, data.Content)
	return &discordgo.Message{ID: "msg-1", Content: data.Content}, nil
}

func TestDiscordApprovalProvider_CanHandle(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{give: "discord:ch:usr", want: true},
		{give: "discord:123:456", want: true},
		{give: "telegram:123:456", want: false},
		{give: "slack:ch:usr", want: false},
		{give: "", want: false},
	}

	state := &discordgo.State{}
	state.User = &discordgo.User{ID: "bot-1"}
	sess := &MockApprovalSession{MockSession: MockSession{State: state}}
	p := NewApprovalProvider(sess, 30*time.Second)

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if got := p.CanHandle(tt.give); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.give, got, tt.want)
			}
		})
	}
}

func TestDiscordApprovalProvider_Approve(t *testing.T) {
	state := &discordgo.State{}
	state.User = &discordgo.User{ID: "bot-1"}
	sess := &MockApprovalSession{MockSession: MockSession{State: state}}
	p := NewApprovalProvider(sess, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-1",
		ToolName:   "exec",
		SessionKey: "discord:chan-1:user-1",
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
	p.HandleInteraction(&discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
			Data: discordgo.MessageComponentInteractionData{
				CustomID: "approve:test-req-1",
			},
		},
	})

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

func TestDiscordApprovalProvider_Deny(t *testing.T) {
	state := &discordgo.State{}
	state.User = &discordgo.User{ID: "bot-1"}
	sess := &MockApprovalSession{MockSession: MockSession{State: state}}
	p := NewApprovalProvider(sess, 5*time.Second)

	req := approval.ApprovalRequest{
		ID:         "test-req-2",
		ToolName:   "fs_delete",
		SessionKey: "discord:chan-1:user-1",
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

	p.HandleInteraction(&discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
			Data: discordgo.MessageComponentInteractionData{
				CustomID: "deny:test-req-2",
			},
		},
	})

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

func TestDiscordApprovalProvider_Timeout(t *testing.T) {
	state := &discordgo.State{}
	state.User = &discordgo.User{ID: "bot-1"}
	sess := &MockApprovalSession{MockSession: MockSession{State: state}}
	p := NewApprovalProvider(sess, 100*time.Millisecond)

	req := approval.ApprovalRequest{
		ID:         "test-req-3",
		ToolName:   "exec",
		SessionKey: "discord:chan-1:user-1",
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
