# Runtime Admission Observe Wave One Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a toggleable observe-only graph-store admission layer for the current supported runtime graph inputs so Lango can classify unknown predicates, preserve original producer identity in telemetry, emit validator-source telemetry, and measure aggregate graph write-failure baselines without changing existing write routing yet.

**Architecture:** This plan implements only `Change A / Phase A1` from the adaptive ontology growth design. It instruments runtime producers, computes admission decisions with the ontology validator closure, publishes graph-admission observability events, and records those metrics in the collector and status page. It does **not** filter, rewrite, or reroute writes; all existing enqueue and direct-store behavior stays intact.

**Tech Stack:** Go, Ent-backed ontology service, BoltDB graph store, synchronous event bus, observability collector, cockpit status page, OpenSpec experimental workflow

---

## Scope

This plan covers only the **observe-only graph-store sub-slice** from `Change A / Phase A1`:

- event-bus `TriplesExtractedEvent` batches whose source is one of `conversation_analysis`, `session_learning`, `learning`, or `proactive_librarian`
- successful `content.saved` extracted triples, with observe-only telemetry emitted under the synthetic `content_saved_extractor` source label
- extractor-local dropped-unknown telemetry for the `content.saved` extraction path
- graph admission, `unmapped-source`, validator-source, and aggregate graph-write-failure telemetry plus cockpit status surfacing
- config and settings needed to turn the observe-only slice off or on

This plan explicitly does **not** cover:

- `content.saved` extractor prompt widening or allowlist bypass
- `CLI graph import`
- `AssertFact`, ontology tools, ontology actions, and P2P fact assertion paths
- any `enforce`-mode filtering or write dropping
- unknown type classification beyond basic type-hint observability
- dormant direct-store producer paths not used in the default runtime
- adaptive shadow growth, schema candidate persistence, or replay hardening

## File Structure

### New Files

- `internal/graph/admission.go`
  Observe-only admission policy primitives, producer enums, and decision records.
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
- `internal/graph/buffer.go`
  Publish graph write failure baseline telemetry for batched graph writes.
- `internal/eventbus/observability_events.go`
  Add the graph admission and baseline event types.
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
  Document graph admission telemetry for the supported runtime graph inputs.
- `README.md`
  Update the high-level feature list.

### Existing Tests To Extend

- `internal/config/loader_test.go`
- `internal/cli/settings/forms_impl_test.go`
- `internal/observability/collector_test.go`
- `internal/graph/extractor_test.go`

## Task 1: Scaffold The OpenSpec Change

**Files:**
- Create: `openspec/changes/runtime-admission-boundary-hardening/proposal.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/design.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/tasks.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/specs/graph-store/spec.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/specs/eventbus/spec.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/specs/cockpit-status-page/spec.md`
- Create: `openspec/changes/runtime-admission-boundary-hardening/specs/system-feedback/spec.md`
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

Dynamic runtime graph inputs currently surface unknown-predicate failures only after graph-store validation runs. Before changing write behavior, the runtime needs an observe-only admission boundary that classifies the supported event-bus producer sources and the `content.saved` extraction path without changing current write routing.

## What Changes

- Add an observe-only graph admission policy for the supported runtime graph inputs in this slice.
- Publish graph admission telemetry, extractor dropped-unknown baselines, graph write-failure baselines, `unmapped-source` telemetry, event-bus producer-source / producer-group attribution, the synthetic `content_saved_extractor` source label, and validator-source tags to observability and cockpit status.
- Reuse one ontology predicate validator closure as the primary predicate-validity source across admission classification and graph-store validation when ontology is available. If ontology is unavailable, graph-store validation falls back to the built-in hardcoded graph predicate validator and observe-only admission degrades to an `unvalidated` observation mode.
- Fix the observe-only admission decision taxonomy at the batch level and record triple-level counts for `known`, `unknown`, and `unvalidated` predicates.

## Terminology

- **Producer source**: the stable runtime label taken from `TriplesExtractedEvent.Source`.
- **Telemetry source label**: a stable non-event-bus telemetry label. In this slice the only synthetic label is `content_saved_extractor`.
- **Producer group**: the fallback-confidence configuration group used for event-bus producer sources. In this slice the only groups are `learning` and `librarian`.
- **Validator-source tag**: the stable telemetry tag naming the predicate validator source used for observe-only classification.
- The stable ontology-backed validator-source value in this slice is `ontology_registry`.
- **Unmapped source**: a raw `TriplesExtractedEvent.Source` label that is outside the supported event-bus producer-source set in this slice and therefore is not assigned to one of the known producer groups.
- **Admission decision taxonomy**: `known`, `unknown`, and `unvalidated` triple counts computed for each observed batch. `unvalidated` is used only when validator-based classification is unavailable.
- **Unavailable validator source**: the stable `validator-source` value used when validator-based classification is unavailable and the batch is observed as fully `unvalidated`.

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

