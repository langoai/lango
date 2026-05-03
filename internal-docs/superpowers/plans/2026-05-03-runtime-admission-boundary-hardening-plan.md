# Runtime Admission Boundary Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an observe/enforce admission boundary for dynamic runtime graph triple producers so unknown predicates no longer reach the graph store unchecked and no longer roll back valid triples in the same batch.

**Architecture:** This plan implements only Change A from the adaptive ontology growth design. The runtime keeps fast paths for deterministic seeded-predicate emitters, and adds a single `graph.AdmissionPolicy` in front of dynamic event-driven producers. The policy runs in `observe` mode by default, emits decision telemetry, and can be switched to `enforce` mode to filter unknown predicates before they reach `GraphBuffer`.

**Tech Stack:** Go, Ent-backed ontology service, BoltDB graph store, Cobra CLI settings, OpenSpec experimental workflow

---

## Scope

This plan intentionally covers only the **runtime app path**:

- `internal/app/wiring_graph.go` dynamic `TriplesExtractedEvent` subscriber path
- `internal/graph/*` admission policy core
- ontology governance config needed to select `observe` vs `enforce`
- runtime settings and public docs for the new config

This plan does **not** cover:

- CLI `internal/cli/graph/import_cmd.go`
- Change B adaptive shadow growth
- Change C promotion/replay hardening

Those should be separate plans because they require different rollback boundaries and, for CLI import, different bootstrap wiring.

## File Structure

### New Files

- `internal/graph/admission.go`
  Runtime admission primitives: mode enum, producer enum, admission candidate, default-confidence table, batch admission policy.
- `internal/graph/admission_test.go`
  Unit tests for observe mode, enforce mode, mixed batches, and producer confidence fallback.
- `internal/app/wiring_graph_test.go`
  App-level tests for the dynamic event subscriber helper that turns `TriplesExtractedEvent` into admitted graph triples.
- `openspec/changes/runtime-admission-boundary-hardening/proposal.md`
  OpenSpec change proposal for Change A runtime hardening slice.
- `openspec/changes/runtime-admission-boundary-hardening/design.md`
  OpenSpec design summary for the runtime-only scope.
- `openspec/changes/runtime-admission-boundary-hardening/tasks.md`
  Task checklist for the implementation slice.
- `openspec/changes/runtime-admission-boundary-hardening/specs/knowledge-graph/spec.md`
  Delta spec for graph admission behavior.
- `openspec/changes/runtime-admission-boundary-hardening/specs/ontology-registry/spec.md`
  Delta spec for validator-source-of-truth and observe/enforce behavior.
- `openspec/changes/runtime-admission-boundary-hardening/specs/cli-settings/spec.md`
  Delta spec for configuration and settings UI exposure.

### Modified Files

- `internal/config/types_ontology.go`
  Add governance admission settings for mode and producer fallback confidence.
- `internal/config/loader.go`
  Set default values for the new ontology governance admission settings.
- `internal/cli/settings/forms_ontology.go`
  Expose admission settings in TUI settings.
- `internal/cli/tuicore/state_update.go`
  Persist new settings fields from the form into config state.
- `internal/app/wiring_graph.go`
  Build a policy instance and route only dynamic event-driven triple producers through it.
- `docs/configuration.md`
  Document the new config keys and observe/enforce semantics.
- `docs/features/ontology.md`
  Document the runtime admission boundary and that Change A is observe-first.
- `docs/features/knowledge-graph.md`
  Document that dynamic triples are filtered before `GraphBuffer` in enforce mode.
- `README.md`
  Update the graph/ontology feature summary to mention admission mode.

### Existing Tests To Extend

- `internal/config/loader_test.go`
- `internal/cli/settings/forms_impl_test.go`
- `internal/graph/bolt_store_test.go`

## Task 1: Scaffold The OpenSpec Change

**Files:**
- Create: `openspec/changes/runtime-admission-boundary-hardening/proposal.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/design.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/tasks.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/specs/knowledge-graph/spec.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/specs/ontology-registry/spec.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/specs/cli-settings/spec.md`

