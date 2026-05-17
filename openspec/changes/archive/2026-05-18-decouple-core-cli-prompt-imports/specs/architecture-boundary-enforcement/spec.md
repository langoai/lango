## ADDED Requirements

### Requirement: Non-CLI internal packages do not import CLI helpers

Production packages under `internal/**` outside `internal/cli/**` SHALL NOT import `internal/cli/**` packages.

#### Scenario: Non-CLI internal package imports CLI helper

- **WHEN** a production package such as `internal/security/passphrase` or `internal/bootstrap` imports `github.com/langoai/lango/internal/cli/prompt`
- **THEN** `go test ./internal/archtest/...` SHALL fail with a boundary violation identifying the offending source and dependency

#### Scenario: CLI packages remain allowed to import CLI helpers

- **WHEN** packages under `internal/cli/**` or the `cmd/lango` entrypoint import `internal/cli/**`
- **THEN** the boundary test SHALL allow those imports
