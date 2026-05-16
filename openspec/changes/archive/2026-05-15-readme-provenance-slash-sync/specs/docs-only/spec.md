## ADDED Requirements

### Requirement: README provenance inventory uses slash-separated subcommand slices
The README internal CLI inventory SHALL describe the current provenance subcommand slices in slash-separated form.

#### Scenario: Hyphen-compressed provenance shorthand stays removed
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL use slash-separated subcommand slices for checkpoint, session, attribution, and bundle commands
