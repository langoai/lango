## ADDED Requirements

### Requirement: MCPToolset parity evaluation against current MCP adapter
The spike SHALL evaluate ADK `mcptoolset.New()` (`tool/mcptoolset/set.go:27`) against current `internal/mcp/` adapter for tool exposure parity.

#### Scenario: Naming contract parity check
- **WHEN** the naming parity is evaluated
- **THEN** the analysis SHALL determine whether `mcp__{serverName}__{toolName}` naming convention can be maintained with ADK MCPToolset (natively or via wrapper)

#### Scenario: Approval path parity check
- **WHEN** the approval parity is evaluated
- **THEN** the analysis SHALL determine whether `RequireConfirmationProvider` can express always-allow grant, payment auto-approve, and P2P owner approval policies currently implemented by `WithApproval` middleware

#### Scenario: Safety metadata parity check
- **WHEN** the safety metadata parity is evaluated
- **THEN** the analysis SHALL determine whether per-tool safety level (`SafetyLevelSafe`/`Moderate`/`Dangerous`) can be propagated through ADK MCPToolset tools

#### Scenario: Output truncation parity check
- **WHEN** the output truncation parity is evaluated
- **THEN** the analysis SHALL determine whether `maxOutputTokens` truncation can be applied to MCPToolset tool results

#### Scenario: Event publication parity check
- **WHEN** the event publication parity is evaluated
- **THEN** the analysis SHALL determine whether tool call/result events can reach the event bus via ADK `AfterToolCallback` or equivalent mechanism

### Requirement: Adoption decision with concrete pass/fail criteria
The spike SHALL produce a pass/fail table for all 5 parity conditions. MCPToolset adoption SHALL be recommended only when ALL 5 conditions pass.

#### Scenario: All conditions pass
- **WHEN** all 5 parity conditions (naming, approval, safety, truncation, event publication) pass
- **THEN** the spike SHALL recommend MCPToolset adoption with a migration plan

#### Scenario: Any condition fails
- **WHEN** one or more parity conditions fail
- **THEN** the spike SHALL recommend keeping the current `internal/mcp/` adapter
- **AND** SHALL document which conditions failed and what ADK changes would be needed
