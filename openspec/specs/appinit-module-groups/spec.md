# appinit-module-groups Specification

## Purpose
TBD - created by archiving change config-bootstrap-regression-fixes. Update Purpose after archive.
## Requirements
### Requirement: Five module group implementations
The system SHALL define five module groups — Foundation, Intelligence, Automation, Network, and Extension — each implementing the `Module` interface. Each module group SHALL wrap existing wiring functions from `internal/app/wiring*.go` and organize them into a cohesive initialization unit.

#### Scenario: All five modules are registered
- **WHEN** the application initializes via the module system
- **THEN** Foundation, Intelligence, Automation, Network, and Extension modules SHALL all be registered with the module builder

### Requirement: Foundation module
The Foundation module SHALL initialize core infrastructure: config validation, DB client, session store, embedding store, memory store, and lifecycle registry. It SHALL declare `Provides` keys for all components it initializes and SHALL have no `DependsOn` keys.

#### Scenario: Foundation provides core components
- **WHEN** Foundation module's `Init()` is called
- **THEN** it SHALL return a `ModuleResult` containing the session store, embedding store, memory store, and related tools
- **THEN** its `Provides()` SHALL include keys such as `"session_store"`, `"embedding_store"`, `"memory_store"`

#### Scenario: Foundation has no dependencies
- **WHEN** Foundation module's `DependsOn()` is called
- **THEN** it SHALL return an empty slice

### Requirement: Intelligence module
The Intelligence module SHALL initialize AI-related components: AI provider, model adapter, context-aware model, graph store, and graph RAG service. It SHALL declare dependencies on Foundation-provided components.

#### Scenario: Intelligence depends on Foundation
- **WHEN** Intelligence module's `DependsOn()` is called
- **THEN** it SHALL include keys provided by the Foundation module (e.g., `"embedding_store"`, `"memory_store"`)

#### Scenario: Intelligence provides AI components
- **WHEN** Intelligence module's `Init()` is called
- **THEN** it SHALL return a `ModuleResult` containing model-related components and tools
- **THEN** its `Provides()` SHALL include keys such as `"model_adapter"`, `"graph_store"`

### Requirement: Automation module
The Automation module SHALL initialize automation subsystems: cron scheduler, background task manager, and workflow engine. It SHALL declare dependencies on Foundation and Intelligence components.

#### Scenario: Automation depends on Foundation and Intelligence
- **WHEN** Automation module's `DependsOn()` is called
- **THEN** it SHALL include keys from both Foundation (e.g., `"session_store"`) and Intelligence (e.g., `"model_adapter"`)

#### Scenario: Automation provides scheduler components
- **WHEN** Automation module's `Init()` is called
- **THEN** it SHALL return a `ModuleResult` containing cron, background, and workflow components and their associated tools

### Requirement: Network module
The Network module SHALL initialize networking subsystems: P2P discovery, A2A protocol, MCP client, and channel adapters. It SHALL declare dependencies on Foundation and Intelligence components.

#### Scenario: Network depends on Foundation and Intelligence
- **WHEN** Network module's `DependsOn()` is called
- **THEN** it SHALL include keys from Foundation and Intelligence modules

#### Scenario: Network provides networking components
- **WHEN** Network module's `Init()` is called
- **THEN** it SHALL return a `ModuleResult` containing P2P, A2A, MCP, and channel adapter components and their associated tools

### Requirement: Extension module
The Extension module SHALL initialize extension points: tool middleware chain, dispatcher tools, and catalog entries for all tool categories. It SHALL declare dependencies on all other modules since it aggregates their outputs.

#### Scenario: Extension depends on all other modules
- **WHEN** Extension module's `DependsOn()` is called
- **THEN** it SHALL include keys from Foundation, Intelligence, Automation, and Network modules

#### Scenario: Extension provides final tool set and catalog
- **WHEN** Extension module's `Init()` is called
- **THEN** it SHALL return a `ModuleResult` containing the finalized tool set, middleware-wrapped tools, and complete `CatalogEntries`

### Requirement: Each module wraps existing wiring functions
Each module's `Init()` method SHALL delegate to the corresponding existing wiring functions (e.g., `initFoundation()`, `initIntelligence()`) rather than reimplementing initialization logic.

#### Scenario: Foundation delegates to initFoundation wiring
- **WHEN** Foundation module's `Init()` is called
- **THEN** it SHALL call the existing `initFoundation()` (or equivalent wiring function) internally

### Requirement: Each module returns tools and components and CatalogEntries
Each module's `Init()` SHALL return a `ModuleResult` containing: initialized components, tools provided by the module, and `CatalogEntries` describing the tools for the catalog system.

#### Scenario: ModuleResult contains CatalogEntries
- **WHEN** any module's `Init()` completes successfully
- **THEN** the returned `ModuleResult` SHALL include a non-nil `CatalogEntries` field listing all tools the module provides with their category and description

