## Overview

Use Cobra command streams for both output and confirmation input in `lango memory clear`.

## Decisions

- preserve current prompt and success strings
- replace direct `fmt.Print*` output with `cmd.OutOrStdout()`
- replace direct `os.Stdin` scanning with `cmd.InOrStdin()`
- verify the contract using command-level tests that set both output and input buffers

## Risks

- none beyond ensuring tests cover both interactive and `--force` paths
