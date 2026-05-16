## ADDED Requirements

### Requirement: Cmd entrypoint exit routing stays disciplined
Top-level binary entrypoints SHALL not call `os.Exit(...)` directly except where explicit seam declarations intentionally define the default process-exit function.

#### Scenario: Cmd entrypoints reject direct os.Exit outside seams
- **WHEN** a non-test Go file under `cmd/` reintroduces a direct `os.Exit(...)` reference outside the approved seam declaration lines
- **THEN** an executable repository test SHALL fail
