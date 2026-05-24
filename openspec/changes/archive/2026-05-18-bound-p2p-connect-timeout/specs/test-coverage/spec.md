## ADDED Requirements

### Requirement: P2P connect timeout coverage stays executable
Executable tests SHALL cover `lango p2p connect` command-context cancellation, timeout selection, and cleanup on connect failure.

#### Scenario: P2P connect context coverage blocks regressions
- **WHEN** P2P connect tests run
- **THEN** they SHALL fail if connect uses `context.Background()` instead of a command-derived context
- **AND** they SHALL fail if configured positive `p2p.handshakeTimeout` is not used
- **AND** they SHALL fail if a shorter parent command deadline is reported as the configured timeout
- **AND** they SHALL fail if an earlier configured timeout is reported as a later parent command deadline
- **AND** they SHALL fail if the 30 second fallback is not used for invalid timeout values
- **AND** they SHALL fail if cleanup is skipped after connect failure
