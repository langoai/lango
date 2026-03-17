## MODIFIED Requirements

### Requirement: All CLI commands use shared loaders
All CLI commands that require bootstrap (config get, config set, serve, doctor, etc.) SHALL use the shared loader functions. No CLI command SHALL call bootstrap directly outside the shared loader package.

#### Scenario: Serve command uses shared loader
- **WHEN** `lango serve` is executed
- **THEN** `serveCmd()` SHALL use `cliboot.BootResult()` instead of calling `bootstrap.Run()` directly

#### Scenario: No direct bootstrap calls in cmd/ package
- **WHEN** the codebase is audited
- **THEN** no file in `cmd/` SHALL call `bootstrap.Run()` directly; all bootstrap access SHALL go through the shared loader
