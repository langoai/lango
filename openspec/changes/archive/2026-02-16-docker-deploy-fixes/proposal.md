## Why

When running lango in a Docker container, 3 issues occur: (1) the exec tool's approval chain fails entirely in environments without a TTY, causing tool usage to be denied, (2) go-rod cannot find the system chromium binary, preventing browser automation, (3) WORKDIR `/app` is owned by root so non-root users cannot write to it.

## What Changes

- Add `HeadlessProvider` to the approval system that auto-approves tool executions in headless environments with WARN-level audit logging
- Add `HeadlessAutoApprove` config field to `InterceptorConfig` (default: `false`, fail-closed)
- Add `BrowserBin` config field to `BrowserToolConfig` for explicit browser binary path
- Update `ensureBrowser()` to use `launcher.LookPath()` for automatic system chromium detection
- Change Dockerfile `WORKDIR` from `/app` to `/home/lango` (user home directory, writable)
- Remove unused `ENV ROD_BROWSER` from Dockerfile

## Capabilities

### New Capabilities
- `headless-approval`: Auto-approve provider for headless/Docker environments where no TTY or companion is available

### Modified Capabilities
- `ai-privacy-interceptor`: Add `HeadlessAutoApprove` config option for headless fallback behavior
- `tool-browser`: Add `BrowserBin` config and `LookPath()` auto-detection for system-installed browsers
- `docker-deployment`: Fix WORKDIR permissions and remove unused ROD_BROWSER env var

## Impact

- `internal/config/types.go`: New fields in `InterceptorConfig` and `BrowserToolConfig`
- `internal/approval/headless.go`: New file — `HeadlessProvider` implementation
- `internal/approval/headless_test.go`: New file — tests
- `internal/app/app.go`: Wiring for `HeadlessProvider` and `BrowserBin`
- `internal/tools/browser/browser.go`: `BrowserBin` field and `LookPath()` logic
- `Dockerfile`: WORKDIR change, ROD_BROWSER removal
