## Overview

Use the Cobra command writer as the only output sink for `lango memory list` and `lango memory status`.

## Decisions

- preserve existing human-readable and JSON payload shapes
- replace direct `fmt.Print*`, `json.NewEncoder(os.Stdout)`, and `tabwriter.NewWriter(os.Stdout, ...)` usage with `cmd.OutOrStdout()`
- verify the contract using temp-database command-level tests that capture `cmd.SetOut(...)`

## Risks

- no user-visible behavior change beyond making capture deterministic for wrappers and tests
