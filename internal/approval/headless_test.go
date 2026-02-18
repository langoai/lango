package approval

import (
	"context"
	"testing"
	"time"
)

func TestHeadlessProvider_AlwaysApproves(t *testing.T) {
	p := &HeadlessProvider{}

	req := ApprovalRequest{
		ID:         "test-1",
		ToolName:   "exec",
		SessionKey: "any:session",
		CreatedAt:  time.Now(),
	}

	resp, err := p.RequestApproval(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Approved {
		t.Error("expected HeadlessProvider to approve")
	}
}

func TestHeadlessProvider_CanHandleAlwaysFalse(t *testing.T) {
	p := &HeadlessProvider{}

	if p.CanHandle("any:session") {
		t.Error("HeadlessProvider.CanHandle should always return false")
	}
}

func TestHeadlessProvider_InterfaceCompliance(t *testing.T) {
	var _ Provider = (*HeadlessProvider)(nil)
}

func TestCompositeProvider_HeadlessFallback(t *testing.T) {
	comp := NewCompositeProvider()
	comp.SetTTYFallback(&HeadlessProvider{})

	req := ApprovalRequest{
		ID:         "test-1",
		ToolName:   "exec",
		SessionKey: "unknown:session",
		CreatedAt:  time.Now(),
	}

	resp, err := comp.RequestApproval(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Approved {
		t.Error("expected headless fallback to approve")
	}
}
