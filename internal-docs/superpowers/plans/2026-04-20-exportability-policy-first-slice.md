# Exportability Policy First Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the first runnable exportability-policy slice for `knowledge exchange v1`, including source tagging, artifact evaluation, and durable audit-style decision receipts.

**Architecture:** This slice intentionally avoids a full policy-rule DSL, human-approval UI, or provenance-bundle embedding. Instead, it introduces a source-primary exportability evaluator, persists source-class metadata on knowledge assets, adds one evaluation tool that emits exportability receipts into audit storage, and exposes the new policy surface through minimal operator-facing status and documentation.

**Tech Stack:** Go, Ent, Cobra CLI, MkDocs/Markdown docs, existing knowledge store (`internal/knowledge/*`), security config (`internal/config/*`), app meta tools (`internal/app/tools_meta.go`), audit log schema (`internal/ent/schema/audit_log.go`)

---

## Scope Split

The design doc covers a broader policy model than is reasonable for one implementation slice.

This first slice covers only:

- source-primary exportability evaluation,
- knowledge-asset source tagging,
- draft/final decision states,
- durable exportability receipts in audit storage,
- minimal operator inspection surface,
- truthful docs for the new policy model.

This slice does **not** implement:

- full user-authored policy-rule DSL,
- human override workflow UI,
- provenance-bundle embedding of exportability receipts,
- dispute-ready receipt unification across audit, provenance, and settlement,
- non-knowledge asset registries.

## OpenSpec Precondition

Before touching implementation code, create or refresh an OpenSpec change for this slice. Use a narrow change name such as `exportability-policy-first-slice`.

The implementation session must end with the repository's required OpenSpec workflow:

- `ff`
- `apply`
- `verify`
- `sync`
- `archive`

## File Map

- Create: `internal/exportability/types.go`
  - Core enums and receipt types for exportability evaluation.
- Create: `internal/exportability/evaluator.go`
  - Source-primary evaluator with mixed-source and metadata-conflict rules.
- Create: `internal/exportability/evaluator_test.go`
  - Unit tests for `exportable`, `blocked`, and `needs-human-review`.
- Modify: `internal/config/types_security.go`
  - Add `ExportabilityConfig` and related policy types under `SecurityConfig`.
- Modify: `internal/config/loader.go`
  - Set default exportability config in `DefaultConfig()`.
- Modify: `internal/ent/schema/knowledge.go`
  - Persist `source_class` and `asset_label`.
- Modify: `internal/ent/schema/audit_log.go`
  - Add `exportability_decision` audit action.
- Modify: generated Ent files under `internal/ent/...`
  - Regenerated after schema changes.
- Modify: `internal/knowledge/types.go`
  - Add source-class and asset-label fields to `KnowledgeEntry`.
- Modify: `internal/knowledge/store.go`
  - Persist/read new knowledge metadata and treat exportability-class changes as version-significant.
- Modify: `internal/knowledge/store_test.go`
  - Add tests for source tagging roundtrip and version bump on classification change.
- Modify: `internal/app/tools_meta.go`
  - Extend `save_knowledge` params with tagging fields and add `evaluate_exportability`.
- Create: `internal/app/tools_meta_exportability_test.go`
  - Cover tool definition and evaluation handler behavior with a real test store.
- Modify: `internal/app/tools_parity_test.go`
  - Add the new meta tool to parity expectations.
- Modify: `internal/cli/security/status.go`
  - Surface exportability-policy status in `lango security status`.
- Modify: `internal/cli/security/security_test.go`
  - Assert the security command still exposes the expected structure.
- Create: `docs/security/exportability.md`
  - Canonical operator doc for source classes, decision flow, and receipts.
- Modify: `docs/security/index.md`
  - Link the new exportability doc from the security landing page.
- Modify: `docs/architecture/trust-security-policy-audit.md`
  - Add post-implementation notes once this slice lands.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark exportability policy as landed for the first slice and point to the operator doc.
- Modify: `README.md`
  - Add a short, truthful note about exportability evaluation for early knowledge exchange.