- [ ] **Step 1: Create the change scaffold**

Run:

```bash
openspec new change "runtime-admission-boundary-hardening"
```

Expected: a new directory at `openspec/changes/runtime-admission-boundary-hardening/`

- [ ] **Step 2: Write the proposal**

Use this content for `openspec/changes/runtime-admission-boundary-hardening/proposal.md`:

```markdown
# Runtime Admission Boundary Hardening

## Why

Dynamic graph triples discovered by runtime LLM producers currently reach `GraphBuffer` and `BoltStore` without a single admission boundary. Unknown predicates can therefore abort valid batch writes and produce repeated `batch graph update error` logs.

## What Changes

- Add a runtime admission policy with `observe` and `enforce` modes.
- Route dynamic `TriplesExtractedEvent` paths through admission before `GraphBuffer`.
- Keep deterministic seeded-predicate emitters on a fast path.
- Expose admission mode and default producer confidence in ontology governance config and settings.

## Out Of Scope

- CLI `graph import`
- adaptive shadow growth and schema candidate persistence
- replay and promotion hardening
```

- [ ] **Step 3: Write the design**

Use this content for `openspec/changes/runtime-admission-boundary-hardening/design.md`:

```markdown
# Design

This change implements only the runtime admission boundary hardening slice from the internal design memo.

Dynamic producer path:

1. Convert `TriplesExtractedEvent` payloads into `graph.AdmissionCandidate` values.
2. Evaluate candidates with `graph.AdmissionPolicy`.
3. In `observe` mode, emit decision telemetry and pass all triples through unchanged.
4. In `enforce` mode, drop unknown predicates before they reach `GraphBuffer`.

Deterministic seeded-predicate producers such as content-saved containment triples and memory graph hooks remain on a fast path and continue to rely on the existing ontology and graph validators.
```

- [ ] **Step 4: Write the delta specs**

Use these minimal deltas:

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/knowledge-graph/spec.md -->
### Requirement: Dynamic triples pass through an admission boundary
Runtime triples emitted from `TriplesExtractedEvent` producers SHALL be evaluated by an admission policy before they are enqueued to `GraphBuffer`.

#### Scenario: Observe mode preserves writes
- **WHEN** admission mode is `observe`
- **THEN** the system SHALL emit admission telemetry and still enqueue the original triples

#### Scenario: Enforce mode filters unknown predicates
- **WHEN** admission mode is `enforce`
- **AND** a batch contains both valid and unknown predicates
- **THEN** the system SHALL enqueue only the valid triples
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/ontology-registry/spec.md -->
### Requirement: Admission and graph validation share one predicate source
The runtime SHALL use the ontology service predicate validator closure as the single predicate-validity source for dynamic triple admission and graph-store validation.
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/cli-settings/spec.md -->
### Requirement: Ontology governance exposes admission mode
The settings surface SHALL expose an ontology governance admission mode with values `observe` and `enforce`, plus fallback confidence defaults for dynamic producers.
```

- [ ] **Step 5: Record tasks and commit**

Use this starter checklist in `openspec/changes/runtime-admission-boundary-hardening/tasks.md`:

```markdown
# Tasks

- [ ] Add ontology governance admission config and defaults
- [ ] Add graph admission policy and tests
- [ ] Route dynamic app runtime producers through admission
- [ ] Update docs and settings
- [ ] Verify with targeted tests plus `go build ./...` and `go test ./...`
```

Run:

```bash
git add openspec/changes/runtime-admission-boundary-hardening
git commit -m "openspec: scaffold runtime admission boundary hardening"
```

Expected: commit succeeds with the new change scaffold

## Task 2: Add Ontology Governance Admission Config

**Files:**
- Modify: `internal/config/types_ontology.go`
- Modify: `internal/config/loader.go`
- Modify: `internal/cli/settings/forms_ontology.go`
- Modify: `internal/cli/tuicore/state_update.go`
- Test: `internal/config/loader_test.go`
- Test: `internal/cli/settings/forms_impl_test.go`

- [ ] **Step 1: Write failing config tests**

Add tests like these:

```go
func TestDefaultConfig_OntologyAdmissionDefaults(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "observe", cfg.Ontology.Governance.AdmissionMode)
	assert.InDelta(t, 0.60, cfg.Ontology.Governance.LearningDefaultConfidence, 0.001)
	assert.InDelta(t, 0.50, cfg.Ontology.Governance.LibrarianDefaultConfidence, 0.001)
}

