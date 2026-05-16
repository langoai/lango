## Overview

The P2P package already defines `executeP2PCmd(...)`, so the remaining workspace guidance and help tests can use it directly without introducing new helpers.

## Decisions

### Standardize on package-local command capture

All P2P workspace regressions now capture output through the package-local command helper.

## Non-Goals

- No runtime behavior changes
- No migration of unrelated P2P tests in this change
