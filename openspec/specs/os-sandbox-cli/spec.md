## ADDED Requirements

### Requirement: Sandbox status command
The system SHALL provide `lango sandbox status` that displays sandbox configuration, platform capabilities, and active isolator name.

#### Scenario: Status on macOS with sandbox enabled
- **WHEN** `lango sandbox status` is run on macOS with `sandbox.enabled: true`
- **THEN** output SHALL show enabled=true, platform=darwin, Seatbelt=available, active isolator=seatbelt

#### Scenario: Status shows allowedNetworkIPs warning on Linux
- **WHEN** `lango sandbox status` is run on Linux with `allowedNetworkIPs` configured
- **THEN** output SHALL include a warning that `allowedNetworkIPs` is macOS-only

### Requirement: Sandbox test command
The system SHALL provide `lango sandbox test` that runs smoke tests verifying filesystem write restriction and read permission.

#### Scenario: Test on platform with sandbox
- **WHEN** `lango sandbox test` is run with available sandbox
- **THEN** it SHALL verify write to /etc is blocked and read from /etc/hosts succeeds

#### Scenario: Test on platform without sandbox
- **WHEN** `lango sandbox test` is run without available sandbox
- **THEN** it SHALL print "OS sandbox not available on this platform"
