## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Approval request moves the tool row into awaiting-approval state
- **WHEN** a tool approval request interrupts a running tool invocation
- **THEN** the latest matching tool transcript row SHALL switch to `awaiting approval`

#### Scenario: Approval denial moves the tool row into canceled state
- **WHEN** the operator denies the interrupted tool request
- **THEN** the matching tool transcript row SHALL switch to `canceled`

#### Scenario: Approval grant restores the tool row to running state
- **WHEN** the operator approves the interrupted tool request
- **THEN** the matching tool transcript row SHALL switch back to `running` until completion
