## Context

The gateway server's `Start()` returns bind/listen errors, but the application lifecycle wrapper runs `Start()` in a goroutine and immediately returns `nil`. This hides deterministic startup failures such as an occupied port from `App.Start()` and from the CLI.

## Decision

Split gateway startup into two steps:

- `Listen()` binds the configured address and prepares the HTTP server synchronously.
- `Serve(listener)` blocks on the accepted listener and preserves the current graceful-shutdown semantics.

`Start()` remains the public blocking convenience method by calling `Listen()` and then `Serve(listener)`. The application lifecycle component will call `Listen()` synchronously, then run `Serve(listener)` in its managed goroutine.

## Tradeoffs

This introduces a small lifecycle API surface on `gateway.Server`, but avoids preflight bind races and duplicated address/server setup in `internal/app`. It also keeps direct gateway users compatible because `Start()` keeps the same behavior.
