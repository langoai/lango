## Overview

Use the Cobra command writer as the only output sink for `lango agent graph`.

## Decisions

- preserve current text and JSON payload shapes
- reuse the existing seeded `traceBootLoader` test helper
- keep edge rendering as plain text while moving all writes to `cmd.OutOrStdout()`

## Risks

- none beyond test and docs updates, since the graph computation itself is unchanged