func TestApplyForm_OntologyAdmissionFields(t *testing.T) {
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "ontology_gov_admission_mode", Type: tuicore.InputSelect, Value: "enforce"})
	form.AddField(&tuicore.Field{Key: "ontology_gov_learning_conf", Type: tuicore.InputText, Value: "0.65"})
	form.AddField(&tuicore.Field{Key: "ontology_gov_librarian_conf", Type: tuicore.InputText, Value: "0.55"})

	s := tuicore.NewConfigStateWith(config.DefaultConfig())
	s.ApplyForm(form)

	assert.Equal(t, "enforce", s.Current.Ontology.Governance.AdmissionMode)
	assert.InDelta(t, 0.65, s.Current.Ontology.Governance.LearningDefaultConfidence, 0.001)
	assert.InDelta(t, 0.55, s.Current.Ontology.Governance.LibrarianDefaultConfidence, 0.001)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/config ./internal/cli/settings -run 'OntologyAdmission|OntologyAdmissionFields'
```

Expected: FAIL with missing fields like `AdmissionMode` / unknown form keys

- [ ] **Step 3: Implement the config fields and defaults**

Apply these changes:

```go
// internal/config/types_ontology.go
type OntologyGovernanceConfig struct {
	Enabled                     bool    `mapstructure:"enabled" json:"enabled,omitempty"`
	MaxNewPerDay                int     `mapstructure:"maxNewPerDay" json:"maxNewPerDay,omitempty"`
	QuarantinePeriodHrs         int     `mapstructure:"quarantinePeriodHrs" json:"quarantinePeriodHrs,omitempty"`
	ShadowModeDurationHrs       int     `mapstructure:"shadowModeDurationHrs" json:"shadowModeDurationHrs,omitempty"`
	MinUsageForPromotion        int     `mapstructure:"minUsageForPromotion" json:"minUsageForPromotion,omitempty"`
	SchemaExplosionBudget       int     `mapstructure:"schemaExplosionBudget" json:"schemaExplosionBudget,omitempty"`
	AdmissionMode               string  `mapstructure:"admissionMode" json:"admissionMode,omitempty"`
	LearningDefaultConfidence   float64 `mapstructure:"learningDefaultConfidence" json:"learningDefaultConfidence,omitempty"`
	LibrarianDefaultConfidence  float64 `mapstructure:"librarianDefaultConfidence" json:"librarianDefaultConfidence,omitempty"`
}
```

```go
// internal/config/loader.go inside DefaultConfig()
Ontology: OntologyConfig{
	ACL: OntologyACLConfig{
		P2PPermission: "read",
	},
	Governance: OntologyGovernanceConfig{
		AdmissionMode:              "observe",
		LearningDefaultConfidence:  0.60,
		LibrarianDefaultConfidence: 0.50,
	},
},
```

```go
// internal/cli/settings/forms_ontology.go
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_admission_mode",
	Label:       "    Admission Mode",
	Type:        tuicore.InputSelect,
	Value:       cfg.Ontology.Governance.AdmissionMode,
	Options:     []string{"observe", "enforce"},
	Description: "Observe only or enforce filtering for dynamic runtime graph triples",
	VisibleWhen: isGovEnabled,
})
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_learning_conf",
	Label:       "    Learning Default Confidence",
	Type:        tuicore.InputText,
	Value:       fmt.Sprintf("%.2f", cfg.Ontology.Governance.LearningDefaultConfidence),
	Placeholder: "0.60",
	Description: "Fallback confidence for dynamic learning graph triples",
	VisibleWhen: isGovEnabled,
})
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_librarian_conf",
	Label:       "    Librarian Default Confidence",
	Type:        tuicore.InputText,
	Value:       fmt.Sprintf("%.2f", cfg.Ontology.Governance.LibrarianDefaultConfidence),
	Placeholder: "0.50",
	Description: "Fallback confidence for dynamic librarian graph triples",
	VisibleWhen: isGovEnabled,
})
```

```go
// internal/cli/tuicore/state_update.go
case "ontology_gov_admission_mode":
	s.Current.Ontology.Governance.AdmissionMode = f.Value
