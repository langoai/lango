## Overview

The smartaccount root help regression is a single test, so the smallest useful change is to add a tiny package-local helper and stop depending on the repo-wide stdout interception utility.

## Decisions

### Use a package-local help capture helper

The new helper captures `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` only and is sufficient for the `--help` surface.

## Non-Goals

- No runtime behavior changes
- No migration of other smartaccount tests in this change
