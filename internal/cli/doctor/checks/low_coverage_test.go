package checks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusString_AllKnownValuesAndUnknown(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPass, "pass"},
		{StatusWarn, "warn"},
		{StatusFail, "fail"},
		{StatusSkip, "skip"},
		{Status(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.String())
		})
	}
}

func TestAllChecks_ReturnsExpectedDoctorChecks(t *testing.T) {
	got := AllChecks()
	require.NotEmpty(t, got)

	names := make([]string, 0, len(got))
	for _, check := range got {
		names = append(names, check.Name())
	}

	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			t.Fatalf("duplicate check name %q", name)
		}
		seen[name] = true
	}

	for _, name := range []string{
		"Approval System",
		"Economy Layer",
		"Graph Store",
		"Observability",
		"Observational Memory",
		"Output Scanning",
		"P2P Workspaces",
		"Proactive Librarian",
		"Smart Contracts",
		"Tool Hooks",
	} {
		assert.Contains(t, seen, name)
	}
}

func TestFeatureStatusToDoctorResult_StatusAndMessagePrecedence(t *testing.T) {
	tests := []struct {
		name       string
		status     types.FeatureStatus
		wantStatus Status
		wantMsg    string
	}{
		{
			name:       "healthy defaults to healthy message",
			status:     types.FeatureStatus{Name: "Graph", Enabled: true, Healthy: true},
			wantStatus: StatusPass,
			wantMsg:    "Graph is healthy",
		},
		{
			name:       "disabled with reason is warning",
			status:     types.FeatureStatus{Name: "RAG", Enabled: false, Healthy: true, Reason: "missing vectors"},
			wantStatus: StatusWarn,
			wantMsg:    "missing vectors",
		},
		{
			name:       "unhealthy enabled is failure",
			status:     types.FeatureStatus{Name: "Embedding", Enabled: true, Healthy: false, Suggestion: "configure provider"},
			wantStatus: StatusFail,
			wantMsg:    "configure provider",
		},
		{
			name:       "disabled without reason is skip",
			status:     types.FeatureStatus{Name: "Librarian", Enabled: false, Healthy: true},
			wantStatus: StatusSkip,
			wantMsg:    "Librarian is healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FeatureStatusToDoctorResult(tt.status)
			assert.Equal(t, tt.status.Name, result.Name)
			assert.Equal(t, tt.wantStatus, result.Status)
			assert.Equal(t, tt.wantMsg, result.Message)
		})
	}
}

