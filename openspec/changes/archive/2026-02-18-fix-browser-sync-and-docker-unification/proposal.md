## Why

The Docker slim + browser sidecar architecture causes a fatal `sync: unlock of unlocked mutex` crash. The root cause is `browser.go:ensureBrowser()` resetting `sync.Once` via value assignment (`t.browserOnce = sync.Once{}`), which creates a data race in concurrent environments. Additionally, the multi-image/sidecar Docker setup adds unnecessary operational complexity. Unifying to a single image with Chromium always included simplifies deployment and eliminates the remote browser connection code entirely.

## What Changes

- **BREAKING**: Remove `RemoteBrowserURL` config field from `BrowserToolConfig` and `browser.Config`
- **BREAKING**: Remove `ROD_BROWSER_WS` environment variable support and remote browser connection code
- Replace `sync.Once` + value-reset pattern with `sync.Mutex` + `bool` guard for retryable browser initialization
- Simplify Dockerfile to always include Chromium (remove `WITH_BROWSER` build arg)
- Simplify `docker-compose.yml` to a single `lango` service (remove `lango-browser`, `lango-sidecar`, `chrome` services and all profiles)
- Remove `docker-build-browser`, `docker-up-browser`, `docker-up-sidecar` Makefile targets
- Delete `remote-browser` spec (capability fully removed)

## Capabilities

### New Capabilities

(none)

### Modified Capabilities
- `tool-browser`: Replace `sync.Once` with `sync.Mutex` + `bool` pattern; remove remote browser config and connection code
- `docker-deployment`: Single image with Chromium always included; single compose service without profiles

## Impact

- `internal/tools/browser/browser.go`: Struct fields, `ensureBrowser()`, `initBrowser()`, `Close()` rewritten
- `internal/config/types.go`: `BrowserToolConfig.RemoteBrowserURL` field removed
- `internal/app/app.go`: `RemoteBrowserURL` wiring removed from browser config construction
- `Dockerfile`: Conditional Chromium install replaced with unconditional install
- `docker-compose.yml`: Reduced to single service, no profiles
- `Makefile`: Browser-specific targets removed, compose commands simplified
- `README.md`: Docker section simplified, remote browser docs removed
- `openspec/specs/remote-browser/`: Entire spec deleted
