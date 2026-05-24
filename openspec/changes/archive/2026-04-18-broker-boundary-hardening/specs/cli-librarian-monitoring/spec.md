## ADDED Requirements

### Requirement: Librarian inquiries command uses storage reader
The `lango librarian inquiries` command MUST read pending inquiries through a storage facade reader instead of querying Ent directly from the CLI layer.

#### Scenario: Inquiries command reads through facade
- **WHEN** the user runs `lango librarian inquiries`
- **THEN** the command loads pending inquiry records from the storage facade reader
