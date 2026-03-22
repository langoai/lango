### Requirement: Registry Names accessor
The `lifecycle.Registry` SHALL expose a `Names()` method that returns the names of all registered components in registration order.

#### Scenario: Names returns registered component names
- **WHEN** components "alpha", "beta", "gamma" are registered in order
- **THEN** `Names()` SHALL return `["alpha", "beta", "gamma"]`

#### Scenario: Names returns empty slice for empty registry
- **WHEN** no components are registered
- **THEN** `Names()` SHALL return an empty slice

### Requirement: Catalog builder correctly processes entries
The `buildCatalogFromEntries()` helper SHALL create a `toolcatalog.Catalog` that reflects the provided entries with correct category names, enabled states, and tool registrations.

#### Scenario: Basic entry processing
- **WHEN** 3 entries are provided (2 enabled with tools, 1 disabled without tools)
- **THEN** `ToolCount()` SHALL equal the total number of tools from enabled entries
- **THEN** `ListCategories()` SHALL contain all 3 category names
- **THEN** enabled categories SHALL have `Enabled == true`, disabled SHALL have `Enabled == false`

#### Scenario: Duplicate category accumulates tools
- **WHEN** 2 entries share the same category name "mcp" with different tools
- **THEN** `ToolNamesForCategory("mcp")` SHALL contain tools from both entries

### Requirement: Post-build lifecycle registration
The `registerPostBuildLifecycle()` function SHALL register a "gateway" component and one "channel-N" component per configured channel.

#### Scenario: No channels configured
- **WHEN** `App.Channels` is empty
- **THEN** `registry.Names()` SHALL equal `["gateway"]`

#### Scenario: Multiple channels configured
- **WHEN** `App.Channels` contains 2 channels
- **THEN** `registry.Names()` SHALL equal `["gateway", "channel-0", "channel-1"]`

### Requirement: Default config parity
Calling `app.New()` with `config.DefaultConfig()` and a valid `bootstrap.Result` SHALL produce an application with the expected default state.

#### Scenario: Default enabled categories
- **WHEN** `app.New()` is called with default config
- **THEN** `ToolCatalog.ListCategories()` SHALL include "exec", "filesystem", "output" as enabled

#### Scenario: Default disabled categories
- **WHEN** `app.New()` is called with default config
- **THEN** categories "browser", "crypto", "secrets", "meta", "graph", "rag", "memory", "agent_memory", "librarian", "mcp", "observability" SHALL be disabled

#### Scenario: Default tool count
- **WHEN** `app.New()` is called with default config
- **THEN** `ToolCatalog.ToolCount()` SHALL be at least 11

#### Scenario: Default lifecycle components
- **WHEN** `app.New()` is called with default config
- **THEN** `registry.Names()` SHALL contain "gateway"
- **THEN** `registry.Names()` SHALL NOT contain "p2p-node", "cron-scheduler", "mcp-manager", "channel-0"

#### Scenario: Default non-nil fields
- **WHEN** `app.New()` is called with default config
- **THEN** `Store`, `Gateway`, `ToolCatalog`, `Agent` SHALL NOT be nil

#### Scenario: Default nil fields
- **WHEN** `app.New()` is called with default config
- **THEN** `P2PNode`, `CronScheduler`, `MCPManager`, `KnowledgeStore` SHALL be nil

### Requirement: Features enabled parity
Calling `app.New()` with knowledge, graph, memory, and cron enabled SHALL produce additional categories and lifecycle components.

#### Scenario: Additional enabled categories
- **WHEN** knowledge, graph, observational memory, and cron are enabled
- **THEN** categories "meta", "graph", "memory", "cron" SHALL be enabled

#### Scenario: Disabled features remain disabled
- **WHEN** background and workflow are not enabled
- **THEN** categories "background" and "workflow" SHALL remain disabled

#### Scenario: Feature lifecycle components
- **WHEN** knowledge, graph, memory, and cron are enabled
- **THEN** `registry.Names()` SHALL contain "memory-buffer", "graph-buffer", "cron-scheduler"

#### Scenario: Feature field population
- **WHEN** knowledge, graph, memory, and cron are enabled
- **THEN** `KnowledgeStore`, `MemoryStore`, `GraphStore`, `CronScheduler` SHALL NOT be nil

#### Scenario: Unrelated features remain nil
- **WHEN** P2P and MCP are not enabled
- **THEN** `P2PNode` and `MCPManager` SHALL remain nil

### Requirement: Extracted tool builders have parity coverage
The test suite SHALL include parity coverage for extracted tool builders so refactors do not silently change tool names or remove handlers.

#### Scenario: Extracted builders expose expected tool names
- **WHEN** the builder parity tests run
- **THEN** extracted builder functions SHALL return the expected tool names in stable order for the covered packages

#### Scenario: Extracted builders have non-nil handlers
- **WHEN** parity tests inspect tools returned by extracted builders
- **THEN** every tool SHALL have a non-nil handler

#### Scenario: Extracted builders avoid duplicate names
- **WHEN** parity tests inspect a builder result set
- **THEN** tool names within that result SHALL be unique
