## Overview

Use the Cobra command writer as the only output sink for `lango graph query`.

## Decisions

- preserve existing query semantics and payload shapes
- replace direct `fmt.Print*`, `json.NewEncoder(os.Stdout)`, and `tabwriter.NewWriter(os.Stdout, ...)` usage with `cmd.OutOrStdout()`
- verify the contract with command-level tests that capture `cmd.SetOut(...)`

## Risks

- none beyond test and documentation updates
