## Overview

Both CLI packages already had local helpers (`executeLearningCmd`, `executeLibrarianCmd`) that capture command writers directly. The remaining global-capture cases were limited to error-path assertions, so this is a surgical test-only cleanup.

## Decisions

### Use existing local helpers only

No new test infrastructure is introduced. The change simply standardizes all assertions in these two packages on the helpers they already define.

## Non-Goals

- No runtime behavior changes
- No migration of unrelated packages in this change
