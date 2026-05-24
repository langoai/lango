## Context

`telegram.New` builds the Bot API client with `cfg.HTTPClient` when provided, otherwise it currently constructs `&http.Client{}`. That default has no timeout, so network stalls can block startup during `getMe` or later Bot API calls. The channel's long-poll update request uses a 60 second Telegram timeout, so the HTTP client timeout must be longer than that long-poll window.

## Decision

Use an internal default HTTP client timeout of 90 seconds for Telegram Bot API requests when `Config.HTTPClient` is nil. This bounds stalled requests while leaving enough headroom for the existing 60 second long-poll update timeout.

## Approach

- Add `defaultTelegramHTTPClientTimeout = 90 * time.Second`.
- Add a small resolver that returns the injected client unchanged or creates the default client with the bounded timeout.
- Use the resolver in `New` before constructing the Bot API wrapper.
- Test the resolver directly so the default and injection semantics are pinned without reaching real Telegram.

## Non-Goals

- Do not expose a new config option in this slice.
- Do not change Telegram long-poll timeout behavior.
- Do not change file download timeout behavior; downloads already use a request context with a 30 second deadline.
