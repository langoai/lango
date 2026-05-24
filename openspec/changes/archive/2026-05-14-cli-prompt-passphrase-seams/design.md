## Overview

The hidden-input prompt surface is small, so package-level seams are enough. Production continues to use stdout, the real stdin file descriptor, and `term.ReadPassword`, while tests can replace each dependency independently.

## Decisions

### Add output/fd/read seams

`Passphrase(...)` now depends on three package-level seams:
- prompt output writer
- stdin fd lookup
- password reader function

### Keep public API unchanged

`Passphrase(...)` and `PassphraseConfirm(...)` signatures stay the same. The seams only improve testability.

## Non-Goals

- No change to hidden-input runtime behavior
- No broader prompt API redesign
