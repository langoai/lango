# Runtime Admission Observe Wave One Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an observe-only runtime admission layer for app-path dynamic graph triple producers so Lango can classify unknown predicates and types, emit admission telemetry, and measure failure baselines without changing existing write routing yet.

**Architecture:** This plan implements only `Change A / Phase A1` from the adaptive ontology growth design. It instruments runtime producers, computes admission decisions with the ontology validator closure, publishes graph-admission observability events, and records those metrics in the collector and status page. It does **not** filter, rewrite, or reroute writes; all existing enqueue and direct-store behavior stays intact.

**Tech Stack:** Go, Ent-backed ontology service, BoltDB graph store, synchronous event bus, observability collector, cockpit status page, OpenSpec experimental workflow

---

## Scope

This plan covers only the **observe-only runtime app sub-slice** from `Change A / Phase A1`:

- `TriplesExtractedEvent` producers that already flow through `internal/app/wiring_graph.go`
- `GraphEngine` direct store writes when the event-bus handoff is not used
- extractor-local dropped-unknown telemetry for the `content.saved` async extraction path
- graph admission telemetry and status surfaces
- config and settings needed to turn observe mode on and off

This plan explicitly does **not** cover:

- `content.saved` extractor prompt widening or allowlist bypass
- `CLI graph import`
- `AssertFact`, ontology tools, ontology actions, and P2P fact assertion paths
- any `enforce`-mode filtering or write dropping
- adaptive shadow growth, schema candidate persistence, or replay hardening

## File Structure

### New Files

- `internal/graph/admission.go`
  Observe-only admission policy primitives, producer enums, decision records, and telemetry event payloads.
- `internal/graph/admission_test.go`
  Unit tests for producer confidence fallback and observe-mode decision recording.
- `internal/eventbus/observability_events_test.go`
  Event type tests for the new graph admission event.
- `internal/app/wiring_graph_test.go`
  Tests for event-source-to-producer mapping and observe-only telemetry hook behavior.

### Modified Files

- `internal/config/types_ontology.go`
  Add observe-mode config and producer fallback confidence fields.
- `internal/config/loader.go`
  Set defaults for the new config.
- `internal/cli/settings/forms_ontology.go`
  Expose the observe-mode config in settings.
- `internal/cli/tuicore/state_update.go`
  Persist the new settings fields.
- `internal/app/modules.go`
  Build the shared observe-mode admission policy after ontology init and before knowledge init.
- `internal/app/wiring_graph.go`
  Observe `TriplesExtractedEvent` batches before existing enqueue behavior.
- `internal/graph/extractor.go`
  Publish observe-only dropped-unknown telemetry for the current `content.saved` extraction path without widening extraction behavior.
- `internal/app/wiring_knowledge.go`
  Pass the shared observe-mode policy into `GraphEngine`.
- `internal/graph/extractor.go`
  Publish observe-only dropped-unknown telemetry for the current extractor path without widening extraction behavior.
- `internal/learning/graph_engine.go`
  Observe direct `AddTriples` writes before existing direct-store behavior.
- `internal/eventbus/observability_events.go`
  Add the graph admission event type.
- `internal/app/wiring_observability.go`
  Subscribe the metrics collector to graph admission events.
- `internal/observability/types.go`
  Add graph admission fields to `SystemSnapshot`.
- `internal/observability/collector.go`
  Record graph admission metrics, dropped-unknown extractor metrics, graph write failure metrics, and include them in snapshots and reset behavior.
- `internal/cli/cockpit/pages/status.go`
  Render graph admission metrics on the status page.
- `docs/configuration.md`
  Document the new observe-mode config.
- `docs/features/ontology.md`
  Document that A1 is observe-only and does not change routing.
- `docs/features/knowledge-graph.md`
  Document graph admission telemetry for dynamic runtime producers.
- `README.md`
  Update the high-level feature list.

### Existing Tests To Extend

- `internal/config/loader_test.go`
- `internal/cli/settings/forms_impl_test.go`
- `internal/learning/graph_engine_test.go`
- `internal/observability/collector_test.go`
- `internal/graph/extractor_test.go`

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

Dynamic runtime graph producers currently produce unknown-predicate failures that surface too late at graph validation time. Before changing write behavior, the runtime needs an observe-only admission layer to classify those batches and measure where the failures come from.

## What Changes

