# Approval Flow First Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the first `knowledge exchange v1` approval-flow slice centered on `artifact release approval`, structured decision states, and audit-backed release outcome records.

**Architecture:** This slice intentionally does not build a full dispute engine or human approval UI. It introduces a dedicated approval-flow domain model, a leader-agent-facing artifact release approval evaluator, structured outcome records for `reject` and `request-revision`, and a narrow automatic-settlement coupling for `approve`. It reuses the landed exportability receipt rather than replacing it.

**Tech Stack:** Go, Ent, Cobra CLI, existing knowledge/exportability code (`internal/knowledge/*`, `internal/exportability/*`), meta tools (`internal/app/tools_meta.go`), audit log schema (`internal/ent/schema/audit_log.go`), MkDocs/Markdown docs

---

## Scope Split

The approval-flow design covers two approval objects:

- upfront payment approval
- artifact release approval

This first slice covers only the parts needed to make `artifact release approval` real and useful:

- a typed approval-flow domain package,
- structured artifact release decision states,
- release outcome records,
- audit-backed approval receipts,
- a minimal agent-facing release-approval tool,
- truthful operator docs.

This slice does **not** implement:

- full human escalation UI,
- dispute orchestration,
- partial settlement execution,
- milestone or team approval flows,
- complete upfront payment approval runtime.

## OpenSpec Precondition

Before touching implementation code, create or refresh an OpenSpec change for this slice. Use a narrow change name such as `approval-flow-first-slice`.

The implementation session must end with the repository's required OpenSpec workflow:

- `ff`
- `apply`
- `verify`
- `sync`
- `archive`

## File Map

- Create: `internal/approvalflow/types.go`
  - Approval objects, release states, issue classes, fulfillment grades, and outcome-record types.
- Create: `internal/approvalflow/release.go`
  - Core `artifact release approval` evaluator using artifact scope, exportability receipt, and transaction context.
- Create: `internal/approvalflow/release_test.go`
  - Unit tests for `approve`, `reject`, `request-revision`, and `escalate`.
- Modify: `internal/ent/schema/audit_log.go`
  - Add approval-flow audit actions such as `artifact_release_approval`.
- Modify: generated Ent files under `internal/ent/...`
  - Regenerated after schema changes.
- Modify: `internal/app/tools_meta.go`
  - Add a new `approve_artifact_release` meta tool.
- Create: `internal/app/tools_meta_approvalflow_test.go`
  - Real-store tests for the new release-approval tool.
- Modify: `internal/app/tools_parity_test.go`
  - Add `approve_artifact_release` to parity expectations.
- Modify: `internal/knowledge/store_test.go`
  - Add audit-log action coverage for the new approval action if stored through `SaveAuditLog`.
- Create: `docs/security/approval-flow.md`
  - Canonical operator doc for artifact release approval states and release outcome records.
- Modify: `docs/security/index.md`
  - Link the new approval-flow doc.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark approval flow as first-slice landed and distinguish remaining gaps.
- Modify: `docs/architecture/trust-security-policy-audit.md`
  - Add post-implementation notes under the approval row.
- Modify: `README.md`
  - Add a short truthful note about structured artifact release approval.
- Modify: `mkdocs.yml`
  - Add the new security doc to nav.

## Task 1: Introduce The Approval-Flow Domain Model

**Files:**
- Create: `internal/approvalflow/types.go`
- Create: `internal/approvalflow/release.go`
- Create: `internal/approvalflow/release_test.go`

- [ ] **Step 1: Write the failing approval-flow tests**

Create `internal/approvalflow/release_test.go`:

