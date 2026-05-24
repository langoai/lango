## Why

Recent config CLI cleanup fixed several public-documentation drifts, but those examples can easily regress because they live across README and feature docs. The repository should enforce the current config CLI contract mechanically.

## What Changes

- add an executable docs guard that rejects stale `config get --format json` examples
- reject profile-less `config export` and `config import` examples in public docs
- sync docs-only and test-coverage specs to describe the guard

## Impact

- prevents high-friction copy-paste failures in public setup docs
- keeps config CLI examples aligned with the real command contract
