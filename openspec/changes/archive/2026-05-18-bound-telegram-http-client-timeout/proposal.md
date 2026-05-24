## Why

The Telegram channel currently creates a default `http.Client` with no timeout when no client is injected. A stalled Bot API request can therefore block channel startup or runtime API calls indefinitely, which is not acceptable for production reliability.

## What Changes

- Add a bounded default Telegram Bot API HTTP client timeout.
- Keep injected `Config.HTTPClient` behavior unchanged for tests and custom deployments.
- Add regression coverage for the default timeout and injected-client preservation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `channel-telegram`: Telegram Bot API requests must use a bounded default HTTP client timeout when no client is supplied.

## Impact

- Runtime code: `internal/channels/telegram`
- Tests: `internal/channels/telegram`
- No public CLI, config, or documentation changes.