- Add an observe-only graph admission policy for runtime app producers.
- Publish graph admission telemetry, dropped-unknown extractor telemetry, and graph write failure baselines to observability and cockpit status.
- Share one ontology predicate validator closure across graph admission and graph-store validation.

## Out Of Scope

- write filtering or dropping
- CLI import
- `AssertFact`/ontology fact assertion paths
- adaptive shadow growth
```

- [ ] **Step 3: Write the design**

Use this content for `openspec/changes/runtime-admission-boundary-hardening/design.md`:

```markdown
# Design

This change implements only `Change A / Phase A1`.

The runtime computes admission decisions for:
- `TriplesExtractedEvent` batches
- `GraphEngine` direct store writes
- app-path extracted triples only where the current runtime already surfaces them without widening extractor behavior
- extractor-local dropped-unknown events for the current `content.saved` extraction path

In all cases, current write routing remains unchanged. The policy is observe-only: it records what would be admitted or rejected, emits telemetry, and lets the existing enqueue/store path continue unchanged.
```

- [ ] **Step 4: Write the delta specs**

Use these deltas:

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/knowledge-graph/spec.md -->
### Requirement: Observe-only admission for runtime dynamic producers
Runtime dynamic graph triple producers SHALL compute admission decisions before existing graph write steps, but SHALL preserve current write routing in observe mode.

#### Scenario: TriplesExtractedEvent observe path
- **WHEN** a `TriplesExtractedEvent` batch is processed
- **THEN** the runtime SHALL record graph-admission telemetry
- **AND** the runtime SHALL still enqueue the original triples

#### Scenario: GraphEngine direct-write observe path
- **WHEN** `GraphEngine` writes directly to the graph store without using the event bus
- **THEN** the runtime SHALL record graph-admission telemetry
- **AND** the runtime SHALL still call the existing direct `AddTriples` path
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/ontology-registry/spec.md -->
### Requirement: Shared predicate validity source
The runtime SHALL use the ontology service predicate validator closure as the shared predicate-validity source for observe-only admission decisions and graph-store validation.
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/cli-settings/spec.md -->
### Requirement: Observe-mode admission settings
The settings surface SHALL expose runtime admission observe-mode configuration and per-producer fallback confidence defaults.
```

- [ ] **Step 5: Record tasks and commit**

Use this content for `openspec/changes/runtime-admission-boundary-hardening/tasks.md`:

```markdown
# Tasks

- [ ] Add observe-mode admission config and defaults
- [ ] Add graph admission policy and telemetry event types
- [ ] Observe runtime event-bus and direct-store producer paths plus extractor dropped-unknown baseline
- [ ] Add graph admission, dropped-unknown, unmapped-source, and graph-write-failure metrics to observability and cockpit status
- [ ] Update docs and verify
```

Run:

```bash
git add openspec/changes/runtime-admission-boundary-hardening
git commit -m "openspec: scaffold runtime admission observe wave one"
```

Expected: commit succeeds

## Task 2: Add Observe-Mode Admission Config

**Files:**
- Modify: `internal/config/types_ontology.go`
- Modify: `internal/config/loader.go`
- Modify: `internal/cli/settings/forms_ontology.go`
- Modify: `internal/cli/tuicore/state_update.go`
- Test: `internal/config/loader_test.go`
- Test: `internal/cli/settings/forms_impl_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests like these:

```go
func TestDefaultConfig_OntologyAdmissionObserveDefaults(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "observe", cfg.Ontology.Governance.AdmissionMode)
	assert.InDelta(t, 0.60, cfg.Ontology.Governance.LearningDefaultConfidence, 0.001)
	assert.InDelta(t, 0.50, cfg.Ontology.Governance.LibrarianDefaultConfidence, 0.001)
}

