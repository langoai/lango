## Why

Compact chat `status` rows are now width-safe, but they still accept raw content with embedded newlines. That means multiline error or warning text can spill into multiple lines and break the compact-row contract.

## What Changes

- Normalize compact status and approval row content to single-line-safe display text before truncation.
- Add regressions for multiline status and approval event content.
- Record the single-line compact-row contract in OpenSpec.

## Impact

- Prevents multiline status text from breaking compact transcript rows.
- Keeps compact transcript surfaces consistent under noisy error messages.
