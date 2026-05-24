## 1. UX Fix

- [x] 1.1 Expose `Enter` as a confirm help binding while retry confirmation is pending.
- [x] 1.2 Update retry action wording to mention both confirm keys.
- [x] 1.3 Add regression coverage for the dual confirm path.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require both confirm keys.
- [x] 2.2 Update cockpit feature docs to describe `Enter` as an alternate confirm key.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
