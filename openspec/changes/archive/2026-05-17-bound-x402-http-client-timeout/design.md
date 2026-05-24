## Context

`payment_x402_fetch` obtains its client from `Interceptor.HTTPClient(ctx)`. The interceptor registers the Coinbase SDK payment handler and wraps a base `*http.Client`. Today that base client has no timeout, so the wrapped client inherits an unbounded transport deadline.

## Decision

Introduce a package-level default X402 HTTP client timeout and use it when creating the wrapped client. Keep this as an internal default rather than a public config field for this change, because the immediate production risk is the unbounded default.

## Tradeoffs

A fixed default is less flexible than a configurable setting, but it is safer than an unbounded client and keeps this change focused. Future work can expose config if operators need environment-specific tuning.