func TestUpdateConfigFromForm_OntologyAdmissionObserveFields(t *testing.T) {
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "ontology_gov_admission_mode", Type: tuicore.InputSelect, Value: "observe"})
	form.AddField(&tuicore.Field{Key: "ontology_gov_learning_conf", Type: tuicore.InputText, Value: "0.65"})
	form.AddField(&tuicore.Field{Key: "ontology_gov_librarian_conf", Type: tuicore.InputText, Value: "0.55"})

	state := tuicore.NewConfigStateWith(config.DefaultConfig())
	state.UpdateConfigFromForm(&form)

	assert.Equal(t, "observe", state.Current.Ontology.Governance.AdmissionMode)
	assert.InDelta(t, 0.65, state.Current.Ontology.Governance.LearningDefaultConfidence, 0.001)
	assert.InDelta(t, 0.55, state.Current.Ontology.Governance.LibrarianDefaultConfidence, 0.001)
}
```

- [ ] **Step 2: Run the tests to verify failure**

Run:

```bash
go test ./internal/config ./internal/cli/settings -run 'OntologyAdmissionObserve'
```

Expected: FAIL with missing config fields

- [ ] **Step 3: Implement the config**

Apply these changes:

```go
// internal/config/types_ontology.go
type OntologyGovernanceConfig struct {
	Enabled                    bool    `mapstructure:"enabled" json:"enabled,omitempty"`
	MaxNewPerDay               int     `mapstructure:"maxNewPerDay" json:"maxNewPerDay,omitempty"`
	QuarantinePeriodHrs        int     `mapstructure:"quarantinePeriodHrs" json:"quarantinePeriodHrs,omitempty"`
	ShadowModeDurationHrs      int     `mapstructure:"shadowModeDurationHrs" json:"shadowModeDurationHrs,omitempty"`
	MinUsageForPromotion       int     `mapstructure:"minUsageForPromotion" json:"minUsageForPromotion,omitempty"`
	SchemaExplosionBudget      int     `mapstructure:"schemaExplosionBudget" json:"schemaExplosionBudget,omitempty"`
	AdmissionMode              string  `mapstructure:"admissionMode" json:"admissionMode,omitempty"`
	LearningDefaultConfidence  float64 `mapstructure:"learningDefaultConfidence" json:"learningDefaultConfidence,omitempty"`
	LibrarianDefaultConfidence float64 `mapstructure:"librarianDefaultConfidence" json:"librarianDefaultConfidence,omitempty"`
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
	Options:     []string{"observe"},
	Description: "Observe-only runtime admission mode for graph triple producers",
	VisibleWhen: admissionVisible,
})
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_learning_conf",
	Label:       "    Learning Default Confidence",
	Type:        tuicore.InputText,
	Value:       fmt.Sprintf("%.2f", cfg.Ontology.Governance.LearningDefaultConfidence),
	Placeholder: "0.60",
	Description: "Fallback confidence for learning-derived triple events",
	VisibleWhen: admissionVisible,
})
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_librarian_conf",
	Label:       "    Librarian Default Confidence",
	Type:        tuicore.InputText,
	Value:       fmt.Sprintf("%.2f", cfg.Ontology.Governance.LibrarianDefaultConfidence),
	Placeholder: "0.50",
	Description: "Fallback confidence for librarian-derived triple events",
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
go test ./internal/config ./internal/cli/settings -run 'OntologyAdmissionObserve'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/config/types_ontology.go internal/config/loader.go internal/cli/settings/forms_ontology.go internal/cli/tuicore/state_update.go internal/config/loader_test.go internal/cli/settings/forms_impl_test.go
git commit -m "feat: add runtime admission observe config"
```

## Task 3: Add The Observe-Only Admission Policy And Event Type

**Files:**
- Create: `internal/graph/admission.go`
- Test: `internal/graph/admission_test.go`
- Modify: `internal/eventbus/observability_events.go`
- Test: `internal/eventbus/observability_events_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests like these:

```go
func TestAdmissionPolicy_ObserveBatchRecordsUnknown(t *testing.T) {
	p := NewAdmissionPolicy(AdmissionConfig{
		Validator: func(name string) bool { return name == CausedBy },
		DefaultConfidence: map[AdmissionProducer]float64{
			ProducerConversationAnalysis: 0.60,
			ProducerUnknownSource:        0.40,
		},
	}, zap.NewNop().Sugar())

	result := p.ObserveBatch([]AdmissionCandidate{{
		Triple:   Triple{Subject: "a", Predicate: "invented_rel", Object: "b"},
		Producer: ProducerConversationAnalysis,
		Source:   "conversation_analysis",
	}})

	assert.Len(t, result.Records, 1)
	assert.Equal(t, DecisionObservedUnknown, result.Records[0].Decision)
	assert.Len(t, result.Forwarded, 1)
}

func TestGraphAdmissionBatchEvent_EventName(t *testing.T) {
	evt := eventbus.GraphAdmissionBatchEvent{}
	assert.Equal(t, eventbus.EventGraphAdmissionBatch, evt.EventName())
}
```

