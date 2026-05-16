## Overview

Use Cobra command streams for both output and confirmation input in `lango security keyring clear`.

## Decisions

- preserve current prompt, abort, and success strings
- allow non-stdin injected readers in tests and wrappers while still requiring `--force` for truly non-interactive terminal usage
- keep provider deletion warnings on `cmd.ErrOrStderr()`

## Risks

- minimal: behavior only changes in how streams are wired, not in the command contract itself