- Modify: `mkdocs.yml`
  - Add the new security doc to nav.

## Task 1: Introduce The Core Exportability Evaluator

**Files:**
- Create: `internal/exportability/types.go`
- Create: `internal/exportability/evaluator.go`
- Create: `internal/exportability/evaluator_test.go`
- Modify: `internal/config/types_security.go`
- Modify: `internal/config/loader.go`

- [ ] **Step 1: Write the failing evaluator tests**

Create `internal/exportability/evaluator_test.go`:

```go
package exportability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluate_PublicAndUserExportableSourcesAllowExport(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageFinal, []SourceRef{
		{AssetID: "pub-1", AssetLabel: "docs/api", Class: ClassPublic},
		{AssetID: "usr-1", AssetLabel: "user/wiki", Class: ClassUserExportable},
	})

	assert.Equal(t, StateExportable, receipt.State)
	assert.Equal(t, "allowed_user_exportable", receipt.PolicyCode)
}

func TestEvaluate_PrivateSourceBlocksExport(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageFinal, []SourceRef{
		{AssetID: "usr-1", AssetLabel: "user/wiki", Class: ClassUserExportable},
		{AssetID: "priv-1", AssetLabel: "private/chat", Class: ClassPrivateConfidential},
	})

	assert.Equal(t, StateBlocked, receipt.State)
	assert.Equal(t, "blocked_private_source", receipt.PolicyCode)
}

func TestEvaluate_MissingSourceMetadataRequiresHumanReview(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageDraft, []SourceRef{
		{AssetID: "unknown-1", AssetLabel: "mystery", Class: ""},
	})

	assert.Equal(t, StateNeedsHumanReview, receipt.State)
	assert.Equal(t, "review_metadata_conflict", receipt.PolicyCode)
}
```

- [ ] **Step 2: Run the new evaluator tests and confirm they fail**

Run:

```bash
go test ./internal/exportability/... -count=1
```

Expected:

```text
FAIL
```

with undefined symbol errors for `Policy`, `StageFinal`, `Evaluate`, and the source/decision enums.

- [ ] **Step 3: Implement the types and evaluator**

Create `internal/exportability/types.go`:

```go
package exportability

type SourceClass string

const (
	ClassPublic              SourceClass = "public"
	ClassUserExportable      SourceClass = "user-exportable"
	ClassPrivateConfidential SourceClass = "private-confidential"
)

type DecisionStage string

const (
	StageDraft DecisionStage = "draft"
	StageFinal DecisionStage = "final"
)

type DecisionState string

const (
	StateExportable       DecisionState = "exportable"
	StateBlocked          DecisionState = "blocked"
	StateNeedsHumanReview DecisionState = "needs-human-review"
)

type Policy struct {
	Enabled bool
}

type SourceRef struct {
	AssetID    string
	AssetLabel string
	Class      SourceClass
}

type LineageSummary struct {
	AssetID    string      `json:"asset_id"`
	AssetLabel string      `json:"asset_label"`
	Class      SourceClass `json:"class"`
	Rule       string      `json:"rule"`
}

type Receipt struct {
	Stage       DecisionStage   `json:"stage"`
	State       DecisionState   `json:"state"`
	PolicyCode  string          `json:"policy_code"`
	Explanation string          `json:"explanation"`
	Lineage     []LineageSummary `json:"lineage"`
}
```

Create `internal/exportability/evaluator.go`:

