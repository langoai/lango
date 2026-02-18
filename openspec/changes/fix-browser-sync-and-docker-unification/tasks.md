## 1. Browser Init Crash Fix

- [x] 1.1 Replace `sync.Once` + `browserErr` fields with `sync.Mutex` + `initDone bool` in `Tool` struct (`internal/tools/browser/browser.go`)
- [x] 1.2 Rewrite `ensureBrowser()` to use mutex-based guard pattern with automatic retry on failure
- [x] 1.3 Update `Close()` to reset `initDone = false` under `initMu` lock instead of `browserOnce = sync.Once{}`
- [x] 1.4 Remove `RemoteBrowserURL` field from `browser.Config` struct
- [x] 1.5 Remove remote browser connection code from `initBrowser()` (config check + `ROD_BROWSER_WS` env var + remote connect block)
- [x] 1.6 Remove unused `"os"` import

## 2. Config and Wiring Cleanup

- [x] 2.1 Remove `RemoteBrowserURL` field from `BrowserToolConfig` in `internal/config/types.go`
- [x] 2.2 Remove `RemoteBrowserURL` wiring from `browser.Config{}` construction in `internal/app/app.go`

## 3. Docker Simplification

- [x] 3.1 Remove `ARG WITH_BROWSER=false` from Dockerfile (both builder and runtime stages)
- [x] 3.2 Replace conditional Chromium install with unconditional `apt-get install chromium`
- [x] 3.3 Remove `lango-browser`, `lango-sidecar`, `chrome` services from `docker-compose.yml`
- [x] 3.4 Remove all `profiles` configuration from `docker-compose.yml`

## 4. Makefile Cleanup

- [x] 4.1 Remove `docker-build-browser` target
- [x] 4.2 Remove `docker-up-browser` and `docker-up-sidecar` targets
- [x] 4.3 Simplify `docker-up` to `docker compose up -d` (no profile flag)
- [x] 4.4 Simplify `docker-down` to `docker compose down` (no multi-profile flags)
- [x] 4.5 Remove deleted targets from `.PHONY` list

## 5. Documentation

- [x] 5.1 Remove `tools.browser.remoteBrowserUrl` row from README config table
- [x] 5.2 Update `tools.browser.enabled` description to remove "or remote browser"
- [x] 5.3 Replace Docker "Image Variants" and "Compose Profiles" sections with single image/compose docs
- [x] 5.4 Remove "Remote Browser Support" section from README

## 6. OpenSpec Updates

- [x] 6.1 Update `openspec/specs/docker-deployment/spec.md` — single image, no profiles
- [x] 6.2 Update `openspec/specs/tool-browser/spec.md` — mutex pattern, remove remote browser scenario
- [x] 6.3 Delete `openspec/specs/remote-browser/spec.md`

## 7. Verification

- [x] 7.1 Run `go build ./...` — confirm compilation
- [x] 7.2 Run `go test ./...` — confirm all tests pass
