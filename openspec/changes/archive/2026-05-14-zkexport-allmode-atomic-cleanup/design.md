## Overview

The command already knows each target path as it iterates `--all` exports. Tracking successful paths and cleaning them up on the first later failure is enough to make the run atomic from the output-directory perspective.

## Decision

- Append each successfully exported path to a local slice
- On the first later failure, remove all previously created paths before returning the error
- Leave single-circuit behavior unchanged

## Consequences

- Failed `--all` runs no longer leave behind a partially-populated verifier directory
- Regression coverage distinguishes cleanup of the failing file from cleanup of earlier successful files
