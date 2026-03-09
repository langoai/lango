package health

import (
	"context"
	"errors"
	"testing"
)

// stubChecker is a test helper that returns a fixed status.
type stubChecker struct {
	name   string
	status Status
	msg    string
}

func (s *stubChecker) Name() string { return s.name }

func (s *stubChecker) Check(_ context.Context) ComponentHealth {
	return ComponentHealth{
		Name:    s.name,
		Status:  s.status,
		Message: s.msg,
	}
}

func TestRegistry_Empty(t *testing.T) {
	r := NewRegistry()
	result := r.CheckAll(context.Background())
	if result.Status != StatusHealthy {
		t.Errorf("Status = %q, want %q", result.Status, StatusHealthy)
	}
	if len(result.Components) != 0 {
		t.Errorf("Components = %d, want 0", len(result.Components))
	}
}

func TestRegistry_AllHealthy(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubChecker{name: "a", status: StatusHealthy})
	r.Register(&stubChecker{name: "b", status: StatusHealthy})

	result := r.CheckAll(context.Background())
	if result.Status != StatusHealthy {
		t.Errorf("Status = %q, want %q", result.Status, StatusHealthy)
	}
	if len(result.Components) != 2 {
		t.Errorf("Components = %d, want 2", len(result.Components))
	}
}

func TestRegistry_DegradedWins(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubChecker{name: "a", status: StatusHealthy})
	r.Register(&stubChecker{name: "b", status: StatusDegraded})

	result := r.CheckAll(context.Background())
	if result.Status != StatusDegraded {
		t.Errorf("Status = %q, want %q", result.Status, StatusDegraded)
	}
}

func TestRegistry_UnhealthyWins(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubChecker{name: "a", status: StatusHealthy})
	r.Register(&stubChecker{name: "b", status: StatusDegraded})
	r.Register(&stubChecker{name: "c", status: StatusUnhealthy})

	result := r.CheckAll(context.Background())
	if result.Status != StatusUnhealthy {
		t.Errorf("Status = %q, want %q", result.Status, StatusUnhealthy)
	}
}

func TestRegistry_Status(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubChecker{name: "a", status: StatusDegraded})

	got := r.Status(context.Background())
	if got != StatusDegraded {
		t.Errorf("Status() = %q, want %q", got, StatusDegraded)
	}
}

func TestDatabaseCheck(t *testing.T) {
	tests := []struct {
		give       string
		pingErr    error
		wantStatus Status
		wantMsg    string
	}{
		{
			give:       "healthy when ping succeeds",
			pingErr:    nil,
			wantStatus: StatusHealthy,
			wantMsg:    "connected",
		},
		{
			give:       "unhealthy when ping fails",
			pingErr:    errors.New("connection refused"),
			wantStatus: StatusUnhealthy,
			wantMsg:    "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			check := NewDatabaseCheck(func(_ context.Context) error {
				return tt.pingErr
			})
			result := check.Check(context.Background())
			if result.Name != "database" {
				t.Errorf("Name = %q, want %q", result.Name, "database")
			}
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", result.Status, tt.wantStatus)
			}
			if result.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", result.Message, tt.wantMsg)
			}
		})
	}
}

func TestMemoryCheck(t *testing.T) {
	check := NewMemoryCheck(0) // no threshold
	result := check.Check(context.Background())
	if result.Name != "memory" {
		t.Errorf("Name = %q, want %q", result.Name, "memory")
	}
	if result.Status != StatusHealthy {
		t.Errorf("Status = %q, want %q", result.Status, StatusHealthy)
	}
	if result.Metadata == nil {
		t.Fatal("Metadata is nil")
	}
	for _, key := range []string{"allocMB", "sysMB", "goroutines"} {
		if _, ok := result.Metadata[key]; !ok {
			t.Errorf("Metadata missing key %q", key)
		}
	}
}

func TestProviderCheck(t *testing.T) {
	tests := []struct {
		give       string
		pingErr    error
		wantStatus Status
		wantMsg    string
	}{
		{
			give:       "healthy when reachable",
			pingErr:    nil,
			wantStatus: StatusHealthy,
			wantMsg:    "reachable",
		},
		{
			give:       "degraded when unreachable",
			pingErr:    errors.New("timeout"),
			wantStatus: StatusDegraded,
			wantMsg:    "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			check := NewProviderCheck("openai", func(_ context.Context) error {
				return tt.pingErr
			})
			result := check.Check(context.Background())
			if result.Name != "provider.openai" {
				t.Errorf("Name = %q, want %q", result.Name, "provider.openai")
			}
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", result.Status, tt.wantStatus)
			}
			if result.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", result.Message, tt.wantMsg)
			}
		})
	}
}
