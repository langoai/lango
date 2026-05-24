## Overview

Use the Cobra command writer as the only output sink for `lango security keyring status`.

## Decisions

- preserve current text and JSON payload shapes
- add a `detectSecureProvider` seam so tests can control environment-dependent keyring detection
- verify the contract with command-level tests that capture `cmd.SetOut(...)`

## Risks

- none beyond test updates, since the output content remains the same