case "ontology_gov_learning_conf":
	if v, err := strconv.ParseFloat(strings.TrimSpace(f.Value), 64); err == nil {
		s.Current.Ontology.Governance.LearningDefaultConfidence = v
	}
case "ontology_gov_librarian_conf":
	if v, err := strconv.ParseFloat(strings.TrimSpace(f.Value), 64); err == nil {
		s.Current.Ontology.Governance.LibrarianDefaultConfidence = v
	}
```

- [ ] **Step 4: Run the tests again**

Run:

```bash
go test ./internal/config ./internal/cli/settings -run 'OntologyAdmission|OntologyAdmissionFields'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/config/types_ontology.go internal/config/loader.go internal/cli/settings/forms_ontology.go internal/cli/tuicore/state_update.go internal/config/loader_test.go internal/cli/settings/forms_impl_test.go
git commit -m "feat: add ontology admission config"
```

## Task 3: Implement The Admission Policy Core

**Files:**
- Create: `internal/graph/admission.go`
- Test: `internal/graph/admission_test.go`

- [ ] **Step 1: Write failing admission tests**

Use tests like these:

```go
func TestAdmissionPolicy_ObserveModePassesUnknown(t *testing.T) {
	p := NewAdmissionPolicy(AdmissionConfig{
		Mode:      AdmissionModeObserve,
		Validator: func(name string) bool { return name == CausedBy },
		DefaultConfidence: map[AdmissionProducer]float64{
			ProducerLearningEvent: 0.60,
		},
	}, zap.NewNop().Sugar())

	result := p.AdmitBatch([]AdmissionCandidate{{
		Triple:   Triple{Subject: "a", Predicate: "invented_rel", Object: "b"},
		Producer: ProducerLearningEvent,
		Source:   "learning",
	}})

	assert.Len(t, result.Admitted, 1)
	assert.Equal(t, DecisionObservedUnknown, result.Records[0].Decision)
}

func TestAdmissionPolicy_EnforceModeFiltersUnknown(t *testing.T) {
	p := NewAdmissionPolicy(AdmissionConfig{
		Mode:      AdmissionModeEnforce,
		Validator: func(name string) bool { return name == CausedBy },
		DefaultConfidence: map[AdmissionProducer]float64{
			ProducerLearningEvent: 0.60,
		},
	}, zap.NewNop().Sugar())

	result := p.AdmitBatch([]AdmissionCandidate{
		{Triple: Triple{Subject: "ok", Predicate: CausedBy, Object: "tool"}, Producer: ProducerLearningEvent, Source: "learning"},
		{Triple: Triple{Subject: "bad", Predicate: "invented_rel", Object: "tool"}, Producer: ProducerLearningEvent, Source: "learning"},
	})

	assert.Len(t, result.Admitted, 1)
	assert.Len(t, result.Rejected, 1)
	assert.Equal(t, "invented_rel", result.Rejected[0].Candidate.Triple.Predicate)
}

