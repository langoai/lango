## Overview

Use the Cobra command writer as the only output sink for `lango security status` and its shared render helper.

## Decisions

- preserve current human-readable and JSON payload shapes
- change `renderStatus` to accept an `io.Writer`
- pass `cmd.OutOrStdout()` from both non-interactive and full-bootstrap status paths
- replace stdout capture tests with buffer-based tests

## Risks

- minimal: only the output sink changes, not the payload content