- [ ] **Step 2: Run the tests to verify failure**

Run:

```bash
go test ./internal/graph ./internal/eventbus -run 'ObserveBatch|GraphAdmissionBatchEvent'
```

Expected: FAIL because `ObserveBatch`, admission types, and graph admission event type do not exist

- [ ] **Step 3: Implement the observe-only policy**

Create `internal/graph/admission.go` with this shape:

```go
package graph

import "go.uber.org/zap"

type AdmissionProducer string

const (
	ProducerConversationAnalysis AdmissionProducer = "conversation_analysis"
	ProducerSessionLearning      AdmissionProducer = "session_learning"
	ProducerUnknownSource        AdmissionProducer = "unknown_source"
	ProducerLibrarianEvent       AdmissionProducer = "proactive_librarian"
	ProducerGraphEngine          AdmissionProducer = "graph_engine"
)

type AdmissionDecision string

const (
	DecisionObservedKnown   AdmissionDecision = "observed_known"
	DecisionObservedUnknown AdmissionDecision = "observed_unknown"
)

type AdmissionCandidate struct {
	Triple     Triple
	Producer   AdmissionProducer
	Source     string
	Confidence *float64
}

type AdmissionRecord struct {
	Candidate  AdmissionCandidate
	Decision   AdmissionDecision
	Confidence float64
}

type AdmissionBatchEvent struct {
	Producer        AdmissionProducer
	Candidates      int
	ObservedKnown   int
	ObservedUnknown int
	ObservedTypeHints int
	DroppedUnknown  int
}

type AdmissionObserveResult struct {
	Records   []AdmissionRecord
	Forwarded []Triple
	Event     AdmissionBatchEvent
}

type AdmissionConfig struct {
	Validator         PredicateValidatorFunc
	DefaultConfidence map[AdmissionProducer]float64
	Observe           func(AdmissionBatchEvent)
}

type AdmissionPolicy struct {
	cfg    AdmissionConfig
	logger *zap.SugaredLogger
}

func NewAdmissionPolicy(cfg AdmissionConfig, logger *zap.SugaredLogger) *AdmissionPolicy {
	return &AdmissionPolicy{cfg: cfg, logger: logger}
}

func (p *AdmissionPolicy) ObserveBatch(candidates []AdmissionCandidate) AdmissionObserveResult {
	result := AdmissionObserveResult{
		Forwarded: make([]Triple, 0, len(candidates)),
		Records:   make([]AdmissionRecord, 0, len(candidates)),
	}
	if len(candidates) > 0 {
		result.Event.Producer = candidates[0].Producer
	}
	result.Event.Candidates = len(candidates)

	for _, c := range candidates {
		conf := p.defaultConfidence(c)
		known := p.cfg.Validator == nil || p.cfg.Validator(c.Triple.Predicate)
		decision := DecisionObservedUnknown
		if known {
			decision = DecisionObservedKnown
			result.Event.ObservedKnown++
		} else {
			result.Event.ObservedUnknown++
		}
		if c.Triple.SubjectType != "" || c.Triple.ObjectType != "" {
			result.Event.ObservedTypeHints++
		}
		result.Records = append(result.Records, AdmissionRecord{
			Candidate:  c,
			Decision:   decision,
			Confidence: conf,
		})
		result.Forwarded = append(result.Forwarded, c.Triple)
	}

	p.logger.Infow("graph admission observe batch",
		"producer", result.Event.Producer,
		"candidates", result.Event.Candidates,
		"known", result.Event.ObservedKnown,
		"unknown", result.Event.ObservedUnknown,
	)
	if p.cfg.Observe != nil {
		p.cfg.Observe(result.Event)
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

func ObserveTriples(policy *AdmissionPolicy, producer AdmissionProducer, source string, triples []Triple) []Triple {
	if policy == nil {
		return triples
	}
	candidates := make([]AdmissionCandidate, 0, len(triples))
	for _, triple := range triples {
		candidates = append(candidates, AdmissionCandidate{
			Triple:   triple,
			Producer: producer,
			Source:   source,
		})
	}
	result := policy.ObserveBatch(candidates)
	return result.Forwarded
}
```

And add this event:

```go
// internal/eventbus/observability_events.go
const EventGraphAdmissionBatch = "graph.admission.batch"

type GraphAdmissionBatchEvent struct {
	Producer        string
	Candidates      int
	ObservedKnown   int
	ObservedUnknown int
	ObservedTypeHints int
	DroppedUnknown int
}

func (e GraphAdmissionBatchEvent) EventName() string { return EventGraphAdmissionBatch }
```

