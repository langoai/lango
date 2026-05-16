## Why

The README internal CLI inventory already exposes the provenance subcommand slices, but it still uses hyphen-compressed shorthand like `list-create-show` and `show-report` instead of the slash-separated form used in the architecture inventory.

## What Changes

- update the README internal tree provenance row to slash-separated subcommand slices
- sync the existing remaining-inventory guard and main specs with that slash-form wording

## Impact

- clearer README provenance inventory wording
- better parity with the architecture inventory
- stronger regression protection against stale hyphen shorthand
