package approval

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// mockProvider is a test provider that handles a specific prefix.
type mockProvider struct {
	prefix  string
	result  bool
	err     error
	called  bool
	callMu  sync.Mutex
}

func (m *mockProvider) RequestApproval(_ context.Context, _ ApprovalRequest) (ApprovalResponse, error) {
	m.callMu.Lock()
	m.called = true
	m.callMu.Unlock()
	return ApprovalResponse{Approved: m.result}, m.err
}

func (m *mockProvider) CanHandle(sessionKey string) bool {
	if m.prefix == "" {
		return false
	}
	return len(sessionKey) >= len(m.prefix) && sessionKey[:len(m.prefix)] == m.prefix
}

func (m *mockProvider) wasCalled() bool {
	m.callMu.Lock()
	defer m.callMu.Unlock()
	return m.called
}

func TestCompositeProvider_RoutesByPrefix(t *testing.T) {
	tests := []struct {
		give       string
		wantResult bool
		wantCalled string // which provider name should be called
	}{
		{give: "telegram:123:456", wantResult: true, wantCalled: "telegram"},
		{give: "discord:ch:usr", wantResult: true, wantCalled: "discord"},
		{give: "slack:ch:usr", wantResult: true, wantCalled: "slack"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tg := &mockProvider{prefix: "telegram:", result: true}
			dc := &mockProvider{prefix: "discord:", result: true}
			sl := &mockProvider{prefix: "slack:", result: true}

			comp := NewCompositeProvider()
			comp.Register(tg)
			comp.Register(dc)
			comp.Register(sl)

			req := ApprovalRequest{
				ID:         "test-1",
				ToolName:   "exec",
				SessionKey: tt.give,
				CreatedAt:  time.Now(),
			}

			resp, err := comp.RequestApproval(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Approved != tt.wantResult {
				t.Errorf("got approved=%v, want %v", resp.Approved, tt.wantResult)
			}

			providers := map[string]*mockProvider{
				"telegram": tg,
				"discord":  dc,
				"slack":    sl,
			}
			for name, p := range providers {
				if name == tt.wantCalled && !p.wasCalled() {
					t.Errorf("expected %s provider to be called", name)
				}
				if name != tt.wantCalled && p.wasCalled() {
					t.Errorf("expected %s provider NOT to be called", name)
				}
			}
		})
	}
}

func TestCompositeProvider_TTYFallback(t *testing.T) {
	tty := &mockProvider{result: true}

	comp := NewCompositeProvider()
	comp.SetTTYFallback(tty)

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
		t.Error("expected TTY fallback to approve")
	}
	if !tty.wasCalled() {
		t.Error("expected TTY fallback to be called")
	}
}

func TestCompositeProvider_FailClosed(t *testing.T) {
	comp := NewCompositeProvider()

	req := ApprovalRequest{
		ID:         "test-1",
		ToolName:   "exec",
		SessionKey: "unknown:session",
		CreatedAt:  time.Now(),
	}

	resp, err := comp.RequestApproval(context.Background(), req)
	if err == nil {
		t.Fatal("expected error when no provider available")
	}
	if resp.Approved {
		t.Error("expected fail-closed (deny) when no provider available")
	}
}

func TestCompositeProvider_ProviderError(t *testing.T) {
	errProvider := &mockProvider{
		prefix: "telegram:",
		result: false,
		err:    fmt.Errorf("connection lost"),
	}

	comp := NewCompositeProvider()
	comp.Register(errProvider)

	req := ApprovalRequest{
		ID:         "test-1",
		ToolName:   "exec",
		SessionKey: "telegram:123:456",
		CreatedAt:  time.Now(),
	}

	resp, err := comp.RequestApproval(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}
	if resp.Approved {
		t.Error("expected denial on error")
	}
}

func TestCompositeProvider_ConcurrentRequests(t *testing.T) {
	slow := &mockProvider{prefix: "telegram:", result: true}

	comp := NewCompositeProvider()
	comp.Register(slow)

	var wg sync.WaitGroup
	results := make([]ApprovalResponse, 10)
	errs := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := ApprovalRequest{
				ID:         fmt.Sprintf("req-%d", idx),
				ToolName:   "exec",
				SessionKey: "telegram:123:456",
				CreatedAt:  time.Now(),
			}
			results[idx], errs[idx] = comp.RequestApproval(context.Background(), req)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		if errs[i] != nil {
			t.Errorf("request %d: unexpected error: %v", i, errs[i])
		}
		if !results[i].Approved {
			t.Errorf("request %d: expected approval", i)
		}
	}
}

func TestCompositeProvider_CanHandleAlwaysTrue(t *testing.T) {
	comp := NewCompositeProvider()
	if !comp.CanHandle("anything") {
		t.Error("CompositeProvider.CanHandle should always return true")
	}
}

func TestCompositeProvider_FirstMatchWins(t *testing.T) {
	first := &mockProvider{prefix: "telegram:", result: true}
	second := &mockProvider{prefix: "telegram:", result: false}

	comp := NewCompositeProvider()
	comp.Register(first)
	comp.Register(second)

	req := ApprovalRequest{
		ID:         "test-1",
		ToolName:   "exec",
		SessionKey: "telegram:123:456",
		CreatedAt:  time.Now(),
	}

	resp, err := comp.RequestApproval(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Approved {
		t.Error("expected first provider to win (approve)")
	}
	if !first.wasCalled() {
		t.Error("expected first provider to be called")
	}
	if second.wasCalled() {
		t.Error("expected second provider NOT to be called")
	}
}

func TestGatewayProvider(t *testing.T) {
	tests := []struct {
		give           string
		hasCompanions  bool
		approveResult  bool
		approveErr     error
		wantCanHandle  bool
		wantApproved   bool
		wantErr        bool
	}{
		{
			give:          "with companions, approved",
			hasCompanions: true,
			approveResult: true,
			wantCanHandle: true,
			wantApproved:  true,
		},
		{
			give:          "with companions, denied",
			hasCompanions: true,
			approveResult: false,
			wantCanHandle: true,
			wantApproved:  false,
		},
		{
			give:          "no companions",
			hasCompanions: false,
			wantCanHandle: false,
		},
		{
			give:          "with companions, error",
			hasCompanions: true,
			approveErr:    fmt.Errorf("timeout"),
			wantCanHandle: true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			gw := &mockGatewayApprover{
				companions: tt.hasCompanions,
				result:     tt.approveResult,
				err:        tt.approveErr,
			}
			p := NewGatewayProvider(gw)

			if got := p.CanHandle("any"); got != tt.wantCanHandle {
				t.Errorf("CanHandle = %v, want %v", got, tt.wantCanHandle)
			}

			if !tt.wantCanHandle {
				return
			}

			req := ApprovalRequest{
				ID:       "test-1",
				ToolName: "exec",
			}
			resp, err := p.RequestApproval(context.Background(), req)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if resp.Approved != tt.wantApproved {
				t.Errorf("approved = %v, want %v", resp.Approved, tt.wantApproved)
			}
		})
	}
}

type mockGatewayApprover struct {
	companions bool
	result     bool
	err        error
}

func (m *mockGatewayApprover) HasCompanions() bool {
	return m.companions
}

func (m *mockGatewayApprover) RequestApproval(_ context.Context, _ string) (ApprovalResponse, error) {
	return ApprovalResponse{Approved: m.result}, m.err
}
