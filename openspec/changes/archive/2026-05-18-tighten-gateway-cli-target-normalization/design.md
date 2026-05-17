## Context

`clihttp.ResolveGatewayAddr` is now the shared gateway default resolver for gateway-backed CLI commands. It handles explicit address, configured server host/port, and localhost fallback. Two small production polish gaps remain:

- explicit values are whitespace-trimmed but not trailing-slash-normalized
- `lango status` passes the resolved probe address to `collectStatus`, but `collectStatus` still computes the displayed `Gateway` field from config only

## Decisions

### Normalize Explicit Addresses

Update the shared resolver so an explicit address like `http://127.0.0.1:18789/` resolves to `http://127.0.0.1:18789`. This avoids double-slash request paths when callers append `/health`, `/metrics`, or `/alerts`.

### Status Display Target

Add a narrow status collection variant that accepts the already resolved gateway target. Keep existing `collectStatus(cfg, profile, addr)` for compatibility, but have it delegate to the new variant with the config-derived display target. The root status command should call the variant with the resolved probe target so JSON/table output matches what was actually probed.

### Scope

Do not change config default behavior or introduce a separate advertised-address model. This change only normalizes explicit CLI input and makes status output truthful for the target chosen in this command invocation.

## Risks

- Some tests assert the status `Gateway` field from config. Preserve that behavior for direct `collectStatus` callers unless they use the new explicit-target path.
- Stripping all trailing slashes means unusual gateway base paths are not supported. Current CLI design treats `--addr` as an origin, not a path-prefixed API root, so this is acceptable and consistent with existing gateway clients.
