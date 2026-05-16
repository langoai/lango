## Overview

`lango security change-passphrase` depends on interactive prompts, envelope unwrap/rewrap, and optional keyfile/keyring updates. A narrow seam around the full command execution is enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce a small change-passphrase execution seam

The seam performs the existing interactive and storage workflow in production and returns the final success message. Tests can replace it with a deterministic stub and avoid prompting or touching real envelope state.

### Keep warnings on stderr

Success output moves to `cmd.OutOrStdout()`, while existing warning messages for keyfile/keyring update issues remain on stderr.

## Non-Goals

- No change to passphrase rotation semantics
- No change to keyfile/keyring warning behavior
