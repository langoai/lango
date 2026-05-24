## Why

`lango security status` already routes its main output through Cobra writers, but the non-interactive passphrase acquisition warning still wrote directly to process-global stderr. That weakened wrapper capture and deterministic testing for the graceful-degrade path.

## What Changes

- Add seams for non-interactive passphrase acquisition and the status warning writer
- Route the warning through the seam instead of process-global stderr
- Add regression coverage for the injected warning path

## Impact

- Makes the security status degrade path more testable
- Aligns the remaining warning path with the broader command-stream hardening work
