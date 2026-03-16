## Why

`app.New()` transitioned from a monolithic 900-line function to an `appinit.Builder.Build()`-based 5-module system. The original plan required a parity gate before dead code deletion, but it was skipped. Without verification tests, regressions in catalog registration, lifecycle ordering, or field wiring could go undetected.

## What Changes

- Add `Names()` method to `lifecycle.Registry` for introspecting registered component names
- Add Layer 1 unit tests for pure helper functions (`buildCatalogFromEntries`, `registerPostBuildLifecycle`)
- Add Layer 2 integration tests that call real `app.New()` with default and feature-enabled configs, verifying catalog categories, lifecycle components, and app field population

## Capabilities

### New Capabilities
- `parity-verification`: Test suite verifying that the module-based `app.New()` produces identical catalog, lifecycle, and field state as expected for default and feature-enabled configurations

### Modified Capabilities

## Impact

- `internal/lifecycle/registry.go`: New `Names()` method (additive, no breaking changes)
- `internal/lifecycle/registry_test.go`: Additional test cases
- `internal/app/parity_test.go`: New test file covering helper functions and integration parity
- No API changes, no dependency additions
