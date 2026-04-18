## ADDED Requirements

### Requirement: Alerts route uses storage alert reader
The observability alerts route MUST query alert history through a storage-provided alert reader instead of a raw ent client.

#### Scenario: Alerts endpoint reads through storage facade
- **WHEN** the `/alerts` route is requested
- **THEN** the route queries alert records through the storage facade alert reader
- **AND** it does not issue ad hoc ent queries from the route layer
