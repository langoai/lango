## Why

`lango librarian status` and `lango librarian inquiries` still write human-readable and JSON output directly to process stdout instead of the Cobra command writer. That weakens wrapper capture and leaves the librarian inspection surface inconsistent with the hardened CLI baseline.

## What Changes

- route `librarian status` table and JSON output through `cmd.OutOrStdout()`
- route `librarian inquiries` table and JSON output through the same writer
- add command-level writer capture tests for representative librarian flows
- sync librarian CLI docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for librarian inspection commands
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
