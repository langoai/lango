## Overview

Use the Cobra command writer as the only output sink for `lango agent tools`.

## Decisions

- preserve current category list and JSON payload shapes
- replace direct `fmt.Print*`, `json.NewEncoder(os.Stdout)`, and `tabwriter.NewWriter(os.Stdout, ...)` usage with `cmd.OutOrStdout()`
- verify the contract with command-level tests that capture `cmd.SetOut(...)`

## Risks

- none beyond test updates, since the output content remains the same
