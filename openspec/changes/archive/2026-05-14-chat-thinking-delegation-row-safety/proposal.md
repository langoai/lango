## Why

Recent transcript hardening covered compact status, approval, recovery, and system rows, but `thinking` and `delegation` rows still have width and single-line gaps. Long previews, actor names, or multiline reasons can still spill past the transcript width or break the one-line contract.

## What Changes

- Make `thinking` transcript rows width-safe and single-line-safe.
- Make `delegation` transcript rows width-safe and single-line-safe.
- Add regressions for narrow-width and multiline rendering.
- Record the rendering contract in OpenSpec and downstream feature docs.

## Impact

- Prevents narrow-terminal overflow for two more transcript event surfaces.
- Keeps runtime transcript rows compact and visually stable under noisy payloads.
