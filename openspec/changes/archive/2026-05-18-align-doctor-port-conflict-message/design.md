## Overview

The doctor server port check already detects occupied ports by attempting `net.Listen`. The root problem is the user-facing failure message: it reports the port as "not available" even when the actionable condition is that the configured port is already in use.

## Decisions

- Preserve the existing detection mechanism and `StatusFail` behavior.
- Change only the failure message to `Port <port> in use`.
- Keep the original network error in `Details` so platform-specific diagnostics remain available.

## Testing

- Add a RED test that reserves an IPv4 loopback port, runs `NetworkCheck`, and expects `StatusFail` with `Port <port> in use`.
- Run focused doctor tests, full Go build/test, diff check, and OpenSpec validation.
