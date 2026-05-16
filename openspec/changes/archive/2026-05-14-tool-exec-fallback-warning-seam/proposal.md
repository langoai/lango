## Why

The exec tool's fail-open warning is operator-visible and covered by docs/specs, but the implementation still wrote directly to process stderr, which made the one-shot warning contract harder to verify deterministically.

## What Changes

- Add a seam-aware stderr writer for exec fallback warnings
- Add regression coverage proving the warning fires only once per process
- Update docs and OpenSpec to reflect the deterministic warning path

## Impact

- Improves confidence in a security-relevant fail-open signal
- Keeps runtime behavior unchanged while making the contract testable
