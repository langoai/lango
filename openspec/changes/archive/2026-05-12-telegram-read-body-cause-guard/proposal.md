## Why

The Telegram download path already distinguishes HTTP failures and empty bodies, but it does not yet directly guard the body-read failure path. Without that regression, a real I/O interruption could be flattened into a generic error or confused with the empty-body contract.

## What Changes

- Add Telegram download coverage for body-read failures.
- Require the error to preserve both the high-level `read file body` label and the underlying I/O cause.
- Sync the production-readiness spec to make that failure shape explicit.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: Telegram file download failures now have a direct regression for body-read cause preservation.

## Impact

- Affected code: `internal/channels/telegram/telegram_download_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
