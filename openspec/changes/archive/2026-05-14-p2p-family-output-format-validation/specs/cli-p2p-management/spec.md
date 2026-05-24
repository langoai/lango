## ADDED Requirements
### Requirement: Remaining P2P operator commands stay explicit and validated
`lango p2p firewall list`, `zkp status`, `zkp circuits`, `team list`, `team status`, `workspace create`, `workspace list`, `workspace status`, and `git log` SHALL accept `--output table|json` and reject unknown values before bootstrap-dependent work.

#### Scenario: Guidance and inspection commands reject unknown output before bootstrap
- **WHEN** a user runs any of those commands with `--output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work
