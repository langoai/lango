## Overview

Use the Cobra command writer as the output sink for the non-error `lango security keyring store` paths that do not depend on interactive input.

## Decisions

- preserve current error behavior and interactive prompt path
- route the already-stored and successful storage messages through `cmd.OutOrStdout()`
- verify the already-stored path with a stub keyring provider and a minimal successful bootloader

## Risks

- minimal, because only the output sink changes for existing success paths
