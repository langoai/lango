## Overview

`storagebroker.Start` should keep the same public API while avoiding a hard
dependency on process-global stderr inside command construction. The change is
limited to the broker client boundary.

## Design

- Add an unexported package-level stderr seam initialized to `os.Stderr`.
- Move command construction into an unexported helper that accepts the
  executable path and stderr writer.
- Keep `Start(ctx)` responsible for resolving `os.Executable`, invoking the
  helper, starting the command, and creating the client.
- Continue to create stdin/stdout pipes and mark them close-on-exec exactly as
  before.

## Error Handling

The helper will preserve existing error wrapping for stdin and stdout pipe
setup. `Start` will continue wrapping executable resolution and process start
errors with the existing messages.

## Testing

Add a unit test that constructs a broker command with a `bytes.Buffer` stderr
writer and asserts `cmd.Stderr` is exactly that writer. This verifies the seam
without launching a subprocess or depending on global stdio reassignment.
