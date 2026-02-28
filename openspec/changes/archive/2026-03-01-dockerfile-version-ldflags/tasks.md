## 1. Dockerfile Modification

- [x] 1.1 Add `ARG VERSION=dev` and `ARG BUILD_TIME=unknown` before the `go build` command in the builder stage
- [x] 1.2 Update `go build -ldflags` to include `-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}`

## 2. Verification

- [x] 2.1 Build Docker image with explicit `--build-arg VERSION=1.0.0 --build-arg BUILD_TIME=2026-03-01` and verify `lango version` output
- [x] 2.2 Build Docker image without build args and verify default `dev`/`unknown` output
