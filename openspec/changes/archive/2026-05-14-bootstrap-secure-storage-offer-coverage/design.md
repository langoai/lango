## Overview

The secure-storage offer sits inside bootstrap credential acquisition and depends on passphrase source, hardware-provider availability, a confirmation prompt, and stderr reporting. Adding a few narrow seams is enough to make the path deterministic under test.

## Decisions

### Add small bootstrap seams

`acquirePassphrase`, `confirmStorePass`, and `bootstrapErrWriter` are introduced as package-level variables so tests can stub passphrase acquisition, confirmation, and stderr capture independently.

### Keep runtime behavior unchanged

Production still uses `passphrase.Acquire`, `prompt.Confirm`, and `os.Stderr`. The seams exist only to support deterministic coverage.

## Non-Goals

- No change to bootstrap passphrase acquisition semantics
- No change to hardware keyring selection logic
