## Context

`clihttp.ResolveGatewayAddr` centralizes gateway CLI fallback behavior, but the final host/port formatting still uses plain string interpolation. Plain interpolation is not safe for IPv6 URLs. Other surfaces also duplicate gateway address assembly: serve startup summary, P2P provenance, doctor websocket reachability, and gateway listen address.

## Decisions

### Shared Formatter

Add a small internal gateway-address helper with standard-library `net.JoinHostPort` formatting:

- client HTTP URLs use fallback host `localhost` and fallback port `18789`
- explicit CLI addresses remain trimmed and trailing-slash-normalized by `clihttp`
- listen addresses use bracket-safe `host:port` formatting and preserve configured bind hosts
- doctor reachability URLs map wildcard bind hosts (`0.0.0.0`, `::`) to loopback because clients cannot dial wildcard binds reliably

### P2P Provenance

Change P2P provenance `gatewayAddr` to call `clihttp.ResolveGatewayAddr(addr, boot.Config)`. This aligns push/fetch with metrics, alerts, status, and bg gateway command behavior.

### Test Surface

Add focused tests for:

- shared CLI resolver IPv6 URL formatting
- P2P provenance explicit trailing-slash normalization
- P2P provenance configured gateway fallback
- gateway address helper listen/dial formatting behavior

### Documentation

Update `docs/cli/p2p.md` near the provenance command section. The docs should state that `--addr` overrides the configured gateway, is normalized when supplied, and otherwise falls back to configured server host/port.

## Risks

- Wildcard bind handling has two valid perspectives: display the configured bind host, or dial loopback. Preserve existing status/startup display semantics and only use loopback for doctor reachability probes.
- This does not move P2P provenance POST handling to `clihttp.PostJSON`; it only aligns address resolution to keep the change focused.
