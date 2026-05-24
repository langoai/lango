## Purpose

Capability spec for relevance-auto-adjust. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: RelevanceAdjuster service
The system SHALL provide a `RelevanceAdjuster` that subscribes to `ContextInjectedEvent` and adjusts knowledge `relevance_score` based on injection history. It SHALL support shadow mode (log only) and active mode (write to DB). It SHALL have a configurable warmup period before adjustments begin (process-local counter).

#### Scenario: Shadow mode — no writes
- **WHEN** mode="shadow" and items are injected
- **THEN** the adjuster SHALL log what would be adjusted but NOT write to the database

#### Scenario: Active mode — boost injected items
- **WHEN** mode="active", warmup complete, and user_knowledge items are injected
- **THEN** the adjuster SHALL call BoostRelevanceScore for each unique key with boostDelta

#### Scenario: Warmup period
- **WHEN** turnCount <= warmupTurns
- **THEN** no adjustments SHALL occur regardless of mode

#### Scenario: Turn-level dedup
- **WHEN** the same key appears multiple times in one ContextInjectedEvent
- **THEN** it SHALL be boosted once only

#### Scenario: Skip non-knowledge items
- **WHEN** items have Layer != "user_knowledge"
- **THEN** they SHALL NOT be boosted

### Requirement: Global periodic decay
Every `decayInterval` turns, the adjuster SHALL subtract `decayDelta` from all latest-version knowledge entries globally (cross-session). Floor at minScore to prevent undershoot. Order: decay fires before boost in the same turn.

#### Scenario: Periodic decay runs before same-turn boosts
- **WHEN** a turn lands on the configured decay interval
- **THEN** the adjuster SHALL decay all latest-version knowledge scores before applying any boosts for that turn

### Requirement: Rollback toggle
Setting mode from "active" to "shadow" SHALL immediately stop all DB writes. `ResetAllRelevanceScores` SHALL set all latest-version scores to 1.0 for hard rollback.

#### Scenario: Switching to shadow mode stops database writes
- **WHEN** auto-adjust mode changes from `active` to `shadow`
- **THEN** the adjuster SHALL stop writing relevance changes to the database immediately

#### Scenario: ResetAllRelevanceScores performs hard rollback
- **WHEN** `ResetAllRelevanceScores` is invoked
- **THEN** all latest-version relevance scores SHALL be reset to `1.0`

### Requirement: AutoAdjustConfig
`RetrievalConfig` SHALL include `AutoAdjust AutoAdjustConfig` with fields: Enabled, Mode, BoostDelta, DecayDelta, DecayInterval, MinScore, MaxScore, WarmupTurns.

#### Scenario: RetrievalConfig exposes auto-adjust knobs
- **WHEN** retrieval configuration is loaded
- **THEN** it SHALL expose the documented `AutoAdjustConfig` fields for enablement, mode, deltas, interval, score bounds, and warmup turns

### Requirement: RelevanceStore interface
The `retrieval` package SHALL define `RelevanceStore` interface with: BoostRelevanceScore, DecayAllRelevanceScores, ResetAllRelevanceScores. `*knowledge.Store` SHALL satisfy it.

#### Scenario: knowledge.Store satisfies RelevanceStore
- **WHEN** `*knowledge.Store` is used by the relevance adjuster
- **THEN** it SHALL satisfy the `RelevanceStore` interface
