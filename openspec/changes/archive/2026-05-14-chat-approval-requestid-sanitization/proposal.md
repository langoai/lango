## Why

Approval transcript events already sanitize event text and request-summary previews, but the compact request-id annotation still truncates the raw ID directly. A malformed request ID can still inject control sequences into the transcript traceability surface.

## What Changes

- Sanitize approval request IDs before compacting them for transcript annotations.
- Add regression coverage for escaped and multiline request IDs.
- Record the request-id sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Completes the plain-text rendering baseline for approval transcript traceability.
- Prevents malformed request IDs from destabilizing compact approval annotations.
