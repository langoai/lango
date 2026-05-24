## Why

The workflow CLI validate command repeated the same pretty-JSON writer setup for success and error payloads. That duplication increases maintenance cost without changing user-facing behavior.

## What Changes

- route workflow CLI validate JSON responses through a shared package-local helper
- keep existing payload shapes unchanged while removing repeated encoder setup

## Impact

- lower maintenance cost for the workflow CLI package
- one place to adjust pretty-JSON writer behavior if needed later
