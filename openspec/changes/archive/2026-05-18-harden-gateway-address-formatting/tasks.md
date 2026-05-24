## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for gateway address formatting hardening
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [x] 2.1 Add failing CLI gateway resolver test for IPv6 host formatting
- [x] 2.2 Add failing P2P provenance test for explicit trailing-slash address normalization
- [x] 2.3 Add failing P2P provenance test for configured gateway fallback
- [x] 2.4 Add failing P2P provenance docs guard for gateway address wording
- [x] 2.5 Add doctor companion check coverage for wildcard bind hosts

## 3. Implementation

- [x] 3.1 Introduce shared gateway address formatting helper
- [x] 3.2 Route CLI resolver, serve summary, doctor check, and gateway listen formatting through the helper
- [x] 3.3 Route P2P provenance gateway address resolution through the shared CLI resolver
- [x] 3.4 Update P2P CLI docs for `--addr` override, normalization, and configured fallback

## 4. Review And Verification

- [x] 4.1 Run focused gateway formatter and P2P provenance tests
- [x] 4.2 Run subagent review for resolver/docs/spec coverage
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
