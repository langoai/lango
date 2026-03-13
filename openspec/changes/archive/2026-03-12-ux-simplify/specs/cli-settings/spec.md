## MODIFIED Requirements

### Requirement: Settings section reorganization
The Settings TUI SHALL organize categories into 8 sections: Core, AI & Knowledge, Automation, Payment & Account, P2P & Economy, Integrations, Security, and an unnamed action section (Save/Cancel).

#### Scenario: Core section
- **WHEN** settings menu is displayed
- **THEN** Core section contains: Providers, Agent, Channels, Tools, Server (advanced), Session (advanced)

#### Scenario: Automation section
- **WHEN** settings menu is displayed in advanced mode
- **THEN** Automation section contains: Cron Scheduler, Background Tasks, Workflow Engine

#### Scenario: P2P & Economy section
- **WHEN** settings menu is displayed in advanced mode
- **THEN** P2P & Economy section contains P2P Network, P2P Workspace, P2P ZKP, P2P Pricing, P2P Owner Protection, P2P Sandbox, Economy, Economy Risk, Economy Negotiation, Economy Escrow, On-Chain Escrow, Economy Pricing

#### Scenario: Hidden sections in basic mode
- **WHEN** settings menu is in basic mode (default)
- **THEN** sections with only advanced categories (e.g., P2P & Economy with all advanced items) are hidden entirely
