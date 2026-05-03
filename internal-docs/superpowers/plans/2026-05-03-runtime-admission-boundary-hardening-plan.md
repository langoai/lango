# Runtime Admission Boundary Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an observe/enforce admission boundary for dynamic or untrusted runtime graph triple producers, including event-driven triples, `content.saved` extractor output, and `GraphEngine` direct writes, so unknown predicates no longer reach the graph store unchecked and no longer roll back valid triples in the same batch.

**Architecture:** This plan implements only Change A from the adaptive ontology growth design. The runtime keeps fast paths for deterministic seeded-predicate emitters, and adds a single `graph.AdmissionPolicy` in front of dynamic event-driven producers. The policy runs in `observe` mode by default, emits decision telemetry, and can be switched to `enforce` mode to filter unknown predicates before they reach `GraphBuffer`.

**Tech Stack:** Go, Ent-backed ontology service, BoltDB graph store, Cobra CLI settings, OpenSpec experimental workflow

---

## Scope

This plan intentionally covers only the **runtime app path**:

- `internal/app/wiring_graph.go` dynamic `TriplesExtractedEvent` subscriber path
- `internal/app/wiring_graph.go` `content.saved` extractor output path
- `internal/learning/graph_engine.go` direct graph writes when the event bus handoff is not used
- `internal/graph/*` admission policy core
- runtime observability event emission for admission decisions
- metrics collector and status surface updates for admission telemetry
- ontology governance config needed to select `observe` vs `enforce`
- runtime settings and public docs for the new config

This plan does **not** cover:

- CLI `internal/cli/graph/import_cmd.go`
- `AssertFact` / ontology tool / ontology action growth-enabled paths
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
- `internal/eventbus/observability_events_test.go`
  Event type tests for graph admission observability events.
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
  Build a policy instance, route dynamic event-driven triples through it, and publish observe/enforce telemetry.
- `internal/app/modules.go`
  Build the admission policy after ontology initialization and before knowledge initialization so both `GraphEngine` and graph wiring can share it.
- `internal/app/wiring_knowledge.go`
  Pass the prebuilt runtime admission policy into `GraphEngine`.
- `internal/graph/extractor.go`
  Add an extractor mode that allows parsed triples to reach admission without pre-dropping unknown predicates.
- `internal/eventbus/observability_events.go`
  Add graph admission observability event types.
- `internal/app/wiring_observability.go`
  Subscribe the metrics collector to graph admission observability events.
- `internal/observability/collector.go`
  Record graph admission batch counters in the central metrics collector.
- `internal/cli/cockpit/pages/status.go`
  Render graph admission metrics on the status page.
- `internal/learning/graph_engine.go`
  Route direct runtime graph writes through the shared admission policy when configured.
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
- `internal/graph/extractor_test.go`
- `internal/learning/graph_engine_test.go`

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
- Route `content.saved` extractor output through the same admission boundary.
- Publish admission observability events in `observe` and `enforce` modes.
- Keep deterministic seeded-predicate emitters on a fast path.
- Expose admission mode and default producer confidence in ontology governance config and settings.

## Out Of Scope

- CLI `graph import`
- `AssertFact`, ontology tools, ontology actions, and P2P fact assertion paths
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
2. Convert `content.saved` extractor triples into `graph.AdmissionCandidate` values without pre-dropping unknown predicates inside the extractor.
3. Route `GraphEngine` direct writes through the same admission policy when no event-bus publish path is used.
4. Evaluate candidates with `graph.AdmissionPolicy`.
5. In `observe` mode, emit admission observability events and pass all triples through unchanged.
6. In `enforce` mode, drop unknown predicates before they reach `GraphBuffer`.

