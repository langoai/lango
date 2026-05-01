## ADDED Requirements

### Requirement: Built-in teammate durability audit
Before archive, the implementation SHALL classify built-in teammate durability for spawn submission, run status transitions, projection sync markers, approval-blocked conditions, and recovery states using one of three verdicts: recorded, not recorded but harmless, or not recorded and follow-up required.

#### Scenario: Audit verdict is recorded before archive
- **WHEN** the hard-cut change is prepared for archive
- **THEN** the implementation SHALL record one verdict for each audited built-in teammate durability item
- **AND** each verdict SHALL be one of: recorded, not recorded but harmless, or not recorded and follow-up required
