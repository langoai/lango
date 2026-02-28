## Context

Lango currently uses a Makefile for local builds (`make build`, `make build-all`) with no CI/CD pipeline for releases. The project requires CGO (`CGO_ENABLED=1`) for SQLite, sqlite-vec, and macOS Keychain dependencies, making standard GoReleaser cross-compilation impossible. Only a docs deployment workflow (`docs.yml`) exists in GitHub Actions.

## Goals / Non-Goals

**Goals:**
- Automate multi-platform binary releases on tag push (linux/darwin × amd64/arm64)
- Produce standard and extended (KMS) build variants per platform
- Generate SHA256 checksums and conventional-commit-based changelogs
- Provide CI validation (build, test, lint) on PRs and main branch pushes
- Enable local release testing via Makefile targets

**Non-Goals:**
- Docker image builds (existing `Dockerfile` and `docker-build` target remain separate)
- Homebrew tap formula publishing (future iteration)
- Windows support (no CGO toolchain readily available)
- Code signing for macOS binaries in CI (existing `make codesign` remains manual)

## Decisions

### 1. Native Runner Matrix + Split/Merge Strategy

**Decision**: Use GitHub Actions native runners per platform instead of cross-compilation.

**Rationale**: CGO requires platform-native C toolchains. Cross-compilation with CGO needs complex toolchain setup (musl-cross, osxcross) that is fragile and hard to maintain. Native runners build natively with zero toolchain complexity.

**Alternatives considered**:
- Cross-compilation with Docker + osxcross: complex setup, Apple SDK licensing concerns
- Zig as C cross-compiler: experimental GoReleaser support, not production-ready

**Implementation**: GoReleaser `--split` in matrix jobs → `--merge` in final release job.

### 2. Two Build Variants (Standard + Extended)

**Decision**: Ship two binaries per platform — standard (default stubs) and extended (`-tags kms_all`).

**Rationale**: Most users don't need AWS/GCP/Azure/PKCS11 KMS. Separating keeps the standard binary smaller and avoids pulling in cloud SDK dependencies at runtime.

### 3. GoReleaser v2 Schema

**Decision**: Use `version: 2` schema for `.goreleaser.yaml`.

**Rationale**: v2 is the current stable schema with better split/merge support and the configuration format that will be maintained going forward.

### 4. Linux Dependencies via apt-get

**Decision**: Install `libsqlite3-dev` via apt-get on Linux runners only.

**Rationale**: macOS runners ship with system SQLite framework. Linux runners need the development headers explicitly. Keeping dependencies minimal reduces build time.

## Risks / Trade-offs

- **[Runner availability]** → ARM64 runners (`ubuntu-24.04-arm`, `macos-14`) are newer GitHub Actions offerings. If unavailable, builds fail gracefully; releases can be re-triggered after runner availability is restored.
- **[Build time]** → 4 parallel matrix jobs × 2 build variants = 8 builds total. Each job runs independently so wall-clock time equals the slowest single build (~5 min). Total compute time is higher but acceptable for release frequency.
- **[No Windows support]** → CGO on Windows CI is complex (MSYS2, MinGW). Excluded for now; can be added later with a dedicated Windows runner + toolchain step.
- **[Artifact size]** → Split/merge produces 8 tar.gz archives + checksums. GitHub Releases handles this well; no concern for storage limits.
