## Context

Recent gateway-backed CLI work made `lango bg` resolve the target process from `--addr` or the configured server address. Other gateway-backed commands still carry local hardcoded defaults:

- `internal/cli/metrics` defaults `--addr` to `http://localhost:18789`
- `internal/cli/alerts` defaults `--addr` to `http://localhost:18789`
- `internal/cli/status` displays the configured gateway but probes the hardcoded flag default unless the user overrides it

This creates inconsistent behavior for users who run Lango on a custom host or port.

## Decisions

### Shared Address Resolver

Add one shared CLI helper that accepts an explicit address and `*config.Config`:

- if explicit `--addr` is non-empty, return it unchanged
- otherwise use `cfg.Server.Host` and `cfg.Server.Port`
- blank host falls back to `localhost`
- zero port falls back to `18789`
- nil config falls back to `http://localhost:18789`

This keeps the fallback centralized and avoids copying address logic across commands.

### Metrics And Alerts Constructors

Keep existing `NewMetricsCmd()` and `NewAlertsCmd()` wrappers for tests or embedded callers. Add config-loader-aware constructors for root wiring:

- `NewMetricsCmdWithConfig(configLoader func() (*config.Config, error))`
- `NewAlertsCmdWithConfig(configLoader func() (*config.Config, error))`

The commands should load config only after output format validation and only when no explicit `--addr` was supplied. This preserves the current fast failure behavior for invalid `--output` and keeps explicit remote gateway calls independent of config loading.

### Status Probe

`NewStatusCmd` already bootstraps configuration before collecting status. Change the `--addr` default to empty and resolve the probe address from the boot config inside the root status command. The displayed `gateway` field and live probe should therefore point at the same default address.

### Documentation

Dedicated CLI docs should say that `--addr` is optional and overrides the configured server host/port. They should not present `http://localhost:18789` as the only default, except as the fallback when config is missing or blank.

## Risks

- Loading config for metrics or alerts when `--addr` is omitted may introduce bootstrap errors that were previously hidden by the hardcoded default. This is preferable because the command is now honoring configured runtime state; explicit `--addr` remains available for no-config scripts.
- Binding host and reachable client host are not always equivalent when the server binds `0.0.0.0`. This change intentionally follows existing config semantics and current `lango bg` behavior rather than inventing separate advertised-address configuration.
