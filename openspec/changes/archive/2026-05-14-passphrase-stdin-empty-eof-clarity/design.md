## Overview

The stdin-pipe path uses the shared raw line reader and then applies passphrase-specific trimming. The remaining issue is classification: empty EOF currently bubbles as a low-level read failure instead of the higher-level "empty passphrase" condition.

## Decision

- Preserve raw read failures for non-EOF errors when no bytes were read
- Treat `io.EOF` with no bytes read the same as an empty passphrase line
- Preserve the successful partial-line behavior when stdin provides a passphrase without a trailing newline

## Consequences

- Empty pipes produce a clearer error
- CI and wrapper callers get more stable, intention-revealing behavior
- The stdin path remains built on the shared raw line reader
