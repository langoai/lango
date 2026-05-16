## Why

The settings TUI still treated legacy SQLCipher database encryption as an editable configuration surface even though the current runtime ignores those fields and relies on broker-managed payload protection instead.

## What Changes

- render the legacy DB encryption settings form as a read-only deprecation notice
- add a reusable read-only TUI field type and help footer behavior for informational forms
- stop applying legacy DB encryption form keys back into runtime config state
- add regression tests for the read-only form and read-only TUI field behavior
- sync the main settings specs

## Impact

- removes a misleading operator-facing control from `lango settings`
- keeps legacy config visibility for old profiles without implying SQLCipher support
- reduces the chance that deprecated config fields are mistaken for active security controls
