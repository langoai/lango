## Purpose

Capability spec for structured-findings-taxonomy. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Shared category mappers
The `knowledge` package SHALL provide exported `MapKnowledgeCategory(string) (entknowledge.Category, error)` and `MapLearningCategory(string) (entlearning.Category, error)` functions as the single source of truth for category mapping. Both SHALL return error for unrecognized types (case-sensitive, no fallback).

#### Scenario: Valid knowledge category mapping
- **WHEN** `MapKnowledgeCategory` is called with any of "rule", "definition", "preference", "fact", "pattern", "correction"
- **THEN** the corresponding `entknowledge.Category` value SHALL be returned with nil error

#### Scenario: Invalid knowledge category
- **WHEN** `MapKnowledgeCategory` is called with an unrecognized string (e.g., "FACT", "unknown")
- **THEN** an empty category and error containing "unrecognized knowledge type" SHALL be returned

#### Scenario: Valid learning category mapping
- **WHEN** `MapLearningCategory` is called with any of "correction", "pattern", "tool_error", "provider_error", "timeout", "permission"
- **THEN** the corresponding `entlearning.Category` value SHALL be returned with nil error

#### Scenario: Invalid learning category
- **WHEN** `MapLearningCategory` is called with an unrecognized string
- **THEN** an empty category and error containing "unrecognized learning type" SHALL be returned

### Requirement: Temporal classification tags
All LLM knowledge extraction prompts SHALL request a `temporal` field with values "evergreen" (always-true knowledge) or "current_state" (may change over time). When temporal is non-empty, it SHALL be stored as a tag `"temporal:<value>"` on the KnowledgeEntry.

#### Scenario: Evergreen fact tagged
- **WHEN** an analyzer extracts a fact with `temporal: "evergreen"`
- **THEN** the saved KnowledgeEntry SHALL have `"temporal:evergreen"` in its Tags slice

#### Scenario: Current-state fact tagged
- **WHEN** an analyzer extracts a fact with `temporal: "current_state"`
- **THEN** the saved KnowledgeEntry SHALL have `"temporal:current_state"` in its Tags slice

#### Scenario: Empty temporal skipped
- **WHEN** an analyzer extracts a fact with empty temporal field
- **THEN** no temporal tag SHALL be added to the KnowledgeEntry

### Requirement: Dual-save for pattern/correction
When ANY knowledge ingestion path encounters type `pattern` or `correction`, the system SHALL save the entry as BOTH a Knowledge entry (primary) AND a Learning entry (backward compat). This applies to ConversationAnalyzer, SessionLearner, ProactiveBuffer, and InquiryProcessor.

#### Scenario: Pattern dual-saved from conversation analysis
- **WHEN** ConversationAnalyzer extracts a result with type "pattern"
- **THEN** it SHALL create a Knowledge entry with category "pattern"
- **AND** it SHALL create a Learning entry with category "general"

#### Scenario: Correction dual-saved from librarian
- **WHEN** ProactiveBuffer extracts a correction
- **THEN** it SHALL create a Knowledge entry with category "correction"
- **AND** it SHALL create a Learning entry with category "user_correction" and fix field populated

### Requirement: 6-category standardized prompts
ALL LLM analyzer prompts SHALL request all 6 knowledge types: "rule", "definition", "preference", "fact", "pattern", "correction". No analyzer SHALL request a subset.

#### Scenario: ConversationAnalyzer prompt uses all 6 types
- **WHEN** the ConversationAnalyzer prompt is examined
- **THEN** the "type" field SHALL list all 6 categories

#### Scenario: InquiryProcessor prompt uses all 6 types
- **WHEN** the InquiryProcessor answer detection prompt is examined
- **THEN** the knowledge category field SHALL include "pattern" and "correction" in addition to the original 4

### Requirement: Learning mapper error return
`mapLearningCategory()` SHALL return `(entlearning.Category, error)` instead of silently defaulting to `CategoryGeneral` for unrecognized types.

#### Scenario: Unknown learning type returns error
- **WHEN** `mapLearningCategory` is called with an unrecognized type
- **THEN** it SHALL return an empty category and a non-nil error
- **AND** it SHALL NOT silently return `CategoryGeneral`
