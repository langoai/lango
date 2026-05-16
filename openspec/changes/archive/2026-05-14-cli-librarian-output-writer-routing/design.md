## Overview

Use the Cobra command writer as the only output sink for librarian inspection commands.

## Decisions

- preserve current text and JSON payloads
- replace direct `fmt.Print*`, `json.NewEncoder(os.Stdout)`, and `tabwriter.NewWriter(os.Stdout, ...)` usage with `cmd.OutOrStdout()`
- verify the contract with command-level tests that capture `cmd.SetOut(...)`

## Risks

- no behavioral risk beyond making output capture deterministic for wrappers and tests
