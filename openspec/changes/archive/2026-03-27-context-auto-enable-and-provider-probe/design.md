## Context

Steps 2-7 built context-engineering pipeline. All subsystems require explicit enable. Step 8 adds auto-enable: "deps detectable AND not explicitly disabled → enable."

## Goals / Non-Goals

**Goals:**
- Auto-enable Knowledge/Memory/Retrieval when config-level deps available
- Auto-detect embedding provider from configured providers (conservative policy)
- Track explicit keys through both config.Load() and configstore/bootstrap paths
- Enrich FeatureStatus with auto-enable diagnostics

**Non-Goals:**
- Auto-enable Librarian (cost surprise), Graph (heavy infra)
- Runtime capability detection (config-level only)
- `*bool` migration (use explicitKeys instead)

## Decisions

### DD1: explicitKeys over *bool migration
Use `collectExplicitKeys(configPath, keys)` to detect user-set keys from raw config file. No type changes to `Enabled bool` fields. Zero caller migration needed.

### DD2: Shared resolver callable from both paths
`ResolveContextAutoEnable(cfg, explicitKeys)` called from both `config.Load()` and `bootstrap.phaseLoadProfile()`. Config-level check: `cfg.Session.DatabasePath != ""` (hasDBPath).

### DD3: ProbeEmbeddingProvider conservative policy
Local-first (Ollama preferred), single-remote-only (one OpenAI/Gemini → auto-select), multiple-remote → no auto-select (cost surprise prevention).

### DD4: profilePayload wraps Config + ExplicitKeys
Stored inside encrypted profile. Legacy profiles without ExplicitKeys → nil → auto-enable all detectable features (intentional migration behavior).

### DD5: Resolution order
unmarshal → collectExplicitKeys → ApplyContextProfile → ResolveContextAutoEnable → PostLoad(validate)

## Risks / Trade-offs

- **[Legacy migration]** → Old profiles auto-enable features on first load after Step 8. Intentional behavior, documented.
- **[hasDBPath approximation]** → `DatabasePath != ""` implies EntStore at runtime but doesn't guarantee it. Conservative naming.
- **[Save path explicitKeys]** → Settings/onboard save with nil explicitKeys for now (TODO for full tracking).
