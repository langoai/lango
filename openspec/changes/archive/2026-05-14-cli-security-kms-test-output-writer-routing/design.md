## Overview

`lango security kms test` validates a KMS encrypt/decrypt roundtrip and emits progress lines as it proceeds. A small provider-constructor seam is enough to make command-level tests deterministic.

## Decisions

### Add a KMS provider constructor seam

The command now resolves KMS providers through a package-level variable. Tests can replace it with a simple fake provider that performs reversible in-memory transforms.

### Route all non-error output through the Cobra writer

- Roundtrip start line uses `fmt.Fprintf(cmd.OutOrStdout(), ...)`
- Encrypt/decrypt progress lines use `fmt.Fprintf(cmd.OutOrStdout(), ...)`
- Final success line uses `fmt.Fprintln(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to KMS connectivity semantics
- No change to ciphertext/plaintext validation rules