```go
package exportability

func Evaluate(policy Policy, stage DecisionStage, refs []SourceRef) Receipt {
	lineage := make([]LineageSummary, 0, len(refs))
	if !policy.Enabled {
		return Receipt{
			Stage:       stage,
			State:       StateNeedsHumanReview,
			PolicyCode:  "review_policy_disabled",
			Explanation: "Exportability policy is disabled.",
			Lineage:     lineage,
		}
	}

	hasUserExportable := false
	for _, ref := range refs {
		rule := "source_class_ok"
		switch ref.Class {
		case "":
			rule = "metadata_missing"
		case ClassPrivateConfidential:
			rule = "highest_sensitivity_wins"
		case ClassUserExportable:
			hasUserExportable = true
		}
		lineage = append(lineage, LineageSummary{
			AssetID:    ref.AssetID,
			AssetLabel: ref.AssetLabel,
			Class:      ref.Class,
			Rule:       rule,
		})
	}

	for _, ref := range refs {
		if ref.Class == "" {
			return Receipt{
				Stage:       stage,
				State:       StateNeedsHumanReview,
				PolicyCode:  "review_metadata_conflict",
				Explanation: "Source metadata is incomplete or conflicting.",
				Lineage:     lineage,
			}
		}
		if ref.Class == ClassPrivateConfidential {
			return Receipt{
				Stage:       stage,
				State:       StateBlocked,
				PolicyCode:  "blocked_private_source",
				Explanation: "Artifact includes a private-confidential source.",
				Lineage:     lineage,
			}
		}
	}

	code := "allowed_public_only"
	if hasUserExportable {
		code = "allowed_user_exportable"
	}
	return Receipt{
		Stage:       stage,
		State:       StateExportable,
		PolicyCode:  code,
		Explanation: "Artifact is exportable under source-based policy.",
		Lineage:     lineage,
	}
}
```

Modify `internal/config/types_security.go`:

```go
type ExportabilityConfig struct {
	Enabled bool `mapstructure:"enabled" json:"enabled"`
}

type SecurityConfig struct {
	Interceptor   InterceptorConfig   `mapstructure:"interceptor" json:"interceptor"`
	Exportability ExportabilityConfig `mapstructure:"exportability" json:"exportability"`
	Signer        SignerConfig        `mapstructure:"signer" json:"signer"`
	DBEncryption  DBEncryptionConfig  `mapstructure:"dbEncryption" json:"dbEncryption"`
	KMS           KMSConfig           `mapstructure:"kms" json:"kms"`
}
```

Modify `internal/config/loader.go` defaults:

```go
Security: SecurityConfig{
	Interceptor: InterceptorConfig{
		Enabled:        true,
		ApprovalPolicy: ApprovalPolicyDangerous,
		Presidio: PresidioConfig{
			URL:            "http://localhost:5002",
			ScoreThreshold: 0.7,
			Language:       "en",
		},
	},
	Exportability: ExportabilityConfig{
		Enabled: true,
	},
	// ...
},
```

- [ ] **Step 4: Run the targeted tests and make sure they pass**

Run:

```bash
go test ./internal/exportability/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the evaluator slice**

Run:

```bash
git add internal/exportability/types.go internal/exportability/evaluator.go internal/exportability/evaluator_test.go internal/config/types_security.go internal/config/loader.go
git -c commit.gpgsign=false commit -m "feat: add exportability evaluator core"
```

## Task 2: Persist Source Tagging On Knowledge Assets

**Files:**
- Modify: `internal/ent/schema/knowledge.go`
- Modify: `internal/knowledge/types.go`
- Modify: `internal/knowledge/store.go`
- Modify: `internal/knowledge/store_test.go`
- Modify: `internal/app/tools_meta.go`
- Modify: generated Ent files under `internal/ent/...`

- [ ] **Step 1: Write the failing knowledge-store tests**

Append to `internal/knowledge/store_test.go`:

```go
func TestSaveKnowledge_PersistsSourceClassAndAssetLabel(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.SaveKnowledge(ctx, "", KnowledgeEntry{
		Key:         "exportable-note",
		Category:    "fact",
		Content:     "Reusable public summary",
		Source:      "user/wiki",
		SourceClass: "user-exportable",
		AssetLabel:  "User Wiki",
	})
	require.NoError(t, err)

	got, err := store.GetKnowledge(ctx, "exportable-note")
	require.NoError(t, err)
	require.Equal(t, "user-exportable", got.SourceClass)
	require.Equal(t, "User Wiki", got.AssetLabel)
}

