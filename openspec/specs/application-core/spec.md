## Requirements

### Requirement: Application Bootstrap
The system SHALL initialize all core components through a centralized application entry point (`internal/app`). Security initialization (crypto provider, passphrase) SHALL be skipped entirely when `security.signer.provider` is not configured. The application MUST NOT fail to start due to missing security configuration.

#### Scenario: Startup Sequence
- **WHEN** the application starts
- **THEN** it SHALL load configuration
- **THEN** it SHALL initialize the SQLite session store via Ent
- **THEN** it SHALL initialize knowledge components (Store, Engine, Registry) if knowledge is enabled and using Ent store
- **THEN** it SHALL initialize the Agent Runtime with the session store, configured tools, and optional knowledge-augmented model adapter
- **THEN** it SHALL initialize Channels and Gateway, injecting the Agent
- **THEN** it SHALL start all background services (Gateway, Channels)

#### Scenario: Startup Without Security
- **WHEN** the application starts with no `security.signer.provider` configured
- **THEN** it SHALL skip all security initialization (passphrase, crypto provider, secrets tool, crypto tool)
- **THEN** the agent SHALL still start and respond to messages through channels

#### Scenario: Graceful Shutdown
- **WHEN** the application receives a termination signal (SIGINT/SIGTERM)
- **THEN** it SHALL stop the Gateway server
- **THEN** it SHALL stop all active Channels
- **THEN** it SHALL close the Database connection
- **THEN** it SHALL allow a grace period for active requests to complete

### Requirement: Component Wiring
The system SHALL inject dependencies between components to enable communication. The application SHALL be organized into focused files: `app.go` (lifecycle orchestration, Start/Stop), `wiring.go` (component initialization: supervisor, session store, agent, gateway), `tools.go` (tool definitions and registration for exec and filesystem), `channels.go` (channel initialization and message handlers), `types.go` (type definitions). No single file SHALL exceed 200 lines.

#### Scenario: Tool Registration
- **WHEN** the Agent is initialized
- **THEN** the system SHALL create and register Tool instances (FS, Exec) based on configuration
- **AND** if knowledge is enabled, SHALL register meta-tools (save_knowledge, search_knowledge, save_learning, search_learnings, create_skill, list_skills)
- **AND** if knowledge is enabled, SHALL wrap all tools with the learning engine observer
- **THEN** it SHALL NOT register browser, crypto, or secrets tools

#### Scenario: Agent Injection into Gateway
- **WHEN** the Gateway is initialized
- **THEN** it SHALL receive a reference to the active Agent Runtime
- **AND** it SHALL use this reference to delegate "chat.message" RPC calls to the Agent

#### Scenario: Agent Injection into Channels
- **WHEN** a Channel (Telegram, Discord, Slack) is initialized
- **THEN** it SHALL receive a reference to the active Agent Runtime
- **AND** it SHALL execute the Agent for incoming messages that match the channel's criteria
- **AND** each channel handler SHALL route messages through `adk.Agent.Run()` as the sole execution path

### Requirement: Knowledge Component Initialization
The system SHALL initialize the knowledge subsystem when enabled.

#### Scenario: Knowledge enabled with Ent store
- **WHEN** `knowledge.enabled` is true and the session store is Ent-based
- **THEN** the system SHALL create a `knowledge.Store` with the Ent client and configured limits
- **AND** SHALL create a `learning.Engine` with the Store
- **AND** SHALL create a `skill.Registry` with the Store
- **AND** SHALL call `registry.LoadSkills` to load active skills from the database

#### Scenario: Knowledge disabled
- **WHEN** `knowledge.enabled` is false or the session store is not Ent-based
- **THEN** the system SHALL skip knowledge initialization
- **AND** the agent SHALL operate without knowledge augmentation

#### Scenario: Skill registry initialization failure
- **WHEN** the skill registry fails to initialize (e.g., home directory not found)
- **THEN** the system SHALL log a warning and skip the knowledge system
- **AND** the agent SHALL operate without knowledge augmentation

### Requirement: Context-Aware Agent
The system SHALL augment the agent's model adapter with context retrieval when knowledge is enabled.

#### Scenario: Knowledge-augmented model
- **WHEN** knowledge components are initialized
- **THEN** the system SHALL wrap the standard `ModelAdapter` with a `ContextAwareModelAdapter`
- **AND** the context-aware adapter SHALL retrieve relevant context before each LLM call

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
