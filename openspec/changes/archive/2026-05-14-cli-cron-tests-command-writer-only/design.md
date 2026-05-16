## Overview

The cron CLI already exposes a local `executeCronCommand(...)` helper that captures the Cobra command writers. The remaining tests can use that helper directly instead of the repo-wide stdout interception utility.

## Decisions

### Standardize on local command capture for cron CLI tests

Error-path assertions now inspect the returned error from `executeCronCommand(...)` instead of relying on `testutil.ExecCmd`.

## Non-Goals

- No runtime behavior changes
- No migration of other packages in this change