func TestSaveKnowledge_SourceClassificationChangeCreatesNewVersion(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entry := KnowledgeEntry{
		Key:         "source-policy",
		Category:    "fact",
		Content:     "same content",
		Source:      "chat",
		SourceClass: "private-confidential",
		AssetLabel:  "Private Chat",
	}
	require.NoError(t, store.SaveKnowledge(ctx, "", entry))

	entry.SourceClass = "user-exportable"
	entry.AssetLabel = "Approved Summary"
	require.NoError(t, store.SaveKnowledge(ctx, "", entry))

	got, err := store.GetKnowledge(ctx, "source-policy")
	require.NoError(t, err)
	require.Equal(t, 2, got.Version)
	require.Equal(t, "user-exportable", got.SourceClass)
	require.Equal(t, "Approved Summary", got.AssetLabel)
}
```

- [ ] **Step 2: Run the new knowledge-store tests and confirm they fail**

Run:

```bash
go test ./internal/knowledge/... -run 'TestSaveKnowledge_(PersistsSourceClassAndAssetLabel|SourceClassificationChangeCreatesNewVersion)' -count=1
```

Expected:

```text
FAIL
```

with unknown-field errors for `SourceClass` and `AssetLabel`.

- [ ] **Step 3: Extend the schema, domain types, and knowledge-save surface**

Modify `internal/ent/schema/knowledge.go`:

```go
field.Enum("source_class").
	Values("public", "user-exportable", "private-confidential").
	Default("private-confidential"),
field.String("asset_label").
	Optional(),