func TestAdmissionPolicy_DefaultConfidenceByProducer(t *testing.T) {
	p := NewAdmissionPolicy(AdmissionConfig{
		Mode:      AdmissionModeObserve,
		Validator: func(name string) bool { return true },
		DefaultConfidence: map[AdmissionProducer]float64{
			ProducerLearningEvent:  0.60,
			ProducerLibrarianEvent: 0.50,
		},
	}, zap.NewNop().Sugar())

	result := p.AdmitBatch([]AdmissionCandidate{{
		Triple:   Triple{Subject: "a", Predicate: CausedBy, Object: "b"},
		Producer: ProducerLibrarianEvent,
		Source:   "librarian",
	}})

	assert.InDelta(t, 0.50, result.Records[0].Confidence, 0.001)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/graph -run 'AdmissionPolicy'
```

Expected: FAIL because `NewAdmissionPolicy`, `AdmissionCandidate`, and related types do not exist

- [ ] **Step 3: Write the minimal admission policy**

Create `internal/graph/admission.go` with this shape:

```go
package graph

import "go.uber.org/zap"

type AdmissionMode string

const (
	AdmissionModeObserve AdmissionMode = "observe"
	AdmissionModeEnforce AdmissionMode = "enforce"
)

type AdmissionProducer string

const (
	ProducerLearningEvent  AdmissionProducer = "learning_event"
	ProducerLibrarianEvent AdmissionProducer = "librarian_event"
)

type AdmissionDecision string

const (
	DecisionAdmitted        AdmissionDecision = "admitted"
	DecisionObservedUnknown AdmissionDecision = "observed_unknown"
	DecisionRejectedUnknown AdmissionDecision = "rejected_unknown"
)

type AdmissionCandidate struct {
	Triple      Triple
	Producer    AdmissionProducer
	Source      string
	Confidence  *float64
	SessionKey  string
	TurnID      string
}

type AdmissionRecord struct {
	Candidate  AdmissionCandidate
	Decision   AdmissionDecision
	Confidence float64
}

type RejectedCandidate struct {
	Candidate AdmissionCandidate
	Reason    string
}

type AdmissionResult struct {
	Admitted []Triple
	Rejected []RejectedCandidate
	Records  []AdmissionRecord
}

type AdmissionConfig struct {
	Mode              AdmissionMode
	Validator         PredicateValidatorFunc
	DefaultConfidence map[AdmissionProducer]float64
}

type AdmissionPolicy struct {
	cfg    AdmissionConfig
	logger *zap.SugaredLogger
}

func NewAdmissionPolicy(cfg AdmissionConfig, logger *zap.SugaredLogger) *AdmissionPolicy {
	return &AdmissionPolicy{cfg: cfg, logger: logger}
}

func (p *AdmissionPolicy) AdmitBatch(candidates []AdmissionCandidate) AdmissionResult {
	var result AdmissionResult
	for _, c := range candidates {
		conf := p.defaultConfidence(c)
		known := p.cfg.Validator == nil || p.cfg.Validator(c.Triple.Predicate)
		switch {
		case known:
			result.Admitted = append(result.Admitted, c.Triple)
			result.Records = append(result.Records, AdmissionRecord{Candidate: c, Decision: DecisionAdmitted, Confidence: conf})
		case p.cfg.Mode == AdmissionModeObserve:
			result.Admitted = append(result.Admitted, c.Triple)
			result.Records = append(result.Records, AdmissionRecord{Candidate: c, Decision: DecisionObservedUnknown, Confidence: conf})
		default:
			result.Rejected = append(result.Rejected, RejectedCandidate{Candidate: c, Reason: "unknown predicate"})
			result.Records = append(result.Records, AdmissionRecord{Candidate: c, Decision: DecisionRejectedUnknown, Confidence: conf})
		}
	}
	return result
}

func (p *AdmissionPolicy) defaultConfidence(c AdmissionCandidate) float64 {
	if c.Confidence != nil {
		return *c.Confidence
	}
	if v, ok := p.cfg.DefaultConfidence[c.Producer]; ok {
		return v
	}
	return 0.50
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
go test ./internal/graph -run 'AdmissionPolicy'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/graph/admission.go internal/graph/admission_test.go
git commit -m "feat: add graph admission policy"
```

## Task 4: Route Dynamic Event Producers Through Admission

**Files:**
- Modify: `internal/app/wiring_graph.go`
- Test: `internal/app/wiring_graph_test.go`

- [ ] **Step 1: Write failing app wiring tests**

Add tests like these:

```go
func TestAdmitExtractedTriples_ObserveModePreservesUnknown(t *testing.T) {
	p := graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Mode:      graph.AdmissionModeObserve,
		Validator: func(name string) bool { return name == graph.CausedBy },
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerLearningEvent: 0.60,
		},
	}, zap.NewNop().Sugar())

	admitted := admitExtractedTriples(p, eventbus.TriplesExtractedEvent{
		Source: "learning",
		Triples: []eventbus.Triple{{
			Subject: "a", Predicate: "invented_rel", Object: "b",
		}},
	})

	require.Len(t, admitted, 1)
	assert.Equal(t, "invented_rel", admitted[0].Predicate)
}

