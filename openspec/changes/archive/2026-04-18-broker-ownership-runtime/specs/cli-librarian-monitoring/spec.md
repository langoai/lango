## ADDED Requirements

### Requirement: Librarian inquiries support broker-backed runtime reads
The `lango librarian inquiries` command MUST remain functional when bootstrap is broker-owned and runtime reads come from broker-backed storage.

#### Scenario: Librarian inquiries under broker-owned runtime
- **WHEN** broker-backed runtime storage is active
- **THEN** `lango librarian inquiries` still returns pending inquiry records through the broker-backed reader path
