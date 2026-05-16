## Overview

The prompt package already exposes seams for prompt output, stdin fd lookup, and password reading. The smallest useful next step is to use those seams to verify the successful confirmation path directly.

## Decisions

### Reuse the existing prompt seams

No new seams are introduced. The success regression simply reuses the existing output/fd/read seams and asserts their coordinated behavior.

## Non-Goals

- No runtime behavior changes
- No prompt API redesign
