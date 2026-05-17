## Why

Gateway-backed CLI commands should honor the same configured server host and port that `lango serve` uses. Today `lango metrics`, `lango alerts`, and the live probe inside `lango status` still default to `http://localhost:18789`, so custom `server.host` or `server.port` configurations can make the CLI silently target the wrong process.

## What Changes

- make `lango metrics` and all metrics subcommands resolve the gateway from explicit `--addr` first, then `server.host`/`server.port`, then the localhost/18789 fallback
- make `lango alerts list` and `lango alerts summary` use the same configured gateway address resolution
- make `lango status` display and probe the same configured gateway address when `--addr` is omitted
- keep `--addr` as an explicit override for scripting and remote gateways
- update public CLI docs and executable coverage so future gateway-backed commands do not regress to hardcoded defaults

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `observability`: root metrics and metrics breakdown commands resolve their default gateway from configuration
- `policy-observability`: policy metrics command resolves its default gateway from configuration
- `alerting`: alerts CLI commands resolve their default gateway from configuration
- `cli-status-dashboard`: status dashboard live probe resolves its default gateway from configuration
- `downstream-docs-sync`: CLI reference docs describe the configured-gateway default accurately
- `test-coverage`: executable tests cover configured-gateway defaults for metrics, alerts, status, and docs

## Impact

- touches CLI packages: `internal/cli/metrics`, `internal/cli/alerts`, `internal/cli/status`, and root wiring in `cmd/lango`
- likely adds a small shared helper in `internal/cli/clihttp` for deriving a gateway URL from `config.Config`
- updates dedicated public docs under `docs/cli/metrics.md`, `docs/cli/alerts.md`, and `docs/cli/status.md`
- changes default CLI behavior only when no explicit `--addr` is supplied; explicit overrides remain compatible
