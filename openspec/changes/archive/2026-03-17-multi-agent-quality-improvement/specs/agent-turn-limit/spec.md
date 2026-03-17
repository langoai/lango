## MODIFIED Requirements

### Requirement: Dynamic turn budget expansion
The agent Run() loop SHALL dynamically expand the turn budget when multi-agent task complexity is detected.

#### Scenario: Planner involvement triggers expansion
- **WHEN** a delegation event targets the "planner" agent
- **THEN** the turn budget SHALL be expanded to 150% of the original value

#### Scenario: Three or more delegations trigger expansion
- **WHEN** 3 or more delegation events occur in a single run
- **THEN** the turn budget SHALL be expanded to 150% of the original value

#### Scenario: Two or more unique agents trigger expansion
- **WHEN** delegations target 2 or more distinct non-orchestrator agents
- **THEN** the turn budget SHALL be expanded to 150% of the original value

#### Scenario: Single expansion only
- **WHEN** the budget has already been expanded once
- **THEN** subsequent delegation patterns SHALL NOT trigger additional expansion

#### Scenario: No expansion for simple tasks
- **WHEN** only 1 delegation occurs to 1 unique agent and planner is not involved
- **THEN** the turn budget SHALL remain at the original value

#### Scenario: Expansion is logged
- **WHEN** budget expansion is triggered
- **THEN** the system SHALL log the old max, new max, unique agent count, delegation count, and planner involvement

### Requirement: Multi-tier wrap-up budget
The wrap-up mechanism SHALL allow a configurable number of turns after the budget is exceeded.

#### Scenario: Default wrap-up budget
- **WHEN** the turn budget is not expanded
- **THEN** the wrap-up budget SHALL be 1 turn

#### Scenario: Expanded wrap-up budget
- **WHEN** the turn budget is expanded due to multi-agent complexity
- **THEN** the wrap-up budget SHALL be 3 turns

#### Scenario: Hard stop after wrap-up exhausted
- **WHEN** all wrap-up turns are consumed
- **THEN** the agent SHALL return an error indicating the turn limit was exceeded

#### Scenario: Delegation events not counted as turns
- **WHEN** an event is a pure delegation transfer (TransferToAgent is non-empty)
- **THEN** it SHALL NOT be counted toward the turn limit
