## MODIFIED Requirements

### Requirement: All CLI commands use shared loaders
All CLI commands that require bootstrap (config get, config set, run, doctor, settings, onboard, etc.) SHALL use the shared loader functions. No production CLI package SHALL call bootstrap directly outside the shared loader package.

#### Scenario: Config commands use shared loader
- **WHEN** `config get` or `config set` is executed
- **THEN** the command SHALL use the shared loader's `Config()` or `BootResult()` function

#### Scenario: Run command uses shared loader
- **WHEN** `lango run` is executed
- **THEN** the command SHALL use the shared loader's `BootResult()` function

#### Scenario: Serve command uses shared loader
- **WHEN** `lango serve` is executed
- **THEN** `serveCmd()` SHALL use `cliboot.BootResult()` instead of calling `bootstrap.Run()` directly

#### Scenario: Interactive setup commands use shared loader
- **WHEN** `lango settings`, `lango onboard`, or `lango doctor` needs bootstrap state
- **THEN** it SHALL use the shared `cliboot` loader instead of constructing `bootstrap.Options` locally

#### Scenario: No direct bootstrap calls in production CLI packages
- **WHEN** the codebase is audited
- **THEN** no non-test production file under `internal/cli` or `cmd/` SHALL call `bootstrap.Run()` directly except the shared loader package
