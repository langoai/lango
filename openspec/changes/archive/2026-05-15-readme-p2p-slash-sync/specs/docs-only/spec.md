## ADDED Requirements

### Requirement: README P2P inventory uses slash-separated subcommand slices
The README internal CLI inventory SHALL describe the current P2P subcommand slices in slash-separated form.

#### Scenario: Hyphen-compressed P2P shorthand stays removed
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL use slash-separated subcommand slices for firewall, session, sandbox, workspace, git, provenance, team, and ZKP
