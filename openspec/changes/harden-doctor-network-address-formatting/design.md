## Overview

The doctor network check should validate the same listen address the gateway server would use. The code currently rebuilds that address with `fmt.Sprintf("%s:%d", host, port)`, which is not valid for IPv6 literals and diverges from the shared `internal/gatewayaddr.ListenAddress` helper.

## Decisions

- Reuse `gatewayaddr.ListenAddress(host, port)` in `NetworkCheck.Run`.
- Preserve existing fallback semantics: blank or nil config uses `localhost:18789`; positive configured ports are honored.
- Keep the test at the doctor check boundary so regressions are caught if the check stops using the shared formatter.

## Data Flow

1. `NetworkCheck.Run` receives `config.Config`.
2. It resolves host and port with existing fallback logic.
3. It formats the listen address through `gatewayaddr.ListenAddress`.
4. It calls `net.Listen("tcp", addr)` and reports the same formatted address in `Result.Details`.

## Error Handling

Existing failure behavior remains unchanged: listen failures return `StatusFail` with the original network error in `Details`.

## Testing

- Add a RED test showing `NetworkCheck` fails to pass for `server.host = "::1"` before the fix.
- Add coverage for bracketed IPv6 input to prevent double-bracketing regressions.
- Run focused doctor tests, full Go build/test, diff check, and OpenSpec validation.