Producer terminology in this slice is fixed as follows:
- **Producer source** = the stable runtime source label taken from `TriplesExtractedEvent.Source`.
- **Telemetry source label** = a stable synthetic label for a non-event-bus observed path.
- **Producer group** = the fallback-confidence configuration group for event-bus producer sources.
- `conversation_analysis`, `session_learning`, and `learning` map to the `learning` producer group.
- `proactive_librarian` maps to the `librarian` producer group.
- `content_saved_extractor` is the only synthetic telemetry source label in this slice and is used on observe-only telemetry emitted for returned triples and dropped-unknown baselines produced by the `content.saved` extraction path.
- Any other raw `TriplesExtractedEvent.Source` value remains visible as an `unmapped-source` telemetry signal and still follows the same graph write operation without observe-only admission classification.

The runtime computes observe-only admission decisions for:
- supported event-bus batches whose `TriplesExtractedEvent.Source` is one of `conversation_analysis`, `session_learning`, `learning`, or `proactive_librarian`
- triples returned from the `content.saved` extraction path; observe-only telemetry for that path SHALL use the synthetic `content_saved_extractor` source label and SHALL publish those telemetry events on the runtime event bus

Separately, this slice records a pre-admission extractor baseline for dropped-unknown events emitted by the `content.saved` extraction path before graph admission runs and tagged with the stable `content_saved_extractor` telemetry source label.

Each observe-only admission decision is batch-scoped. For every observed triple slice, the runtime computes:
- `batch_count = 1` for the observed slice
- `known_count` = number of triples whose predicate is accepted by validator-based classification
- `unknown_count` = number of triples whose predicate is rejected as unknown by validator-based classification
- `unvalidated_count` = number of triples left unclassified because validator-based classification is unavailable

`known_count + unknown_count + unvalidated_count` SHALL equal the number of triples in the observed slice.

When validator-based classification is unavailable, the runtime still emits one observe-only admission observation for the batch with:
- `known_count = 0`
- `unknown_count = 0`
- `unvalidated_count = len(observed slice)`
- `validator_source = "unavailable"`

This slice does not introduce new producer families or new source adapters. Runtime admission config is limited to mode selection (`off` or `observe`) plus fallback confidence defaults for the learning producer group and the librarian producer group. These settings are stored under the existing `ontology.governance.*` namespace for config compatibility, but they always remain directly visible and editable on the runtime admission settings surface rather than inheriting governance-enabled gating semantics.