```go
package approvalflow

import (
	"testing"

	"github.com/langoai/lango/internal/exportability"
	"github.com/stretchr/testify/assert"
)

func TestApproveArtifactRelease_ApproveOnExportableScopeMatch(t *testing.T) {
	outcome := ApproveArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel: "research memo",
		RequestedScope: "research memo",
		Exportability: exportability.Receipt{
			State: exportability.StateExportable,
		},
	})

	assert.Equal(t, DecisionApprove, outcome.Decision)
	assert.Equal(t, SettlementAutoRelease, outcome.SettlementHint)
}

func TestApproveArtifactRelease_RequestRevisionOnScopeMismatch(t *testing.T) {
	outcome := ApproveArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel: "rough draft",
		RequestedScope: "final design memo",
		Exportability: exportability.Receipt{
			State: exportability.StateExportable,
		},
	})

	assert.Equal(t, DecisionRequestRevision, outcome.Decision)
	assert.Equal(t, IssueScopeMismatch, outcome.Issue)
}

func TestApproveArtifactRelease_EscalateOnNeedsHumanReview(t *testing.T) {
	outcome := ApproveArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel: "sensitive memo",
		RequestedScope: "sensitive memo",
		Exportability: exportability.Receipt{
			State: exportability.StateNeedsHumanReview,
		},
	})

	assert.Equal(t, DecisionEscalate, outcome.Decision)
	assert.Equal(t, IssuePolicy, outcome.Issue)
}

func TestApproveArtifactRelease_RejectOnBlockedOverrideAttempt(t *testing.T) {
	outcome := ApproveArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel: "blocked memo",
		RequestedScope: "blocked memo",
		Exportability: exportability.Receipt{
			State: exportability.StateBlocked,
		},
		OverrideRequested: false,
	})

	assert.Equal(t, DecisionReject, outcome.Decision)
	assert.Equal(t, IssuePolicy, outcome.Issue)
}
```

- [ ] **Step 2: Run the new tests and confirm they fail**

Run:

```bash
go test ./internal/approvalflow/... -count=1
```

Expected:

```text
FAIL
```

with undefined symbol errors for `ApproveArtifactRelease`, `ArtifactReleaseInput`, and the approval enums.

- [ ] **Step 3: Implement the minimal domain model**

Create `internal/approvalflow/types.go`:

```go
package approvalflow

import "github.com/langoai/lango/internal/exportability"

type ApprovalObject string

const (
	ObjectUpfrontPayment  ApprovalObject = "upfront_payment"
	ObjectArtifactRelease ApprovalObject = "artifact_release"
)

type Decision string

const (
	DecisionApprove         Decision = "approve"
	DecisionReject          Decision = "reject"
	DecisionRequestRevision Decision = "request-revision"
	DecisionEscalate        Decision = "escalate"
)

type IssueClass string

const (
	IssueScopeMismatch IssueClass = "scope_mismatch"
	IssueQuality       IssueClass = "quality_issue"
	IssuePolicy        IssueClass = "policy_issue"
)

type FulfillmentGrade string

const (
	FulfillmentNone        FulfillmentGrade = "none"
	FulfillmentPartial     FulfillmentGrade = "partial"
	FulfillmentSubstantial FulfillmentGrade = "substantial"
)

type SettlementHint string

const (
	SettlementAutoRelease SettlementHint = "auto_release"
	SettlementHold        SettlementHint = "hold"
	SettlementReview      SettlementHint = "review"
)

type ArtifactReleaseInput struct {
	ArtifactLabel      string
	RequestedScope     string
	Exportability      exportability.Receipt
	OverrideRequested  bool
	HighRisk           bool
}

type ArtifactReleaseOutcome struct {
	Object          ApprovalObject   `json:"object"`
	Decision        Decision         `json:"decision"`
	Reason          string           `json:"reason"`
	Issue           IssueClass       `json:"issue,omitempty"`
	Fulfillment     FulfillmentGrade `json:"fulfillment,omitempty"`
	FulfillmentRatio float64         `json:"fulfillment_ratio,omitempty"`
	SettlementHint  SettlementHint   `json:"settlement_hint"`
}
```

