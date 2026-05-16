## Why

Recovery transcript rows already sanitize `causeClass` during rendering, but `appendRecovery()` still stores raw `causeClass` metadata in the transcript model. That leaves recovery entries one step out of sync with the display-safe baseline used by the rest of the chat transcript.

## What Changes

- Sanitize stored recovery `causeClass` metadata at append time.
- Add regression coverage for stored sanitized recovery metadata.
- Record the append-time coherence contract in OpenSpec.

## Impact

- Aligns recovery transcript storage with its already-hardened rendering path.
- Reduces the risk of future raw-metadata regressions in alternate transcript consumers.
