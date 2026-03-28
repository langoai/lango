package app

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	cronpkg "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- stubSessionStore ---

// stubSessionStore is a minimal session.Store that is NOT an EntStore.
// Used to test branches that require EntStore and should return nil otherwise.
type stubSessionStore struct{}

func (s *stubSessionStore) Create(_ *session.Session) error              { return nil }
func (s *stubSessionStore) Get(_ string) (*session.Session, error)       { return nil, nil }
func (s *stubSessionStore) Update(_ *session.Session) error              { return nil }
func (s *stubSessionStore) Delete(_ string) error                        { return nil }
func (s *stubSessionStore) AppendMessage(_ string, _ session.Message) error { return nil }
func (s *stubSessionStore) AnnotateTimeout(_ string, _ string) error     { return nil }
func (s *stubSessionStore) Close() error                                 { return nil }
func (s *stubSessionStore) GetSalt(_ string) ([]byte, error)             { return nil, nil }
func (s *stubSessionStore) SetSalt(_ string, _ []byte) error                              { return nil }
func (s *stubSessionStore) ListSessions(_ context.Context) ([]session.SessionSummary, error) { return nil, nil }

// --- initCron ---

func TestInitCron_DisabledReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Cron.Enabled = false

	result := initCron(cfg, &stubSessionStore{}, &App{Config: cfg})

	assert.Nil(t, result, "expected nil scheduler when cron is disabled")
}

func TestInitCron_NonEntStoreReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Cron.Enabled = true

	result := initCron(cfg, &stubSessionStore{}, &App{Config: cfg})

	assert.Nil(t, result, "expected nil scheduler when store is not EntStore")
}

func TestInitCron_DisabledBranch_TableDriven(t *testing.T) {
	tests := []struct {
		give        string
		giveCronOn  bool
		giveStore   session.Store
		wantNil     bool
	}{
		{
			give:       "disabled config returns nil",
			giveCronOn: false,
			giveStore:  &stubSessionStore{},
			wantNil:    true,
		},
		{
			give:       "enabled but non-EntStore returns nil",
			giveCronOn: true,
			giveStore:  &stubSessionStore{},
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Cron.Enabled = tt.giveCronOn

			result := initCron(cfg, tt.giveStore, &App{Config: cfg})

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// --- initBackground ---

func TestInitBackground_DisabledReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Background.Enabled = false

	result := initBackground(cfg, &App{Config: cfg})

	assert.Nil(t, result, "expected nil manager when background is disabled")
}

func TestInitBackground_DisabledBranch_TableDriven(t *testing.T) {
	tests := []struct {
		give     string
		giveOn   bool
		wantNil  bool
	}{
		{
			give:    "disabled returns nil",
			giveOn:  false,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Background.Enabled = tt.giveOn

			result := initBackground(cfg, &App{Config: cfg})

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// --- initWorkflow ---

func TestInitWorkflow_DisabledReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Workflow.Enabled = false

	result := initWorkflow(cfg, &stubSessionStore{}, &App{Config: cfg}, nil)

	assert.Nil(t, result, "expected nil engine when workflow is disabled")
}

func TestInitWorkflow_NonEntStoreReturnsNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Workflow.Enabled = true

	result := initWorkflow(cfg, &stubSessionStore{}, &App{Config: cfg}, nil)

	assert.Nil(t, result, "expected nil engine when store is not EntStore")
}

func TestInitWorkflow_DisabledBranch_TableDriven(t *testing.T) {
	tests := []struct {
		give          string
		giveWorkflowOn bool
		giveStore     session.Store
		wantNil       bool
	}{
		{
			give:           "disabled config returns nil",
			giveWorkflowOn: false,
			giveStore:      &stubSessionStore{},
			wantNil:        true,
		},
		{
			give:           "enabled but non-EntStore returns nil",
			giveWorkflowOn: true,
			giveStore:      &stubSessionStore{},
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Workflow.Enabled = tt.giveWorkflowOn

			result := initWorkflow(cfg, tt.giveStore, &App{Config: cfg}, nil)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// --- agentRunnerAdapter ---

func TestAgentRunnerAdapter_ImplementsInterfaces(t *testing.T) {
	// Verify agentRunnerAdapter satisfies all three AgentRunner interfaces.
	var _ cronpkg.AgentRunner = (*agentRunnerAdapter)(nil)
	var _ background.AgentRunner = (*agentRunnerAdapter)(nil)
	var _ workflow.AgentRunner = (*agentRunnerAdapter)(nil)
}

func TestAgentRunnerAdapter_HoldsAppReference(t *testing.T) {
	a := &App{Config: config.DefaultConfig()}
	adapter := &agentRunnerAdapter{app: a}

	require.NotNil(t, adapter.app)
	assert.Same(t, a, adapter.app, "adapter should hold reference to the provided App")
}

func TestAgentRunnerAdapter_RunMethodSignature(t *testing.T) {
	// Verify the Run method exists with the correct signature by obtaining it
	// via the interface. If the signature mismatches, this will not compile.
	a := &App{Config: config.DefaultConfig()}
	adapter := &agentRunnerAdapter{app: a}

	// Assign to each interface to confirm signature compatibility.
	var cronRunner cronpkg.AgentRunner = adapter
	var bgRunner background.AgentRunner = adapter
	var wfRunner workflow.AgentRunner = adapter

	require.NotNil(t, cronRunner)
	require.NotNil(t, bgRunner)
	require.NotNil(t, wfRunner)
}
