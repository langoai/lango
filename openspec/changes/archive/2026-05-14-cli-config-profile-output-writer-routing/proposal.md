## Why

The profile-management subcommands under `lango config` still write prompts, confirmations, and list/import output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves the profile CLI inconsistent with the already-hardened config surfaces.

## What Changes

- route `config list`, `create`, `use`, `delete`, and `import` output through the Cobra command writer
- route delete confirmation through `cmd.InOrStdin()` / `cmd.OutOrStdout()`
- add command-level capture coverage for profile-management flows
- sync config CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for profile management
- keeps user-visible output unchanged while aligning with the rest of the CLI
