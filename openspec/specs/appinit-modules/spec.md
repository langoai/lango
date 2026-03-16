## Purpose

Module interface with topological sort for declarative app initialization.
## Requirements
### Requirement: Module interface
The system SHALL define a Module interface with Name(), Provides(), DependsOn(), Enabled(), and Init() methods for declarative initialization units.

#### Scenario: Module declares dependencies
- **WHEN** a module's DependsOn() returns ["session_store"]
- **THEN** the builder SHALL ensure the session_store provider runs first

### Requirement: Topological sort with cycle detection
TopoSort SHALL order modules so dependencies are initialized before dependents, and SHALL return an error if cycles are detected.

#### Scenario: A depends on B depends on C
- **WHEN** modules A->B->C are sorted
- **THEN** order SHALL be C, B, A

#### Scenario: Cycle detected
- **WHEN** A depends on B and B depends on A
- **THEN** TopoSort SHALL return an error naming the involved modules

### Requirement: Disabled module exclusion
TopoSort SHALL exclude modules where Enabled() returns false, and SHALL ignore dependencies on keys provided only by disabled modules.

#### Scenario: Disabled module skipped
- **WHEN** module B is disabled and A depends on B's key
- **THEN** A SHALL still be included (dependency treated as optional)

### Requirement: Builder with resolver
The Builder SHALL execute modules in topological order and provide a Resolver that allows later modules to access values provided by earlier modules.

#### Scenario: Resolver passes values between modules
- **WHEN** module A provides key "store" with value X
- **THEN** module B's Init can call resolver.Resolve("store") and receive X

### Requirement: BuildResult aggregation
Build SHALL aggregate all module Tools and Components into a single BuildResult.

#### Scenario: Two modules contribute tools
- **WHEN** module A provides 3 tools and module B provides 2 tools
- **THEN** BuildResult.Tools SHALL contain all 5 tools

### Requirement: CatalogEntry in ModuleResult
The `ModuleResult` struct SHALL include a `CatalogEntries` field of type `[]CatalogEntry` that allows each module to declare tool catalog metadata. Each `CatalogEntry` SHALL contain category name, description, config key, enabled flag, and associated tools.

#### Scenario: Module returns catalog entries
- **WHEN** a module's `Init()` returns a `ModuleResult`
- **THEN** the result SHALL have a `CatalogEntries` field aggregated by the builder

#### Scenario: CatalogEntry metadata
- **WHEN** a `CatalogEntry` is inspected
- **THEN** it SHALL contain Category, Description, ConfigKey, Enabled, and Tools fields

### Requirement: Expanded Provides constants
The `appinit` package SHALL define at least 8 additional `Provides` key constants beyond the original set, covering supervisor, skills, economy, contract, smart account, observability, MCP, and workspace.

#### Scenario: New constants defined
- **WHEN** the `appinit` package is inspected
- **THEN** it SHALL contain `ProvidesSupervisor`, `ProvidesSkills`, `ProvidesEconomy`, `ProvidesContract`, `ProvidesSmartAccount`, `ProvidesObservability`, `ProvidesMCP`, `ProvidesWorkspace`

#### Scenario: Modules use typed constants
- **WHEN** any module declares `Provides()` or `DependsOn()`
- **THEN** it SHALL reference only `appinit.Provides` typed constants

