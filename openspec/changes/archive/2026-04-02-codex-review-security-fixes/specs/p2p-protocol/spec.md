## MODIFIED Requirements

### Requirement: Safety gate wiring
The application MUST wire `SetSafetyGate()` on the P2P handler during initialization when a tool catalog is available. The safety checker MUST use `ToolCatalog.GetToolSafetyLevel()`. If `ParseSafetyLevel` returns false for the configured `MaxSafetyLevel`, the system MUST fall back to `SafetyLevelModerate`.

#### Scenario: Safety gate wired at startup
- **WHEN** the application initializes with P2P enabled and a tool catalog available
- **THEN** `SetSafetyGate` SHALL be called on the P2P handler with the catalog-based checker

#### Scenario: Invalid MaxSafetyLevel defaults to moderate
- **WHEN** `p2p.maxSafetyLevel` is empty or an invalid string
- **THEN** the safety gate SHALL use `SafetyLevelModerate` as the threshold

### Requirement: Paid tool safety ordering
In `handleToolInvokePaid`, the safety-level gate MUST execute before the payment gate. Tools blocked by safety level SHALL be denied without initiating payment processing.

#### Scenario: Blocked tool not charged
- **WHEN** a P2P peer invokes a paid tool that exceeds `maxSafetyLevel`
- **THEN** the handler SHALL return `ResponseStatusDenied` with `ErrToolSafetyBlocked`
- **AND** the payment gate SHALL NOT be consulted
