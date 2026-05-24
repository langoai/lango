## 1. Tests
- [x] 1.1 Add failing `config set` tests for preserving existing explicit keys, marking context-related set paths explicit, and avoiding mutation on invalid paths.
- [x] 1.2 Add failing `onboard` tests for preserving existing explicit keys and saving preset explicit keys.

## 2. Implementation
- [x] 2.1 Update `configcmd.NewSetCmd` to carry explicit keys from load to save.
- [x] 2.2 Update root `config set` wiring to pass `boot.ExplicitKeys` and save the updated explicit-key map.
- [x] 2.3 Preserve existing cleanup and error behavior.
- [x] 2.4 Update `onboard` load/save flow to carry explicit keys for existing and preset-backed profiles.

## 3. Specs and Verification
- [x] 3.1 Sync main `config-cli-commands` and `config-system` specs.
- [x] 3.2 Validate the OpenSpec change in strict mode.
- [x] 3.3 Run focused config command tests.
- [x] 3.4 Run `go build ./...` and `go test ./...`.
- [x] 3.5 Run subagent-driven review.
- [x] 3.6 Archive the OpenSpec change and commit this scoped unit separately.
