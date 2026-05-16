## Overview

The economy CLI suite already uses standard Cobra command execution semantics, so a package-local output capture helper is enough for both success and help/guidance assertions.

## Decisions

### Standardize on package-local command capture

All touched economy CLI tests capture output through a local helper instead of the repo-wide stdout interception utility.

## Non-Goals

- No runtime behavior changes
- No migration of unrelated packages in this change
