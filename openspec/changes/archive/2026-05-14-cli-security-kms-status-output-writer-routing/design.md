## Overview

Use the Cobra command writer as the only output sink for `lango security kms status`.

## Decisions

- preserve the current KMS status payload, including error-state rendering when KMS support is not compiled
- test via config-backed bootloaders rather than adding a new KMS provider seam
- leave other KMS subcommands for later; this change is status-only

## Risks

- none beyond output sink changes, since status computation itself is unchanged