Create `internal/approvalflow/release.go`:

```go
package approvalflow

func ApproveArtifactRelease(in ArtifactReleaseInput) ArtifactReleaseOutcome {
	out := ArtifactReleaseOutcome{
		Object:         ObjectArtifactRelease,
		SettlementHint: SettlementHold,
	}

	if in.Exportability.State == "needs-human-review" || in.HighRisk {
		out.Decision = DecisionEscalate
		out.Issue = IssuePolicy
		out.Reason = "Artifact release requires human escalation."
		out.SettlementHint = SettlementReview
		return out
	}

	if in.Exportability.State == "blocked" {
		if in.OverrideRequested {
			out.Decision = DecisionEscalate
			out.Issue = IssuePolicy
			out.Reason = "Blocked artifact override requires human approval."
			out.SettlementHint = SettlementReview
			return out
		}
		out.Decision = DecisionReject
		out.Issue = IssuePolicy
		out.Fulfillment = FulfillmentNone
		out.Reason = "Artifact release blocked by exportability policy."
		return out
	}

	if in.ArtifactLabel != in.RequestedScope {
		out.Decision = DecisionRequestRevision
		out.Issue = IssueScopeMismatch
		out.Fulfillment = FulfillmentPartial
		out.Reason = "Submitted artifact does not match requested scope."
		out.SettlementHint = SettlementReview
		return out
	}

	out.Decision = DecisionApprove
	out.Reason = "Artifact release approved."
	out.Fulfillment = FulfillmentSubstantial
	out.FulfillmentRatio = 1.0
	out.SettlementHint = SettlementAutoRelease
	return out
}
```

- [ ] **Step 4: Run the targeted tests and make sure they pass**

Run:

```bash
go test ./internal/approvalflow/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the domain-model slice**

Run:

```bash
git add internal/approvalflow/types.go internal/approvalflow/release.go internal/approvalflow/release_test.go
git -c commit.gpgsign=false commit -m "feat: add approval flow domain model"
```

## Task 2: Add Artifact Release Approval Receipts

**Files:**
- Modify: `internal/ent/schema/audit_log.go`
- Modify: generated Ent files under `internal/ent/...`
- Modify: `internal/knowledge/store_test.go`
- Modify: `internal/app/tools_meta.go`
- Create: `internal/app/tools_meta_approvalflow_test.go`
- Modify: `internal/app/tools_parity_test.go`

- [ ] **Step 1: Write the failing tool and audit tests**

Create `internal/app/tools_meta_approvalflow_test.go`:

```go
package app

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildMetaTools_IncludesApproveArtifactRelease(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil)
	names := toolNamesUnsorted(tools)
	assert.Contains(t, names, "approve_artifact_release")
}

