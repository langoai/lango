## Overview

Use the Cobra command writer as the only output sink for `lango mcp add` and `lango mcp remove`.

## Decisions

- preserve existing confirmation text
- replace direct `fmt.Printf` calls with `fmt.Fprintf(cmd.OutOrStdout(), ...)`
- verify the contract with project-scope tests that capture `cmd.SetOut(...)`

## Risks

- none beyond test expectation updates, because output content does not change
