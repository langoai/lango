## Purpose

Default per-agent IDENTITY.md prompt files for all 7 sub-agents, embedded in `prompts/agents/<name>/IDENTITY.md`. These provide reference defaults for the per-agent prompt customization system.

## Requirements

### Requirement: Default IDENTITY.md for each sub-agent
The system SHALL provide a default `prompts/agents/<name>/IDENTITY.md` file for each of the 7 sub-agents: operator, navigator, vault, librarian, automator, planner, chronicler.

#### Scenario: All 7 IDENTITY.md files exist
- **WHEN** the prompts directory is inspected
- **THEN** each of `prompts/agents/{operator,navigator,vault,librarian,automator,planner,chronicler}/IDENTITY.md` SHALL exist

#### Scenario: Content matches agentSpecs instruction
- **WHEN** a sub-agent IDENTITY.md is loaded
- **THEN** its content SHALL match the corresponding `agentSpecs[].Instruction` from `internal/orchestration/tools.go`, including the What You Do, Input Format, Output Format, and Constraints sections

### Requirement: IDENTITY.md embedded in binary
The IDENTITY.md files SHALL be included in the embedded prompts filesystem via the existing `go:embed` directive in the prompts package.

#### Scenario: Prompt builder loads default identity
- **WHEN** `agent.promptsDir` is not configured and multi-agent mode is enabled
- **THEN** the prompt builder SHALL use the embedded default IDENTITY.md for each sub-agent
