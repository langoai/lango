## Why

The quick references and dedicated run docs already describe `lango run journal <run-id>`, but the README internal tree and architecture inventory still stop at a bare `journal`.

## What Changes

- update the README internal tree and architecture inventory to include the `journal <run-id>` placeholder
- sync the existing remaining-inventory guard and main specs with that placeholder-aware run contract

## Impact

- more truthful run inventory docs
- clearer operator guidance for RunLedger journal inspection
- stronger regression protection against placeholder loss
