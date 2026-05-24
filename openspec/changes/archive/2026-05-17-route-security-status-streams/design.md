# Design: Security Status Stream Seams

## Scope

This change is limited to the non-interactive `lango security status` path and
its DB-status helper. It does not change the `--full` bootstrap path, status
rendering, passphrase acquisition behavior, or storage-broker protocol.

## Approach

`newStatusCmd` will pass `cmd.ErrOrStderr()` to `runStatusNonInteractive`, and
`runStatusNonInteractive` will pass that writer to `readDBStatusNonInteractive`.
The helper will use that writer for non-interactive passphrase warnings instead
of a package-global `os.Stderr` writer.

The status helper will also call package-local dependency seams for secure
provider detection and storage-broker startup. The secure-provider seam defaults
to the existing security package detector, and the broker seam defaults to
`storagebroker.Start`. Tests can replace both seams to verify wiring without
touching platform keyring/TPM discovery or launching a real broker child process.

## Verification

- A command-level test captures the warning on Cobra stderr and verifies stdout
  remains render-only.
- A helper-level test proves the secure-provider seam is passed into
  non-interactive acquisition.
- Helper-level tests prove the broker starter seam is used and broker-start
  failure gracefully degrades.
- Existing status rendering and invalid-output tests continue to pass.
