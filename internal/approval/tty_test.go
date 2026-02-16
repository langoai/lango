package approval

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestTTYProvider_NonTerminal_ReturnsError(t *testing.T) {
	// CI/test environments typically do not have a terminal attached to stdin,
	// so RequestApproval should return an error indicating TTY is unavailable.
	p := &TTYProvider{}

	req := ApprovalRequest{
		ID:         "test-tty-1",
		ToolName:   "exec",
		SessionKey: "tty:local",
		CreatedAt:  time.Now(),
	}

	approved, err := p.RequestApproval(context.Background(), req)
	if approved {
		t.Error("expected TTYProvider to deny in non-terminal environment")
	}
	if err == nil {
		t.Fatal("expected error for non-terminal stdin, got nil")
	}
	if !strings.Contains(err.Error(), "not a terminal") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTTYProvider_CanHandleAlwaysFalse(t *testing.T) {
	p := &TTYProvider{}
	if p.CanHandle("any:session") {
		t.Error("TTYProvider.CanHandle should always return false")
	}
}

func TestTTYProvider_InterfaceCompliance(t *testing.T) {
	var _ Provider = (*TTYProvider)(nil)
}
