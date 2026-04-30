## MODIFIED Requirements

### Requirement: Application Bootstrap
The system SHALL initialize all core components through a centralized application entry point (`internal/app`) assembled by `cmd/lango`. Security initialization (crypto provider, passphrase) SHALL be skipped entirely when `security.signer.provider` is not configured. The application MUST NOT fail to start due to missing security configuration. `cmd/lango` SHALL be the only production package that imports `internal/app`; CLI packages under `internal/cli/**` SHALL NOT import `internal/app` and SHALL receive narrow interfaces, function providers, DTOs, or app-independent helpers from `cmd/lango`. Production import scans SHALL allow `github.com/langoai/lango/internal/app` only from `cmd/lango/main.go` and `cmd/lango/dead_letter_status.go`.

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

#### Scenario: Production app import boundary
- **WHEN** production imports are scanned
- **THEN** imports of `github.com/langoai/lango/internal/app` are allowed only from `cmd/lango/main.go` and `cmd/lango/dead_letter_status.go`

#### Scenario: CLI package does not import app
- **WHEN** production imports are scanned
- **THEN** packages under `internal/cli/**` do not import `github.com/langoai/lango/internal/app`
