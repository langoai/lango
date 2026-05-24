## MODIFIED Requirements

### Requirement: Gateway CLI docs default wording guard stays executable
Repository-level docs guards SHALL prevent gateway-backed CLI docs from presenting localhost/18789 as the only default when the command now honors configured server host and port, and SHALL prevent status docs from omitting the explicit `--addr` probe/display contract.

#### Scenario: Status explicit address docs guard remains covered
- **WHEN** public status CLI docs are checked
- **THEN** executable tests SHALL fail if they omit that explicit `--addr` probes the normalized address
- **AND** executable tests SHALL fail if they omit that status output reports the same normalized gateway target
