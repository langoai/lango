## Why

The root `lango status` table path writes directly to process stdout instead of the Cobra command writer. That breaks output capture in tests, wrappers, and non-interactive tooling even though sibling status subcommands already respect the command output stream.

## What Changes

- route root status table output through `cmd.OutOrStdout()`
- add a command-level regression test that captures root status output
- sync status CLI spec and docs with the command-writer contract

## Impact

- restores consistent CLI output routing across status commands
- improves testability and wrapper integration without changing user-facing content
