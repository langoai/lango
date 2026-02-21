## ADDED Requirements

### Requirement: Valid and Values methods on existing enums
The system SHALL add `Valid() bool` and `Values() []T` methods to all existing typed enum types.

#### Scenario: ApprovalPolicy enum methods
- **WHEN** `config/types.go` defines `ApprovalPolicy`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: PIICategory enum methods
- **WHEN** `agent/pii_pattern.go` defines `PIICategory`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: KeyType enum methods
- **WHEN** `security/key_registry.go` defines `KeyType`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: SectionID enum methods
- **WHEN** `prompt/section.go` defines `SectionID`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: StreamEventType enum methods
- **WHEN** `provider/provider.go` defines `StreamEventType`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: SafetyLevel enum methods
- **WHEN** `agent/runtime.go` defines `SafetyLevel`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: Background Status enum methods
- **WHEN** `background/task.go` defines `Status`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: ContextLayer enum methods
- **WHEN** `knowledge/types.go` defines `ContextLayer`
- **THEN** it SHALL have `Valid()` and `Values()` methods

### Requirement: Package-local enum types
The system SHALL convert untyped string constants to typed enums with `Valid()`/`Values()` in their respective packages.

#### Scenario: graph.Predicate typed enum
- **WHEN** `graph/store.go` defines predicate constants
- **THEN** they SHALL be typed as `Predicate string` with `Valid()` and `Values()`

#### Scenario: knowledge.KnowledgeCategory typed enum
- **WHEN** `knowledge/types.go` defines category constants
- **THEN** they SHALL be typed as `KnowledgeCategory string` with `Valid()` and `Values()`

#### Scenario: skill.SkillStatus and SkillType typed enums
- **WHEN** `skill/types.go` defines status and type constants
- **THEN** they SHALL be typed enums with `Valid()` and `Values()`