Deterministic seeded-predicate producers such as content-saved containment triples and memory graph hooks remain on a fast path and continue to rely on the existing ontology and graph validators.
```

- [ ] **Step 4: Write the delta specs**

Use these minimal deltas:

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/knowledge-graph/spec.md -->
### Requirement: Dynamic triples pass through an admission boundary
Runtime triples emitted from `TriplesExtractedEvent` producers, `content.saved` extractor output, and `GraphEngine` direct-write paths SHALL be evaluated by an admission policy before they are enqueued to `GraphBuffer` or written directly to the graph store.

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
- [ ] Add graph admission policy, telemetry, and tests
- [ ] Route dynamic app runtime producers and extractor output through admission
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

func TestUpdateConfigFromForm_OntologyAdmissionFields(t *testing.T) {
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "ontology_gov_admission_mode", Type: tuicore.InputSelect, Value: "enforce"})
	form.AddField(&tuicore.Field{Key: "ontology_gov_learning_conf", Type: tuicore.InputText, Value: "0.65"})
	form.AddField(&tuicore.Field{Key: "ontology_gov_librarian_conf", Type: tuicore.InputText, Value: "0.55"})

	s := tuicore.NewConfigStateWith(config.DefaultConfig())
	s.UpdateConfigFromForm(&form)

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
admissionVisible := func() bool { return enabled.Checked }

form.AddField(&tuicore.Field{
	Key:         "ontology_gov_admission_mode",
	Label:       "    Admission Mode",
	Type:        tuicore.InputSelect,
	Value:       cfg.Ontology.Governance.AdmissionMode,
	Options:     []string{"observe", "enforce"},
	Description: "Observe only or enforce filtering for dynamic runtime graph triples",
	VisibleWhen: admissionVisible,
})
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_learning_conf",
	Label:       "    Learning Default Confidence",
	Type:        tuicore.InputText,
	Value:       fmt.Sprintf("%.2f", cfg.Ontology.Governance.LearningDefaultConfidence),
	Placeholder: "0.60",
	Description: "Fallback confidence for dynamic learning graph triples",
	VisibleWhen: admissionVisible,
})
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_librarian_conf",
	Label:       "    Librarian Default Confidence",
	Type:        tuicore.InputText,
	Value:       fmt.Sprintf("%.2f", cfg.Ontology.Governance.LibrarianDefaultConfidence),
	Placeholder: "0.50",
	Description: "Fallback confidence for dynamic librarian graph triples",
	VisibleWhen: admissionVisible,
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

## Task 3: Implement The Admission Policy Core, Observe Telemetry, And Concurrency-Safe Stats

**Files:**
- Create: `internal/graph/admission.go`
- Test: `internal/graph/admission_test.go`
- Modify: `internal/eventbus/observability_events.go`
- Test: `internal/eventbus/observability_events_test.go`

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

func TestAdmissionPolicy_StatsTrackObserveAndReject(t *testing.T) {
	p := NewAdmissionPolicy(AdmissionConfig{
		Mode:      AdmissionModeEnforce,
		Validator: func(name string) bool { return name == CausedBy },
		DefaultConfidence: map[AdmissionProducer]float64{
			ProducerLearningEvent: 0.60,
		},
		Observe: func(AdmissionBatchEvent) {},
	}, zap.NewNop().Sugar())

	p.AdmitBatch([]AdmissionCandidate{
		{Triple: Triple{Subject: "ok", Predicate: CausedBy, Object: "tool"}, Producer: ProducerLearningEvent, Source: "learning"},
		{Triple: Triple{Subject: "bad", Predicate: "invented_rel", Object: "tool"}, Producer: ProducerLearningEvent, Source: "learning"},
	})

	stats := p.Snapshot()
	assert.Equal(t, int64(1), stats.RejectedUnknown)
	assert.Equal(t, int64(1), stats.AdmittedKnown)
}

func TestGraphAdmissionBatchEvent_EventName(t *testing.T) {
	evt := eventbus.GraphAdmissionBatchEvent{}
	assert.Equal(t, eventbus.EventGraphAdmissionBatch, evt.EventName())
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

import (
	"sync/atomic"

	"go.uber.org/zap"
)

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

type AdmissionStats struct {
	AdmittedKnown   atomic.Int64
	ObservedUnknown atomic.Int64
	RejectedUnknown atomic.Int64
	TotalCandidates atomic.Int64
	TotalBatches    atomic.Int64
}

type AdmissionStatsSnapshot struct {
	AdmittedKnown   int64
	ObservedUnknown int64
	RejectedUnknown int64
	TotalCandidates int64
	TotalBatches    int64
}

type AdmissionBatchEvent struct {
	Mode            AdmissionMode
	Producer        AdmissionProducer
	Candidates      int
	Admitted        int
	Rejected        int
	ObservedUnknown int
}

type AdmissionConfig struct {
	Mode              AdmissionMode
	Validator         PredicateValidatorFunc
	DefaultConfidence map[AdmissionProducer]float64
	Observe           func(AdmissionBatchEvent)
}

type AdmissionPolicy struct {
	cfg    AdmissionConfig
	logger *zap.SugaredLogger
	stats  AdmissionStats
}

func NewAdmissionPolicy(cfg AdmissionConfig, logger *zap.SugaredLogger) *AdmissionPolicy {
	return &AdmissionPolicy{cfg: cfg, logger: logger}
}

func (p *AdmissionPolicy) AdmitBatch(candidates []AdmissionCandidate) AdmissionResult {
	var result AdmissionResult
	p.stats.TotalBatches.Add(1)
	p.stats.TotalCandidates.Add(int64(len(candidates)))
	observedUnknown := 0
	producer := ProducerLearningEvent
	if len(candidates) > 0 {
		producer = candidates[0].Producer
	}
	for _, c := range candidates {
		conf := p.defaultConfidence(c)
		known := p.cfg.Validator == nil || p.cfg.Validator(c.Triple.Predicate)
		switch {
		case known:
			result.Admitted = append(result.Admitted, c.Triple)
			result.Records = append(result.Records, AdmissionRecord{Candidate: c, Decision: DecisionAdmitted, Confidence: conf})
			p.stats.AdmittedKnown.Add(1)
		case p.cfg.Mode == AdmissionModeObserve:
			result.Admitted = append(result.Admitted, c.Triple)
			result.Records = append(result.Records, AdmissionRecord{Candidate: c, Decision: DecisionObservedUnknown, Confidence: conf})
			p.stats.ObservedUnknown.Add(1)
			observedUnknown++
		default:
			result.Rejected = append(result.Rejected, RejectedCandidate{Candidate: c, Reason: "unknown predicate"})
			result.Records = append(result.Records, AdmissionRecord{Candidate: c, Decision: DecisionRejectedUnknown, Confidence: conf})
			p.stats.RejectedUnknown.Add(1)
		}
	}
	event := AdmissionBatchEvent{
		Mode:            p.cfg.Mode,
		Producer:        producer,
		Candidates:      len(candidates),
		Admitted:        len(result.Admitted),
		Rejected:        len(result.Rejected),
		ObservedUnknown: observedUnknown,
	}
	p.logger.Infow("graph admission batch",
		"mode", p.cfg.Mode,
		"producer", producer,
		"candidates", len(candidates),
		"admitted", len(result.Admitted),
		"rejected", len(result.Rejected),
	)
	if p.cfg.Observe != nil {
		p.cfg.Observe(event)
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

func (p *AdmissionPolicy) Snapshot() AdmissionStatsSnapshot {
	return AdmissionStatsSnapshot{
		AdmittedKnown:   p.stats.AdmittedKnown.Load(),
		ObservedUnknown: p.stats.ObservedUnknown.Load(),
		RejectedUnknown: p.stats.RejectedUnknown.Load(),
		TotalCandidates: p.stats.TotalCandidates.Load(),
		TotalBatches:    p.stats.TotalBatches.Load(),
	}
}
```