In all cases, observe-only mode MUST NOT drop, rewrite, or reroute the original triple slice. The policy classifies that **original triple slice** against the shared predicate-validity source, emits event-bus admission telemetry tagged with stable producer-source values plus producer-group identifiers and validator-source tags, emits `content.saved` admission telemetry tagged with the synthetic `content_saved_extractor` telemetry source label and validator-source tag, preserves unknown event-bus source labels as `unmapped-source` signals instead of collapsing them, and then leaves the original triple slice to proceed through the same graph write operation it would have used without observe-only admission. Extractor dropped-unknown baselines and aggregate graph write-failure baselines remain separate telemetry families rather than admission decisions.
```

- [ ] **Step 4: Write the delta specs**

Use these deltas:

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/graph-store/spec.md -->
## ADDED Requirements

### Requirement: Observe-only admission for supported runtime graph inputs
The Phase A1 graph store observe slice SHALL compute admission decisions only for the supported runtime graph inputs in this slice: event-bus `TriplesExtractedEvent` batches whose `Source` is `conversation_analysis`, `session_learning`, `learning`, or `proactive_librarian`, plus triples already returned from the `content.saved` extraction path. Observe-only telemetry for that extraction path SHALL use the synthetic `content_saved_extractor` source label. Observe mode SHALL NOT drop, rewrite, or reroute the original triple slice before graph write execution.

For this slice, `conversation_analysis`, `session_learning`, and `learning` are the supported learning-group producer sources, `proactive_librarian` is the supported librarian-group producer source, and `content_saved_extractor` is a separate synthetic telemetry source label for the extraction path.

For every observed triple slice in this requirement:
- `batch_count` SHALL equal `1`
- `known_count + unknown_count + unvalidated_count` SHALL equal the number of triples in that slice

#### Scenario: Event-bus triple producer source is observed
- **WHEN** a supported `TriplesExtractedEvent` producer-source batch is processed in observe mode
- **THEN** the runtime SHALL compute an observe-only admission decision for the original triple slice
- **AND** that decision SHALL classify the slice into `known_count`, `unknown_count`, and `unvalidated_count`
- **AND** the runtime SHALL preserve the original triples unchanged for graph write execution

#### Scenario: Content-saved extraction source is observed
- **WHEN** the `content.saved` extraction path returns triples in observe mode
- **THEN** the runtime SHALL compute an observe-only admission decision for the original extracted triple slice
- **AND** that decision SHALL classify the slice into `known_count`, `unknown_count`, and `unvalidated_count`
- **AND** the runtime SHALL preserve the original extracted triples unchanged for graph write execution

#### Scenario: Unsupported event-bus source keeps the existing write path
- **WHEN** observe mode receives a `TriplesExtractedEvent` batch whose `Source` is outside the supported event-bus producer-source set for this slice
- **THEN** the runtime SHALL skip observe-only admission classification for that batch
- **AND** the runtime SHALL preserve the original triples unchanged for graph write execution

#### Scenario: Validator unavailable keeps observe-only admission in unvalidated observation mode
- **WHEN** ontology is disabled or ontology initialization fails before observe-only admission can classify a supported runtime graph input
- **THEN** the runtime SHALL still compute one batch-scoped observe-only admission decision for that graph input
- **AND** all triples in that slice SHALL contribute to `unvalidated_count`
- **AND** `known_count` and `unknown_count` SHALL both equal `0`
- **AND** the decision SHALL use `validator_source = "unavailable"`
- **AND** the runtime SHALL preserve the original triple slice unchanged for graph write execution
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/eventbus/spec.md -->
## ADDED Requirements

### Requirement: Graph admission-related event types are defined
The eventbus package SHALL define the event types required to represent observe-only graph admission telemetry and the related non-admission baseline signals for this slice.

#### Scenario: Event-bus admission event shape carries source, group, and validator fields
- **WHEN** a supported `TriplesExtractedEvent` producer-source batch is classified in observe mode
- **THEN** the graph-admission event shape SHALL include that producer-source identifier, its producer-group identifier, and the validator-source identifier used for predicate checks
- **AND** it SHALL use the stable validator-source value `ontology_registry` when classification uses the ontology service validator closure
- **AND** it SHALL include `batch_count`, `known_count`, `unknown_count`, and `unvalidated_count` fields for that observed slice

#### Scenario: Content-saved admission event shape carries synthetic source and validator fields
- **WHEN** the `content.saved` extraction path returns triples that are classified in observe mode
- **THEN** the graph-admission event shape SHALL include the synthetic `content_saved_extractor` source label and the validator-source identifier used for predicate checks
- **AND** it SHALL NOT synthesize an event-bus producer-group identifier for that telemetry
- **AND** it SHALL include `batch_count`, `known_count`, `unknown_count`, and `unvalidated_count` fields for that observed slice

#### Scenario: Validator-unavailable mode emits unvalidated admission observations
- **WHEN** ontology is disabled or ontology initialization fails before observe-only admission can classify a batch through the shared validator closure
- **THEN** the runtime SHALL emit an `unvalidated` graph-admission observation for that batch with `validator_source = "unavailable"`
- **AND** that observation SHALL preserve the batch-scoped `batch_count`, `known_count`, `unknown_count`, and `unvalidated_count` fields
- **AND** it SHALL aggregate that batch under the `unavailable` validator-source grouping key

### Requirement: Non-admission baseline events remain separate telemetry families
Extractor dropped-unknown baselines, `unmapped-source` signals, and aggregate graph write-failure baselines SHALL remain separate telemetry families rather than graph-admission telemetry.

#### Scenario: Unmapped event-bus source is surfaced explicitly
- **WHEN** observe mode receives a `TriplesExtractedEvent.Source` label that is outside the supported event-bus producer-source set and therefore not assigned to a known producer group in this slice
- **THEN** the runtime SHALL record an `unmapped-source` telemetry signal carrying the original raw source label
- **AND** it SHALL NOT widen observe-only admission classification beyond the supported runtime graph inputs in this slice

#### Scenario: Extractor dropped-unknown baseline stays pre-admission
- **WHEN** the `content.saved` extraction path rejects an unknown predicate before graph admission runs
- **THEN** the runtime SHALL record dropped-unknown baseline telemetry for the synthetic `content_saved_extractor` source label
- **AND** it SHALL NOT imply that graph admission dropped the triple

#### Scenario: Graph write-failure baseline stays aggregate
- **WHEN** a batched graph write fails while observe mode is enabled
- **THEN** the runtime SHALL record an aggregate graph write-failure baseline event for that failed batch
- **AND** it SHALL NOT require admission source, producer-group, or validator-source tags on that aggregate failure baseline
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/ontology-registry/spec.md -->
## ADDED Requirements

### Requirement: Shared predicate validity source
The runtime SHALL use the ontology service predicate validator closure as the primary predicate-validity source for observe-only admission decisions and graph-store validation.

#### Scenario: Graph admission and graph-store validation use the same validator closure
- **WHEN** observe-only admission and graph-store validation both perform predicate checks
- **THEN** the runtime SHALL obtain those checks from the same ontology service predicate validator closure

#### Scenario: Ontology init failure preserves existing graph validation behavior
- **WHEN** ontology is disabled or ontology initialization fails
- **THEN** graph-store validation SHALL continue to use the built-in hardcoded graph predicate validator
- **AND** observe-only admission SHALL switch to the stable validator-source value `unavailable` rather than blocking current graph writes
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/system-feedback/spec.md -->
## ADDED Requirements

### Requirement: Observe-only graph admission telemetry is emitted for runtime feedback
The runtime SHALL emit observe-only graph admission telemetry and non-admission baseline signals for the supported runtime graph inputs in this slice.

#### Scenario: Supported inputs emit telemetry on the runtime event bus
- **WHEN** observe-only admission processes a supported runtime graph input
- **THEN** event-bus admission telemetry SHALL preserve producer-source, producer-group, and validator-source identity
- **AND** `content_saved_extractor` telemetry SHALL preserve its synthetic source label and validator-source identity without inventing a producer-group

#### Scenario: Non-admission baseline signals remain distinct
- **WHEN** the runtime emits dropped-unknown, `unmapped-source`, or graph write-failure baseline feedback
- **THEN** those signals SHALL remain distinct from graph-admission telemetry
- **AND** they SHALL preserve the source identities required by their respective contracts

#### Scenario: Validator-unavailable mode emits unvalidated admission observations
- **WHEN** ontology is disabled or ontology initialization fails before observe-only admission can classify a batch through the shared validator closure
- **THEN** the runtime SHALL emit an `unvalidated` graph-admission observation for that batch with `validator_source = "unavailable"`
- **AND** that observation SHALL preserve the batch-scoped `batch_count`, `known_count`, `unknown_count`, and `unvalidated_count` fields
- **AND** it SHALL aggregate that batch under the `unavailable` validator-source grouping key

### Requirement: Observe-only graph admission metrics are aggregated into runtime feedback snapshots
The runtime feedback snapshot SHALL aggregate observe-only graph admission metrics into stable metric families for downstream surfaces.

#### Scenario: Admission batch metrics are aggregated by source and validator identity
- **WHEN** graph-admission telemetry is recorded for a supported runtime graph input
- **THEN** the runtime feedback snapshot SHALL aggregate one observed batch count for that telemetry event
- **AND** event-bus admission metrics SHALL remain grouped by producer-source, producer-group, and validator-source identity
- **AND** `content_saved_extractor` admission metrics SHALL remain grouped by the synthetic source label and validator-source identity without inventing a producer-group
- **AND** the snapshot SHALL aggregate `known_count`, `unknown_count`, and `unvalidated_count` totals from those batch events

#### Scenario: Non-admission baselines preserve their counting units
- **WHEN** dropped-unknown or graph write-failure baseline telemetry is recorded
- **THEN** extractor dropped-unknown metrics SHALL aggregate one dropped-triple count per rejected triple
- **AND** graph write-failure baseline metrics SHALL aggregate one failed-batch count per failed graph write attempt

#### Scenario: Unmapped and validator identities remain visible in snapshots
- **WHEN** the runtime feedback snapshot is built
- **THEN** unmapped-source metrics SHALL remain grouped by raw source label
- **AND** validator-source metrics SHALL remain grouped by validator-source identifier, including the stable `unavailable` value when classification could not run
- **AND** both groupings SHALL use batch counts rather than triple counts
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/cockpit-status-page/spec.md -->
## ADDED Requirements

### Requirement: Graph admission metrics are surfaced on the cockpit status page
The cockpit status page SHALL surface observe-only graph admission metrics from the runtime feedback snapshot, including graph-admission counts grouped by source and validator identity, extractor dropped-unknown baselines, unmapped-source counts, and aggregate graph write-failure baselines.

#### Scenario: Status page renders graph admission metrics
- **WHEN** the cockpit status page is rendered while observe mode metrics are available
- **THEN** it SHALL display event-bus graph-admission counts grouped by supported producer source and producer group, plus a separate grouped view for the synthetic `content_saved_extractor` source label
- **AND** it SHALL display validator-source as a grouping key on graph-admission metrics rather than as a separate independent metric family
- **AND** it SHALL display extractor dropped-unknown, unmapped-source, and aggregate graph write-failure baseline counts as distinct metrics
- **AND** it SHALL preserve raw `unmapped-source` identity by grouping those counts by raw source label
- **AND** it SHALL preserve validator-source identity by grouping those counts by validator-source identifier
- **AND** it SHALL display `known`, `unknown`, and `unvalidated` triple totals for graph-admission decisions
```

