## ADDED Requirements

### Requirement: Public P2P tool isolation defaults stay synchronized

Public configuration documentation SHALL describe selected P2P tool isolation defaults using `DefaultConfig()` as the source of truth.

#### Scenario: Public docs match DefaultConfig P2P tool isolation defaults

- **WHEN** maintainers update `DefaultConfig().P2P.ToolIsolation`
- **THEN** executable documentation guards SHALL verify that `README.md` and `docs/configuration.md` document the selected P2P tool isolation defaults consistently with `DefaultConfig()`
- **AND** the guards SHALL fail if `p2p.toolIsolation.maxMemoryMB` or other guarded defaults drift in either public document

#### Scenario: Runtime behavior remains unchanged

- **WHEN** the documentation drift is corrected
- **THEN** the `DefaultConfig()` P2P tool isolation values SHALL remain unchanged