func TestApproveArtifactReleaseTool_EscalatesNeedsHumanReview(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	store := knowledge.NewStore(client, zap.NewNop().Sugar())
	ctx := context.Background()

	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, config.DefaultConfig())
	var tool *agent.Tool
	for _, tdef := range tools {
		if tdef.Name == "approve_artifact_release" {
			tool = tdef
			break
		}
	}
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"artifact_label": "sensitive memo",
		"requested_scope": "sensitive memo",
		"exportability_state": "needs-human-review",
	})
	require.NoError(t, err)
	payload := got.(map[string]interface{})
	assert.Equal(t, "escalate", payload["decision"])
}
```

Append to `internal/knowledge/store_test.go`:

```go
func TestSaveAuditLog_ArtifactReleaseApprovalAction(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.SaveAuditLog(ctx, AuditEntry{
		SessionKey: "sess-approval",
		Action:     "artifact_release_approval",
		Actor:      "agent",
		Target:     "artifact:research-memo",
		Details: map[string]interface{}{
			"decision": "approve",
		},
	})
	require.NoError(t, err)
}
```

- [ ] **Step 2: Run the new tests and confirm they fail**

Run:

```bash
go test ./internal/app/... ./internal/knowledge/... -run 'Test(BuildMetaTools_IncludesApproveArtifactRelease|ApproveArtifactReleaseTool_EscalatesNeedsHumanReview|SaveAuditLog_ArtifactReleaseApprovalAction)' -count=1
```

Expected:

```text
FAIL
```

because the new action and tool do not exist yet.

- [ ] **Step 3: Implement the audit action and meta tool**

Modify `internal/ent/schema/audit_log.go` enum values:

```go
"artifact_release_approval",
```

Add to `internal/app/tools_meta.go`:

```go
{
	Name:        "approve_artifact_release",
	Description: "Approve, reject, revise, or escalate an artifact release decision for knowledge exchange",
	SafetyLevel: agent.SafetyLevelModerate,
	Capability: agent.ToolCapability{
		Category: "knowledge",
		Activity: agent.ActivityWrite,
	},
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"artifact_label": map[string]interface{}{"type": "string"},
			"requested_scope": map[string]interface{}{"type": "string"},
			"exportability_state": map[string]interface{}{
				"type": "string",
				"enum": []string{"exportable", "blocked", "needs-human-review"},
			},
			"override_requested": map[string]interface{}{"type": "boolean"},
			"high_risk": map[string]interface{}{"type": "boolean"},
		},
		"required": []string{"artifact_label", "requested_scope", "exportability_state"},
	},
	Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		artifactLabel, err := toolparam.RequireString(params, "artifact_label")
		if err != nil {
			return nil, err
		}
		requestedScope, err := toolparam.RequireString(params, "requested_scope")
		if err != nil {
			return nil, err
		}
		state, err := toolparam.RequireString(params, "exportability_state")
		if err != nil {
			return nil, err
		}
		overrideRequested, _ := params["override_requested"].(bool)
		highRisk, _ := params["high_risk"].(bool)

		outcome := approvalflow.ApproveArtifactRelease(approvalflow.ArtifactReleaseInput{
			ArtifactLabel:     artifactLabel,
			RequestedScope:    requestedScope,
			Exportability:     exportability.Receipt{State: exportability.DecisionState(state)},
			OverrideRequested: overrideRequested,
			HighRisk:          highRisk,
		})

		_ = store.SaveAuditLog(ctx, knowledge.AuditEntry{
			SessionKey: session.SessionKeyFromContext(ctx),
			Action:     "artifact_release_approval",
			Actor:      "agent",
			Target:     "artifact:" + artifactLabel,
			Details: map[string]interface{}{
				"decision":          outcome.Decision,
				"reason":            outcome.Reason,
				"issue":             outcome.Issue,
				"fulfillment":       outcome.Fulfillment,
				"fulfillment_ratio": outcome.FulfillmentRatio,
				"settlement_hint":   outcome.SettlementHint,
			},
		})

		return map[string]interface{}{
			"artifact_label":    artifactLabel,
			"decision":          outcome.Decision,
			"reason":            outcome.Reason,
			"issue":             outcome.Issue,
			"fulfillment":       outcome.Fulfillment,
			"fulfillment_ratio": outcome.FulfillmentRatio,
			"settlement_hint":   outcome.SettlementHint,
		}, nil
	},
},
```

Modify `internal/app/tools_parity_test.go` meta-tool names to include:

```go
"approve_artifact_release",
```

- [ ] **Step 4: Regenerate Ent and rerun targeted tests**

Run:

```bash
go generate ./internal/ent/...
go test ./internal/app/... ./internal/knowledge/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the approval-receipt slice**

Run:

```bash
git add internal/ent/schema/audit_log.go internal/app/tools_meta.go internal/app/tools_meta_approvalflow_test.go internal/app/tools_parity_test.go internal/knowledge/store_test.go internal/ent
git -c commit.gpgsign=false commit -m "feat: add artifact release approval receipts"
```

