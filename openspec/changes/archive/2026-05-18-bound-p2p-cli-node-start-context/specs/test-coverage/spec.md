## ADDED Requirements

### Requirement: P2P CLI node startup context coverage stays executable
Executable tests SHALL cover command-context propagation into shared P2P CLI node startup.

#### Scenario: P2P CLI startup context tests block regressions
- **WHEN** P2P CLI tests run
- **THEN** they SHALL fail if representative status, peers, discover, identity, session, connect, or disconnect paths detach ephemeral node startup from `cmd.Context()`
- **AND** they SHALL fail if `initP2PDeps` starts `internal/p2p.Node` without passing the caller context
- **AND** they SHALL fail if ephemeral P2P CLI cleanup returns before startup worker goroutines registered with the startup wait group finish

### Requirement: P2P CLI docs guard covers startup cancellation
Executable docs quality coverage SHALL fail when public P2P CLI docs omit the command-scoped ephemeral node startup cancellation contract.

#### Scenario: P2P docs guard checks startup cancellation contract
- **WHEN** docs quality tests run
- **THEN** `docs/cli/p2p.md` SHALL be checked for the command-scoped ephemeral node startup cancellation contract
