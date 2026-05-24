## Why

`lango config create` supports `--preset`, but the public config CLI docs still describe only the default-profile path. That hides a faster setup workflow that directly improves first-run usability.

## What Changes

- document `lango config create <name> [--preset <name>]` in the public config CLI docs
- mention preset-backed profile creation in the CLI index

## Impact

- better discoverability for the fast-start preset workflow
- closer alignment between the public docs and the implemented CLI
