## ADDED Requirements

### Requirement: Top-level startup stream docs stay current
Public CLI docs SHALL describe the current startup stream routing for top-level interactive entrypoints when that routing is part of the tested contract.

#### Scenario: Workbench and cockpit docs mention stderr seam routing
- **WHEN** bare `lango` workbench and `lango cockpit` use seam-aware stderr startup notices
- **THEN** the public core CLI docs SHALL mention that startup notice routing
