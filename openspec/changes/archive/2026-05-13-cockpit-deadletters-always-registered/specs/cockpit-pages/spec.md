## ADDED Requirements

### Requirement: Dead Letters remains registered as a cockpit degraded surface
The cockpit SHALL keep Dead Letters available as a registered page route even when the dead-letter bridge is unavailable, relying on page-level unavailable messaging instead of suppressing page registration.

#### Scenario: Dead Letters page remains registered without bridge callbacks
- **WHEN** cockpit startup wiring runs without a ready dead-letter bridge
- **THEN** the Dead Letters page route SHALL still be registered

#### Scenario: Dead Letters degraded page reports missing list function
- **WHEN** the registered Dead Letters page is activated without a configured list callback
- **THEN** activation SHALL yield a load error that explains the dead-letter list function is not configured
