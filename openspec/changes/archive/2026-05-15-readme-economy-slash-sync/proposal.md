## Why

The README internal CLI inventory still leaves the escrow part of the economy row in a hyphen-compressed form: `escrow status-list-show/sentinel status`. The current architecture inventory and real command surface are clearer in slash-separated form.

## What Changes

- update the README internal tree economy row to use `escrow status/list/show/sentinel status`
- sync the existing economy inventory guard and main docs-only spec with that slash-form contract
- archive the follow-up after verification

## Impact

- cleaner README inventory wording
- better parity with the architecture inventory
- less ambiguity about the escrow command paths
