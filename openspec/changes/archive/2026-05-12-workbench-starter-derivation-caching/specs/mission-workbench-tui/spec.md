## ADDED Requirements

### Requirement: Workbench starter prompt derivation is cached per page instance

The standalone workbench SHALL derive its starter prompt contract from workspace inputs once per Mission Control page instance rather than repeatedly recomputing it during ordinary render work.

#### Scenario: Cached starter prompt contract backs render-time copy
- **WHEN** a Mission Control workbench page is created for a given workspace input
- **THEN** the page SHALL cache the derived starter prompt set and default starter prompt for that page instance
- **AND** ordinary render-time quick-start copy SHALL reuse that cached contract