```

Modify `internal/knowledge/types.go`:

```go
type KnowledgeEntry struct {
	Key         string
	Category    entknowledge.Category
	Content     string
	Tags        []string
	Source      string
	SourceClass string
	AssetLabel  string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```

Modify `internal/knowledge/store.go`:

```go
if existing.Category == entry.Category &&
	existing.Content == entry.Content &&
	existing.Source == entry.Source &&
	existing.SourceClass == entry.SourceClass &&
	existing.AssetLabel == entry.AssetLabel {
	return nil
}
```

and on create/update:

```go
if entry.Source != "" {
	builder.SetSource(entry.Source)
}
if entry.SourceClass != "" {
	builder.SetSourceClass(entry.SourceClass)
}
if entry.AssetLabel != "" {
	builder.SetAssetLabel(entry.AssetLabel)
}
```

and on reads:

```go
SourceClass: k.SourceClass,
AssetLabel:  k.AssetLabel,
```

Modify `internal/app/tools_meta.go` save_knowledge parameters:

```go
"source_class": map[string]interface{}{
	"type": "string",
	"enum": []string{"public", "user-exportable", "private-confidential"},
	"description": "Exportability source classification for this knowledge asset",
},
"asset_label": map[string]interface{}{
	"type": "string",
	"description": "Human-readable label for source-lineage and receipts",
},
```

and populate the entry:

```go
sourceClass := toolparam.OptionalString(params, "source_class", "private-confidential")
assetLabel := toolparam.OptionalString(params, "asset_label", key)

entry := knowledge.KnowledgeEntry{
	Key:         key,
	Category:    cat,
	Content:     content,
	Tags:        tags,
	Source:      source,
	SourceClass: sourceClass,
	AssetLabel:  assetLabel,
}
```

- [ ] **Step 4: Regenerate Ent and rerun targeted tests**

Run:

```bash
go generate ./internal/ent/...
go test ./internal/knowledge/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the source-tagging slice**

Run:

```bash
git add internal/ent/schema/knowledge.go internal/knowledge/types.go internal/knowledge/store.go internal/knowledge/store_test.go internal/app/tools_meta.go internal/ent
git -c commit.gpgsign=false commit -m "feat: persist knowledge exportability tags"
```

## Task 3: Add Artifact Evaluation And Audit Receipts

**Files:**
- Modify: `internal/ent/schema/audit_log.go`
- Modify: generated Ent files under `internal/ent/...`
- Modify: `internal/knowledge/store_test.go`
- Modify: `internal/knowledge/store.go`
- Modify: `internal/app/tools_meta.go`
- Create: `internal/app/tools_meta_exportability_test.go`
- Modify: `internal/app/tools_parity_test.go`

- [ ] **Step 1: Write the failing tool and audit tests**

Create `internal/app/tools_meta_exportability_test.go`:

```go
package app

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildMetaTools_IncludesEvaluateExportability(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil)
	names := toolNamesUnsorted(tools)
	assert.Contains(t, names, "evaluate_exportability")
}

func TestEvaluateExportabilityTool_BlocksPrivateKnowledgeSource(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	store := knowledge.NewStore(client, zap.NewNop().Sugar())
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "", knowledge.KnowledgeEntry{
		Key:         "private-note",
		Category:    "fact",
		Content:     "private source",
		Source:      "chat",
		SourceClass: "private-confidential",
		AssetLabel:  "Private Chat",
	}))

	var evalTool *agent.Tool
	for _, tool := range buildMetaTools(store, nil, nil, config.SkillConfig{}, nil) {
		if tool.Name == "evaluate_exportability" {
			evalTool = tool
			break
		}
	}
	require.NotNil(t, evalTool)

	got, err := evalTool.Handler(ctx, map[string]interface{}{
		"artifact_label": "draft memo",
		"source_keys":    []string{"private-note"},
		"stage":          "final",
	})
	require.NoError(t, err)
	payload := got.(map[string]interface{})
	assert.Equal(t, "blocked", payload["state"])
	assert.Equal(t, "blocked_private_source", payload["policy_code"])
}
```

Append to `internal/knowledge/store_test.go`:

```go
func TestSaveAuditLog_ExportabilityDecisionAction(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.SaveAuditLog(ctx, AuditEntry{
		SessionKey: "sess-export",
		Action:     "exportability_decision",
		Actor:      "agent",
		Target:     "artifact:draft-memo",
		Details: map[string]interface{}{
			"state":       "blocked",
			"policy_code": "blocked_private_source",
		},
	})
	require.NoError(t, err)
}
```

- [ ] **Step 2: Run the new tests and confirm they fail**

Run:

```bash
go test ./internal/app/... ./internal/knowledge/... -run 'Test(BuildMetaTools_IncludesEvaluateExportability|EvaluateExportabilityTool_BlocksPrivateKnowledgeSource|SaveAuditLog_ExportabilityDecisionAction)' -count=1
```

Expected:

```text
FAIL
```

because the new tool and audit action do not exist yet.

- [ ] **Step 3: Add the audit action, batch lookup helper, and evaluation tool**

Modify `internal/ent/schema/audit_log.go`:

```go
"exportability_decision",
```

Add to `internal/knowledge/store.go`:

```go
func (s *Store) GetKnowledgeBatch(ctx context.Context, keys []string) ([]KnowledgeEntry, error) {
	out := make([]KnowledgeEntry, 0, len(keys))
	for _, key := range keys {
		entry, err := s.GetKnowledge(ctx, key)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			out = append(out, *entry)
		}
	}
	return out, nil
}
```

Add to `internal/app/tools_meta.go`:

```go
{
	Name:        "evaluate_exportability",
	Description: "Evaluate whether an artifact is exportable from the source lineage of knowledge assets",
	SafetyLevel: agent.SafetyLevelModerate,
	Capability: agent.ToolCapability{
		Category: "knowledge",
		Activity: agent.ActivityQuery,
	},
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"artifact_label": map[string]interface{}{"type": "string"},
			"source_keys": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"stage": map[string]interface{}{
				"type": "string",
				"enum": []string{"draft", "final"},
			},
		},
		"required": []string{"artifact_label", "source_keys", "stage"},
	},
	Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		artifactLabel, err := toolparam.RequireString(params, "artifact_label")
		if err != nil {
			return nil, err
		}
		sourceKeys := toolparam.StringSlice(params, "source_keys")
		stage, err := toolparam.RequireString(params, "stage")
		if err != nil {
			return nil, err
		}

		entries, err := store.GetKnowledgeBatch(ctx, sourceKeys)
		if err != nil {
			return nil, fmt.Errorf("load source knowledge: %w", err)
		}
		refs := make([]exportability.SourceRef, 0, len(entries))
		for _, entry := range entries {
			refs = append(refs, exportability.SourceRef{
				AssetID:    entry.Key,
				AssetLabel: entry.AssetLabel,
				Class:      exportability.SourceClass(entry.SourceClass),
			})
		}

		receipt := exportability.Evaluate(
			exportability.Policy{Enabled: cfg.Security.Exportability.Enabled},
			exportability.DecisionStage(stage),
			refs,
		)

		_ = store.SaveAuditLog(ctx, knowledge.AuditEntry{
			SessionKey: session.SessionKeyFromContext(ctx),
			Action:     "exportability_decision",
			Actor:      "agent",
			Target:     "artifact:" + artifactLabel,
			Details: map[string]interface{}{
				"stage":       receipt.Stage,
				"state":       receipt.State,
				"policy_code": receipt.PolicyCode,
				"explanation": receipt.Explanation,
				"lineage":     receipt.Lineage,
			},
		})

		return map[string]interface{}{
			"artifact_label": artifactLabel,
			"stage":          receipt.Stage,
			"state":          receipt.State,
			"policy_code":    receipt.PolicyCode,
			"explanation":    receipt.Explanation,
			"lineage":        receipt.Lineage,
		}, nil
	},
},
```

Modify `internal/app/tools_parity_test.go` expected names to include:

```go
"evaluate_exportability",
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

- [ ] **Step 5: Commit the evaluation-tool slice**

Run:

```bash
git add internal/ent/schema/audit_log.go internal/knowledge/store.go internal/knowledge/store_test.go internal/app/tools_meta.go internal/app/tools_meta_exportability_test.go internal/app/tools_parity_test.go internal/ent
git -c commit.gpgsign=false commit -m "feat: add exportability evaluation receipts"
```

## Task 4: Add Minimal Operator Surface And Truthful Documentation

**Files:**
- Modify: `internal/cli/security/status.go`
- Modify: `internal/cli/security/security_test.go`
- Create: `docs/security/exportability.md`
- Modify: `docs/security/index.md`
- Modify: `docs/architecture/trust-security-policy-audit.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `README.md`
- Modify: `mkdocs.yml`

- [ ] **Step 1: Write the failing security-status test**

Append to `internal/cli/security/security_test.go`:

```go
func TestRenderStatus_IncludesExportabilityPolicy(t *testing.T) {
	out := statusOutput{
		SignerProvider: "local",
		ApprovalPolicy: "dangerous",
		Interceptor:    "enabled",
		PIIRedaction:   "enabled",
		DBEncryption:   "disabled",
		DBAvailable:    false,
		ExportabilityEnabled: true,
	}

	buf := new(bytes.Buffer)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	require.NoError(t, renderStatus(out, false))
	require.NoError(t, w.Close())
	_, _ = io.Copy(buf, r)
	assert.Contains(t, buf.String(), "Exportability")
}
```

- [ ] **Step 2: Run the CLI test and confirm it fails**

Run:

```bash
go test ./internal/cli/security/... -run 'TestRenderStatus_IncludesExportabilityPolicy' -count=1
```

Expected:

```text
FAIL
```

because `statusOutput` does not yet include exportability fields.

- [ ] **Step 3: Implement status output and docs**

Modify `internal/cli/security/status.go`:

```go
type statusOutput struct {
	SignerProvider        string                `json:"signer_provider"`
	EncryptionKeys        int                   `json:"encryption_keys"`
	StoredSecrets         int                   `json:"stored_secrets"`
	Interceptor           string                `json:"interceptor"`
	PIIRedaction          string                `json:"pii_redaction"`
	ApprovalPolicy        string                `json:"approval_policy"`
	ExportabilityEnabled  bool                  `json:"exportability_enabled"`
	DBEncryption          string                `json:"db_encryption"`
	Envelope              envelopeSection       `json:"envelope"`
	IdentityBundle        identityBundleSection `json:"identity_bundle"`
	DBAvailable           bool                  `json:"db_available"`
	// ...
}
```

and in `renderStatus`:

```go
fmt.Printf("  Approval Policy:    %s\n", s.ApprovalPolicy)
fmt.Printf("  Exportability:      %v\n", s.ExportabilityEnabled)
```

Create `docs/security/exportability.md` with:

```md
# Exportability

Lango evaluates exportability at the artifact level for early knowledge exchange.

## Source Classes

- `public`
- `user-exportable`
- `private-confidential`

## Decision States

- `exportable`
- `blocked`
- `needs-human-review`

## Current First Slice

This first slice is source-primary:

- mixed artifacts use highest sensitivity wins,
- private-confidential sources block export by default,
- receipts are written as `exportability_decision` audit entries.
```

Modify `docs/security/index.md` quick links:

```md
- [Exportability](exportability.md) -- Source classes, artifact evaluation, and decision receipts
```

Modify `docs/architecture/trust-security-policy-audit.md` post-implementation notes under the privacy row:

```md
### Post-Implementation Notes

- The first slice now has explicit source classes, artifact-level evaluation, and audit-backed exportability receipts.
- Policy-rule DSL, human override UI, and dispute-ready receipt unification remain follow-on work.
```

Modify `docs/architecture/p2p-knowledge-exchange-track.md` follow-on section so `exportability policy` is no longer listed as unstarted, and instead references the first landed slice plus the remaining follow-on gaps.

Modify `README.md` in the security / knowledge-exchange section with one short truthful paragraph explaining that early external deliverables can be evaluated through source-based exportability rules.

Modify `mkdocs.yml`:

```yaml
  - Security:
    - security/index.md
    - Encryption & Secrets: security/encryption.md
    - PII Redaction: security/pii-redaction.md
    - Tool Approval: security/tool-approval.md
    - Exportability: security/exportability.md
    - Authentication: security/authentication.md
```

- [ ] **Step 4: Run documentation and CLI verification**

Run:

```bash
go test ./internal/cli/security/... -count=1
python3 -m mkdocs build --strict
```

Expected:

```text
ok
```

and MkDocs exits `0`.

- [ ] **Step 5: Commit the operator-surface slice**

Run:

```bash
git add internal/cli/security/status.go internal/cli/security/security_test.go docs/security/exportability.md docs/security/index.md docs/architecture/trust-security-policy-audit.md docs/architecture/p2p-knowledge-exchange-track.md README.md mkdocs.yml
git -c commit.gpgsign=false commit -m "docs: surface exportability policy"
```

## Task 5: Full Verification And OpenSpec Closeout

**Files:**
- Modify: `openspec/changes/exportability-policy-first-slice/*` or create the change if missing
- Sync: `openspec/specs/*` as required by the implemented delta

- [ ] **Step 1: Verify the full repository build and test suite**

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

- [ ] **Step 2: Create or refresh the OpenSpec change artifacts**

If no change exists yet, create one for this slice and make sure proposal/design/tasks/specs reflect:

- source classes on knowledge assets,
- source-primary evaluator,
- exportability decision receipts,
- security-status/docs surface.

Use the repository's required workflow:

```bash
$openspec-new-change
$openspec-ff-change
```

- [ ] **Step 3: Apply, verify, sync, and archive the change**

Run:

```bash
$openspec-apply-change
$openspec-archive-change
```

Choose sync when the final implementation matches the delta specs and main-spec merge is safe.

- [ ] **Step 4: Confirm the worktree is clean**

Run:

```bash
git status --short
```

Expected:

```text
[no output]
```

- [ ] **Step 5: Commit the OpenSpec closeout if the archive/sync step changed tracked files**

Run:

```bash
git add openspec/specs openspec/changes/archive
git -c commit.gpgsign=false commit -m "specs: archive exportability policy first slice"
```
