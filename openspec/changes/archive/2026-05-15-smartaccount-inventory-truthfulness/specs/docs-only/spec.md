## ADDED Requirements

### Requirement: Smart-account inventory docs stay aligned with the current command surface
The public smart-account inventory docs SHALL include the currently implemented session, module, policy, and paymaster subcommands rather than abbreviated subsets.

#### Scenario: Smart-account inventory stays truthful
- **WHEN** a maintainer updates `docs/cli/smartaccount.md`, `docs/architecture/project-structure.md`, or the README internal tree inventory
- **THEN** those docs SHALL include `session create/revoke`, `module install`, `policy set`, and `paymaster approve`
