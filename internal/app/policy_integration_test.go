package app

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/toolchain"
	execpkg "github.com/langoai/lango/internal/tools/exec"
)

// mockApprovalProvider tracks whether RequestApproval was called.
type mockApprovalProvider struct {
	called atomic.Int32
}

func (m *mockApprovalProvider) RequestApproval(_ context.Context, _ approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	m.called.Add(1)
	return approval.ApprovalResponse{Approved: true}, nil
}

func (m *mockApprovalProvider) CanHandle(_ string) bool { return true }

func (m *mockApprovalProvider) wasCalled() bool { return m.called.Load() > 0 }

// buildTestChain creates a tool with WithPolicy (outermost) and WithApproval,
// matching the production middleware application order.
func buildTestChain(t *testing.T, ap approval.Provider) (*agent.Tool, *atomic.Int32) {
	t.Helper()

	var executorCalled atomic.Int32
	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			executorCalled.Add(1)
			return "executed", nil
		},
	}

	guard := execpkg.NewCommandGuard([]string{"~/.lango"})
	classifier := func(cmd string) (string, execpkg.ReasonCode) {
		lower := strings.ToLower(strings.TrimSpace(cmd))
		if strings.HasPrefix(lower, "lango ") {
			return "blocked: lango CLI", execpkg.ReasonLangoCLI
		}
		return "", execpkg.ReasonNone
	}
	pe := execpkg.NewPolicyEvaluator(guard, classifier, nil,
		execpkg.WithCatastrophicPatterns(toolchain.DefaultBlockedPatterns()))

	// Apply in production order: approval first (inner), then policy (outer).
	ic := config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous}
	gs := approval.NewGrantStore()
	tool = toolchain.Chain(tool, toolchain.WithApproval(ic, ap, gs, nil, nil))
	tool = toolchain.Chain(tool, execpkg.WithPolicy(pe))

	return tool, &executorCalled
}

func TestPolicyIntegration_CatastrophicBlockedBeforeApproval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give           string
		wantApproval   bool
		wantExecution  bool
	}{
		// Catastrophic: blocked by policy before approval
		{give: "rm -rf /", wantApproval: false, wantExecution: false},
		{give: `eval "rm -rf /"`, wantApproval: false, wantExecution: false},
		{give: "dd if=/dev/zero of=/dev/sda", wantApproval: false, wantExecution: false},
		// Kill verb: blocked by policy before approval
		{give: "kill 1234", wantApproval: false, wantExecution: false},
		{give: `sh -c "kill 1234"`, wantApproval: false, wantExecution: false},
		// Shell-wrapped lango CLI: blocked by policy before approval
		{give: `bash -c "lango security"`, wantApproval: false, wantExecution: false},
		// Clean command: reaches approval (and gets approved by mock)
		{give: "ls -la", wantApproval: true, wantExecution: true},
		// Opaque non-catastrophic: observe passes through to approval
		{give: `eval "echo hello"`, wantApproval: true, wantExecution: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			ap := &mockApprovalProvider{}
			tool, executorCalled := buildTestChain(t, ap)

			result, err := tool.Handler(context.Background(), map[string]interface{}{"command": tt.give})
			require.NoError(t, err)

			assert.Equal(t, tt.wantApproval, ap.wasCalled(),
				"approval provider called for %q", tt.give)
			assert.Equal(t, tt.wantExecution, executorCalled.Load() > 0,
				"executor called for %q", tt.give)

			if !tt.wantApproval && !tt.wantExecution {
				br, ok := result.(*execpkg.BlockedResult)
				require.True(t, ok, "expected BlockedResult for %q, got %T", tt.give, result)
				assert.True(t, br.Blocked)
			}
		})
	}
}
