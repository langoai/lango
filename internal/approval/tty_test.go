package approval

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestTTYProvider_NonTerminal_ReturnsError(t *testing.T) {
	// CI/test environments typically do not have a terminal attached to stdin,
	// so RequestApproval should return an error indicating TTY is unavailable.
	origIsTerminal := ttyIsTerminal
	t.Cleanup(func() { ttyIsTerminal = origIsTerminal })
	ttyIsTerminal = func() bool { return false }

	p := &TTYProvider{}

	req := ApprovalRequest{
		ID:         "test-tty-1",
		ToolName:   "exec",
		SessionKey: "tty:local",
		CreatedAt:  time.Now(),
	}

	resp, err := p.RequestApproval(context.Background(), req)
	if resp.Approved {
		t.Error("expected TTYProvider to deny in non-terminal environment")
	}
	if err == nil {
		t.Fatal("expected error for non-terminal stdin, got nil")
	}
	if !strings.Contains(err.Error(), "not a terminal") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTTYProvider_ApproveOnceUsesInjectedStreams(t *testing.T) {
	origIsTerminal := ttyIsTerminal
	origInput := ttyInput
	origError := ttyError
	t.Cleanup(func() {
		ttyIsTerminal = origIsTerminal
		ttyInput = origInput
		ttyError = origError
	})

	ttyIsTerminal = func() bool { return true }
	ttyInput = bytes.NewBufferString("y\n")
	var errBuf bytes.Buffer
	ttyError = &errBuf

	p := &TTYProvider{}
	req := ApprovalRequest{
		ID:         "test-tty-approve",
		ToolName:   "exec",
		SessionKey: "tty:local",
		Summary:    "Delete: /tmp/test",
		CreatedAt:  time.Now(),
	}

	resp, err := p.RequestApproval(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Approved || resp.AlwaysAllow {
		t.Fatalf("expected single approval, got %+v", resp)
	}
	if resp.Provider != "tty" {
		t.Fatalf("expected provider tty, got %q", resp.Provider)
	}
	if !strings.Contains(errBuf.String(), "Sensitive tool 'exec' requires approval.") {
		t.Fatalf("expected approval banner, got %q", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Delete: /tmp/test") {
		t.Fatalf("expected summary in prompt, got %q", errBuf.String())
	}
}

func TestTTYProvider_AlwaysAllowUsesInjectedStreams(t *testing.T) {
	origIsTerminal := ttyIsTerminal
	origInput := ttyInput
	origError := ttyError
	t.Cleanup(func() {
		ttyIsTerminal = origIsTerminal
		ttyInput = origInput
		ttyError = origError
	})

	ttyIsTerminal = func() bool { return true }
	ttyInput = bytes.NewBufferString("always\n")
	ttyError = &bytes.Buffer{}

	p := &TTYProvider{}
	req := ApprovalRequest{
		ID:         "test-tty-always",
		ToolName:   "exec",
		SessionKey: "tty:local",
		CreatedAt:  time.Now(),
	}

	resp, err := p.RequestApproval(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Approved || !resp.AlwaysAllow {
		t.Fatalf("expected always-allow approval, got %+v", resp)
	}
}

func TestTTYProvider_DenyUsesInjectedStreams(t *testing.T) {
	origIsTerminal := ttyIsTerminal
	origInput := ttyInput
	origError := ttyError
	t.Cleanup(func() {
		ttyIsTerminal = origIsTerminal
		ttyInput = origInput
		ttyError = origError
	})

	ttyIsTerminal = func() bool { return true }
	ttyInput = bytes.NewBufferString("n\n")
	ttyError = &bytes.Buffer{}

	p := &TTYProvider{}
	req := ApprovalRequest{
		ID:         "test-tty-deny",
		ToolName:   "exec",
		SessionKey: "tty:local",
		CreatedAt:  time.Now(),
	}

	resp, err := p.RequestApproval(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Approved || resp.AlwaysAllow {
		t.Fatalf("expected denial, got %+v", resp)
	}
}

func TestTTYProvider_EOFDeniesWithoutError(t *testing.T) {
	origIsTerminal := ttyIsTerminal
	origInput := ttyInput
	origError := ttyError
	t.Cleanup(func() {
		ttyIsTerminal = origIsTerminal
		ttyInput = origInput
		ttyError = origError
	})

	ttyIsTerminal = func() bool { return true }
	ttyInput = bytes.NewBuffer(nil)
	var errBuf bytes.Buffer
	ttyError = &errBuf

	p := &TTYProvider{}
	req := ApprovalRequest{
		ID:         "test-tty-eof",
		ToolName:   "exec",
		SessionKey: "tty:local",
		Summary:    "Delete: /tmp/test",
		CreatedAt:  time.Now(),
	}

	resp, err := p.RequestApproval(context.Background(), req)
	if err != nil {
		t.Fatalf("expected EOF denial without error, got %v", err)
	}
	if resp.Approved || resp.AlwaysAllow {
		t.Fatalf("expected denial, got %+v", resp)
	}
	if resp.Provider != "tty" {
		t.Fatalf("expected provider tty, got %q", resp.Provider)
	}
	if !strings.Contains(errBuf.String(), "[y/a/N]") {
		t.Fatalf("expected prompt output, got %q", errBuf.String())
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
