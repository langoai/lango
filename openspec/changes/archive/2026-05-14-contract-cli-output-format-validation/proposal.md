## Why

The contract CLI family used a boolean `--output` flag while most other operator-facing commands use explicit string formats such as `table|json`. That made the UX inconsistent and prevented actionable validation of unknown values before config loading.

## What Changes

- change `lango contract read`, `call`, and `abi load` to use `--output table|json`
- reject unknown output values before config loading
- sync contract CLI docs and spec text to the normalized output-format contract

## Impact

- more consistent CLI UX across Lango commands
- fail-fast validation for invalid output formats
- easier wrapper and operator expectations around contract command output behavior
