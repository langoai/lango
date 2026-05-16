## Overview

The graph CLI suite already defines `executeGraphCommand(...)`, so the remaining global-capture cases can be replaced directly without introducing new helpers or touching runtime behavior.

## Decisions

### Standardize on package-local graph command capture

Both success and error-path graph assertions now rely on the same local command-writer helper.

## Non-Goals

- No runtime behavior changes
- No migration of unrelated packages in this change
