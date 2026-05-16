## Why

The config profile export command still used its own ad-hoc pretty-JSON encoder setup even after the shared CLI JSON helper was introduced elsewhere. That duplication adds maintenance noise without changing behavior.

## What Changes

- route config profile export JSON output through the shared CLI JSON helper
- keep existing payload shapes unchanged while removing repeated encoder setup

## Impact

- lower maintenance cost for config profile CLI code
- one place to adjust pretty-JSON writer behavior if needed later
