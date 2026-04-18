---
title: Installation
---

# Installation

## Prerequisites

| Requirement | Details |
|---|---|
| **Go** | 1.25 or later |
| **Git** | For cloning the repository |
| **C compiler** | Optional. Only needed when building legacy `vec` integrations that still depend on CGO. |

### Optional C Compiler Setup

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

!!! info "Default Runtime"

    The default runtime uses a pure-Go SQLite driver with FTS5 enabled. CGO is not required for normal builds. A C compiler is only needed if you explicitly build optional legacy `vec` integrations.

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
go install github.com/langoai/lango/cmd/lango@latest
```

!!! note "`make build` vs `go install`"

    `make build` and `go install` both use the default runtime with FTS5 enabled. Optional legacy `vec` integrations are not part of the default build.

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
