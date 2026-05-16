## Overview

Use the Cobra command writer as the only output sink for `lango graph export`.

## Decisions

- preserve the current runtime behavior: no positional file argument, output streamed to stdout, `json|csv` formats only
- replace direct `json.NewEncoder(os.Stdout)` and `csv.NewWriter(os.Stdout)` usage with `cmd.OutOrStdout()`
- verify the contract with command-level tests that capture `cmd.SetOut(...)`

## Risks

- none beyond bringing docs/specs in line with the already-implemented runtime behavior
