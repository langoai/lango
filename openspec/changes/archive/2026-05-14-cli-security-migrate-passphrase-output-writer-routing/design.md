## Overview

`lango security migrate-passphrase` mixes deprecation output, interactive prompt flow, and the actual migration helper. A narrow command-level seam plus an explicit writer passed into the migration helper is enough to make output routing deterministic without changing the migration behavior.

## Decisions

### Introduce a small migrate-passphrase execution seam

The seam owns bootstrap, interactive checks, prompt collection, and the call into `migrateSecrets(...)`. Tests can replace it with a deterministic stub and avoid prompts or storage setup.

### Pass the command writer into `migrateSecrets`

The helper now accepts an `io.Writer` so progress messages follow the invoking command stream instead of process-global stdout.

## Non-Goals

- No change to migration semantics
- No change to the deprecation warning staying on stderr
