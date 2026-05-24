# Guard Core Stdio Stream References

## Why

The repository already has executable guards for raw prints and direct standard-stream references in `cmd/` entrypoints and `internal/cli/` production code. General `internal/` packages still rely on manual review to keep direct `os.Stdin`, `os.Stdout`, and `os.Stderr` usage constrained to intentional seams.

That creates drift risk: future core packages can add direct process-global stream references without an executable failure, undermining the seam work already added for passphrase acquisition, logging, sandbox workers, broker stderr, tracing, bootstrap, approval, and exec warnings.

## What Changes

- Add a repository-level test guard for non-CLI `internal/` production code.
- Allow only documented seam/default lines and generated or test-guard packages.
- Add fixture coverage proving the guard rejects a new direct standard-stream reference outside the allowlist.

## Impact

- No production behavior change.
- Future unapproved direct core stdio references fail tests.
- Existing intentional seams remain allowed explicitly by file and line fragment.
