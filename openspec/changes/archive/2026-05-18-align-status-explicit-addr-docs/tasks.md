## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for status explicit address docs alignment
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [x] 2.1 Add a failing docs guard for status explicit `--addr` probe/display wording

## 3. Documentation

- [x] 3.1 Update `docs/cli/status.md` to describe normalized explicit `--addr` probe/display behavior
- [x] 3.2 Update the status JSON example so `gateway` matches a custom explicit target

## 4. Review And Verification

- [x] 4.1 Run the focused docs guard test
- [x] 4.2 Run subagent review for docs/spec/test coverage
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
