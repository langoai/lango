## Overview

Use the Cobra command writer as the only output sink for `lango graph add`.

## Decisions

- preserve current success text and JSON payload shape
- replace direct `fmt.Printf` and `json.NewEncoder(os.Stdout)` usage with `cmd.OutOrStdout()`
- verify the contract with command-level tests that capture `cmd.SetOut(...)`

## Risks

- none beyond test and documentation updates