Add this event type:

```go
// internal/eventbus/observability_events.go
const EventGraphAdmissionBatch = "graph.admission.batch"

type GraphAdmissionBatchEvent struct {
	Mode            string
	Producer        string
	Candidates      int
	Admitted        int
	Rejected        int
	ObservedUnknown int
}

func (e GraphAdmissionBatchEvent) EventName() string { return EventGraphAdmissionBatch }
```

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
go test ./internal/graph ./internal/eventbus -run 'AdmissionPolicy|GraphAdmissionBatchEvent'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/graph/admission.go internal/graph/admission_test.go internal/eventbus/observability_events.go internal/eventbus/observability_events_test.go
git commit -m "feat: add graph admission policy"
```

## Task 4: Route Dynamic Event Producers And Extractor Output Through Admission

**Files:**
- Modify: `internal/app/wiring_graph.go`
- Modify: `internal/app/modules.go`
- Modify: `internal/graph/extractor.go`
- Test: `internal/app/wiring_graph_test.go`
- Test: `internal/graph/extractor_test.go`

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

func TestProducerForExtractedEvent_UsesSourceSpecificDefaults(t *testing.T) {
	assert.Equal(t, graph.ProducerLearningEvent, producerForExtractedEvent("learning"))
	assert.Equal(t, graph.ProducerLibrarianEvent, producerForExtractedEvent("proactive_librarian"))
}

func TestExtractor_PassthroughModePreservesUnknownPredicate(t *testing.T) {
	logger := zap.NewNop().Sugar()
	e := NewExtractor(nil, logger, WithoutPredicateValidation())

	triples := e.parseResponse("a|invented_rel|b", "src")
	require.Len(t, triples, 1)
	assert.Equal(t, "invented_rel", triples[0].Predicate)
}

func TestExtractor_PassthroughPromptOmitsFixedAllowlist(t *testing.T) {
	logger := zap.NewNop().Sugar()
	e := NewExtractor(nil, logger, WithoutPredicateValidation())

	assert.NotContains(t, e.systemPrompt(), "Valid predicates:")
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/app -run 'AdmitExtractedTriples|ProducerForExtractedEvent'
go test ./internal/graph -run 'PassthroughModePreservesUnknownPredicate|PassthroughPromptOmitsFixedAllowlist'
```