func TestAdmitExtractedTriples_EnforceModeFiltersUnknown(t *testing.T) {
	p := graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Mode:      graph.AdmissionModeEnforce,
		Validator: func(name string) bool { return name == graph.CausedBy },
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerLearningEvent: 0.60,
		},
	}, zap.NewNop().Sugar())

	admitted := admitExtractedTriples(p, eventbus.TriplesExtractedEvent{
		Source: "learning",
		Triples: []eventbus.Triple{
			{Subject: "ok", Predicate: graph.CausedBy, Object: "tool"},
			{Subject: "bad", Predicate: "invented_rel", Object: "tool"},
		},
	})

	require.Len(t, admitted, 1)
	assert.Equal(t, graph.CausedBy, admitted[0].Predicate)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/app -run 'AdmitExtractedTriples'
```

Expected: FAIL because `admitExtractedTriples` does not exist

- [ ] **Step 3: Implement admission in `wiring_graph.go`**

Add a helper and use it only for dynamic subscriber paths:

```go
func admitExtractedTriples(policy *graph.AdmissionPolicy, evt eventbus.TriplesExtractedEvent) []graph.Triple {
	if policy == nil {
		graphTriples := make([]graph.Triple, len(evt.Triples))
		for i, t := range evt.Triples {
			graphTriples[i] = graph.Triple{
				Subject:     t.Subject,
				Predicate:   t.Predicate,
				Object:      t.Object,
				SubjectType: t.SubjectType,
				ObjectType:  t.ObjectType,
				Metadata:    t.Metadata,
			}
		}
		return graphTriples
	}

	candidates := make([]graph.AdmissionCandidate, 0, len(evt.Triples))
	for _, t := range evt.Triples {
		candidates = append(candidates, graph.AdmissionCandidate{
			Triple: graph.Triple{
				Subject:     t.Subject,
				Predicate:   t.Predicate,
				Object:      t.Object,
				SubjectType: t.SubjectType,
				ObjectType:  t.ObjectType,
				Metadata:    t.Metadata,
			},
			Producer: graph.ProducerLearningEvent,
			Source:   evt.Source,
		})
	}

	result := policy.AdmitBatch(candidates)
	return result.Admitted
}
```

And in `wireGraphCallbacks`:

```go
	var admissionPolicy *graph.AdmissionPolicy
	if ontologyValidator != nil && cfg.Ontology.Governance.AdmissionMode != "" {
		admissionPolicy = graph.NewAdmissionPolicy(graph.AdmissionConfig{
			Mode:      graph.AdmissionMode(cfg.Ontology.Governance.AdmissionMode),
			Validator: ontologyValidator,
			DefaultConfidence: map[graph.AdmissionProducer]float64{
				graph.ProducerLearningEvent:  cfg.Ontology.Governance.LearningDefaultConfidence,
				graph.ProducerLibrarianEvent: cfg.Ontology.Governance.LibrarianDefaultConfidence,
			},
		}, logger())
	}

	eventbus.SubscribeTyped(bus, func(evt eventbus.TriplesExtractedEvent) {
		admitted := admitExtractedTriples(admissionPolicy, evt)
		if len(admitted) == 0 {
			return
		}
		gc.buffer.Enqueue(graph.GraphRequest{Triples: admitted})
	})
