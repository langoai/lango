## Why

The recovery transcript row already advertises a compact one-line event in public docs, but its `causeClass` text is still rendered raw. If upstream recovery metadata contains newlines or excess whitespace, the transcript can break that one-line contract even though the row is otherwise width-clamped.

## What Changes

- Normalize recovery `causeClass` text to a single line before rendering.
- Add regressions for multiline recovery metadata.
- Record the single-line contract in OpenSpec and keep public docs aligned.

## Impact

- Makes the recovery transcript row match its documented compact one-line behavior.
- Prevents noisy orchestration metadata from destabilizing the chat transcript layout.