Expected: FAIL because `admitExtractedTriples`, `producerForExtractedEvent`, `WithoutPredicateValidation`, and `systemPrompt()` do not exist yet

- [ ] **Step 3: Implement admission in `wiring_graph.go`**

Add helpers and use them for both dynamic event subscriber paths and `content.saved` extractor output:

```go
func producerForExtractedEvent(source string) graph.AdmissionProducer {
	switch strings.TrimSpace(source) {
	case "proactive_librarian":
		return graph.ProducerLibrarianEvent
	default:
		return graph.ProducerLearningEvent
	}
}

func publishAdmissionObservation(bus *eventbus.Bus, evt graph.AdmissionBatchEvent) {
	if bus == nil {
		return
	}
	bus.Publish(eventbus.GraphAdmissionBatchEvent{
		Mode:            string(evt.Mode),
		Producer:        string(evt.Producer),
		Candidates:      evt.Candidates,
		Admitted:        evt.Admitted,
		Rejected:        evt.Rejected,
		ObservedUnknown: evt.ObservedUnknown,
	})
}

func admitTriples(policy *graph.AdmissionPolicy, producer graph.AdmissionProducer, source string, triples []graph.Triple) []graph.Triple {
	if policy == nil {
		return triples
	}
	candidates := make([]graph.AdmissionCandidate, 0, len(triples))
	for _, triple := range triples {
		candidates = append(candidates, graph.AdmissionCandidate{
			Triple:   triple,
			Producer: producer,
			Source:   source,
		})
	}
	result := policy.AdmitBatch(candidates)
	return result.Admitted
}

func admitExtractedTriples(policy *graph.AdmissionPolicy, evt eventbus.TriplesExtractedEvent) []graph.Triple {
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
	return admitTriples(policy, producerForExtractedEvent(evt.Source), evt.Source, graphTriples)
}
```

Build the policy before `initKnowledge`, not inside `wireGraphCallbacks`, so both app wiring and other runtime producers can reuse it. Update `graphComponents` to carry the policy:

```go
// internal/app/wiring_graph.go
type graphComponents struct {
	store           graph.Store
	buffer          *graph.GraphBuffer
	ragService      *graph.GraphRAGService
	admissionPolicy *graph.AdmissionPolicy
}
```

```go
// internal/app/modules.go
if gc != nil && ontologyResult != nil && ontologyResult.Service != nil && cfg.Ontology.Enabled {
	gc.admissionPolicy = graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Mode:      graph.AdmissionMode(cfg.Ontology.Governance.AdmissionMode),
		Validator: ontologyResult.Service.PredicateValidator(),
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerLearningEvent:  cfg.Ontology.Governance.LearningDefaultConfidence,
			graph.ProducerLibrarianEvent: cfg.Ontology.Governance.LibrarianDefaultConfidence,
		},
		Observe: func(evt graph.AdmissionBatchEvent) {
			publishAdmissionObservation(m.bus, evt)
		},
	}, logger())
}
```

In `wireGraphCallbacks`:

