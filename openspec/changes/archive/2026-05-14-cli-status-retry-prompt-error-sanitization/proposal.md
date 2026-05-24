## Why

The dead-letter retry command still interpolates raw transaction receipt IDs into confirmation prompts and non-JSON error strings. That leaves one operator-facing status path where control text can leak even though JSON outputs are already normalized.

## What Changes

- sanitize transaction receipt IDs when rendering retry confirmation prompts
- sanitize transaction receipt IDs when building non-JSON retry error strings
- add regression coverage for malformed retry prompt/error text
- sync the status spec and docs with the operator-facing retry text contract

## Impact

- hardens the remaining human-readable retry path in `lango status`
- keeps lookup behavior unchanged while making displayed text safe
