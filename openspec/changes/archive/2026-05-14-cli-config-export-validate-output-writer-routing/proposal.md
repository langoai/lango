## Why

`lango config export` and `lango config validate` still write success output directly to process stdout instead of the Cobra command writer. `config export` also emits its plaintext warning through global stderr. That breaks output capture for wrappers and tests on two important profile-management surfaces.

## What Changes

- route `config export` JSON output through `cmd.OutOrStdout()`
- route the export plaintext warning through `cmd.ErrOrStderr()`
- route `config validate` success output through `cmd.OutOrStdout()`
- add command-level capture coverage for export and validate
- sync config CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for profile export/validation
- keeps user-visible output unchanged while aligning with the rest of the CLI
