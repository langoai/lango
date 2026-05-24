## Purpose

Capability spec for dynamic-budget-reallocation. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: SectionTokens type
The `adk` package SHALL provide a `SectionTokens` struct with fields `Knowledge`, `RAG`, `Memory`, `RunSummary` (all `int`), representing measured token counts per section before truncation.

#### Scenario: SectionTokens carries per-section measurements
- **WHEN** pre-truncation section token counts are measured
- **THEN** the system SHALL represent those counts using `SectionTokens.Knowledge`, `SectionTokens.RAG`, `SectionTokens.Memory`, and `SectionTokens.RunSummary`

### Requirement: ReallocateBudgets method
`ContextBudgetManager` SHALL provide `ReallocateBudgets(measured SectionTokens) SectionBudgets` that implements empty-section redistribution. Sections with measured=0 SHALL donate their entire budget proportionally (by original ratio) to non-empty sections. Non-empty sections SHALL keep their full initial budget plus a proportional share of the surplus. No recursive redistribution SHALL occur.

#### Scenario: All sections present — no reallocation
- **WHEN** all measured values are > 0
- **THEN** budgets SHALL equal the base SectionBudgets (no change)

#### Scenario: One section empty — surplus redistributed
- **WHEN** RAG measured=0 and others > 0
- **THEN** RAG budget SHALL be 0, and Knowledge/Memory/RunSummary SHALL receive proportional shares of the donated budget

#### Scenario: All sections empty — all-zero budgets
- **WHEN** all measured values are 0
- **THEN** all budget fields SHALL be 0 and Degraded SHALL be false

#### Scenario: Degraded passthrough
- **WHEN** base SectionBudgets is degraded
- **THEN** ReallocateBudgets SHALL return the degraded result unchanged

### Requirement: Headroom policy
Headroom (10% allocation) SHALL NOT participate in empty-section redistribution. It SHALL remain as a safety margin throughout.

#### Scenario: Headroom remains untouched during redistribution
- **WHEN** one or more content sections are empty and budgets are reallocated
- **THEN** the 10% headroom allocation SHALL remain reserved
- **AND** it SHALL NOT be donated to non-empty sections

### Requirement: Two-phase GenerateContent
`GenerateContent` SHALL use a two-phase pipeline: Phase 1 retrieves all section data in parallel without budget truncation, Phase 2 measures actual content and calls `ReallocateBudgets`, and Phase 3 SHALL truncate and format each section with reallocated budgets.

#### Scenario: GenerateContent measures before truncating
- **WHEN** `GenerateContent` assembles contextual sections
- **THEN** it SHALL first retrieve section data without token truncation
- **AND** SHALL measure that data before applying reallocated budgets in the formatting phase

### Requirement: Memory pre-limits preserved
Phase 1 retrieval for memory SHALL continue to respect `maxReflections` and `maxObservations` item count limits. These are not token budgets and SHALL remain enforced.

#### Scenario: Memory retrieval still respects item-count caps
- **WHEN** memory data is retrieved during Phase 1
- **THEN** `maxReflections` and `maxObservations` SHALL still cap the number of retrieved items
- **AND** reallocation SHALL NOT bypass those caps

### Requirement: Assembly function split
RAG, GraphRAG, Memory, and RunSummary assembly functions SHALL be split into separate retrieve and format functions. The original monolithic functions SHALL be preserved as thin wrappers for backward compatibility.

| Section | Retrieve | Format |
|---------|----------|--------|
| Memory | retrieveMemoryData | formatMemorySection |
| RAG | retrieveRAGData | formatRAGSection |
| GraphRAG | retrieveGraphRAGData | formatGraphRAGSection |
| RunSummary | retrieveRunSummaryData | formatRunSummarySection |

#### Scenario: Legacy wrappers delegate to retrieve and format helpers
- **WHEN** existing monolithic assembly entry points are called
- **THEN** they SHALL delegate through the new retrieve and format helper split
- **AND** observable behavior SHALL remain backward compatible

### Requirement: Token estimation helpers
The `adk` package SHALL provide token estimation functions for pre-assembly measurement: `estimateKnowledgeTokens`, `estimateRAGResultTokens`, `estimateMemoryTokens`, and `estimateRunSummaryTokens`. All SHALL use `types.EstimateTokens()` with approximate header overhead.

#### Scenario: Token estimation helpers use shared token estimation
- **WHEN** section token counts are estimated before formatting
- **THEN** the helper functions SHALL use `types.EstimateTokens()`
- **AND** SHALL include approximate header overhead in their estimates
