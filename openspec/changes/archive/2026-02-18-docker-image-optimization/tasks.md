## 1. Build Context Optimization

- [x] 1.1 Create `.dockerignore` excluding `.git`, `.claude`, `openspec/`, `bin/`, test artifacts, secrets, and IDE files

## 2. Dockerfile Optimization

- [x] 2.1 Add `ARG WITH_BROWSER=false` at top level and in runtime stage
- [x] 2.2 Add `--no-install-recommends` to all `apt-get install` commands
- [x] 2.3 Remove `curl` from runtime image dependencies
- [x] 2.4 Make Chromium installation conditional on `WITH_BROWSER=true`
- [x] 2.5 Merge user/group creation and data directory setup into single RUN layer
- [x] 2.6 Update HEALTHCHECK to use `lango health` CLI command instead of curl

## 3. Health Check CLI Command

- [x] 3.1 Add `lango health` command to `cmd/lango/main.go` with HTTP GET to `/health`
- [x] 3.2 Add `--port` flag with default value 18789
- [x] 3.3 Set 5-second HTTP client timeout

## 4. Remote Browser Support

- [x] 4.1 Add `RemoteBrowserURL` field to `BrowserToolConfig` in `internal/config/types.go`
- [x] 4.2 Add `RemoteBrowserURL` field to `browser.Config` struct in `internal/tools/browser/browser.go`
- [x] 4.3 Update `ensureBrowser()` to check `RemoteBrowserURL` config, then `ROD_BROWSER_WS` env var, before local launcher
- [x] 4.4 Wire `RemoteBrowserURL` from config to `browser.Config` in `internal/app/app.go`

## 5. Docker Compose Profiles

- [x] 5.1 Add `profiles: [default, browser]` to existing lango service
- [x] 5.2 Add `lango-browser` service with `WITH_BROWSER=true` build arg and `browser` profile
- [x] 5.3 Add `lango-sidecar` service with `ROD_BROWSER_WS` env var and `browser-sidecar` profile
- [x] 5.4 Add `chrome` service using `chromedp/headless-shell:latest` with 512M memory limit and `browser-sidecar` profile

## 6. Verification

- [x] 6.1 Run `go build ./...` and verify success
- [x] 6.2 Run `go test ./...` for changed packages and verify all pass
