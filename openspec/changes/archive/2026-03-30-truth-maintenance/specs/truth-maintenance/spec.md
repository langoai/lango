## ADDED Requirements

### Requirement: Bi-temporal metadata on asserted triples
The system SHALL set `_valid_from`, `_valid_to`, `_recorded_at`, `_recorded_by`, `_source`, and `_confidence` metadata fields on all triples stored via `AssertFact`. Existing triples without these fields SHALL be treated as always-valid (backward compatible).

#### Scenario: Assert a fact with temporal metadata
- **WHEN** `AssertFact(ctx, AssertionInput{Triple: {Subject: "a", Predicate: "related_to", Object: "b"}, Source: "manual", Confidence: 0.9})` is called
- **THEN** the stored triple's metadata contains `_valid_from`, `_recorded_at`, `_recorded_by`, `_source="manual"`, `_confidence="0.9000"`

### Requirement: Cardinality-based conflict detection
The system SHALL detect conflicts when a OneToOne or ManyToOne predicate has an existing valid triple with a different object for the same subject-predicate pair. OneToMany and ManyToMany predicates SHALL NOT trigger conflicts.

#### Scenario: OneToOne conflict
- **GIVEN** a OneToOne predicate "primary_cause" and an existing valid triple `error:x | primary_cause | cause_a`
- **WHEN** `AssertFact` is called with `error:x | primary_cause | cause_b`
- **THEN** a conflict record is created with status "open" and both triples are stored

#### Scenario: ManyToMany no conflict
- **GIVEN** a ManyToMany predicate "caused_by" and existing triple `err:1 | caused_by | tool:a`
- **WHEN** `AssertFact` is called with `err:1 | caused_by | tool:b`
- **THEN** no conflict is created

### Requirement: Source precedence auto-resolution
When a higher-precedence source asserts a conflicting fact, the system SHALL auto-resolve by retracting the existing fact and creating an `auto_resolved` conflict record for audit trail. Precedence: manual(10) > knowledge(8) > correction(7) > llm_extraction(4) > graph_engine(3) > memory_hook(2).

#### Scenario: Auto-resolve by source precedence
- **GIVEN** an existing OneToOne triple from source "llm_extraction"
- **WHEN** a new triple with different object from source "manual" is asserted
- **THEN** the existing triple is retracted, an `auto_resolved` conflict record is created, and no open conflict remains

### Requirement: Fact retraction
`RetractFact` SHALL set `_valid_to` to the current time on the matching triple (soft delete). The retracted triple remains in the store for historical queries via `FactsAt`.

#### Scenario: Retract and time-travel query
- **GIVEN** a triple asserted with ValidFrom=1 hour ago
- **WHEN** `RetractFact` is called, then `FactsAt(subject, now)` and `FactsAt(subject, 30min ago)` are queried
- **THEN** `FactsAt(now)` excludes the triple, `FactsAt(30min ago)` includes it

### Requirement: Conflict resolution
`ResolveConflict` SHALL retract all candidate triples except the winner and mark the conflict as "resolved".

### Requirement: Predicate validation
`AssertFact` SHALL call `ValidateTriple` before any other operation. Unknown or deprecated predicates SHALL be rejected.
