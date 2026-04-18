## 1. Build Surface

- [x] 1.1 Remove `fts5` from default build and test commands in Makefile and Dockerfile
- [x] 1.2 Keep optional legacy `vec` guidance intact without implying FTS5 is tag-gated

## 2. Docs And Tests

- [x] 2.1 Update README and development/installation docs to describe FTS5 as always on
- [x] 2.2 Update FTS5-related test or skip messaging so it no longer instructs users to rebuild with `-tags "fts5"`

## 3. Verification

- [x] 3.1 Verify `go build ./...` and `go test ./...` succeed without the `fts5` tag
- [x] 3.2 Run `openspec validate --type change always-on-fts5`
