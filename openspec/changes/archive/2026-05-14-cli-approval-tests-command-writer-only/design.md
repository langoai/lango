## Overview

The approval CLI already follows the Cobra writer contract. The test suite can therefore use a simple local helper instead of the repo-wide stdout interception utility.

## Decisions

### Use package-local command capture only

The existing `executeApprovalCommand(...)` helper is sufficient for all approval CLI test cases, including success, JSON, and error paths.

## Non-Goals

- No runtime behavior changes
- No broader migration of other packages off `testutil.ExecCmd` in this change