## Task 3: Add Minimal Operator Surface And Docs

**Files:**
- Create: `docs/security/approval-flow.md`
- Modify: `docs/security/index.md`
- Modify: `docs/architecture/trust-security-policy-audit.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `README.md`
- Modify: `mkdocs.yml`

- [ ] **Step 1: Write the new operator doc**

Create `docs/security/approval-flow.md`:

```md
# Approval Flow

Lango's first approval-flow slice for `knowledge exchange v1` is centered on artifact release approval.

## Objects

- `upfront payment`
- `artifact release`

## Release Decisions

- `approve`
- `reject`
- `request-revision`
- `escalate`

## Current First Slice

This slice currently focuses on structured artifact release decisions and audit-backed approval receipts.

It does not yet include:

- human approval UI
- dispute orchestration
- partial settlement execution
```

- [ ] **Step 2: Link and truth-align surrounding docs**

Modify `docs/security/index.md` quick links:

```md
- [Approval Flow](approval-flow.md) -- Structured artifact release approval for knowledge exchange
```

Modify `docs/architecture/trust-security-policy-audit.md` under the approval row with post-implementation notes:

```md
### Post-Implementation Notes

- The first approval-flow slice now has explicit artifact release decision states and audit-backed approval receipts.
- Upfront payment approval, human escalation UI, and dispute orchestration remain follow-on work.
```

Modify `docs/architecture/p2p-knowledge-exchange-track.md` so approval flow is no longer entirely pending; distinguish landed artifact release approval from remaining follow-on items.

Modify `README.md` with one short truthful note that early knowledge exchange now has structured artifact release approval states layered on top of exportability.

Modify `mkdocs.yml`:

```yaml
  - Security:
    - security/index.md
    - Encryption & Secrets: security/encryption.md
    - PII Redaction: security/pii-redaction.md
    - Tool Approval: security/tool-approval.md
    - Exportability: security/exportability.md
    - Approval Flow: security/approval-flow.md
    - Authentication: security/authentication.md
```

- [ ] **Step 3: Run docs verification**

Run:

```bash
python3 -m mkdocs build --strict
```

Expected:

```text
Documentation built
```

with exit code `0`.

- [ ] **Step 4: Commit the docs slice**

Run:

```bash
git add docs/security/approval-flow.md docs/security/index.md docs/architecture/trust-security-policy-audit.md docs/architecture/p2p-knowledge-exchange-track.md README.md mkdocs.yml
git -c commit.gpgsign=false commit -m "docs: add approval flow operator surface"
```

## Task 4: Full Verification And OpenSpec Closeout

**Files:**
- Modify: `openspec/changes/approval-flow-first-slice/*` or create the change if missing
- Sync: `openspec/specs/*` as required by the implemented delta

- [ ] **Step 1: Verify the full repository**

Run:

```bash
go test ./...
go build ./...
python3 -m mkdocs build --strict
```

Expected:

```text
ok
```

with all commands exiting `0`.

- [ ] **Step 2: Create or refresh the OpenSpec change**

If no change exists yet, create one and make sure proposal/design/tasks/specs cover:

- approval-flow domain model
- artifact release approval tool
- audit-backed approval receipts
- approval-flow operator docs

Use:

```bash
$openspec-new-change
$openspec-ff-change
```

- [ ] **Step 3: Apply, sync, and archive**

Run:

```bash
$openspec-apply-change
$openspec-archive-change
```

If direct archive automation collides with already-synced specs, perform agent-driven sync first and then move the change into the dated archive path.

- [ ] **Step 4: Confirm a clean worktree**

Run:

```bash
git status --short
```

Expected:

```text
[no output]
```

- [ ] **Step 5: Commit OpenSpec closeout if needed**

Run:

```bash
git add openspec/specs openspec/changes/archive
git -c commit.gpgsign=false commit -m "specs: archive approval flow first slice"
```
