## Overview

Use the Cobra command writer as the only output sink for `lango agent hooks`.

## Decisions

- preserve current text and JSON payloads
- replace direct `fmt.Print*` and `json.NewEncoder(os.Stdout)` usage with `cmd.OutOrStdout()`
- remove process-global stdout swapping from command-level tests in favor of direct writer capture

## Risks

- none beyond test updates, since output content does not change