- [ ] **Step 4: Run the tests again**

Run:

```bash
go test ./internal/graph ./internal/eventbus -run 'ObserveBatch|GraphAdmissionBatchEvent'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/graph/admission.go internal/graph/admission_test.go internal/eventbus/observability_events.go internal/eventbus/observability_events_test.go
git commit -m "feat: add observe-only graph admission policy"
```

## Task 4: Observe Runtime Event Producers Without Changing Routing

**Files:**
- Modify: `internal/app/modules.go`
- Modify: `internal/app/wiring_graph.go`
- Modify: `internal/graph/extractor.go`
- Test: `internal/app/wiring_graph_test.go`
- Test: `internal/graph/extractor_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests like these:

```go
func TestProducerForExtractedEvent_MapsKnownSources(t *testing.T) {
	assert.Equal(t, graph.ProducerConversationAnalysis, producerForExtractedEvent("conversation_analysis"))
	assert.Equal(t, graph.ProducerSessionLearning, producerForExtractedEvent("session_learning"))
	assert.Equal(t, graph.ProducerGraphEngine, producerForExtractedEvent("learning"))
	assert.Equal(t, graph.ProducerLibrarianEvent, producerForExtractedEvent("proactive_librarian"))
	assert.Equal(t, graph.ProducerUnknownSource, producerForExtractedEvent("new_source"))
}

func TestObserveExtractedTriples_PreservesOriginalTriples(t *testing.T) {
	p := graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Validator: func(name string) bool { return name == graph.CausedBy },
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerConversationAnalysis: 0.60,
		},
	}, zap.NewNop().Sugar())

	triples := observeExtractedTriples(p, eventbus.TriplesExtractedEvent{
		Source: "conversation_analysis",
		Triples: []eventbus.Triple{{
			Subject: "a", Predicate: "invented_rel", Object: "b",
		}},
	})

	require.Len(t, triples, 1)
	assert.Equal(t, "invented_rel", triples[0].Predicate)
}

func TestExtractor_EmitDroppedUnknownObservation(t *testing.T) {
	logger := zap.NewNop().Sugar()
	var got []DroppedUnknownPredicateEvent
	e := NewExtractor(nil, logger, WithDroppedUnknownObserver(func(evt DroppedUnknownPredicateEvent) {
		got = append(got, evt)
	}))

	triples := e.parseResponse("a|invented_rel|b", "src")
	require.Len(t, triples, 0)
	require.Len(t, got, 1)
	assert.Equal(t, "invented_rel", got[0].Predicate)
}
```

- [ ] **Step 2: Run the tests to verify failure**

Run:

```bash
go test ./internal/app -run 'ProducerForExtractedEvent|ObserveExtractedTriples'
go test ./internal/graph -run 'EmitDroppedUnknownObservation'
```

Expected: FAIL because the observe helpers do not exist

- [ ] **Step 3: Implement observe-only wiring**

Apply changes like these:

```go
// internal/app/wiring_graph.go
type graphComponents struct {
	store           graph.Store
	buffer          *graph.GraphBuffer
	ragService      *graph.GraphRAGService
	admissionPolicy *graph.AdmissionPolicy
}

func producerForExtractedEvent(source string) graph.AdmissionProducer {
	switch strings.TrimSpace(source) {
	case "conversation_analysis":
		return graph.ProducerConversationAnalysis
	case "session_learning":
		return graph.ProducerSessionLearning
	case "learning":
		return graph.ProducerGraphEngine
	case "proactive_librarian":
		return graph.ProducerLibrarianEvent
	default:
		return graph.ProducerUnknownSource
	}
}

func publishAdmissionObservation(bus *eventbus.Bus, evt graph.AdmissionBatchEvent) {
	if bus == nil {
		return
	}
	bus.Publish(eventbus.GraphAdmissionBatchEvent{
		Producer:        string(evt.Producer),
		Candidates:      evt.Candidates,
		ObservedKnown:   evt.ObservedKnown,
		ObservedUnknown: evt.ObservedUnknown,
		ObservedTypeHints: evt.ObservedTypeHints,
		DroppedUnknown: evt.DroppedUnknown,
	})
}

