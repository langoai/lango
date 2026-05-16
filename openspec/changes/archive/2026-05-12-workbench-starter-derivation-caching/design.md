## Overview

This is a behavior-preserving performance and architecture cleanup.

## Design Decisions

### Cache derived prompt state at page construction

The page already receives stable inputs like `workDir`. It now converts those into:

- the starter prompt list
- the default starter prompt

once at construction time, then reuses the cached values throughout render and key-handling paths.

### No user-visible contract changes

The quick-start prompt text, default prompt selection, and submit flow remain unchanged. This change only removes repeated derivation work from the hot render path.
