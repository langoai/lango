## ADDED Requirements

### Requirement: Status and Settings remain registered as cockpit core pages
The cockpit SHALL keep Status and Settings available as core page routes even when their optional backing dependencies are absent, relying on the page-level degraded states instead of suppressing page registration.

#### Scenario: Status page remains registered without metrics or feature providers
- **WHEN** cockpit startup wiring runs with no metrics collector and no feature-status provider
- **THEN** the Status page route SHALL still be registered

#### Scenario: Settings page remains registered without a config-profile store
- **WHEN** cockpit startup wiring runs with no config-profile store
- **THEN** the Settings page route SHALL still be registered
