## Overview

Use the Cobra command writer as the only output sink for `lango graph import`.

## Decisions

- preserve current runtime semantics: file input is JSON only, and `--json` controls output mode
- replace direct `fmt.Print*` and `json.NewEncoder(os.Stdout)` usage with `cmd.OutOrStdout()`
- verify the contract with command-level tests that capture `cmd.SetOut(...)`

## Risks

- none beyond updating docs/specs to reflect the actual command shape
