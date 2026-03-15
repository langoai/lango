## MODIFIED Requirements

### Requirement: Grouped Section Layout
The settings menu SHALL organize categories into named sections using a two-level hierarchical navigation. Level 1 SHALL display 7 named sections (with category counts) plus Save & Cancel. Selecting a section at Level 1 SHALL drill into Level 2 showing only that section's categories. Esc at Level 2 SHALL return to Level 1 with cursor restored. Tab SHALL only toggle Basic/Advanced filtering at Level 2.

The sections SHALL be, in order:
1. **Core** — Providers, Agent, Channels, Tools, Server (advanced), Session (advanced), Logging (advanced), Gatekeeper (advanced), Output Manager (advanced)
2. **AI & Knowledge** — Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store (advanced), Librarian (advanced), Agent Memory (advanced), Multi-Agent (advanced), A2A Protocol (advanced), Hooks (advanced)
3. **Automation** — Cron Scheduler, Background Tasks (advanced), Workflow Engine (advanced)
4. **Payment & Account** — Payment, Smart Account (advanced), SA Session Keys (advanced), SA Paymaster (advanced), SA Modules (advanced)
5. **P2P & Economy** — P2P Network (advanced), P2P Workspace (advanced), P2P ZKP (advanced), P2P Pricing (advanced), P2P Owner Protection (advanced), P2P Sandbox (advanced), Economy (advanced), Economy Risk (advanced), Economy Negotiation (advanced), Economy Escrow (advanced), On-Chain Escrow (advanced), Economy Pricing (advanced)
6. **Integrations** — MCP Settings, MCP Server List (advanced), Observability (advanced)
7. **Security** — Security, Auth (advanced), Security DB Encryption (advanced), Security KMS (advanced)
8. *(untitled)* — Save & Exit, Cancel

#### Scenario: Level 1 section list displayed
- **WHEN** user views the settings menu at Level 1
- **THEN** 7 named section items SHALL be rendered with category counts, followed by a separator and Save & Cancel items

#### Scenario: Level 2 categories displayed
- **WHEN** user selects a section at Level 1
- **THEN** the menu SHALL show only the categories belonging to that section, filtered by the current Basic/Advanced toggle

#### Scenario: Hidden categories in basic mode at Level 2
- **WHEN** settings menu is at Level 2 with basic mode and the section has only advanced categories
- **THEN** a "No basic settings. Press Tab to show all." message SHALL be displayed

### Requirement: User Interface
The settings editor SHALL provide menu-based navigation with a two-level hierarchy, free navigation between categories at Level 2, and shared `tuicore.FormModel` for all forms. Provider and OIDC provider list views SHALL support managing collections. Pressing Esc at Level 1 of StepMenu SHALL navigate back to StepWelcome. Pressing Esc at Level 2 SHALL navigate back to Level 1 with cursor restored. The help bar at Level 1 SHALL omit the Tab hint. The help bar at Level 2 SHALL include the Tab hint.

#### Scenario: Launch settings
- **WHEN** user runs `lango settings`
- **THEN** the editor SHALL display a welcome screen followed by the Level 1 section list

#### Scenario: Save from settings
- **WHEN** user selects "Save & Exit" from Level 1
- **THEN** the configuration SHALL be saved as an encrypted profile

#### Scenario: Esc at Welcome screen quits
- **WHEN** user presses Esc at the Welcome screen (StepWelcome)
- **THEN** the TUI SHALL quit

#### Scenario: Esc at Level 1 navigates back to Welcome
- **WHEN** user presses Esc at Level 1 while not in search mode
- **THEN** the editor SHALL navigate back to StepWelcome without quitting

#### Scenario: Esc at Level 2 navigates back to Level 1
- **WHEN** user presses Esc at Level 2 while not in search mode
- **THEN** the menu SHALL return to Level 1 with cursor restored to the section position

#### Scenario: Esc at Menu during search cancels search
- **WHEN** user presses Esc at the settings menu while search mode is active
- **THEN** the search SHALL be cancelled and the menu SHALL remain at StepMenu

#### Scenario: Ctrl+C always quits
- **WHEN** user presses Ctrl+C at any step
- **THEN** the TUI SHALL quit immediately with Cancelled flag set

#### Scenario: Menu help bar at Level 1
- **WHEN** the settings menu is at Level 1 in normal mode
- **THEN** the help bar SHALL display: Navigate, Select, Search, Back (no Tab hint)

#### Scenario: Menu help bar at Level 2
- **WHEN** the settings menu is at Level 2 in normal mode
- **THEN** the help bar SHALL display: Navigate, Select, Search, Tab (filter label), Back

### Requirement: Breadcrumb navigation in settings editor
The settings editor SHALL display a breadcrumb navigation header that reflects the current editor step. The breadcrumb SHALL use `tui.Breadcrumb()` with the following segments per step:
- **StepWelcome / StepMenu (Level 1)**: "Settings"
- **StepMenu (Level 2)**: "Settings" > section title (from `menu.ActiveSectionTitle()`)
- **StepForm**: "Settings" > form title (from `activeForm.Title`)
- **StepProvidersList**: "Settings" > "Providers"
- **StepAuthProvidersList**: "Settings" > "Auth Providers"
- **StepMCPServersList**: "Settings" > "MCP Servers"

The last breadcrumb segment SHALL be rendered in `Primary` color with bold weight. Preceding segments SHALL be rendered in `Muted` color. Segments SHALL be separated by " > " in `Dim` color.

#### Scenario: Breadcrumb at Level 1
- **WHEN** user is at StepMenu Level 1
- **THEN** the breadcrumb SHALL display "Settings" as a single segment

#### Scenario: Breadcrumb at Level 2
- **WHEN** user is at StepMenu Level 2 for the "Core" section
- **THEN** the breadcrumb SHALL display "Settings > Core"

#### Scenario: Breadcrumb at form
- **WHEN** user is editing the Agent form (StepForm)
- **THEN** the breadcrumb SHALL display "Settings > Agent Configuration"

#### Scenario: Breadcrumb at providers list
- **WHEN** user is at StepProvidersList
- **THEN** the breadcrumb SHALL display "Settings > Providers"

### Requirement: Search Help Bar
The help bar SHALL update based on the current mode and navigation level.

#### Scenario: Level 1 help bar
- **WHEN** the menu is at Level 1 in normal mode
- **THEN** the help bar SHALL display: Navigate, Select, Search (`/`), Back (`Esc`)

#### Scenario: Level 2 help bar
- **WHEN** the menu is at Level 2 in normal mode
- **THEN** the help bar SHALL display: Navigate, Select, Search (`/`), Tab (filter label), Back (`Esc`)

#### Scenario: Search mode help bar
- **WHEN** the menu is in search mode
- **THEN** the help bar SHALL display: Navigate, Select, Cancel (`Esc`)