func observeExtractedTriples(policy *graph.AdmissionPolicy, evt eventbus.TriplesExtractedEvent) []graph.Triple {
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
	return graph.ObserveTriples(policy, producerForExtractedEvent(evt.Source), evt.Source, graphTriples)
}
```

```go
// internal/app/modules.go
if gc != nil && ontologyResult != nil && ontologyResult.Service != nil && cfg.Ontology.Enabled && cfg.Ontology.Governance.AdmissionMode == "observe" {
	gc.admissionPolicy = graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Validator: ontologyResult.Service.PredicateValidator(),
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerConversationAnalysis: cfg.Ontology.Governance.LearningDefaultConfidence,
			graph.ProducerSessionLearning:      cfg.Ontology.Governance.LearningDefaultConfidence,
			graph.ProducerLibrarianEvent:       cfg.Ontology.Governance.LibrarianDefaultConfidence,
			graph.ProducerGraphEngine:          cfg.Ontology.Governance.LearningDefaultConfidence,
			graph.ProducerUnknownSource:        cfg.Ontology.Governance.LearningDefaultConfidence,
		},
		Observe: func(evt graph.AdmissionBatchEvent) {
			publishAdmissionObservation(m.bus, evt)
		},
	}, logger())
}
```

```go
// internal/app/wiring_graph.go
eventbus.SubscribeTyped(bus, func(evt eventbus.TriplesExtractedEvent) {
	graphTriples := observeExtractedTriples(gc.admissionPolicy, evt)
	if len(graphTriples) == 0 {
		return
	}
	gc.buffer.Enqueue(graph.GraphRequest{Triples: graphTriples})
})
```

Add extractor-local baseline telemetry without widening extraction behavior:

```go
// internal/graph/extractor.go
type DroppedUnknownPredicateEvent struct {
	SourceID  string
	Predicate string
	Subject   string
	Object    string
}

func WithDroppedUnknownObserver(fn func(DroppedUnknownPredicateEvent)) ExtractorOption {
	return func(e *Extractor) { e.onDroppedUnknown = fn }
}

type Extractor struct {
	generator        llm.TextGenerator
	validator        PredicateValidatorFunc
	onDroppedUnknown func(DroppedUnknownPredicateEvent)
	logger           *zap.SugaredLogger
}

// inside parseResponse unknown-predicate branch
if !e.isValidPredicate(predicate) {
	if e.onDroppedUnknown != nil {
		e.onDroppedUnknown(DroppedUnknownPredicateEvent{
			SourceID:  sourceID,
			Predicate: predicate,
			Subject:   subject,
			Object:    object,
		})
	}
	e.logger.Warnw("rejected unknown predicate from extraction",
		"predicate", predicate,
		"subject", subject,
		"object", object,
		"source", sourceID,
	)
	continue
}
```

And in `wireGraphCallbacks`:

```go
extractor = graph.NewExtractor(generator, logger(),
	graph.WithDroppedUnknownObserver(func(evt graph.DroppedUnknownPredicateEvent) {
		if gc.admissionPolicy == nil {
			return
		}
		publishAdmissionObservation(bus, graph.AdmissionBatchEvent{
			Producer:        graph.ProducerUnknownSource,
			Candidates:      1,
			ObservedKnown:   0,
			ObservedUnknown: 1,
			ObservedTypeHints: 0,
			DroppedUnknown:  1,
		})
	}),
	graph.WithPredicateValidator(ontologyValidator),
)
```

- [ ] **Step 4: Run the tests again**

Run:

```bash
go test ./internal/app -run 'ProducerForExtractedEvent|ObserveExtractedTriples'
go test ./internal/graph -run 'EmitDroppedUnknownObservation'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/app/modules.go internal/app/wiring_graph.go internal/app/wiring_graph_test.go internal/graph/extractor.go internal/graph/extractor_test.go
git commit -m "feat: observe runtime graph event producers"
```

## Task 5: Observe GraphEngine Direct Writes And Wire Telemetry Consumers

**Files:**
- Modify: `internal/learning/graph_engine.go`
- Modify: `internal/app/wiring_knowledge.go`
- Modify: `internal/app/wiring_observability.go`
- Modify: `internal/observability/types.go`
- Modify: `internal/observability/collector.go`
- Modify: `internal/cli/cockpit/pages/status.go`
- Test: `internal/learning/graph_engine_test.go`
- Test: `internal/observability/collector_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests like these:

