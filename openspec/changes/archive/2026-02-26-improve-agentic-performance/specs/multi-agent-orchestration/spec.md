## MODIFIED Requirements

### Requirement: Hierarchical agent tree with sub-agents
The system SHALL support a multi-agent mode (`agent.multiAgent: true`) that creates an orchestrator root agent with specialized sub-agents: operator, navigator, vault, librarian, automator, planner, and chronicler. The orchestrator SHALL have NO direct tools (`Tools: nil`) and MUST delegate all tool-requiring tasks to sub-agents.

#### Scenario: Default delegation rounds increased to 10
- **WHEN** `MaxDelegationRounds` is zero or unset
- **THEN** the default SHALL be 10 rounds (previously 5)

## ADDED Requirements

### Requirement: Round budget guidance in orchestrator prompt
The orchestrator instruction SHALL include round-budget management guidance that helps the LLM self-regulate delegation efficiency.

#### Scenario: Budget guidance included in prompt
- **WHEN** the orchestrator instruction is built
- **THEN** it SHALL contain guidance categorizing tasks by round cost: simple (1-2), medium (3-5), complex (6-10)

#### Scenario: Prompt includes consolidation advice
- **WHEN** the orchestrator is running low on rounds
- **THEN** the prompt SHALL advise consolidating partial results and providing the best possible answer

#### Scenario: Delegation rules formatting
- **WHEN** the orchestrator instruction is built
- **THEN** the "Maximum N delegation rounds" text SHALL appear as part of the round budget section, not the delegation rules section
