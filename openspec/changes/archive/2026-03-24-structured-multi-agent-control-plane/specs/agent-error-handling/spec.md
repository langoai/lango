## ADDED Requirements

### Requirement: RecoveryPolicy captures inline recovery patterns
The `RecoveryPolicy` in `internal/agentrt/` SHALL capture the recovery patterns currently inline in `adk/agent.go:473-593` as a code policy. It SHALL support REJECT detection, tool churn recovery, learning-based error correction, and missing agent correction — applied as a wrapper around the inner executor without modifying `agent.go`.

#### Scenario: Tool churn recovery via hint retry
- **WHEN** inner executor returns `ErrToolChurn` and recovery budget allows
- **THEN** RecoveryPolicy SHALL return `RecoveryRetryWithHint` adding "do not delegate to same agent" hint

#### Scenario: Learning-based fix applied
- **WHEN** inner executor fails and `ErrorFixProvider.GetFixForError()` returns a fix
- **THEN** RecoveryPolicy SHALL incorporate the fix into the retry input
