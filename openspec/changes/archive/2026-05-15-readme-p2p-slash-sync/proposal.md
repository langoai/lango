## Why

The README internal CLI inventory already exposes the detailed P2P subcommand slices, but it still uses hyphen-compressed shorthand like `list-add-remove` and `status-test-cleanup` instead of the clearer slash-separated form used elsewhere.

## What Changes

- update the README internal tree P2P row to slash-separated subcommand slices
- sync the existing P2P inventory guard and main specs with that slash-form contract

## Impact

- clearer README inventory wording
- better parity with the architecture inventory
- stronger regression protection against stale hyphen shorthand