```go
func TestGraphEngine_RecordFix_ObserveModePreservesDirectWrite(t *testing.T) {
	gs := &fakeGraphStore{}
	logger := zap.NewNop().Sugar()
	p := graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Validator: func(name string) bool { return name == graph.ResolvedBy },
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerGraphEngine: 0.60,
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
	require.Len(t, gs.triples, 2, "observe-only path must preserve original direct write batch")
}

func TestGraphEngine_RecordErrorGraph_ObserveModePreservesDirectWrite(t *testing.T) {
	gs := &fakeGraphStore{}
	logger := zap.NewNop().Sugar()
	p := graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Validator: func(name string) bool { return name == graph.CausedBy },
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerGraphEngine: 0.60,
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
	require.Len(t, gs.triples, 2, "observe-only path must preserve original direct write batch")
}

func TestCollector_RecordGraphAdmissionBatch(t *testing.T) {
	c := NewCollector()
	c.RecordGraphAdmissionBatch("conversation_analysis", 3, 2, 1)
	c.RecordGraphAdmissionWriteFailure()
	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.GraphAdmission.Batches)
	assert.Equal(t, int64(3), snap.GraphAdmission.Candidates)
	assert.Equal(t, int64(1), snap.GraphAdmission.ObservedUnknown)
	assert.Equal(t, int64(1), snap.GraphAdmission.WriteFailures)
}
```

- [ ] **Step 2: Run the tests to verify failure**

Run:

```bash
go test ./internal/learning ./internal/observability -run 'ObserveModePreservesDirectWrite|RecordGraphAdmissionBatch'
```

Expected: FAIL because the direct-write observe path and graph admission metrics do not exist

- [ ] **Step 3: Implement GraphEngine observe path and telemetry consumers**

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