```go
	eventbus.SubscribeTyped(bus, func(evt eventbus.ContentSavedEvent) {
		if !evt.NeedsGraph {
			return
		}
		gc.buffer.Enqueue(graph.GraphRequest{Triples: []graph.Triple{{
			Subject:     evt.Collection + ":" + evt.ID,
			SubjectType: evt.Collection,
			Predicate:   graph.Contains,
			Object:      "collection:" + evt.Collection,
			Metadata:    evt.Metadata,
		}}})

		if extractor != nil && evt.Content != "" {
			go func() {
				ctx := context.Background()
				triples, err := extractor.Extract(ctx, evt.Content, evt.ID)
				if err != nil {
					logger().Debugw("entity extraction error", "id", evt.ID, "error", err)
					return
				}
				admitted := admitTriples(gc.admissionPolicy, graph.ProducerLearningEvent, "content_saved_extractor", triples)
				if len(admitted) > 0 {
					gc.buffer.Enqueue(graph.GraphRequest{Triples: admitted})
				}
			}()
		}
	})

	eventbus.SubscribeTyped(bus, func(evt eventbus.TriplesExtractedEvent) {
		admitted := admitExtractedTriples(gc.admissionPolicy, evt)
		if len(admitted) == 0 {
			return
		}
		gc.buffer.Enqueue(graph.GraphRequest{Triples: admitted})
	})
```

And change extractor setup so unknown predicates reach admission instead of being dropped inside `Extractor`:

```go
// internal/graph/extractor.go
type Extractor struct {
	generator          llm.TextGenerator
	validator          PredicateValidatorFunc
	skipPredicateCheck bool
	logger             *zap.SugaredLogger
}

func WithoutPredicateValidation() ExtractorOption {
	return func(e *Extractor) { e.skipPredicateCheck = true }
}

func (e *Extractor) systemPrompt() string {
	if e.skipPredicateCheck {
		return `You are an entity and relationship extraction system. Given text, extract entities and relationships as triples.

Output format (one triple per line):
SUBJECT|PREDICATE|OBJECT

Rules:
- Extract factual relationships only
- Use concise entity names and concise snake_case predicates
- Skip trivial or obvious relationships
- Maximum 10 triples per extraction
- If no meaningful relationships found, output NONE`
	}
	return extractionSystemPrompt
}

func (e *Extractor) isValidPredicate(p string) bool {
	if e.skipPredicateCheck {
		return true
	}
	if e.validator != nil {
		return e.validator(p)
	}
	return defaultIsValidPredicate(p)
}

func (e *Extractor) Extract(ctx context.Context, content, sourceID string) ([]Triple, error) {
	if content == "" {
		return nil, nil
	}
	userPrompt := fmt.Sprintf("Extract entities and relationships from:\n\n%s", content)
	response, err := e.generator.GenerateText(ctx, e.systemPrompt(), userPrompt)
	if err != nil {
		return nil, fmt.Errorf("generate extraction: %w", err)
	}
	return e.parseResponse(response, sourceID), nil
}
```

And in `wireGraphCallbacks`:

```go
var opts []graph.ExtractorOption
if gc != nil && gc.admissionPolicy != nil {
	opts = append(opts, graph.WithoutPredicateValidation())
} else if ontologyValidator != nil {
	opts = append(opts, graph.WithPredicateValidator(ontologyValidator))
}
extractor = graph.NewExtractor(generator, logger(), opts...)
```

- [ ] **Step 4: Run tests**

Run:

```bash
go test ./internal/app ./internal/graph -run 'AdmitExtractedTriples|ProducerForExtractedEvent|PassthroughModePreservesUnknownPredicate'
go test ./internal/graph -run 'PassthroughPromptOmitsFixedAllowlist'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/app/wiring_graph.go internal/app/wiring_graph_test.go
git add internal/app/modules.go internal/graph/extractor.go internal/graph/extractor_test.go
git commit -m "feat: admit runtime graph extraction before enqueue"
```

## Task 5: Wire Admission Telemetry And GraphEngine Direct Writes

**Files:**
- Modify: `internal/learning/graph_engine.go`
- Modify: `internal/app/wiring_knowledge.go`
- Modify: `internal/app/wiring_observability.go`
- Modify: `internal/observability/collector.go`
- Modify: `internal/cli/cockpit/pages/status.go`
- Test: `internal/learning/graph_engine_test.go`

- [ ] **Step 1: Write failing tests for GraphEngine and telemetry consumers**

Add tests like these:

