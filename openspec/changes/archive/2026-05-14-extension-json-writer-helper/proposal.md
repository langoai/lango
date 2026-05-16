## Why

The extension CLI package still repeated pretty-JSON encoder setup across its inspect and list renderers. That duplication adds maintenance cost without any user-facing benefit.

## What Changes

- route extension CLI pretty-printed JSON responses through a shared package-local helper
- keep existing payload shapes and output modes unchanged while removing repeated encoder setup

## Impact

- lower maintenance cost for extension CLI JSON rendering
- one place to adjust pretty-JSON writer behavior later if needed
