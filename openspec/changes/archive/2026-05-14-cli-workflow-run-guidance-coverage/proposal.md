## Why

`lango workflow run` now routes non-error output through the command writer, but only the schedule-not-implemented path is covered by regression tests. The direct-execution fallback guidance for unavailable runtime and disabled workflow engine remains weakly verified.

## What Changes

- Add command-level regression coverage for the server-unavailable guidance path
- Add command-level regression coverage for the workflow-engine-disabled guidance path
- Update docs and OpenSpec to describe these validated fallback messages explicitly

## Impact

- Improves confidence that workflow run remains operator-friendly when direct execution is not available
- Extends coverage of a recently hardened output surface without changing behavior