```go
func TestGraphEngine_RecordFix_AdmissionEnforceFiltersUnknown(t *testing.T) {
	gs := &fakeGraphStore{}
	logger := zap.NewNop().Sugar()
	p := graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Mode:      graph.AdmissionModeEnforce,
		Validator: func(name string) bool { return name == graph.ResolvedBy },
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerLearningEvent: 0.60,
		},
	}, logger)

	ge := &GraphEngine{
		Engine:          &Engine{store: nil, logger: logger},
		graphStore:      gs,
		admissionPolicy: p,
		propagation:     0.3,
		logger:          logger,
	}

	ge.RecordFix(context.Background(), "timeout error", "increase timeout", "session-1")
	require.Len(t, gs.triples, 1, "admission should filter LearnedFrom when validator disallows it")
}

func TestGraphEngine_RecordErrorGraph_AdmissionEnforceFiltersUnknown(t *testing.T) {
	gs := &fakeGraphStore{}
	logger := zap.NewNop().Sugar()
	p := graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Mode:      graph.AdmissionModeEnforce,
		Validator: func(name string) bool { return name == graph.CausedBy },
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerLearningEvent: 0.60,
		},
	}, logger)

	ge := &GraphEngine{
		Engine:          &Engine{store: nil, logger: logger},
		graphStore:      gs,
		admissionPolicy: p,
		propagation:     0.3,
		logger:          logger,
	}

	ge.recordErrorGraph(context.Background(), "s1", "tool1", fmt.Errorf("boom"))
	require.Len(t, gs.triples, 1, "admission should filter InSession when validator disallows it")
}
```

```go
func TestCollector_RecordGraphAdmissionBatch(t *testing.T) {
	c := NewCollector()
	c.RecordGraphAdmissionBatch("observe", "learning_event", 3, 3, 0, 1)

	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.GraphAdmission.Batches)
	assert.Equal(t, int64(3), snap.GraphAdmission.Candidates)
	assert.Equal(t, int64(1), snap.GraphAdmission.ObservedUnknown)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/learning ./internal/observability -run 'GraphEngine_RecordFix_AdmissionEnforceFiltersUnknown|GraphEngine_RecordErrorGraph_AdmissionEnforceFiltersUnknown|RecordGraphAdmissionBatch'
```

Expected: FAIL because the direct-write admission path and collector metrics do not exist

- [ ] **Step 3: Implement shared policy use and telemetry consumer wiring**

Apply changes like these:

```go
// internal/learning/graph_engine.go
type GraphEngine struct {
	*Engine
	graphStore      graph.Store
	bus             *eventbus.Bus
	admissionPolicy *graph.AdmissionPolicy
	propagation     float64
	logger          *zap.SugaredLogger
}

func (e *GraphEngine) SetAdmissionPolicy(p *graph.AdmissionPolicy) {
	e.admissionPolicy = p
}

func (e *GraphEngine) addTriples(ctx context.Context, triples []graph.Triple) error {
	if e.graphStore == nil {
		return nil
	}
	if e.admissionPolicy == nil {
		return e.graphStore.AddTriples(ctx, triples)
	}
	candidates := make([]graph.AdmissionCandidate, 0, len(triples))
	for _, triple := range triples {
		candidates = append(candidates, graph.AdmissionCandidate{
			Triple:   triple,
			Producer: graph.ProducerLearningEvent,
			Source:   "graph_engine",
		})
	}
	result := e.admissionPolicy.AdmitBatch(candidates)
	if len(result.Admitted) == 0 {
		return nil
	}
	return e.graphStore.AddTriples(ctx, result.Admitted)
}
```

```go
// internal/app/wiring_knowledge.go
if gc != nil {
	graphEngine := learning.NewGraphEngine(kStore, gc.store, kLogger)
	if gc.admissionPolicy != nil {
		graphEngine.SetAdmissionPolicy(gc.admissionPolicy)
	}
	graphEngine.SetEventBus(bus)
	observer = graphEngine
}
```

```go
// internal/observability/collector.go
type GraphAdmissionMetrics struct {
	Batches         int64
	Candidates      int64
	Admitted        int64
	Rejected        int64
	ObservedUnknown int64
}

type MetricsCollector struct {
	mu             sync.RWMutex
	startedAt      time.Time
	totalTokens    TokenUsageSummary
	sessions       map[string]*SessionMetric
	agents         map[string]*AgentMetric
	toolExecs      int64
	tools          map[string]*ToolMetric
	policyBlocks   int64
	policyObserves int64
	policyByReason map[string]int64
	graphAdmission GraphAdmissionMetrics
	MaxSessions    int
}

func (c *MetricsCollector) RecordGraphAdmissionBatch(mode, producer string, candidates, admitted, rejected, observedUnknown int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graphAdmission.Batches++
	c.graphAdmission.Candidates += int64(candidates)
	c.graphAdmission.Admitted += int64(admitted)
	c.graphAdmission.Rejected += int64(rejected)
	c.graphAdmission.ObservedUnknown += int64(observedUnknown)
}

// Extend the existing Snapshot() implementation by copying the current
// graphAdmission counters into the returned SystemSnapshot.
```

