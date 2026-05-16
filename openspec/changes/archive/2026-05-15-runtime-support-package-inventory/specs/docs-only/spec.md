## ADDED Requirements

### Requirement: Architecture and README inventory docs include current runtime-support packages
The public inventory docs SHALL include shipped runtime-support packages that back exportability, receipt progression, storage brokering, stream composition, and dynamic tool plumbing instead of omitting them from the package inventory.

#### Scenario: Runtime-support package rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `exportability/`, `knowledgeruntime/`, `receipts/`, `storagebroker/`, `streamx/`, `tooloutput/`, and `toolparam/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe source-class exportability evaluation, knowledge-exchange runtime branch selection, canonical receipt/event progression, persistent stdio JSON storage brokering, iterator-based stream combinators, TTL-backed tool output retention, and typed tool parameter extraction truthfully
