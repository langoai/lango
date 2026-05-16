## Overview

Use the Cobra command writer as the only human-readable output sink for cron commands.

## Decisions

- keep existing text and table formats unchanged
- replace direct `fmt.Print*` calls with `fmt.Fprint*`/`fmt.Fprintf` against `cmd.OutOrStdout()`
- replace `tabwriter.NewWriter(os.Stdout, ...)` with `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`
- verify routing with command-level tests that set `cmd.SetOut(...)`

## Risks

- tests that only intercept `os.Stdout` would no longer observe routed output
- mitigated by adding dedicated command-writer tests in the cron package
