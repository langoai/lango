## MODIFIED Requirements

### Requirement: Learning suggestions render as actionable proposed missions

Mission Control SHALL back proposed missions with a transient proposal registry instead of rendering raw learning-buffer rows directly. In this Wave 3 slice, `LearningSuggestionEvent` is the only active proposal producer. Proposed missions SHALL remain transient and SHALL move through explicit proposal states such as `suggested`, `preparing`, and `prepared` before acceptance or dismissal.

#### Scenario: Learning suggestion becomes a transient proposal record
- **WHEN** a `LearningSuggestionEvent` is emitted for the current session
- **THEN** Mission Control SHALL create or update one transient proposal record using the suggestion as the proposal source
- **AND** the proposal SHALL NOT create a durable `Mission` row before user acceptance

#### Scenario: Proposed mission no longer depends on raw learning-buffer row shape
- **WHEN** Mission Control renders active proposals
- **THEN** the page SHALL read proposal state from the transient proposal registry
- **AND** the learning buffer MAY remain only as a producer input or compatibility source, not the primary rendered proposal state

### Requirement: Proposed missions can prepare a deterministic brief before acceptance

Mission Control SHALL support deterministic low-risk preparation for transient proposals in this slice. Preparation SHALL generate a prepared brief from source-native evidence already available on the proposal source and SHALL NOT require broad heuristic agent work.

#### Scenario: Proposal reaches prepared with deterministic source-native brief
- **WHEN** a learning-based proposal is prepared
- **THEN** the proposal SHALL move into `prepared`
- **AND** the prepared brief SHALL be derived from source event fields such as proposed rule, rationale, confidence, and other stable source metadata
- **AND** the slice SHALL NOT require an LLM call to generate that prepared brief

#### Scenario: Generic proposal-owned execution remains out of scope
- **WHEN** a proposal moves through `suggested`, `preparing`, or `prepared`
- **THEN** the slice SHALL NOT require launching generic proposal-owned background execution or generic proposal-owned `RunLedger` execution
- **AND** the prepared state SHALL still be satisfiable through deterministic brief generation alone

## ADDED Requirements

### Requirement: Accepting a prepared proposal preserves prepared context in the durable mission path

Accepting a prepared proposal SHALL create the first durable mission row while preserving the prepared brief context that justified the proposal. `proposed` itself SHALL remain transient; the durable mission row SHALL still begin at `prepared` or `active`.

#### Scenario: Prepared proposal acceptance creates durable mission without losing prepared brief
- **WHEN** the user accepts a prepared proposal from Mission Control
- **THEN** the application SHALL create a durable mission through the mission lifecycle write path
- **AND** the resulting durable mission SHALL begin at `prepared` or `active`
- **AND** the prepared brief context SHALL remain available on the resulting mission path instead of being discarded during acceptance

### Requirement: Non-learning proposal producers remain explicitly disabled in this slice

Wave 3 SHALL keep the first proactive slice narrowly scoped. Librarian-gap producers, runtime-failure producers, and other future proposal sources SHALL remain explicit non-goals until they have dedicated adapters and source contracts.

#### Scenario: Librarian and runtime-failure producers do not create proposals in this slice
- **WHEN** librarian gap/inquiry signals or runtime failure signals exist
- **THEN** the Wave 3 first slice SHALL NOT require those signals to create transient proposals
- **AND** Mission Control SHALL NOT imply that those producers are active before their dedicated adapters exist
