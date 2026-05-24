# Align Cron CLI Documented Flags

## Why

The public cron documentation shows CLI examples that do not match the current
command surface:

- `lango cron add --deliver-to ...` is documented, while the CLI only accepts
  `--deliver`.
- Per-job `--timeout` is documented and already supported by the cron domain
  model and agent tool handler, but not by `lango cron add`.
- `pause`, `resume`, `delete`, and `history` are documented with `--id`, while
  the CLI only accepts a positional `id-or-name`.

These mismatches make documented setup commands fail before users reach the
actual scheduler.

## What Changes

- Add documented compatibility flags to the cron CLI without removing existing
  positional and `--deliver` forms.
- Thread `--timeout` into `cron.Job.Timeout` for CLI-created jobs.
- Keep public docs aligned with the supported command surface.

## Impact

- Modified capabilities: `cron-scheduling`, `downstream-docs-sync`.
- CLI behavior becomes more forgiving and matches existing docs.
- No scheduler, storage schema, or runtime execution semantics change.
