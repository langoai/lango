## Why

The production-readiness spec says Telegram file downloads must use HTTP GET with a 30-second timeout, but the existing tests only covered result behavior. That leaves the timeout and request-shape contract weaker than it should be.

## What Changes

- Add a Telegram download regression that inspects the outgoing HTTP request.
- Assert that downloads use HTTP GET.
- Assert that the request context carries a deadline consistent with the 30-second timeout contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: Telegram media download now has executable request-shape coverage for the HTTP GET and timeout contract.

## Impact

- Affected tests: `internal/channels/telegram/telegram_download_test.go`
