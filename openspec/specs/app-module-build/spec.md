# app-module-build Specification

## Purpose
Module-based application initialization via `appinit.Builder.Build()`, replacing the monolithic `app.New()` sequential initializer with 5 decoupled modules (foundation, intelligence, automation, network, extension).

## Requirements
### Requirement: Module-based application initialization via Builder
The `app.New()` function SHALL use `appinit.Builder.Build()` to initialize all application modules in topological dependency order, replacing the monolithic sequential initializer.

#### Scenario: Builder initializes all modules
- **WHEN** `app.New(boot)` is called
- **THEN** it SHALL create a `Builder`, register 5 modules (foundation, intelligence, automation, network, extension), and call `Build(ctx)` which initializes them in dependency order

#### Scenario: Module build failure stops initialization
- **WHEN** any module's `Init()` returns an error during `Build()`
- **THEN** `app.New()` SHALL return the error without proceeding to post-build wiring

### Requirement: BuildResult aggregates CatalogEntries from modules
The `BuildResult` struct SHALL include a `CatalogEntries []CatalogEntry` field that collects all catalog entries returned by modules during the build phase.

#### Scenario: CatalogEntries collected from all modules
- **WHEN** `Build()` completes successfully
- **THEN** `BuildResult.CatalogEntries` SHALL contain the union of all `ModuleResult.CatalogEntries` from every initialized module

### Requirement: ProvidesBaseTools key for inter-module tool sharing
The `appinit` package SHALL define a `ProvidesBaseTools` key that the foundation module uses to expose base tools (exec, filesystem, browser, crypto, secrets) to downstream modules.

#### Scenario: Foundation publishes base tools
- **WHEN** the foundation module initializes successfully
- **THEN** it SHALL store all its tools under the `ProvidesBaseTools` key in the resolver

#### Scenario: Intelligence module receives base tools
- **WHEN** the intelligence module initializes
- **THEN** it SHALL resolve `ProvidesBaseTools` from the resolver and pass the tools to `initSkills()`

### Requirement: Modules return lifecycle ComponentEntry
Each module SHALL return lifecycle `ComponentEntry` values in `ModuleResult.Components` for components that need Start/Stop management, instead of directly calling the lifecycle registry.

#### Scenario: Intelligence module returns buffer components
- **WHEN** the intelligence module initializes with memory, embedding, graph, analysis, or librarian buffers
- **THEN** it SHALL include a `ComponentEntry` for each active buffer in the returned `ModuleResult.Components`

#### Scenario: Automation module returns scheduler components
- **WHEN** the automation module initializes with cron, background, or workflow enabled
- **THEN** it SHALL include a `ComponentEntry` for each active component in the returned `ModuleResult.Components`

#### Scenario: Network module returns P2P and economy components
- **WHEN** the network module initializes with P2P, workspace, economy, or health monitor enabled
- **THEN** it SHALL include `ComponentEntry` values for p2p-node, nonce-cache, workspace-db, workspace-gossip, health-monitor, economy-event-monitor, and economy-dangling-detector (as applicable)

#### Scenario: Extension module returns MCP and observability components
- **WHEN** the extension module initializes with MCP or observability enabled
- **THEN** it SHALL include `ComponentEntry` values for mcp-manager and observability-token-cleanup (as applicable)

### Requirement: Post-build wiring applies cross-cutting middleware
After `builder.Build()`, `app.New()` SHALL apply cross-cutting middleware to all tools in a fixed order: WithLearning, WithOutputManager, WithHooks, WithApproval.

#### Scenario: Middleware applied in correct order
- **WHEN** `app.New()` enters the post-build wiring phase
- **THEN** it SHALL apply WithLearning (if knowledge enabled), then WithOutputManager, then WithHooks, then WithApproval to the combined tools slice

### Requirement: Post-build lifecycle registration
After module build, `app.New()` SHALL register module-returned components and post-build components (gateway, channels) with the lifecycle registry.

#### Scenario: Module components registered before gateway
- **WHEN** `app.New()` registers lifecycle components
- **THEN** all `BuildResult.Components` SHALL be registered before the gateway and channel components

### Requirement: Network module includes workspace initialization
The network module SHALL initialize P2P workspace components (workspace manager, gossip, DB) when P2P and workspace are both enabled, matching the behavior previously in `app.New()`.

#### Scenario: Workspace tools registered when P2P active
- **WHEN** P2P is enabled and workspace is configured
- **THEN** the network module SHALL initialize workspace, register workspace tools, wire the workspace-team bridge, and return workspace lifecycle components

### Requirement: Network module includes team-economy bridges
The network module SHALL wire team-economy bridges (escrow, budget, reputation, shutdown) when both P2P coordinator and economy components are available.

#### Scenario: Team-escrow convenience tools added
- **WHEN** P2P coordinator and escrow engine are both available
- **THEN** the network module SHALL build and include team-escrow convenience tools in its returned tools
