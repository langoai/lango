## 1. Goreleaser Vec Build Tag

- [x] 1.1 Add `vec` tag to `lango` build in `.goreleaser.yaml`
- [x] 1.2 Add `vec` tag to `lango-extended` build in `.goreleaser.yaml`
- [x] 1.3 Verify `goreleaser check` passes

## 2. TTY Guard for TUI Commands

- [x] 2.1 Add `prompt` package import to `cmd/lango/main.go`
- [x] 2.2 Add TTY guard to root command `RunE` (fall back to `cmd.Help()`)
- [x] 2.3 Add TTY guard to `cockpitCmd` `RunE` (return error)
- [x] 2.4 Add TTY guard to `chatCmd` `RunE` (return error)

## 3. Verification

- [x] 3.1 `go build ./...` passes
- [x] 3.2 `go test ./...` passes
