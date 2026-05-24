## Overview

The change-passphrase and recovery-restore commands already use small execution seams. Extending those seams to accept the command error writer is enough to bring the remaining warning paths under command-stream control.

## Decisions

### Pass the command error writer into the execution seams

The seams now accept an `io.Writer` for notices and warnings, while leaving the success message on the command output stream.

### Keep warning semantics unchanged

Only the destination stream changes. The wording and conditions for keyfile/keyring notices remain the same.

## Non-Goals

- No change to passphrase rotation semantics
- No change to recovery semantics
