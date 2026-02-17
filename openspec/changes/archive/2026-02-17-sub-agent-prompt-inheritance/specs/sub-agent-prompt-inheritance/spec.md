## ADDED Requirements

### Requirement: Sub-agents inherit shared prompt sections
In multi-agent mode, each sub-agent's system prompt SHALL include shared Safety and ConversationRules sections from the prompt builder, in addition to the agent's own role instruction.

#### Scenario: Default sub-agent prompt structure
- **WHEN** multi-agent mode is enabled and no per-agent overrides exist
- **THEN** each sub-agent's system prompt SHALL contain: AgentIdentity (priority 150) from spec.Instruction, Safety (priority 200), and ConversationRules (priority 300), ordered by priority

#### Scenario: Sub-agents do not inherit global identity or tool usage
- **WHEN** the shared prompt builder contains SectionIdentity and SectionToolUsage
- **THEN** sub-agents SHALL NOT receive these sections; only Safety and ConversationRules are shared

### Requirement: Per-agent prompt overrides via directory
Users SHALL be able to override or extend prompt sections for individual sub-agents by placing `.md` files in `<promptsDir>/agents/<agentName>/`.

#### Scenario: Override agent identity via IDENTITY.md
- **WHEN** `<promptsDir>/agents/operator/IDENTITY.md` exists with non-empty content
- **THEN** the operator sub-agent's SectionAgentIdentity SHALL use the file content instead of the default spec.Instruction

#### Scenario: Override shared safety for one agent
- **WHEN** `<promptsDir>/agents/operator/SAFETY.md` exists with non-empty content
- **THEN** the operator sub-agent's Safety section SHALL use the file content instead of the shared Safety section

#### Scenario: Add custom section for one agent
- **WHEN** `<promptsDir>/agents/librarian/MY_RULES.md` exists with non-empty content
- **THEN** the librarian sub-agent SHALL include the custom section with priority 900+

#### Scenario: No per-agent directory
- **WHEN** no `<promptsDir>/agents/<agentName>/` directory exists
- **THEN** the sub-agent SHALL use the shared base prompt with its default spec.Instruction unchanged

### Requirement: SubAgentPromptFunc backward compatibility
When `SubAgentPromptFunc` is nil in the orchestration Config, sub-agents SHALL use their original `spec.Instruction` unchanged, preserving backward compatibility.

#### Scenario: Nil SubAgentPromptFunc
- **WHEN** `Config.SubAgentPrompt` is nil
- **THEN** `BuildAgentTree` SHALL pass `spec.Instruction` directly to each sub-agent's Instruction field

### Requirement: Per-agent override does not mutate shared base
Loading per-agent overrides SHALL NOT modify the shared base builder used by other sub-agents.

#### Scenario: Independent sub-agent builders
- **WHEN** operator has a per-agent override and librarian does not
- **THEN** the librarian's prompt SHALL still contain the original shared Safety and ConversationRules, unaffected by operator's overrides