func TestApprovalCheck_Run_DeterministicPolicies(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*config.Config)
		wantStatus Status
		wantMsg    string
	}{
		{
			name: "disabled interceptor skips",
			mutate: func(cfg *config.Config) {
				cfg.Security.Interceptor.Enabled = false
			},
			wantStatus: StatusSkip,
			wantMsg:    "approval system inactive",
		},
		{
			name: "empty policy defaults to dangerous",
			mutate: func(cfg *config.Config) {
				cfg.Security.Interceptor.Enabled = true
				cfg.Security.Interceptor.ApprovalPolicy = ""
			},
			wantStatus: StatusPass,
			wantMsg:    "policy=dangerous",
		},
		{
			name: "none policy warns",
			mutate: func(cfg *config.Config) {
				cfg.Security.Interceptor.Enabled = true
				cfg.Security.Interceptor.ApprovalPolicy = config.ApprovalPolicyNone
			},
			wantStatus: StatusWarn,
			wantMsg:    "all tools auto-approved",
		},
		{
			name: "unknown policy fails",
			mutate: func(cfg *config.Config) {
				cfg.Security.Interceptor.Enabled = true
				cfg.Security.Interceptor.ApprovalPolicy = config.ApprovalPolicy("mystery")
			},
			wantStatus: StatusFail,
			wantMsg:    `Unknown approval policy: "mystery"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			tt.mutate(cfg)

			result := (&ApprovalCheck{}).Run(context.Background(), cfg)

			assert.Equal(t, tt.wantStatus, result.Status)
			assert.Contains(t, result.Message, tt.wantMsg)
		})
	}
}

func TestContractCheck_Run_PaymentConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*config.Config)
		wantStatus Status
		wantMsg    string
	}{
		{
			name:       "payment disabled skips",
			mutate:     func(cfg *config.Config) { cfg.Payment.Enabled = false },
			wantStatus: StatusSkip,
			wantMsg:    "Payment not enabled",
		},
		{
			name: "missing rpc and chain id fails",
			mutate: func(cfg *config.Config) {
				cfg.Payment.Enabled = true
				cfg.Payment.Network.RPCURL = ""
				cfg.Payment.Network.ChainID = 0
			},
			wantStatus: StatusFail,
			wantMsg:    "payment.network.rpcUrl is required",
		},
		{
			name: "rpc and chain id pass",
			mutate: func(cfg *config.Config) {
				cfg.Payment.Enabled = true
				cfg.Payment.Network.RPCURL = "http://127.0.0.1:8545"
				cfg.Payment.Network.ChainID = 84532
			},
			wantStatus: StatusPass,
			wantMsg:    "Contract interaction configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			tt.mutate(cfg)

			result := (&ContractCheck{}).Run(context.Background(), cfg)

			assert.Equal(t, tt.wantStatus, result.Status, result.Message)
			assert.Contains(t, result.Message, tt.wantMsg)
		})
	}
}

func TestEconomyCheck_Run_ValidationBranches(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Budget.DefaultMax = "not-a-decimal"
	cfg.Economy.Risk.HighTrustScore = 0.4
	cfg.Economy.Risk.MediumTrustScore = 0.5
	cfg.Economy.Escrow.Enabled = true
	cfg.Economy.Escrow.MaxMilestones = 0
	cfg.Economy.Negotiate.Enabled = true
	cfg.Economy.Negotiate.MaxRounds = 0
	cfg.Economy.Pricing.Enabled = true
	cfg.Economy.Pricing.MinPrice = "also-bad"

	result := (&EconomyCheck{}).Run(context.Background(), cfg)

	assert.Equal(t, StatusFail, result.Status)
	assert.Contains(t, result.Message, `budget.defaultMax "not-a-decimal" is not a valid decimal`)
	assert.Contains(t, result.Message, "risk.highTrustScore")
	assert.Contains(t, result.Message, "escrow.maxMilestones should be positive")
	assert.Contains(t, result.Message, "negotiate.maxRounds should be positive")
	assert.Contains(t, result.Message, `pricing.minPrice "also-bad" is not a valid decimal`)
}

func TestGraphStoreCheck_Run_ValidationBranches(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.Backend = "neo4j"
	cfg.Graph.DatabasePath = ""
	cfg.Graph.MaxTraversalDepth = 0
	cfg.Graph.MaxExpansionResults = -1

	result := (&GraphStoreCheck{}).Run(context.Background(), cfg)

	assert.Equal(t, StatusFail, result.Status)
	assert.Contains(t, result.Message, `unsupported backend "neo4j"`)
	assert.Contains(t, result.Message, "graph.databasePath is not set")
	assert.Contains(t, result.Message, "graph.maxTraversalDepth should be positive")
	assert.Contains(t, result.Message, "graph.maxExpansionResults should be positive")
}

func TestLibrarianCheck_Run_ProviderFallbackAndWarnings(t *testing.T) {
	t.Run("warns when dependencies are missing", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Librarian.Enabled = true
		cfg.Knowledge.Enabled = false
		cfg.Librarian.Provider = ""
		cfg.Agent.Provider = ""

		result := (&LibrarianCheck{}).Run(context.Background(), cfg)

		assert.Equal(t, StatusWarn, result.Status)
		assert.Contains(t, result.Message, "knowledge.enabled is false")
		assert.Contains(t, result.Message, "no provider configured")
	})

	t.Run("uses agent provider when librarian provider is empty", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Librarian.Enabled = true
		cfg.Knowledge.Enabled = true
		cfg.Librarian.Provider = ""
		cfg.Agent.Provider = "openai"

		result := (&LibrarianCheck{}).Run(context.Background(), cfg)

		assert.Equal(t, StatusPass, result.Status)
		assert.Contains(t, result.Message, "provider=openai")
		assert.Contains(t, result.Message, "model=(default)")
	})
}

func TestObservabilityCheck_Run_FeatureSummaryAndWarnings(t *testing.T) {
	t.Run("warns for invalid enabled subfeature retention", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Observability.Enabled = true
		cfg.Observability.Tokens.PersistHistory = true
		cfg.Observability.Tokens.RetentionDays = 0
		cfg.Observability.Health.Enabled = true
		cfg.Observability.Health.Interval = 0
		cfg.Observability.Audit.Enabled = true
		cfg.Observability.Audit.RetentionDays = -1

		result := (&ObservabilityCheck{}).Run(context.Background(), cfg)

		assert.Equal(t, StatusWarn, result.Status)
		assert.Contains(t, result.Message, "tokens.retentionDays should be positive")
		assert.Contains(t, result.Message, "health.interval should be positive")
		assert.Contains(t, result.Message, "audit.retentionDays should be positive")
	})

	t.Run("summarizes enabled features", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Observability.Enabled = true
		cfg.Observability.Health.Enabled = true
		cfg.Observability.Health.Interval = time.Second
		cfg.Observability.Audit.Enabled = true
		cfg.Observability.Audit.RetentionDays = 30
		cfg.Observability.Metrics.Enabled = true

		result := (&ObservabilityCheck{}).Run(context.Background(), cfg)

		assert.Equal(t, StatusPass, result.Status, result.Message)
		assert.Contains(t, result.Message, "tokens, health, audit, metrics")
	})
}

func TestObservationalMemoryCheck_Run_InvalidThresholdsAndProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ObservationalMemory.Enabled = true
	cfg.ObservationalMemory.MessageTokenThreshold = 0
	cfg.ObservationalMemory.ObservationTokenThreshold = -1
	cfg.ObservationalMemory.MaxMessageTokenBudget = 0
	cfg.ObservationalMemory.Provider = "missing"
	cfg.Providers = map[string]config.ProviderConfig{}

	result := (&ObservationalMemoryCheck{}).Run(context.Background(), cfg)

	assert.Equal(t, StatusFail, result.Status)
	assert.Contains(t, result.Message, "messageTokenThreshold must be positive")
	assert.Contains(t, result.Message, "observationTokenThreshold must be positive")
	assert.Contains(t, result.Message, "maxMessageTokenBudget must be positive")
	assert.Contains(t, result.Message, "provider 'missing' not found")
}

func TestOutputScanningCheck_Run_NoLiveHTTP(t *testing.T) {
	t.Run("disabled interceptor without database skips", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Security.Interceptor.Enabled = false
		cfg.Session.DatabasePath = ""

		result := (&OutputScanningCheck{}).Run(context.Background(), cfg)

		assert.Equal(t, StatusSkip, result.Status)
		assert.Equal(t, "Output interceptor is disabled", result.Message)
	})

	t.Run("presidio health uses httptest server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/health", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := config.DefaultConfig()
		cfg.Security.Interceptor.Enabled = true
		cfg.Security.Interceptor.RedactPII = true
		cfg.Security.Interceptor.Presidio.Enabled = true
		cfg.Security.Interceptor.Presidio.URL = server.URL

		result := (&OutputScanningCheck{}).Run(context.Background(), cfg)

		assert.Equal(t, StatusPass, result.Status, result.Message)
	})
}

func TestToolHooksCheck_Run_ReportsActiveHooks(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Knowledge.Enabled = true
	cfg.Security.Interceptor.Enabled = true

	result := (&ToolHooksCheck{}).Run(context.Background(), cfg)

	assert.Equal(t, StatusPass, result.Status)
	assert.Contains(t, result.Message, "4 tool hooks configured")
	assert.Contains(t, result.Message, "learning_observer")
	assert.Contains(t, result.Message, "approval_gate")
}

func TestWorkspaceCheck_RunAndFix_DataDirectory(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "missing-workspaces")
	cfg := config.DefaultConfig()
	cfg.P2P.Workspace.Enabled = true
	cfg.P2P.Workspace.DataDir = dataDir
	cfg.P2P.Workspace.MaxWorkspaces = 3

	check := &WorkspaceCheck{}
	result := check.Run(context.Background(), cfg)

	assert.Equal(t, StatusWarn, result.Status)
	assert.True(t, result.Fixable)
	assert.Equal(t, "Create directory "+dataDir, result.FixAction)
	assert.Contains(t, result.Message, "Workspace data dir missing")

	fixed := check.Fix(context.Background(), cfg)

	assert.Equal(t, StatusPass, fixed.Status, fixed.Message)

	afterFix := check.Run(context.Background(), cfg)
	if afterFix.Status == StatusWarn && strings.Contains(afterFix.Message, "git binary not found") {
		t.Skip("git is not available in PATH for this test environment")
	}
	assert.Equal(t, StatusPass, afterFix.Status, afterFix.Message)
}

func TestEveryTargetCheck_FixDelegatesToRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Payment.Enabled = true
	cfg.Payment.Network.RPCURL = "http://127.0.0.1:8545"
	cfg.Observability.Enabled = true
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.RedactPII = true

	checks := []Check{
		&ApprovalCheck{},
		&ContractCheck{},
		&EconomyCheck{},
		&GraphStoreCheck{},
		&LibrarianCheck{},
		&ObservabilityCheck{},
		&ObservationalMemoryCheck{},
		&OutputScanningCheck{},
		&ToolHooksCheck{},
	}

	for _, check := range checks {
		t.Run(check.Name(), func(t *testing.T) {
			run := check.Run(context.Background(), cfg)
			fix := check.Fix(context.Background(), cfg)

			assert.True(t, reflect.DeepEqual(run, fix), "Run=%#v Fix=%#v", run, fix)
		})
	}
}
