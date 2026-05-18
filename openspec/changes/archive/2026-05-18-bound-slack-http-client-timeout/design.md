## Context

Slack's Go SDK initializes its API client with `&http.Client{}` unless `slack.OptionHTTPClient` is provided. A zero-timeout HTTP client can hang indefinitely on stalled network operations. Lango already allows injecting `Config.HTTPClient`, so the safest narrow fix is to supply a bounded default client only when no custom client is configured.

## Decision

Add a package-level Slack default HTTP timeout and helper that returns:

- `Config.HTTPClient` unchanged when provided.
- A new `*http.Client` with the bounded timeout otherwise.

Use that helper when constructing the Slack SDK client.

## Tradeoffs

This keeps the default fixed rather than making it user-configurable. That is intentionally narrow for this change: it removes the production risk while preserving the existing public config surface.
