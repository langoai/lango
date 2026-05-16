## Why

The dead-letter filter validation path still echoes raw invalid flag values in non-JSON error strings. That leaves one remaining status CLI error surface where ANSI/OSC control text or embedded newlines from user input can leak directly to the terminal.

## What Changes

- sanitize invalid subtype, family, and RFC3339 flag values before interpolating them into CLI validation errors
- add regression coverage for malformed invalid-flag text
- sync the status spec and docs with the validation-error contract

## Impact

- hardens non-JSON validation feedback without changing validation behavior
- keeps operator-facing status errors aligned with the rest of the plain single-line baseline
