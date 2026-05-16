## Overview

Use the Cobra command writer as the only output sink for `lango agent trace list` and `lango agent trace show`.

## Decisions

- preserve current table and JSON payload shapes
- reuse the existing seeded `traceBootLoader` test helper rather than adding a new fake trace store seam
- fix docs to use `lango agent trace show <trace-id>` explicitly

## Risks

- none beyond test and docs updates, since the trace retrieval logic itself is unchanged
