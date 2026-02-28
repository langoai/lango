# Codebase Structure

## Purpose

Defines file organization conventions and domain-based file splitting rules for the Lango codebase. Ensures large files are split into navigable, domain-focused units within the same Go package without API changes.

## Requirements

### Requirement: Domain-based file splitting for tools
The `internal/app/tools.go` file SHALL be split into domain-focused files within the same package. The orchestrator function `buildTools` and shared utilities SHALL remain in `tools.go`. Each domain builder function SHALL be placed in a file named `tools_<domain>.go`.

#### Scenario: Tools file split into 9 files
- **WHEN** the refactoring is applied to `internal/app/tools.go`
- **THEN** the following files SHALL exist: `tools.go` (orchestrator + utilities), `tools_exec.go`, `tools_filesystem.go`, `tools_browser.go`, `tools_meta.go`, `tools_security.go`, `tools_automation.go`, `tools_p2p.go`, `tools_data.go`

#### Scenario: No API changes after tools split
- **WHEN** any consumer imports `internal/app`
- **THEN** all previously available functions SHALL remain accessible with identical signatures

### Requirement: Domain-based file splitting for wiring
The `internal/app/wiring.go` file SHALL be split into domain-focused files within the same package. Core initialization functions SHALL remain in `wiring.go`. Each domain's component struct and init function SHALL be placed in a file named `wiring_<domain>.go`.

#### Scenario: Wiring file split into 9 files
- **WHEN** the refactoring is applied to `internal/app/wiring.go`
- **THEN** the following files SHALL exist: `wiring.go` (core init), `wiring_knowledge.go`, `wiring_memory.go`, `wiring_embedding.go`, `wiring_graph.go`, `wiring_payment.go`, `wiring_p2p.go`, `wiring_automation.go`, `wiring_librarian.go`

#### Scenario: Component structs co-located with init functions
- **WHEN** a domain has a components struct (e.g., `graphComponents`)
- **THEN** the struct and its associated init function (e.g., `initGraphStore`) SHALL be in the same file

### Requirement: Domain-based file splitting for settings forms
The `internal/cli/settings/forms_impl.go` file SHALL be split into domain-focused files within the same package. Core form constructors and shared helpers SHALL remain in `forms_impl.go`. Each domain's form constructors SHALL be placed in a file named `forms_<domain>.go`.

#### Scenario: Forms file split into 6 files
- **WHEN** the refactoring is applied to `internal/cli/settings/forms_impl.go`
- **THEN** the following files SHALL exist: `forms_impl.go` (core forms + helpers), `forms_knowledge.go`, `forms_automation.go`, `forms_security.go`, `forms_p2p.go`, `forms_agent.go`

#### Scenario: Shared helpers remain in the base file
- **WHEN** helper functions are used across multiple domain files
- **THEN** they SHALL remain in `forms_impl.go` (e.g., `derefBool`, `formatKeyValueMap`, `validatePort`)

### Requirement: Domain-based file splitting for config types
The `internal/config/types.go` file SHALL be split into domain-focused files within the same package. Root config and core infrastructure types SHALL remain in `types.go`. Each domain's types SHALL be placed in a file named `types_<domain>.go`.

#### Scenario: Types file split into 5 files
- **WHEN** the refactoring is applied to `internal/config/types.go`
- **THEN** the following files SHALL exist: `types.go` (root + core), `types_security.go`, `types_knowledge.go`, `types_p2p.go`, `types_automation.go`

#### Scenario: Type methods co-located with types
- **WHEN** a type has associated methods (e.g., `ApprovalPolicy.String()`)
- **THEN** the methods SHALL be in the same file as the type definition

### Requirement: Build and test integrity after refactoring
All code changes SHALL maintain full build and test compatibility. No compilation errors or test failures SHALL be introduced.

#### Scenario: Clean build after each phase
- **WHEN** `go build ./...` is executed after any phase of the refactoring
- **THEN** the build SHALL complete with zero errors

#### Scenario: All tests pass after each phase
- **WHEN** `go test ./...` is executed after any phase of the refactoring
- **THEN** all existing tests SHALL pass without modification

### Requirement: File naming convention
All split files SHALL follow the `<base>_<domain>.go` naming convention where `<base>` is the original file's base name and `<domain>` is a kebab-case domain identifier.

#### Scenario: Consistent naming across packages
- **WHEN** a file is split across any package
- **THEN** the new files SHALL use the pattern `<base>_<domain>.go` (e.g., `tools_p2p.go`, `wiring_graph.go`, `types_security.go`, `forms_agent.go`)
