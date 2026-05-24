## Overview

The P2P package already defines `executeP2PCmd(...)`, so the remaining git guidance tests can use it directly without introducing new helpers.

## Decisions

### Standardize on package-local command capture

All P2P git guidance regressions now capture output through the package-local command helper.

## Non-Goals

- No runtime behavior changes
- No migration of unrelated P2P tests in this change