```

- [ ] **Step 4: Run tests**

Run:

```bash
go test ./internal/app -run 'AdmitExtractedTriples'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/app/wiring_graph.go internal/app/wiring_graph_test.go
git commit -m "feat: admit dynamic graph events before enqueue"
```

## Task 5: Update Downstream Docs And Verify The Slice

**Files:**
- Modify: `README.md`
- Modify: `docs/configuration.md`
- Modify: `docs/features/ontology.md`
- Modify: `docs/features/knowledge-graph.md`

- [ ] **Step 1: Write doc assertions as failing grep checks**

Run:

```bash
rg -n "admissionMode|Learning Default Confidence|observe|enforce" README.md docs/configuration.md docs/features/ontology.md docs/features/knowledge-graph.md
```

Expected: either no matches or incomplete matches for the new runtime admission behavior

- [ ] **Step 2: Update the docs**

Apply content like this:

```md
<!-- docs/configuration.md -->
| `ontology.governance.admissionMode` | `string` | `observe` | Runtime admission mode for dynamic graph triples: `observe` or `enforce` |
| `ontology.governance.learningDefaultConfidence` | `float64` | `0.60` | Fallback confidence for dynamic learning graph triples |
| `ontology.governance.librarianDefaultConfidence` | `float64` | `0.50` | Fallback confidence for dynamic librarian graph triples |
```

```md
<!-- docs/features/knowledge-graph.md -->
Dynamic event-driven graph triples now pass through a runtime admission policy before they reach `GraphBuffer`. In `observe` mode the policy logs decisions without changing writes; in `enforce` mode unknown predicates are dropped before batch persistence.
```

```md
<!-- docs/features/ontology.md -->
The runtime admission boundary for dynamic graph triples is controlled by `ontology.governance.admissionMode`. This Change A slice is observe-first: `observe` preserves writes and emits telemetry, while `enforce` filters unknown predicates before graph batching.
```

```md
<!-- README.md -->
- Runtime graph admission modes (`observe` / `enforce`) for dynamic extracted triples
```

- [ ] **Step 3: Run project verification**

Run:

```bash
go build ./...
go test ./...
```

Expected: both commands exit `0`

- [ ] **Step 4: Verify and close the OpenSpec change**

Run:

```bash
openspec status --change "runtime-admission-boundary-hardening" --json
```

Expected: the change exists and tasks/spec artifacts are present

Then use the repo workflow:

```text
1. Mark completed items in openspec/changes/runtime-admission-boundary-hardening/tasks.md
2. Use openspec-verify-change for runtime-admission-boundary-hardening
3. Sync any delta specs that should land in main specs
4. Archive the change only after verification passes and docs are updated
```

- [ ] **Step 5: Commit**

Run:

```bash
git add README.md docs/configuration.md docs/features/ontology.md docs/features/knowledge-graph.md openspec/changes/runtime-admission-boundary-hardening
git commit -m "docs: document runtime graph admission boundary"
```

## Self-Review

Spec coverage for this plan:

- Dynamic runtime producer admission boundary: covered in Tasks 3 and 4.
- Observe vs enforce mode: covered in Tasks 2, 3, and 4.
- Single validator source of truth: covered in Task 4 by reusing the injected ontology validator closure.
- Batch pre-filtering while retaining atomic writes: covered in Tasks 3 and 4.
- Downstream docs and config surfaces: covered in Tasks 2 and 5.

Intentional gaps:

- CLI `graph import` is not in this plan.
- Adaptive shadow growth, schema candidate persistence, and replay hardening are not in this plan.

Placeholder scan:

- No placeholder markers remain.
- Each code-changing step includes concrete code.
- Commands are explicit and package-scoped before the final full-suite run.

Type consistency checks:

- `AdmissionModeObserve` / `AdmissionModeEnforce`
- `ProducerLearningEvent` / `ProducerLibrarianEvent`
- `AdmissionCandidate`, `AdmissionRecord`, `AdmissionResult`

## Execution Handoff

Plan complete and saved to `internal-docs/superpowers/plans/2026-05-03-runtime-admission-boundary-hardening-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
