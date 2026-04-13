## Purpose

Capability spec for llm-text-generator. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Single TextGenerator interface in internal/llm
The `TextGenerator` interface SHALL be defined exactly once in `internal/llm/text_generator.go`. All consumers (learning, memory, graph, librarian) MUST import from `internal/llm/`.

#### Scenario: No duplicate TextGenerator definitions
- **WHEN** searching for `type TextGenerator interface` across the codebase
- **THEN** exactly one definition exists in `internal/llm/text_generator.go`

#### Scenario: Consumer packages use llm.TextGenerator
- **WHEN** `internal/learning/`, `internal/memory/`, `internal/graph/`, `internal/librarian/` reference TextGenerator
- **THEN** they import `llm.TextGenerator` from `internal/llm/`
