---
title: Installation
---

# Installation

## Prerequisites

| Requirement | Details |
|---|---|
| **Go** | 1.25 or later |
| **CGO** | Must be enabled (`CGO_ENABLED=1`). Required by the `sqlite3` and `sqlite-vec` drivers. |
| **C compiler** | `gcc` or `clang` (needed by CGO) |
| **Git** | For cloning the repository |

### Platform-Specific C Compiler Setup

=== "macOS"

    Xcode Command Line Tools are typically pre-installed. If not:

    ```bash
    xcode-select --install
    ```

=== "Ubuntu / Debian"

    ```bash
    sudo apt-get update && sudo apt-get install -y gcc libsqlite3-dev
    ```

=== "Fedora / RHEL"

    ```bash
    sudo dnf install gcc sqlite-devel
    ```

=== "Alpine Linux"

    ```bash
    apk add gcc musl-dev sqlite-dev
    ```

!!! info "Why CGO?"

    Lango uses SQLite for encrypted configuration storage and `sqlite-vec` for vector similarity search. Both require CGO-enabled builds. The Makefile sets `CGO_ENABLED=1` automatically.

## Build from Source

```bash
git clone https://github.com/langoai/lango.git
cd lango
make build
```

The binary is written to `bin/lango`. To install it into your `$GOPATH/bin`:

```bash
make install
```

## Go Install

You can also install directly with `go install`:

```bash
CGO_ENABLED=1 go install -tags "fts5,vec" github.com/langoai/lango/cmd/lango@latest
```

!!! note "`make build` vs `go install`"

    `make build` sets CGO, build tags (`fts5`, `vec`), and version/build-time LDFLAGS automatically. With `go install` you must pass these flags yourself. If you omit `-tags "fts5,vec"`, full-text search and vector similarity will not be available.

## Verify Installation

```bash
lango version
```

You should see output like:

```
lango v0.2.1
```

## Optional: Browser Tools

Some tools (browser automation) require a Chromium-based browser.

!!! note "Chromium Dependency"

    If you plan to use browser automation tools, ensure a Chromium-based browser (Chrome, Chromium, or Edge) is installed on the system. Lango uses it via the Chrome DevTools Protocol for web page interaction.

## Platform-Specific Builds

The Makefile provides cross-compilation targets:

```bash
# Linux amd64
make build-linux

# macOS arm64 (Apple Silicon)
make build-darwin

# All platforms
make build-all
```

## Next Steps

Once installed, proceed to the [Quick Start](quickstart.md) to configure and launch your agent.
