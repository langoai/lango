## MODIFIED Requirements

### Requirement: RecoveryPolicy captures inline recovery patterns
The `RecoveryPolicy` in `internal/agentrt/` SHALL capture the recovery patterns currently inline in `adk/agent.go:473-593` as a code policy. It SHALL support REJECT detection, tool churn recovery, learning-based error correction, missing agent correction, and specialist-aware reroute recovery — applied as a wrapper around the inner executor without modifying `agent.go`.

#### Scenario: Tool churn recovery via hint retry
- **WHEN** inner executor returns `ErrToolChurn` and recovery budget allows
- **THEN** RecoveryPolicy SHALL return `RecoveryRetryWithHint` adding "do not delegate to same agent" hint

#### Scenario: Learning-based fix applied
- **WHEN** inner executor fails and `ErrorFixProvider.GetFixForError()` returns a fix
- **THEN** RecoveryPolicy SHALL incorporate the fix into the retry input

#### Scenario: Specialist tool error becomes reroute recovery
- **WHEN** inner executor returns `ErrToolError` after a specialist delegation has been observed
- **THEN** RecoveryPolicy SHALL return `RecoveryRetryWithHint`
- **AND** the reroute hint SHALL include the failed specialist name
- **AND** the recovery layer SHALL not issue a blind same-input retry to the same specialist path

#### Scenario: Generic retry remains available before delegation
- **WHEN** inner executor returns a retryable tool error before any specialist delegation has been observed
- **THEN** RecoveryPolicy MAY return `RecoveryRetry`
- **AND** the recovery context SHALL leave `AgentName` empty