func (e *GraphEngine) observeTriples(triples []graph.Triple) []graph.Triple {
	if e.admissionPolicy == nil {
		return triples
	}
	return graph.ObserveTriples(e.admissionPolicy, graph.ProducerGraphEngine, "graph_engine", triples)
}
```

Use that helper in both direct-write branches:

```go
if e.bus != nil {
	e.publishTriples(triples)
} else if e.graphStore != nil {
	if addErr := e.graphStore.AddTriples(ctx, e.observeTriples(triples)); addErr != nil {
		e.logger.Warnw("add error graph triples", "error", addErr)
	}
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
// internal/observability/types.go
type GraphAdmissionSummary struct {
	Batches         int64
	Candidates      int64
	ObservedKnown   int64
	ObservedUnknown int64
	ObservedTypeHints int64
	UnmappedSources int64
	DroppedUnknown int64
	WriteFailures int64
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
// internal/observability/collector.go
type MetricsCollector struct {
	// existing fields...
	graphAdmission GraphAdmissionSummary
}

func (c *MetricsCollector) RecordGraphAdmissionBatch(producer string, candidates, observedKnown, observedUnknown, observedTypeHints, droppedUnknown int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graphAdmission.Batches++
	c.graphAdmission.Candidates += int64(candidates)
	c.graphAdmission.ObservedKnown += int64(observedKnown)
	c.graphAdmission.ObservedUnknown += int64(observedUnknown)
	c.graphAdmission.ObservedTypeHints += int64(observedTypeHints)
	c.graphAdmission.DroppedUnknown += int64(droppedUnknown)
	if producer == "unknown_source" {
		c.graphAdmission.UnmappedSources++
	}
}

func (c *MetricsCollector) RecordGraphAdmissionWriteFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graphAdmission.WriteFailures++
}

// Extend the existing Snapshot() implementation to copy c.graphAdmission.
// Extend Reset() to zero c.graphAdmission.
```

```go
// internal/app/wiring_observability.go
eventbus.SubscribeTyped[eventbus.GraphAdmissionBatchEvent](bus, func(evt eventbus.GraphAdmissionBatchEvent) {
	oc.collector.RecordGraphAdmissionBatch(
		evt.Producer,
		evt.Candidates,
		evt.ObservedKnown,
		evt.ObservedUnknown,
		evt.ObservedTypeHints,
		evt.DroppedUnknown,
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
	b.WriteString(fmt.Sprintf("Observed known: %d\n", ga.ObservedKnown))
	b.WriteString(fmt.Sprintf("Observed unknown: %d\n", ga.ObservedUnknown))
	b.WriteString(fmt.Sprintf("Type hints: %d\n", ga.ObservedTypeHints))
	b.WriteString(fmt.Sprintf("Dropped unknown: %d\n", ga.DroppedUnknown))
	b.WriteString(fmt.Sprintf("Unmapped sources: %d\n", ga.UnmappedSources))
	b.WriteString(fmt.Sprintf("Write failures: %d\n", ga.WriteFailures))
	return b.String()
}
```

- [ ] **Step 4: Run the tests again**

Run:

```bash
go test ./internal/learning ./internal/observability -run 'ObserveModePreservesDirectWrite|RecordGraphAdmissionBatch'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/learning/graph_engine.go internal/learning/graph_engine_test.go internal/app/wiring_knowledge.go internal/app/wiring_observability.go internal/observability/types.go internal/observability/collector.go internal/observability/collector_test.go internal/cli/cockpit/pages/status.go
git commit -m "feat: observe graph admission telemetry and direct writes"
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
rg -n "admissionMode|Learning Default Confidence|observe" README.md docs/configuration.md docs/features/ontology.md docs/features/knowledge-graph.md
```

Expected: either no matches or incomplete matches

- [ ] **Step 2: Update the docs**

Apply content like this:

```md
<!-- docs/configuration.md -->
| `ontology.governance.admissionMode` | `string` | `observe` | Observe-only runtime admission mode for app-path graph producers |
| `ontology.governance.learningDefaultConfidence` | `float64` | `0.60` | Fallback confidence for learning-derived graph events |
| `ontology.governance.librarianDefaultConfidence` | `float64` | `0.50` | Fallback confidence for librarian-derived graph events |
```

```md
<!-- docs/features/knowledge-graph.md -->
Observe-only graph admission now records admission decisions for dynamic runtime producers before the existing graph write path runs. This phase does not drop or rewrite triples yet.
```

```md
<!-- docs/features/ontology.md -->
The runtime admission boundary starts in `observe` mode only. It shares the ontology predicate validator with graph-store validation and emits observability signals without changing write routing.
```

```md
<!-- README.md -->
- Observe-only runtime graph admission telemetry for dynamic graph producers
```

- [ ] **Step 3: Run verification**

Run:

```bash
go build ./...
go test ./...
```

Expected: both commands exit `0`

- [ ] **Step 4: Verify the OpenSpec change**

Run:

```bash
openspec status --change "runtime-admission-boundary-hardening" --json
```

Expected: the change exists and artifacts are present

Then use the repo workflow:

```text
1. Mark completed items in openspec/changes/runtime-admission-boundary-hardening/tasks.md
2. Use openspec-verify-change for runtime-admission-boundary-hardening
3. Sync any delta specs that should land in main specs
4. Archive only after verification passes
```

- [ ] **Step 5: Commit**

Run:

```bash
git add README.md docs/configuration.md docs/features/ontology.md docs/features/knowledge-graph.md openspec/changes/runtime-admission-boundary-hardening
git commit -m "docs: document observe-only graph admission"
```

## Self-Review

Spec coverage:

- Observe-only dynamic runtime producer admission: covered in Tasks 3, 4, and 5.
- Shared validator source of truth: covered in Tasks 3 and 4.
- Graph admission telemetry and status surface: covered in Task 5.
- Config/settings surface: covered in Task 2.
- Public docs for the observe-only slice: covered in Task 6.

Intentional gaps:

- `enforce` mode and write filtering are not in this plan.
- `content.saved` extractor widening beyond observe-only instrumentation is not in this plan.
- `CLI graph import` is not in this plan.
- `AssertFact`, ontology tools, ontology actions, and P2P fact assertion paths are not in this plan.
- adaptive shadow growth, schema candidate persistence, and replay hardening are not in this plan.

Placeholder scan:

- No placeholder markers remain.
- Code-changing steps include concrete code.
- Commands are explicit and runnable.

Type consistency checks:

- `AdmissionProducer`, `AdmissionDecision`, `AdmissionCandidate`, `AdmissionRecord`
- `AdmissionBatchEvent`, `AdmissionObserveResult`
- `GraphAdmissionBatchEvent`, `GraphAdmissionSummary`
- `producerForExtractedEvent`, `observeTriples`, `observeExtractedTriples`, `publishAdmissionObservation`

## Execution Handoff

Plan complete and saved to `internal-docs/superpowers/plans/2026-05-03-runtime-admission-boundary-hardening-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
