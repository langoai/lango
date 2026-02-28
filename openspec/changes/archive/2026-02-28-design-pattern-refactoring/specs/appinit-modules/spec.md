## ADDED Requirements

### Requirement: Module interface
The system SHALL define a Module interface with Name(), Provides(), DependsOn(), Enabled(), and Init() methods for declarative initialization units.

#### Scenario: Module declares dependencies
- **WHEN** a module's DependsOn() returns ["session_store"]
- **THEN** the builder SHALL ensure the session_store provider runs first

### Requirement: Topological sort with cycle detection
TopoSort SHALL order modules so dependencies are initialized before dependents, and SHALL return an error if cycles are detected.

#### Scenario: A depends on B depends on C
- **WHEN** modules A→B→C are sorted
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
