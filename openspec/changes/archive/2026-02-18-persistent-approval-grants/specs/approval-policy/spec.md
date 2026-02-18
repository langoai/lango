## MODIFIED Requirements

### Requirement: Policy-based approval gate
The system SHALL apply tool approval wrapping based on ApprovalPolicy instead of the legacy boolean check. When policy is "none", no wrapping SHALL occur. For all other policies, each tool SHALL be passed through `wrapWithApproval` with the GrantStore for persistent grant tracking.

#### Scenario: None policy skips wrapping
- **WHEN** ApprovalPolicy is "none"
- **THEN** tools SHALL NOT be wrapped with approval logic

#### Scenario: Dangerous policy wraps tools with GrantStore
- **WHEN** ApprovalPolicy is "dangerous"
- **THEN** all tools SHALL be passed through wrapWithApproval with a GrantStore instance

#### Scenario: Granted tool bypasses approval prompt
- **WHEN** a tool has been previously "Always Allowed" in the current session
- **AND** the same tool is invoked again
- **THEN** the tool SHALL execute without prompting the user

#### Scenario: AlwaysAllow response records grant
- **WHEN** a user responds with "Always Allow" to an approval request
- **THEN** the GrantStore SHALL record the grant for that session and tool
- **AND** subsequent invocations of the same tool in the same session SHALL auto-approve

#### Scenario: Deny response does not record grant
- **WHEN** a user denies an approval request
- **THEN** no grant SHALL be recorded
- **AND** future invocations SHALL still prompt for approval
