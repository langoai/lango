## Context

`gateway.Server.Shutdown(ctx)` cancels in-flight request contexts, closes WebSocket clients, and delegates to the underlying `http.Server`. Cleanup and rollback code can call it before the gateway has successfully started. That should be safe because cleanup functions should not need startup phase knowledge.

## Decision

Treat shutdown as a lifecycle-safe operation:

- Always cancel `shutdownCtx` and close known WebSocket clients.
- If no HTTP server has been prepared, return `nil`.
- If an HTTP server exists, delegate to `http.Server.Shutdown(ctx)` and preserve its error behavior.
- Keep running-server graceful shutdown behavior unchanged.

## Tradeoffs

Repeated shutdown after a successful shutdown may still surface the standard library's server state if an HTTP server exists. The production-critical guarantee is no panic and safe cleanup from partially initialized states.
