## Overview

Use the Cobra command writer as the only output sink for `lango agent status`.

## Decisions

- preserve current text and JSON payloads
- replace direct `fmt.Print*` and `json.NewEncoder(os.Stdout)` usage with `cmd.OutOrStdout()`
- update existing status tests to assert writer-based capture directly

## Risks

- none beyond updating test helpers away from process-global stdout interception