```markdown
<!-- openspec/changes/runtime-admission-boundary-hardening/specs/cli-settings/spec.md -->
## ADDED Requirements

### Requirement: Runtime admission settings
The settings surface SHALL expose runtime admission configuration under `ontology.governance.admissionMode`, `ontology.governance.learningDefaultConfidence`, and `ontology.governance.librarianDefaultConfidence`, with values `off` and `observe` for admission mode plus fallback confidence defaults of `0.60` for the learning producer group and `0.50` for the librarian producer group.

These fields SHALL use the existing `ontology.governance.*` config namespace for storage compatibility, but they SHALL always remain directly visible on the runtime admission settings surface rather than inheriting governance-enabled gating semantics.

#### Scenario: Runtime admission config is editable
- **WHEN** an operator edits runtime admission settings
- **THEN** the runtime admission mode SHALL be configurable as `off` or `observe`
- **AND** the learning producer group and librarian producer group SHALL each expose a fallback confidence default

#### Scenario: Runtime admission settings are not hidden behind governance-only gating
- **WHEN** the runtime admission settings surface is rendered
- **THEN** the runtime admission mode and both fallback confidence defaults SHALL remain directly editable within that settings surface
- **AND** the runtime SHALL NOT require a separate governance-enabled toggle before showing those fields

#### Scenario: No extra producer groups are implied
- **WHEN** the runtime admission settings surface is rendered
- **THEN** it SHALL scope fallback confidence defaults only to the learning producer group and the librarian producer group
- **AND** it SHALL NOT imply additional first-slice producer groups
```

- [ ] **Step 5: Record tasks and commit**

Use this content for `openspec/changes/runtime-admission-boundary-hardening/tasks.md`:

