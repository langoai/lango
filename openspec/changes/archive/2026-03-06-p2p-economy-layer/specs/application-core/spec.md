## MODIFIED Requirements

### Requirement: App struct economy fields
The App struct SHALL include 5 economy component fields typed as `interface{}` to avoid importing economy packages in the core types file. Comments SHALL document the concrete types.

#### Scenario: Economy fields present
- **WHEN** App struct is inspected
- **THEN** EconomyBudget, EconomyRisk, EconomyPricing, EconomyNegotiation, EconomyEscrow fields exist as interface{}

### Requirement: Economy initialization in app startup
The app.New() function SHALL call initEconomy() at step 5o (after MCP wiring, before Auth) and assign returned components to App struct fields.

#### Scenario: Economy step in startup
- **WHEN** app.New() executes with economy enabled
- **THEN** initEconomy is called and economy tools are registered in the catalog

### Requirement: P2P protocol negotiate message types
The protocol handler SHALL support RequestNegotiatePropose and RequestNegotiateRespond message types with NegotiatePayload struct. A SetNegotiator setter SHALL follow the existing SetPayGate pattern.

#### Scenario: Negotiate handler set
- **WHEN** SetNegotiator is called with a NegotiateHandler function
- **THEN** the handler routes negotiate requests to the provided function