```go
// internal/observability/types.go
type GraphAdmissionSummary struct {
	Batches         int64
	Candidates      int64
	Admitted        int64
	Rejected        int64
	ObservedUnknown int64
}

type SystemSnapshot struct {
	StartedAt        time.Time
	Uptime           time.Duration
	TokenUsageTotal  TokenUsageSummary
	ToolExecutions   int64
	ToolBreakdown    map[string]ToolMetric
	AgentBreakdown   map[string]AgentMetric
	SessionBreakdown map[string]SessionMetric
	Policy           PolicyMetrics
	GraphAdmission   GraphAdmissionSummary
}
```

```go
// internal/app/wiring_observability.go
eventbus.SubscribeTyped[eventbus.GraphAdmissionBatchEvent](bus, func(evt eventbus.GraphAdmissionBatchEvent) {
	oc.collector.RecordGraphAdmissionBatch(
		evt.Mode,
		evt.Producer,
		evt.Candidates,
		evt.Admitted,
		evt.Rejected,
		evt.ObservedUnknown,
	)
})
```

```go
// internal/cli/cockpit/pages/status.go
sections := []string{
	m.renderFeatureFlags(sectionTitle, divider),
	m.renderTokenUsage(sectionTitle, divider),
	m.renderToolExecution(sectionTitle, divider),
	m.renderGraphAdmission(sectionTitle, divider),
	m.renderSystemInfo(sectionTitle, divider),
}

func (m *StatusPage) renderGraphAdmission(titleStyle lipgloss.Style, divider string) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Graph Admission"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')
	ga := m.snapshot.GraphAdmission
	b.WriteString(fmt.Sprintf("Batches: %d\n", ga.Batches))
	b.WriteString(fmt.Sprintf("Candidates: %d\n", ga.Candidates))
	b.WriteString(fmt.Sprintf("Rejected: %d\n", ga.Rejected))
	b.WriteString(fmt.Sprintf("Observed unknown: %d\n", ga.ObservedUnknown))
	return b.String()
}
```

- [ ] **Step 4: Run tests**

Run:

```bash
go test ./internal/learning ./internal/observability -run 'GraphEngine_RecordFix_AdmissionEnforceFiltersUnknown|GraphEngine_RecordErrorGraph_AdmissionEnforceFiltersUnknown|RecordGraphAdmissionBatch'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/learning/graph_engine.go internal/learning/graph_engine_test.go internal/app/wiring_knowledge.go internal/app/wiring_observability.go internal/observability/collector.go internal/cli/cockpit/pages/status.go
git commit -m "feat: wire graph admission telemetry and direct writes"
```

## Task 6: Update Downstream Docs And Verify The Slice

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

- Dynamic runtime producer admission boundary: covered in Tasks 3, 4, and 5.
- Observe vs enforce mode: covered in Tasks 2, 3, 4, and 5.
- Single validator source of truth: covered in Tasks 4 and 5 by reusing the injected ontology validator closure.
- Batch pre-filtering while retaining atomic writes: covered in Tasks 3, 4, and 5.
- Downstream docs and config surfaces: covered in Tasks 2 and 6.

Intentional gaps:

- CLI `graph import` is not in this plan.
- `AssertFact`, ontology tools, ontology actions, and P2P fact assertion paths are not in this plan.
- Adaptive shadow growth, schema candidate persistence, and replay hardening are not in this plan.

Placeholder scan:

- No placeholder markers remain.
- Each code-changing step includes concrete code.
- Commands are explicit and package-scoped before the final full-suite run.

Type consistency checks:

- `AdmissionModeObserve` / `AdmissionModeEnforce`
- `ProducerLearningEvent` / `ProducerLibrarianEvent`
- `AdmissionCandidate`, `AdmissionRecord`, `AdmissionResult`, `AdmissionBatchEvent`, `AdmissionStatsSnapshot`
- `producerForExtractedEvent`, `publishAdmissionObservation`, `admitTriples`, `WithoutPredicateValidation`

## Execution Handoff

Plan complete and saved to `internal-docs/superpowers/plans/2026-05-03-runtime-admission-boundary-hardening-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
