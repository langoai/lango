## MODIFIED Requirements

### Requirement: Policy-based approval gate
The system SHALL apply tool approval wrapping based on ApprovalPolicy instead of the legacy boolean check. When policy is "none", no wrapping SHALL occur. For all other policies, each tool SHALL be passed through `wrapWithApproval`.

#### Scenario: None policy skips wrapping
- **WHEN** ApprovalPolicy is "none"
- **THEN** tools SHALL NOT be wrapped with approval logic

#### Scenario: Dangerous policy wraps tools
- **WHEN** ApprovalPolicy is "dangerous"
- **THEN** all tools SHALL be passed through wrapWithApproval (which internally checks needsApproval)

#### Scenario: Approval denied with session key present
- **WHEN** a tool approval is denied
- **AND** the session key is present in context
- **THEN** the error message SHALL be `tool '<name>' execution denied: user did not approve the action`

#### Scenario: Approval denied with session key missing
- **WHEN** a tool approval is denied
- **AND** the session key is empty in context
- **THEN** the error message SHALL be `tool '<name>' execution denied: no approval channel available (session key missing)`

## ADDED Requirements

### Requirement: Tool approval system prompt guidance
The system prompt SHALL include a "Tool Approval" section in `TOOL_USAGE.md` that instructs the AI to correctly interpret approval outcomes.

#### Scenario: User denial guidance
- **WHEN** the AI receives "user did not approve the action" error
- **THEN** the system prompt guidance SHALL instruct the AI to inform the user and offer to retry or suggest alternatives

#### Scenario: Missing channel guidance
- **WHEN** the AI receives "no approval channel available" error
- **THEN** the system prompt guidance SHALL instruct the AI to inform the user of a configuration issue
