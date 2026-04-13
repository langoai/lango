## RESTORED Requirements

### Requirement: Config Load returns LoadResult (RESTORED)
`config.Load()` SHALL return `(*LoadResult, error)` where `LoadResult` contains `Config`, `ExplicitKeys`, and `AutoEnabled` fields. The loader SHALL detect explicitly-set context-related keys, apply context profile defaults, and resolve auto-enable before returning.

#### Scenario: Load with explicit keys
- **WHEN** `config.Load()` is called with a config file containing `knowledge.enabled: false`
- **THEN** `LoadResult.ExplicitKeys` SHALL include `"knowledge.enabled": true`
- **AND** `ResolveContextAutoEnable` SHALL NOT override `knowledge.enabled`

### Requirement: Config migration uses LoadResult (RESTORED)
`configstore.MigrateFromJSON()` SHALL use `result.Config` and `result.ExplicitKeys` from `config.Load()` when importing a JSON config file as an encrypted profile.

### Requirement: Settings TUI field handler completeness (RESTORED)
`UpdateConfigFromForm` SHALL handle ALL settings categories registered in `menu.go`. This includes Orchestration, RunLedger, Provenance, OS Sandbox, and TraceStore field handlers that were present in the `dev` branch.

## ADDED Requirements

### Requirement: Context allocation defaults
`DefaultConfig()` SHALL include `Context.Allocation` with defaults: Knowledge=0.30, RAG=0.25, Memory=0.25, RunSummary=0.10, Headroom=0.10. These match the context-budget spec requirement.
