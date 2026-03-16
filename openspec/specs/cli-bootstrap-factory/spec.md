# cli-bootstrap-factory Specification

## Purpose
TBD - created by archiving change config-bootstrap-regression-fixes. Update Purpose after archive.
## Requirements
### Requirement: Shared CLI bootstrap loader package
The system SHALL provide a `cliutil` (or equivalent) package that exposes shared bootstrap loader functions for use by all CLI commands. This package SHALL encapsulate the bootstrap lifecycle so that individual CLI commands do not duplicate bootstrap logic.

#### Scenario: Package is importable by all CLI commands
- **WHEN** a CLI command needs to bootstrap the application
- **THEN** it SHALL import the shared loader package instead of calling bootstrap directly

### Requirement: BootResult returns full bootstrap result
The loader SHALL provide a `BootResult()` function (or equivalent) that performs full bootstrap and returns the complete bootstrap result struct. The caller receives access to the config, DB client, and all other bootstrapped components.

#### Scenario: BootResult succeeds
- **WHEN** `BootResult()` is called with valid configuration
- **THEN** it SHALL return the full bootstrap result containing config, DB client, and other initialized components
- **THEN** the caller is responsible for closing resources via the returned result

#### Scenario: BootResult fails on invalid config
- **WHEN** `BootResult()` is called and bootstrap fails (e.g., missing config file)
- **THEN** it SHALL return an error without leaking any partially initialized resources

### Requirement: Config returns config and closes DB
The loader SHALL provide a `Config()` function (or equivalent) that performs bootstrap, extracts the config, closes the DB client, and returns only the config. This is a convenience function for commands that only need config access.

#### Scenario: Config returns valid config
- **WHEN** `Config()` is called with valid configuration
- **THEN** it SHALL return the fully loaded and validated config
- **THEN** the DB client SHALL be closed before the function returns

#### Scenario: Config does not leak DB connection
- **WHEN** `Config()` is called and the caller uses only the returned config
- **THEN** no DB connection remains open after the function returns

#### Scenario: Config propagates bootstrap errors
- **WHEN** bootstrap fails during `Config()`
- **THEN** the error SHALL be returned and no resources SHALL be leaked

### Requirement: All CLI commands use shared loaders
All CLI commands that require bootstrap (config get, config set, run, doctor, etc.) SHALL use the shared loader functions. No CLI command SHALL call bootstrap directly outside the shared loader package.

#### Scenario: Config commands use shared loader
- **WHEN** `config get` or `config set` is executed
- **THEN** the command SHALL use the shared loader's `Config()` or `BootResult()` function

#### Scenario: Run command uses shared loader
- **WHEN** `lango run` is executed
- **THEN** the command SHALL use the shared loader's `BootResult()` function

#### Scenario: Serve command uses shared loader
- **WHEN** `lango serve` is executed
- **THEN** `serveCmd()` SHALL use `cliboot.BootResult()` instead of calling `bootstrap.Run()` directly

#### Scenario: No direct bootstrap calls in cmd/ package
- **WHEN** the codebase is audited
- **THEN** no file in `cmd/` SHALL call `bootstrap.Run()` directly; all bootstrap access SHALL go through the shared loader

