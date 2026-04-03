## ADDED Requirements

### Requirement: team.BuildEscrowTools remains in p2p/team
The `team.BuildEscrowTools()` function SHALL remain in `internal/p2p/team/` and MUST NOT be moved to `internal/app/`. This respects the existing domain-tool-builders spec requirement that team owns its tool builders.

#### Scenario: team escrow tools do not import app
- **WHEN** `internal/p2p/team/tools_escrow.go` is compiled
- **THEN** it does NOT import `internal/app/`

#### Scenario: Phase 0 does not modify team tool builder ownership
- **WHEN** Phase 0 boundary fixes are applied
- **THEN** `team.BuildEscrowTools()` remains callable from `p2p/team` package
