## Why

The status CLI package still maintained its own pretty-JSON encoder setup even though the repository already has a shared CLI JSON writer helper. Keeping a package-local duplicate increases maintenance cost and lets formatting behavior drift over time.

## What Changes

- route `lango status` pretty-printed JSON output through the shared CLI JSON writer helper
- preserve the current payload shapes and output modes while removing duplicate encoder setup

## Impact

- lower maintenance cost for the status CLI package
- one shared pretty-JSON writer path across more CLI surfaces
