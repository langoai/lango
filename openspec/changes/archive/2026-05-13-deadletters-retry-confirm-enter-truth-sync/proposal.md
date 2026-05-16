## Why

The Dead Letters retry-confirm flow already accepts `Enter` as a submit path, but the help bar and user-facing retry label still imply that only pressing `r` again confirms the request. That makes the visible guidance narrower than the real runtime contract.

## What Changes

- Expose `Enter` as a confirm action while Dead Letters retry confirmation is pending.
- Update the retry action label and docs wording so they mention both confirm keys.
- Add regressions and extend the cockpit-pages spec for the dual confirm path.

## Impact

- Operators can discover the full confirm path from in-product guidance instead of guessing.
- Runtime help, detail text, docs, tests, and spec all describe the same retry-confirm contract.