```markdown
# Tasks

## 1. Runtime Admission Config Surface

- [ ] Add the runtime admission mode config surface with `off` and `observe`.
  Affected artifacts: `specs/cli-settings/spec.md`; runtime config schema and settings wiring when implementation starts.
  Verification: `openspec validate runtime-admission-boundary-hardening --strict`; config serialization/defaulting coverage confirms the new mode values.
- [ ] Add fallback confidence defaults for the learning producer group and the librarian producer group without introducing extra first-slice producer groups.
  Affected artifacts: `specs/cli-settings/spec.md`; runtime config defaults when implementation starts.
  Verification: review confirms only the learning and librarian producer groups are named in this slice and that their defaults are fixed at `0.60` and `0.50`.

## 2. Graph Admission Classification Contract

- [ ] Add the observe-only graph admission policy contract for the supported runtime graph inputs in this slice.
  Affected artifacts: `specs/graph-store/spec.md`; admission policy implementation and runtime wiring when implementation starts.
  Verification: strict OpenSpec validation passes; contract review confirms write routing remains observe-only.
- [ ] Fix the observe-only admission decision taxonomy and counting units.
  Affected artifacts: `proposal.md`, `design.md`, `specs/graph-store/spec.md`, `specs/eventbus/spec.md`, `specs/system-feedback/spec.md`, `specs/cockpit-status-page/spec.md`.
  Verification: review confirms each observed slice is batch-scoped and carries `known_count`, `unknown_count`, and `unvalidated_count`, while baseline families retain their own batch/triple counting units.
- [ ] Define graph-admission telemetry contracts around stable event-bus producer-source identifiers, event-bus producer-group identifiers, the synthetic `content_saved_extractor` telemetry source label, and validator-source tags.
  Affected artifacts: `specs/eventbus/spec.md`; observability event types when implementation starts.
  Verification: review confirms telemetry distinguishes admission classification from downstream write failures.
- [ ] Define the aggregate graph write-failure baseline event contract as a separate downstream write-failure telemetry family.
  Affected artifacts: `specs/eventbus/spec.md`; observability event types when implementation starts.
  Verification: review confirms the aggregate write-failure baseline is not specified as carrying admission source, producer-group, or validator-source tags.

## 3. Supported Producer-Source Coverage

- [ ] Observe `TriplesExtractedEvent` batches only for the explicit event-bus producer-source set `conversation_analysis`, `session_learning`, `learning`, and `proactive_librarian`.
  Affected artifacts: `specs/graph-store/spec.md`; runtime event-bus subscribers when implementation starts.
  Verification: review confirms no other event-bus source labels are normalized into named producer groups in this slice.
- [ ] Observe `content.saved` extraction triples and the extractor dropped-unknown baseline under the separate synthetic `content_saved_extractor` telemetry source label without changing extractor behavior.
  Affected artifacts: `design.md`, `specs/graph-store/spec.md`, `specs/eventbus/spec.md`; extractor/graph wiring when implementation starts.
  Verification: review confirms dropped-unknown telemetry is source-scoped and separate from graph admission decisions.

## 4. Shared Validator Source And Observability Surfacing

- [ ] Reuse the ontology predicate validator closure as the shared predicate-validity source for observe-only admission and graph-store validation.
  Affected artifacts: `specs/ontology-registry/spec.md`; ontology/graph integration when implementation starts.
  Verification: review confirms one validator source is named across both paths and that ontology init failure preserves existing graph validation behavior by degrading observe-only admission to an `unvalidated` observation mode.
- [ ] Surface graph admission, dropped-unknown, unmapped-source, validator-source tag, and graph write-failure metrics in observability and cockpit status.
  Affected artifacts: `specs/eventbus/spec.md`, `specs/system-feedback/spec.md`, `specs/cockpit-status-page/spec.md`, proposal/design references in this change; observability and cockpit surfacing when implementation starts.
  Verification: review confirms event types carry the required identities; runtime feedback snapshots preserve the defined batch/triple counting units and grouped identities; and the cockpit status page renders those grouped metrics.

## 5. Change Documentation And Verification

- [ ] Keep proposal, design, and delta specs aligned on stable producer-source and config terminology.
  Affected artifacts: `proposal.md`, `design.md`, `specs/cli-settings/spec.md`, `specs/graph-store/spec.md`, `specs/ontology-registry/spec.md`.
  Verification: terminology review finds no relative scope wording that depends on transient wiring descriptions and no ambiguous fallback semantics for validator selection.
- [ ] Validate the change scaffolding before implementation work begins.
  Affected artifacts: the full `openspec/changes/runtime-admission-boundary-hardening/**` change set.
  Verification: run `openspec validate runtime-admission-boundary-hardening --strict`.
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
	assert.Equal(t, "off", cfg.Ontology.Governance.AdmissionMode)
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

func TestNewOntologyForm_AdmissionFieldsVisibleWithoutGovernanceEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Ontology.Enabled = true
	cfg.Ontology.Governance.Enabled = false

	form := NewOntologyForm(cfg)
	visible := map[string]bool{}
	for _, field := range form.VisibleFields() {
		visible[field.Key] = true
	}

	assert.True(t, visible["ontology_gov_admission_mode"])
	assert.True(t, visible["ontology_gov_learning_conf"])
	assert.True(t, visible["ontology_gov_librarian_conf"])
}
```

- [ ] **Step 2: Run the tests to verify failure**

Run:

```bash
go test ./internal/config ./internal/cli/settings -run 'OntologyAdmissionObserve|AdmissionFieldsVisibleWithoutGovernanceEnabled'
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
		AdmissionMode:              "off",
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
	Options:     []string{"off", "observe"},
	Description: "Disabled or observe-only runtime admission mode for graph triple producers",
	VisibleWhen: admissionVisible,
})
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_learning_conf",
	Label:       "    Learning Default Confidence",
	Type:        tuicore.InputText,
	Value:       fmt.Sprintf("%.2f", cfg.Ontology.Governance.LearningDefaultConfidence),
	Placeholder: "0.60",
	Description: "Fallback confidence for the learning producer group",
	VisibleWhen: admissionVisible,
})
form.AddField(&tuicore.Field{
	Key:         "ontology_gov_librarian_conf",
	Label:       "    Librarian Default Confidence",
	Type:        tuicore.InputText,
	Value:       fmt.Sprintf("%.2f", cfg.Ontology.Governance.LibrarianDefaultConfidence),
	Placeholder: "0.50",
	Description: "Fallback confidence for the librarian producer group",
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
go test ./internal/config ./internal/cli/settings -run 'OntologyAdmissionObserve|AdmissionFieldsVisibleWithoutGovernanceEnabled'
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
	ProducerContentSavedExtractor AdmissionProducer = "content_saved_extractor"
	ProducerUnknownSource        AdmissionProducer = "unknown_source"
	ProducerLibrarianEvent       AdmissionProducer = "proactive_librarian"
	ProducerGraphEngine          AdmissionProducer = "graph_engine"
)

type AdmissionDecision string

const (
	DecisionObservedKnown       AdmissionDecision = "observed_known"
	DecisionObservedUnknown     AdmissionDecision = "observed_unknown"
	DecisionObservedUnvalidated AdmissionDecision = "observed_unvalidated"
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
	RawSource       string
	Candidates      int
	ObservedKnown   int
	ObservedUnknown int
	ObservedTypeHints int
	ValidatorSource string
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
		result.Event.RawSource = candidates[0].Source
	}
	result.Event.Candidates = len(candidates)
	result.Event.ValidatorSource = "ontology_predicate_validator_closure"

	for _, c := range candidates {
		conf := p.defaultConfidence(c)
		decision := DecisionObservedUnknown
		if p.cfg.Validator == nil {
			decision = DecisionObservedUnvalidated
			result.Event.ValidatorSource = "missing_validator"
			result.Event.ObservedUnknown++
		} else if p.cfg.Validator(c.Triple.Predicate) {
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

	if result.Event.ValidatorSource == "" {
		result.Event.ValidatorSource = "ontology_predicate_validator_closure"
	}

	p.logger.Infow("graph admission observe batch",
		"producer", result.Event.Producer,
		"candidates", result.Event.Candidates,
		"known", result.Event.ObservedKnown,
		"unknown", result.Event.ObservedUnknown,
		"validatorSource", result.Event.ValidatorSource,
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
	_ = policy.ObserveBatch(candidates)
	return triples
}
```

And add this event:

```go
// internal/eventbus/observability_events.go
const EventGraphAdmissionBatch = "graph.admission.batch"
const EventGraphAdmissionWriteFailure = "graph.admission.write_failure"
const EventGraphExtractorDrop = "graph.extractor.drop"

type GraphAdmissionBatchEvent struct {
	Producer        string
	RawSource       string
	Candidates      int
	ObservedKnown   int
	ObservedUnknown int
	ObservedTypeHints int
	ValidatorSource string
}

func (e GraphAdmissionBatchEvent) EventName() string { return EventGraphAdmissionBatch }

type GraphAdmissionWriteFailureEvent struct {
	Source     string
	Candidates int
	ErrorClass string
}

func (e GraphAdmissionWriteFailureEvent) EventName() string { return EventGraphAdmissionWriteFailure }

type GraphExtractorDropEvent struct {
	SourceID  string
	Predicate string
	Subject   string
	Object    string
	Source    string
}

func (e GraphExtractorDropEvent) EventName() string { return EventGraphExtractorDrop }
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

func TestObserveExtractedContentTriples_PreservesReturnedTriples(t *testing.T) {
	p := graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Validator: func(name string) bool { return name == graph.CausedBy },
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerContentSavedExtractor: 0.60,
		},
	}, zap.NewNop().Sugar())

	triples := graph.ObserveTriples(p, graph.ProducerContentSavedExtractor, "content_saved_extractor", []graph.Triple{{
		Subject: "a", Predicate: graph.CausedBy, Object: "b",
	}})

	require.Len(t, triples, 1)
	assert.Equal(t, graph.CausedBy, triples[0].Predicate)
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
go test ./internal/graph -run 'ObserveExtractedContentTriples|EmitDroppedUnknownObservation'
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
	predicateValidator graph.PredicateValidatorFunc
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

func publishAdmissionObservation(bus *eventbus.Bus, rawSource string, evt graph.AdmissionBatchEvent) {
	if bus == nil {
		return
	}
	bus.Publish(eventbus.GraphAdmissionBatchEvent{
		Producer:        string(evt.Producer),
		RawSource:       rawSource,
		Candidates:      evt.Candidates,
		ObservedKnown:   evt.ObservedKnown,
		ObservedUnknown: evt.ObservedUnknown,
		ObservedTypeHints: evt.ObservedTypeHints,
		ValidatorSource: evt.ValidatorSource,
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
// internal/app/modules.go immediately before initKnowledge(...)
if gc != nil && ontologyResult != nil && ontologyResult.Service != nil && cfg.Ontology.Enabled {
	gc.predicateValidator = ontologyResult.Service.PredicateValidator()
}
if gc != nil && gc.predicateValidator != nil && cfg.Ontology.Governance.AdmissionMode == "observe" {
	gc.admissionPolicy = graph.NewAdmissionPolicy(graph.AdmissionConfig{
		Validator: gc.predicateValidator,
		DefaultConfidence: map[graph.AdmissionProducer]float64{
			graph.ProducerConversationAnalysis: cfg.Ontology.Governance.LearningDefaultConfidence,
			graph.ProducerSessionLearning:      cfg.Ontology.Governance.LearningDefaultConfidence,
			graph.ProducerLibrarianEvent:       cfg.Ontology.Governance.LibrarianDefaultConfidence,
			graph.ProducerGraphEngine:          cfg.Ontology.Governance.LearningDefaultConfidence,
			graph.ProducerUnknownSource:        cfg.Ontology.Governance.LearningDefaultConfidence,
		},
		Observe: func(evt graph.AdmissionBatchEvent) {
			publishAdmissionObservation(m.bus, evt.RawSource, evt)
		},
	}, logger())
}

kc, kcStatus := initKnowledge(cfg, store, gc, m.bus, brokerAPI)
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
	SourceTag string
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

func (e *Extractor) Extract(ctx context.Context, content, sourceID, sourceTag string) ([]Triple, error) {
	// existing behavior plus sourceTag forwarded into parseResponse
}

// inside parseResponse unknown-predicate branch
if !e.isValidPredicate(predicate) {
	if e.onDroppedUnknown != nil {
	e.onDroppedUnknown(DroppedUnknownPredicateEvent{
			SourceID:  sourceID,
			SourceTag: sourceTag,
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
		bus.Publish(eventbus.GraphExtractorDropEvent{
			SourceID:  evt.SourceID,
			Predicate: evt.Predicate,
			Subject:   evt.Subject,
			Object:    evt.Object,
			Source:    evt.SourceTag,
		})
	}),
	graph.WithPredicateValidator(gc.predicateValidator),
)

// inside content.saved extraction goroutine after extractor.Extract(...)
sourceTag := fmt.Sprintf("content_saved_extractor:%s:%s", evt.Source, evt.Collection)
triples, err := extractor.Extract(ctx, evt.Content, evt.ID, sourceTag)
observed := graph.ObserveTriples(gc.admissionPolicy, graph.ProducerContentSavedExtractor, sourceTag, triples)
if len(observed) > 0 {
	gc.buffer.Enqueue(graph.GraphRequest{Triples: observed})
}
```

- [ ] **Step 4: Run the tests again**

Run:

```bash
go test ./internal/app -run 'ProducerForExtractedEvent|ObserveExtractedTriples'
go test ./internal/graph -run 'ObserveExtractedContentTriples|EmitDroppedUnknownObservation'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/app/modules.go internal/app/wiring_graph.go internal/app/wiring_graph_test.go internal/graph/extractor.go internal/graph/extractor_test.go
git commit -m "feat: observe runtime graph event producers"
```

## Task 5: Wire Telemetry Consumers And GraphBuffer Failure Baselines

**Files:**
- Modify: `internal/app/wiring_observability.go`
- Modify: `internal/observability/types.go`
- Modify: `internal/observability/collector.go`
- Modify: `internal/graph/buffer.go`
- Modify: `internal/cli/cockpit/pages/status.go`
- Test: `internal/observability/collector_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests like these:

```go
func TestCollector_RecordGraphAdmissionBatch(t *testing.T) {
	c := NewCollector()
	c.RecordGraphAdmissionBatch("conversation_analysis", "conversation_analysis", 3, 2, 1, 1, "ontology_predicate_validator_closure")
	c.RecordGraphExtractorDrop()
	c.RecordGraphAdmissionWriteFailure(3)
	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.GraphAdmission.Batches)
	assert.Equal(t, int64(3), snap.GraphAdmission.Candidates)
	assert.Equal(t, int64(1), snap.GraphAdmission.ObservedUnknown)
	assert.Equal(t, int64(1), snap.GraphAdmission.ObservedTypeHints)
	assert.Equal(t, int64(1), snap.GraphAdmission.DroppedUnknown)
	assert.Equal(t, int64(1), snap.GraphAdmission.WriteFailures)
	assert.Equal(t, int64(3), snap.GraphAdmission.WriteFailureCandidates)
}
```

- [ ] **Step 2: Run the tests to verify failure**

Run:

```bash
go test ./internal/observability -run 'RecordGraphAdmissionBatch'
```

Expected: FAIL because the graph admission metrics and write-failure baseline do not exist

- [ ] **Step 3: Implement telemetry consumers and failure baseline**

Apply changes like these:

```go
// internal/observability/types.go
type GraphAdmissionSummary struct {
	Batches         int64
	Candidates      int64
	ObservedKnown   int64
	ObservedUnknown int64
	ObservedTypeHints int64
	UnmappedSources int64
	UnknownSourceByName map[string]int64
	DroppedUnknown int64
	WriteFailures int64
	WriteFailureCandidates int64
	ValidatorSource string
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

func NewCollector() *MetricsCollector {
	return &MetricsCollector{
		// existing fields...
		graphAdmission: GraphAdmissionSummary{
			UnknownSourceByName: map[string]int64{},
		},
	}
}

func (c *MetricsCollector) RecordGraphAdmissionBatch(producer, rawSource string, candidates, observedKnown, observedUnknown, observedTypeHints int, validatorSource string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graphAdmission.Batches++
	c.graphAdmission.Candidates += int64(candidates)
	c.graphAdmission.ObservedKnown += int64(observedKnown)
	c.graphAdmission.ObservedUnknown += int64(observedUnknown)
	c.graphAdmission.ObservedTypeHints += int64(observedTypeHints)
	if producer == "unknown_source" {
		c.graphAdmission.UnmappedSources++
		c.graphAdmission.UnknownSourceByName[rawSource]++
	}
	c.graphAdmission.ValidatorSource = validatorSource
}

func (c *MetricsCollector) RecordGraphExtractorDrop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graphAdmission.DroppedUnknown++
}

func (c *MetricsCollector) RecordGraphAdmissionWriteFailure(candidates int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graphAdmission.WriteFailures++
	c.graphAdmission.WriteFailureCandidates += int64(candidates)
}

// Extend the existing Snapshot() implementation to copy c.graphAdmission.
// Extend Reset() to zero c.graphAdmission.
```

```go
// internal/app/wiring_observability.go
eventbus.SubscribeTyped[eventbus.GraphAdmissionBatchEvent](bus, func(evt eventbus.GraphAdmissionBatchEvent) {
	oc.collector.RecordGraphAdmissionBatch(
		evt.Producer,
		evt.RawSource,
		evt.Candidates,
		evt.ObservedKnown,
		evt.ObservedUnknown,
		evt.ObservedTypeHints,
		evt.ValidatorSource,
	)
})

eventbus.SubscribeTyped[eventbus.GraphExtractorDropEvent](bus, func(evt eventbus.GraphExtractorDropEvent) {
	oc.collector.RecordGraphExtractorDrop()
})

eventbus.SubscribeTyped[eventbus.GraphAdmissionWriteFailureEvent](bus, func(evt eventbus.GraphAdmissionWriteFailureEvent) {
	oc.collector.RecordGraphAdmissionWriteFailure(evt.Candidates)
})
```

```go
// internal/graph/buffer.go
type GraphWriteFailureObserver func(count int, err error)

type GraphBuffer struct {
	store                Store
	inner                *asyncbuf.BatchBuffer[GraphRequest]
	logger               *zap.SugaredLogger
	writeFailureObserver GraphWriteFailureObserver
}

func (b *GraphBuffer) SetWriteFailureObserver(fn GraphWriteFailureObserver) {
	b.writeFailureObserver = fn
}

func (b *GraphBuffer) processBatch(batch []Triple) {
	ctx := context.Background()
	if err := b.store.AddTriples(ctx, batch); err != nil {
		if b.writeFailureObserver != nil {
			b.writeFailureObserver(len(batch), err)
		}
		b.logger.Errorw("batch graph update error", "count", len(batch), "error", err)
	}
}
```

```go
// internal/app/modules.go after graph store / observability are available
if gc != nil && gc.buffer != nil && m.bus != nil && cfg.Ontology.Governance.AdmissionMode == "observe" {
	gc.buffer.SetWriteFailureObserver(func(count int, err error) {
		m.bus.Publish(eventbus.GraphAdmissionWriteFailureEvent{
			Source:     "graph_buffer_batch",
			Candidates: count,
			ErrorClass: "graph_store_add_triples_failed",
		})
	})
}
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
	b.WriteString(fmt.Sprintf("Unknown source labels: %v\n", ga.UnknownSourceByName))
	b.WriteString(fmt.Sprintf("Write failures: %d\n", ga.WriteFailures))
	b.WriteString(fmt.Sprintf("Failed candidates: %d\n", ga.WriteFailureCandidates))
	b.WriteString(fmt.Sprintf("Validator source: %s\n", ga.ValidatorSource))
	return b.String()
}
```

- [ ] **Step 4: Run the tests again**

Run:

```bash
go test ./internal/observability -run 'RecordGraphAdmissionBatch'
```

Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/app/wiring_observability.go internal/observability/types.go internal/observability/collector.go internal/observability/collector_test.go internal/cli/cockpit/pages/status.go
git add internal/graph/buffer.go internal/app/modules.go
git commit -m "feat: add graph admission telemetry baselines"
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
| `ontology.governance.admissionMode` | `string` | `off` | Disabled or observe-only runtime admission mode for the supported runtime graph producer sources |
| `ontology.governance.learningDefaultConfidence` | `float64` | `0.60` | Fallback confidence for the learning producer group |
| `ontology.governance.librarianDefaultConfidence` | `float64` | `0.50` | Fallback confidence for the librarian producer group |
```

```md
<!-- docs/features/knowledge-graph.md -->
Observe-only graph admission now records admission decisions for the supported runtime graph producer sources before the existing graph write path runs. This phase does not drop or rewrite triples yet.
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
