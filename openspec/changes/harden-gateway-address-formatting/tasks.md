## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for gateway address formatting hardening
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [ ] 2.1 Add failing CLI gateway resolver test for IPv6 host formatting
- [ ] 2.2 Add failing P2P provenance test for explicit trailing-slash address normalization
- [ ] 2.3 Add failing P2P provenance test for configured gateway fallback

## 3. Implementation

- [ ] 3.1 Introduce shared gateway address formatting helper
- [ ] 3.2 Route CLI resolver, serve summary, doctor check, and gateway listen formatting through the helper
- [ ] 3.3 Route P2P provenance gateway address resolution through the shared CLI resolver
- [ ] 3.4 Update P2P CLI docs for `--addr` override, normalization, and configured fallback

## 4. Review And Verification

- [ ] 4.1 Run focused gateway formatter and P2P provenance tests
- [ ] 4.2 Run subagent review for resolver/docs/spec coverage
- [ ] 4.3 Run `go build ./...`
- [ ] 4.4 Run `go test ./...`
- [ ] 4.5 Run `git diff --check`
- [ ] 4.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 4.7 Archive the completed OpenSpec change
