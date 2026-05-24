## Why

`lango doctor` still renders and serializes raw check names, messages, details, fix actions, and structured trace metadata directly from check results. That leaves both operator-facing TUI output and machine-readable JSON output vulnerable to control-sequence leakage from runtime-fed diagnostics.

## What Changes

- sanitize doctor renderer text fields before TUI rendering
- sanitize doctor JSON output fields, including structured trace metadata
- add regression coverage for malformed doctor result text
- sync doctor specs and CLI docs with the new text-sanitization contract

## Impact

- hardens `lango doctor` without changing check semantics
- keeps operator and automation surfaces aligned on the same plain single-line baseline
